package internal

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/rs/zerolog"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"google.golang.org/grpc"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	yaml "gopkg.in/yaml.v3"
)

// VERSION follows semantic versioning.
const VERSION = "0.0.1"

const (
	PermissionsDefault = 0o744
	PermissionWrite    = 0o600
)

var subway *Subway

type Subway struct {
	sync.Mutex

	ConfigurationLocation string `json:"configuration_location"`

	ctx    context.Context
	cancel func()

	Logger    zerolog.Logger `json:"-"`
	StartTime time.Time      `json:"start_time" yaml:"start_time"`

	configurationMu sync.RWMutex   `json:"-"`
	Configuration   *Configuration `json:"configuration" yaml:"configuration"`

	Route *gin.Engine `json:"-"`

	Commands   *sandwich.InteractionCommandable `json:"-"`
	Converters *sandwich.InteractionConverters  `json:"-"`

	Cogs map[string]sandwich.Cog `json:"-"`

	RESTInterface discord.RESTInterface `json:"-"`

	SandwichClient protobuf.SandwichClient `json:"-"`
	GRPCInterface  sandwich.GRPC           `json:"-"`

	PrometheusHandler *ginprometheus.Prometheus `json:"-"`

	EmptySession *discord.Session `json:"-"`

	// Environment Variables.
	publicKey         ed25519.PublicKey
	host              string
	prometheusAddress string
	nginxAddress      string
}

// Configuration represents the configuration file.
type Configuration struct {
	Logging struct {
		Level              string `json:"level" yaml:"level"`
		FileLoggingEnabled bool   `json:"file_logging_enabled" yaml:"file_logging_enabled"`

		EncodeAsJSON bool `json:"encode_as_json" yaml:"encode_as_json"`

		Directory  string `json:"directory" yaml:"directory"`
		Filename   string `json:"filename" yaml:"filename"`
		MaxSize    int    `json:"max_size" yaml:"max_size"`
		MaxBackups int    `json:"max_backups" yaml:"max_backups"`
		MaxAge     int    `json:"max_age" yaml:"max_age"`
		Compress   bool   `json:"compress" yaml:"compress"`
	} `json:"logging" yaml:"logging"`

	Webhooks []string `json:"webhooks" yaml:"webhooks"`
}

func NewSubway(conn grpc.ClientConnInterface, restInterface discord.RESTInterface, logger io.Writer, isReleaseMode bool, configurationLocation, publicKey, host, prometheusAddress, nginxAddress string) (s *Subway, err error) {
	if subway != nil {
		return subway, ErrSubwayAlreadyExists
	}

	s = &Subway{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		ConfigurationLocation: configurationLocation,

		configurationMu: sync.RWMutex{},
		Configuration:   &Configuration{},

		Commands:   sandwich.SetupInteractionCommandable(&sandwich.InteractionCommandable{}),
		Converters: sandwich.NewInteractionConverters(),

		Cogs: make(map[string]sandwich.Cog),

		RESTInterface: restInterface,

		SandwichClient: protobuf.NewSandwichClient(conn),
		GRPCInterface:  sandwich.NewDefaultGRPCClient(),

		PrometheusHandler: ginprometheus.NewPrometheus("gin"),

		host:              host,
		prometheusAddress: prometheusAddress,
		nginxAddress:      nginxAddress,
	}

	s.publicKey, err = hex.DecodeString(publicKey)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}

	s.Lock()
	defer s.Unlock()

	s.ctx, s.cancel = context.WithCancel(context.Background())

	if isReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup sessions
	s.EmptySession = discord.NewSession(s.ctx, "", s.RESTInterface, s.Logger)

	if nginxAddress != "" {
		err = s.Route.SetTrustedProxies([]string{nginxAddress})
		if err != nil {
			return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
		}
	}

	// Load configuration.
	configuration, err := s.LoadConfiguration(s.ConfigurationLocation)
	if err != nil {
		return nil, err
	}

	s.Configuration = configuration

	var writers []io.Writer

	writers = append(writers, logger)

	if s.Configuration.Logging.FileLoggingEnabled {
		if err := os.MkdirAll(s.Configuration.Logging.Directory, PermissionsDefault); err != nil {
			log.Error().Err(err).Str("path", s.Configuration.Logging.Directory).Msg("Unable to create log directory")
		} else {
			lumber := &lumberjack.Logger{
				Filename:   path.Join(s.Configuration.Logging.Directory, s.Configuration.Logging.Filename),
				MaxBackups: s.Configuration.Logging.MaxBackups,
				MaxSize:    s.Configuration.Logging.MaxSize,
				MaxAge:     s.Configuration.Logging.MaxAge,
				Compress:   s.Configuration.Logging.Compress,
			}

			if s.Configuration.Logging.EncodeAsJSON {
				writers = append(writers, lumber)
			} else {
				writers = append(writers, zerolog.ConsoleWriter{
					Out:        lumber,
					TimeFormat: time.Stamp,
					NoColor:    true,
				})
			}
		}
	}

	mw := io.MultiWriter(writers...)
	s.Logger = zerolog.New(mw).With().Timestamp().Logger()
	s.Logger.Info().Msg("Logging configured")

	// Setup gin router.
	s.Route = s.PrepareGin()

	subway = s

	return s, nil
}

// LoadConfiguration handles loading the configuration file.
func (subway *Subway) LoadConfiguration(path string) (configuration *Configuration, err error) {
	subway.Logger.Debug().
		Str("path", path).
		Msg("Loading configuration")

	defer func() {
		if err == nil {
			subway.Logger.Info().Msg("Configuration loaded")
		}
	}()

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return configuration, ErrReadConfigurationFailure
	}

	configuration = &Configuration{}

	err = yaml.Unmarshal(file, configuration)
	if err != nil {
		return configuration, ErrLoadConfigurationFailure
	}

	return configuration, nil
}

// Open sets up any services and starts the webserver.
func (subway *Subway) Open() error {
	subway.StartTime = time.Now().UTC()
	subway.Logger.Info().Msgf("Starting sandwich. Version %s", VERSION)

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

// SetupPrometheus sets up prometheus.
func (subway *Subway) SetupPrometheus() error {
	subway.Logger.Info().Msgf("Serving prometheus at %s", subway.prometheusAddress)

	subway.PrometheusHandler.SetListenAddress(subway.prometheusAddress)
	subway.PrometheusHandler.SetMetricsPath(nil)

	return nil
}

// PrepareGin prepares gin routes and middleware.
func (subway *Subway) PrepareGin() *gin.Engine {
	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare

	_ = router.SetTrustedProxies(nil)

	router.Use(logger.SetLogger())
	router.Use(subway.PrometheusHandler.HandlerFunc())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(gin.Recovery())

	registerRoutes(router)

	return router
}

// Dispatch

// ProcessInteraction processes the interaction that has been registered to the bot.
func (subway *Subway) ProcessInteraction(interaction discord.Interaction) (resp *discord.InteractionResponse, err error) {
	return subway.AsBot().ProcessInteraction(nil, interaction)
}

// Subway commands

func (subway *Subway) AsBot() *sandwich.Bot {
	return &sandwich.Bot{
		Logger:                subway.Logger,
		InteractionCommands:   subway.Commands,
		Cogs:                  subway.Cogs,
		InteractionConverters: subway.Converters,
	}
}

func (subway *Subway) MustRegisterCog(cog sandwich.Cog) {
	if err := subway.RegisterCog(cog); err != nil {
		panic(fmt.Sprintf(`sandwich: RegisterCog(%v): %v`, cog, err.Error()))
	}
}

func (subway *Subway) RegisterCog(cog sandwich.Cog) (err error) {
	cogInfo := cog.CogInfo()

	if _, ok := subway.Cogs[cogInfo.Name]; ok {
		return sandwich.ErrCogAlreadyRegistered
	}

	err = cog.RegisterCog(subway.AsBot())
	if err != nil {
		subway.Logger.Panic().Str("cog", cogInfo.Name).Err(err).Msg("Failed to register sandwich.Cog")

		return
	}

	subway.Cogs[cogInfo.Name] = cog

	subway.Logger.Info().Str("cog", cogInfo.Name).Msg("Loaded sandwich.Cog")

	if cast, ok := cog.(sandwich.CogWithBotLoad); ok {
		subway.Logger.Info().Str("cog", cogInfo.Name).Msg("Cog has BotLoad")

		cast.BotLoad(subway.AsBot())
	}

	if cast, ok := cog.(sandwich.CogWithInteractionCommands); ok {
		interactionCommandable := cast.GetInteractionCommandable()

		subway.Logger.Info().Str("cog", cogInfo.Name).Int("commands", len(interactionCommandable.GetAllCommands())).Msg("Cog has interaction commands")

		subway.RegisterCogInteractionCommandable(cog, interactionCommandable)
	}

	return nil
}

func (bot *Subway) RegisterCogInteractionCommandable(cog sandwich.Cog, interactionCommandable *sandwich.InteractionCommandable) {
	for _, command := range interactionCommandable.GetAllCommands() {
		// Add sandwich.Cog checks to all commands.
		command.Checks = append(interactionCommandable.Checks, command.Checks...)

		bot.Logger.Debug().Str("name", command.Name).Msg("Registering interaction command")

		bot.Commands.MustAddInteractionCommand(command)
	}
}
