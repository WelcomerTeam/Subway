package internal

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/rs/zerolog"

	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
)

// VERSION follows semantic versioning.
const VERSION = "0.0.1"

const (
	PermissionsDefault = 0o744
	PermissionWrite    = 0o600
)

type Subway struct {
	ctx    context.Context
	cancel func()

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

	Commands   *InteractionCommandable `json:"-"`
	Converters *InteractionConverters  `json:"-"`

	Cogs map[string]Cog `json:"-"`

	Route *gin.Engine `json:"-"`

	SandwichClient protobuf.SandwichClient `json:"-"`
	GRPCInterface  sandwich.GRPC           `json:"-"`
	RESTInterface  discord.RESTInterface   `json:"-"`
	EmptySession   *discord.Session        `json:"-"`

	// Environment Variables.
	publicKey         ed25519.PublicKey
	host              string
	prometheusAddress string
	nginxAddress      string

	webhooks []string
}

// SubwayOptions represents the options to create a new subway service.
type SubwayOptions struct {
	SandwichClient protobuf.SandwichClient
	RESTInterface  discord.RESTInterface
	Logger         zerolog.Logger

	GinMode   string
	PublicKey string
	Host      string

	PrometheusAddress string
	NginxAddress      string

	Webhooks []string
}

func NewSubway(options SubwayOptions) (*Subway, error) {
	subway := &Subway{
		Logger: options.Logger,

		RESTInterface:  options.RESTInterface,
		SandwichClient: options.SandwichClient,
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),

		host:              options.Host,
		prometheusAddress: options.PrometheusAddress,
		nginxAddress:      options.NginxAddress,

		Commands:   SetupInteractionCommandable(nil),
		Converters: NewInteractionConverters(),

		Cogs: make(map[string]Cog),
	}

	var err error

	subway.publicKey, err = hex.DecodeString(options.PublicKey)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	subway.ctx, subway.cancel = context.WithCancel(context.Background())

	// Setup sessions
	subway.EmptySession = discord.NewSession(subway.ctx, "", subway.RESTInterface, subway.Logger)

	if options.GinMode != "" {
		gin.SetMode(options.GinMode)
	}

	if options.NginxAddress != "" {
		err = subway.Route.SetTrustedProxies([]string{options.NginxAddress})
		if err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}

	subway.Route = subway.PrepareGin()

	return subway, nil
}

// Open sets up any services and starts the webserver.
func (subway *Subway) Open() error {
	subway.StartTime = time.Now().UTC()
	subway.Logger.Info().Msgf("Starting subway Version %s", VERSION)

	go subway.PublishSimpleWebhook(subway.EmptySession, "Starting subway", "", "Version "+VERSION, EmbedColourSandwich)

	// Setup Prometheus
	go subway.SetupPrometheus()

	subway.Logger.Info().Msgf("Serving http at %s", subway.host)

	err := subway.Route.Run(subway.host)
	if err != nil {
		return fmt.Errorf("failed to run gin: %w", err)
	}

	return nil
}

// Close gracefully closes the backend.
func (subway *Subway) Close() error {
	// TODO

	return nil
}

// PrepareGin prepares gin routes and middleware.
func (subway *Subway) PrepareGin() *gin.Engine {
	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare

	_ = router.SetTrustedProxies(nil)

	router.Use(logger.SetLogger())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(gin.Recovery())

	subway.registerRoutes(router)

	return router
}
