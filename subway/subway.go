package internal

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
)

// VERSION follows semantic versioning.
const VERSION = "1.0.1"

const (
	PermissionsDefault = 0o744
	PermissionWrite    = 0o600

	defaultMaximumInteractionAge = 15 * time.Minute
)

type Subway struct {
	context.Context

	Logger    *slog.Logger `json:"-"`
	StartTime time.Time    `json:"start_time" yaml:"start_time"`

	Commands   *InteractionCommandable `json:"-"`
	Converters *InteractionConverters  `json:"-"`

	Cogs map[string]Cog `json:"-"`

	SandwichClient sandwich_protobuf.SandwichClient `json:"-"`
	RESTInterface  discord.RESTInterface            `json:"-"`
	EmptySession   *discord.Session                 `json:"-"`

	ComponentListenersMu sync.RWMutex
	ComponentListeners   map[string]*ComponentListener

	OnBeforeInteraction InteractionRequestHandler
	OnAfterInteraction  InteractionResponseHandler

	// Environment Variables.
	PublicKeys        *atomic.Pointer[[]ed25519.PublicKey]
	prometheusAddress string
}

// SubwayOptions represents the options to create a new subway service.
type SubwayOptions struct {
	SandwichClient sandwich_protobuf.SandwichClient
	RESTInterface  discord.RESTInterface
	Logger         *slog.Logger

	OnBeforeInteraction InteractionRequestHandler
	OnAfterInteraction  InteractionResponseHandler

	PublicKeys        string
	PrometheusAddress string

	// Maximum age for component listeners. Defaults to 15 minutes.
	// This is the absolute maximum age of a component listener,
	// ignoring a listener with a longer age.
	MaximumInteractionAge time.Duration
}

func NewSubway(ctx context.Context, options SubwayOptions) (*Subway, error) {
	sub := &Subway{
		Context: ctx,

		Logger: options.Logger,

		RESTInterface:  options.RESTInterface,
		SandwichClient: options.SandwichClient,

		ComponentListenersMu: sync.RWMutex{},
		ComponentListeners:   make(map[string]*ComponentListener),

		OnBeforeInteraction: options.OnBeforeInteraction,
		OnAfterInteraction:  options.OnAfterInteraction,

		prometheusAddress: options.PrometheusAddress,

		Commands:   SetupInteractionCommandable(nil),
		Converters: NewInteractionConverters(),

		Cogs: make(map[string]Cog),

		PublicKeys: &atomic.Pointer[[]ed25519.PublicKey]{},
	}

	if err := sub.SetPublicKeys(strings.Split(options.PublicKeys, ","), false); err != nil {
		return nil, fmt.Errorf("failed to set public keys: %w", err)
	}

	// Setup sessions
	sub.EmptySession = discord.NewSession("", sub.RESTInterface)

	if options.MaximumInteractionAge <= 0 {
		options.MaximumInteractionAge = defaultMaximumInteractionAge
	}

	go sub.InteractionCleanupJob(ctx, options.MaximumInteractionAge)

	return sub, nil
}

func (sub *Subway) AddPublicKey(publicKeyString string) error {
	hex, err := hex.DecodeString(publicKeyString)
	if err != nil {
		return fmt.Errorf("failed to decode public key %s: %w", publicKeyString, err)
	}

	if len(hex) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(hex))
	}

	publicKeysPtr := sub.PublicKeys.Load()
	publicKeys := *publicKeysPtr
	publicKeys = append(publicKeys, ed25519.PublicKey(hex))
	sub.PublicKeys.Store(&publicKeys)

	return nil
}

func (sub *Subway) SetPublicKeys(publicKeyStrings []string, ignoreErrors bool) error {
	publicKeys := make([]ed25519.PublicKey, 0)
	sub.PublicKeys.Store(&publicKeys)

	// Setup public keys
	for _, publicKey := range publicKeyStrings {
		hex, err := hex.DecodeString(publicKey)
		if err != nil {
			if ignoreErrors {
				sub.Logger.Error("Failed to decode public key, ignoring", "public_key", publicKey, "error", err)

				continue
			}

			return fmt.Errorf("failed to decode public key %s: %w", publicKey, err)
		}

		if len(hex) != ed25519.PublicKeySize {
			if ignoreErrors {
				sub.Logger.Error("Invalid public key size, ignoring", "public_key", publicKey, "expected", ed25519.PublicKeySize, "got", len(hex))

				continue
			}

			return fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(hex))
		}

		publicKeys = append(publicKeys, ed25519.PublicKey(hex))
	}

	sub.PublicKeys.Store(&publicKeys)

	return nil
}

func (sub *Subway) InteractionCleanupJob(ctx context.Context, maximumAge time.Duration) {
	ticker := time.NewTicker(maximumAge)

	for {
		select {
		case <-ticker.C:
			sub.cleanupInteractions(maximumAge)
		case <-ctx.Done():
			return
		}
	}
}

func (sub *Subway) cleanupInteractions(maximumAge time.Duration) {
	now := time.Now()

	sub.ComponentListenersMu.RLock()

	deletedKeys := []string{}

	for i, k := range sub.ComponentListeners {
		if k.expiresAt.After(now) || k.createdAt.Add(maximumAge).After(now) {
			deletedKeys = append(deletedKeys, i)
		}
	}

	sub.ComponentListenersMu.RUnlock()

	if len(deletedKeys) > 0 {
		sub.ComponentListenersMu.Lock()
		for _, key := range deletedKeys {
			delete(sub.ComponentListeners, key)
		}
		sub.ComponentListenersMu.Unlock()
	}
}

// Listen handles starting up the webserver and services for you.
func (sub *Subway) ListenAndServe(route, host string) error {
	if route == "" {
		route = "/"
	}

	sub.StartTime = time.Now().UTC()
	sub.Logger.Info("Starting subway", "version", VERSION)

	// Setup Prometheus
	go sub.SetupPrometheus()

	sub.Logger.Info("Serving subway", "host", host)

	subwayMux := http.NewServeMux()
	subwayMux.HandleFunc(route, sub.HandleSubwayRequest)

	err := http.ListenAndServe(host, subwayMux)
	if err != nil {
		sub.Logger.Error("Failed to serve subway server", "host", sub.prometheusAddress, "error", err)

		return fmt.Errorf("failed to serve sub: %w", err)
	}

	return nil
}

// SyncCommands syncs all registered commands with the discord API.
// Use sandwichClient.FetchIdentifier to get the token for an identifier.
// Token must have "Bot " added.
func (sub *Subway) SyncCommands(ctx context.Context, token string, applicationID discord.Snowflake) error {
	session := discord.NewSession(token, sub.RESTInterface)

	applicationCommands := sub.Commands.MapApplicationCommands()

	_, err := discord.BulkOverwriteGlobalApplicationCommands(ctx, session, applicationID, applicationCommands)
	if err != nil {
		return fmt.Errorf("failed to bulk overwrite commands: %w", err)
	}

	return nil
}
