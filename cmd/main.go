package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	subway "github.com/WelcomerTeam/Subway/subway"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	configurationLocation := flag.String("configuration", os.Getenv("CONFIGURATION_PATH"), "Path of configuration file")

	grpcAddress := flag.String("grpcAddress", os.Getenv("GRPC_ADDRESS"), "GRPC Address")

	proxyAddress := flag.String("proxyAddress", os.Getenv("PROXY_ADDRESS"), "Twilight proxy Address")
	proxyDebug := flag.Bool("proxyDebug", false, "Enable debug on proxy")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("PROMETHEUS_ADDRESS"), "Prometheus address")

	publicKey := flag.String("publicKey", os.Getenv("PUBLIC_KEY"), "Public key for signature validation")

	host := flag.String("host", os.Getenv("HOST"), "Host")

	isRelease := flag.Bool("release", false, "Release Mode")

	nginxAddress := flag.String("nginxProxy", os.Getenv("NGINX_PROXY"), "NGINX Proxy Address")

	releaseMode := flag.String("ginMode", os.Getenv("GIN_MODE"), "gin mode (release/debug)")

	flag.Parse()

	gin.SetMode(*releaseMode)

	// Setup Rest
	proxyURL, err := url.Parse(*proxyAddress)
	if err != nil {
		panic(fmt.Sprintf("url.Parse(%s): %v", *proxyAddress, err.Error()))
	}

	restInterface := discord.NewTwilightProxy(*proxyURL)
	restInterface.SetDebug(*proxyDebug)

	// TODO: Move into NewBackend(...)
	// Setup GRPC
	grpcConnection, err := grpc.Dial(*grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Sprintf(`grpc.Dial(%s): %v`, *grpcAddress, err.Error()))
	}

	// Setup Logger
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Stamp,
	}

	logger := zerolog.New(writer).With().Timestamp().Logger()

	// Setup app.
	app, err := subway.NewSubway(
		grpcConnection, restInterface, writer, *isRelease, *configurationLocation,
		*publicKey, *host, *prometheusAddress, *nginxAddress)
	if err != nil {
		logger.Panic().Err(err).Msg("Exception creating app")
	}

	err = app.Open()
	if err != nil {
		logger.Warn().Err(err).Msg("Exceptions whilst starting app")
	}

	// Close app.
	err = app.Close()
	if err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing app")
	}

	err = grpcConnection.Close()
	if err != nil {
		logger.Warn().Err(err).Msg("Exception whilst closing grpc client")
	}
}
