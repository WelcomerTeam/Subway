package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"

	"github.com/WelcomerTeam/Discord/discord"
	protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	subway "github.com/WelcomerTeam/Subway/subway"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PermissionsDefault = 0o744
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	sandwichGRPCHost := flag.String("sandwichGRPCHost", os.Getenv("SANDWICH_GRPC_HOST"), "GRPC Address for the Sandwich Daemon service")
	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Address to proxy requests through. This can be 'https://discord.com', if one is not setup.")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debugging requests to the proxy")
	prometheusAddress := flag.String("prometheusAddress", os.Getenv("INTERACTIONS_PROMETHEUS_ADDRESS"), "Prometheus address")

	host := flag.String("host", os.Getenv("INTERACTIONS_HOST"), "Host to serve interactions from")
	publicKeys := flag.String("publicKey", os.Getenv("INTERACTIONS_PUBLIC_KEY"), "Public key(s) for signature validation. Comma delimited.")

	dryRun := flag.Bool("dryRun", false, "When true, will close after setting up the app")

	flag.Parse()

	// Setup Rest
	proxyURL, err := url.Parse(*proxyAddress)
	if err != nil {
		panic(fmt.Errorf("failed to parse proxy address. url.Parse(%s): %w", *proxyAddress, err))
	}

	restInterface := discord.NewTwilightProxy(*proxyURL)
	restInterface.SetDebug(*proxyDebug)

	// Setup GRPC
	grpcConnection, err := grpc.NewClient(*sandwichGRPCHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf(`failed to parse grpcAddress. grpc.Dial(%s): %w`, *sandwichGRPCHost, err))
	}

	// Setup Logger
	level := slog.LevelDebug
	err = level.UnmarshalText([]byte(*loggingLevel))
	if err != nil {
		panic(fmt.Errorf(`failed to parse loggingLevel. level.UnmarshalText(%s): %w`, *loggingLevel, err))
	}

	logger := slog.Default()
	logger.Info("Logging configured")

	context, cancel := context.WithCancel(context.Background())

	// Setup app.
	app, err := subway.NewSubway(context, subway.SubwayOptions{
		SandwichClient:    protobuf.NewSandwichClient(grpcConnection),
		RESTInterface:     restInterface,
		Logger:            logger,
		PublicKeys:        *publicKeys,
		PrometheusAddress: *prometheusAddress,
	})
	if err != nil {
		logger.Error("Error creating subway", "error", err)
	}

	// Register Cogs here. Either via app.RegisterCog or app.MustRegisterCog
	// sub.MustRegisterCog(plugins.NewGeneralCog())

	// Make sure you sync commands. It is recommended to manually initiate this,
	// instead of on every startup. You can use app.SyncCommands() with "Bot " + token.

	// We return if it a dry run. Any issues loading up the bot would've already caused a panic.
	if *dryRun {
		return
	}

	err = app.ListenAndServe("", *host)
	if err != nil {
		logger.Warn("Exception whilst starting app", "error", err)
	}

	cancel()

	err = grpcConnection.Close()
	if err != nil {
		logger.Warn("Exception whilst closing grpc client", "error", err)
	}
}

func MustParseBool(str string) bool {
	boolean, _ := strconv.ParseBool(str)

	return boolean
}

func MustParseInt(str string) int {
	integer, _ := strconv.ParseInt(str, 10, 64)

	return int(integer)
}
