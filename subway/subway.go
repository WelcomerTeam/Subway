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
	sub := &Subway{
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

	sub.publicKey, err = hex.DecodeString(options.PublicKey)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	// Setup sessions
	sub.EmptySession = discord.NewSession(ctx, "", sub.RESTInterface, sub.Logger)

	return sub, nil
}

// Listen handles starting up the webserver and services for you.
func (sub *Subway) ListenAndServe(route, host string) error {
	if route == "" {
		route = "/"
	}

	sub.StartTime = time.Now().UTC()
	sub.Logger.Info().Msgf("Starting subway Version %s", VERSION)

	go sub.PublishSimpleWebhook(sub.EmptySession, "Starting sub", "", "Version "+VERSION, EmbedColourSandwich)

	// Setup Prometheus
	go sub.SetupPrometheus()

	sub.Logger.Info().Msgf("Serving subway at %s", host)

	subwayMux := http.NewServeMux()
	subwayMux.HandleFunc(route, sub.HandleSubwayRequest)

	err := http.ListenAndServe(host, subwayMux)
	if err != nil {
		sub.Logger.Error().Str("host", sub.prometheusAddress).Err(err).Msg("Failed to serve subway server")

		return fmt.Errorf("failed to serve sub: %w", err)
	}

	return nil
}

// SyncCommands syncs all registered commands with the discord API.
// Use sandwichClient.FetchIdentifier to get the token for an identifier.
// Token must have "Bot " added.
func (sub *Subway) SyncCommands(ctx context.Context, token string, applicationID discord.Snowflake) error {
	session := discord.NewSession(ctx, token, sub.RESTInterface, sub.Logger)

	applicationCommands := sub.Commands.MapApplicationCommands()

	_, err := discord.BulkOverwriteGlobalApplicationCommands(session, applicationID, applicationCommands)
	if err != nil {
		return fmt.Errorf("failed to bulk overwrite commands: %w", err)
	}

	return nil
}
