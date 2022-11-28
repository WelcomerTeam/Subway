package internal

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/rs/zerolog"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

// VERSION follows semantic versioning.
const VERSION = "0.2"

const (
	PermissionsDefault = 0o744
	PermissionWrite    = 0o600
)

type Subway struct {
	context.Context

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

	Commands   *InteractionCommandable `json:"-"`
	Converters *InteractionConverters  `json:"-"`

	Cogs map[string]Cog `json:"-"`

	SandwichClient protobuf.SandwichClient `json:"-"`
	GRPCInterface  sandwich.GRPC           `json:"-"`
	RESTInterface  discord.RESTInterface   `json:"-"`
	EmptySession   *discord.Session        `json:"-"`

	OnBeforeInteraction InteractionRequestHandler
	OnAfterInteraction  InteractionResponseHandler

	// Environment Variables.
	publicKey         ed25519.PublicKey
	prometheusAddress string

	webhooks []string
}

// SubwayOptions represents the options to create a new subway service.
type SubwayOptions struct {
	SandwichClient protobuf.SandwichClient
	RESTInterface  discord.RESTInterface
	Logger         zerolog.Logger

	OnBeforeInteraction InteractionRequestHandler
	OnAfterInteraction  InteractionResponseHandler

	PublicKey         string
	PrometheusAddress string

	Webhooks []string
}

func NewSubway(ctx context.Context, options SubwayOptions) (*Subway, error) {
	subway := &Subway{
		Context: ctx,

		Logger: options.Logger,

		RESTInterface:  options.RESTInterface,
		SandwichClient: options.SandwichClient,
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),

		OnBeforeInteraction: options.OnBeforeInteraction,
		OnAfterInteraction:  options.OnAfterInteraction,

		prometheusAddress: options.PrometheusAddress,

		Commands:   SetupInteractionCommandable(nil),
		Converters: NewInteractionConverters(),

		Cogs: make(map[string]Cog),

		webhooks: options.Webhooks,
	}

	var err error

	subway.publicKey, err = hex.DecodeString(options.PublicKey)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	// Setup sessions
	subway.EmptySession = discord.NewSession(ctx, "", subway.RESTInterface, subway.Logger)

	return subway, nil
}

// Listen handles starting up the webserver and services for you.
func (subway *Subway) ListenAndServe(route, host string) error {
	if route == "" {
		route = "/"
	}

	subway.StartTime = time.Now().UTC()
	subway.Logger.Info().Msgf("Starting subway Version %s", VERSION)

	go subway.PublishSimpleWebhook(subway.EmptySession, "Starting subway", "", "Version "+VERSION, EmbedColourSandwich)

	// Setup Prometheus
	go subway.SetupPrometheus()

	subway.Logger.Info().Msgf("Serving subway at %s", host)

	subwayMux := http.NewServeMux()
	subwayMux.HandleFunc(route, subway.HandleSubwayRequest)

	err := http.ListenAndServe(host, subwayMux)
	if err != nil {
		subway.Logger.Error().Str("host", subway.prometheusAddress).Err(err).Msg("Failed to serve subway server")

		return fmt.Errorf("failed to serve subway: %w", err)
	}

	println("D")

	return nil
}

// SyncCommands syncs all registered commands with the discord API.
// Use sandwichClient.FetchIdentifier to get the token for an identifier.
// Token must have "Bot " added.
func (subway *Subway) SyncCommands(ctx context.Context, token string, applicationID discord.Snowflake) error {
	session := discord.NewSession(ctx, token, subway.RESTInterface, subway.Logger)

	applicationCommands := subway.Commands.MapApplicationCommands()

	_, err := discord.BulkOverwriteGlobalApplicationCommands(session, applicationID, applicationCommands)
	if err != nil {
		return fmt.Errorf("failed to bulk overwrite commands: %w", err)
	}

	return nil
}
