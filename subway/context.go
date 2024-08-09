package internal

import (
	"context"
	"net/url"

	"github.com/WelcomerTeam/Discord/discord"
)

type SubwayContextKey int

const (
	SubwayKey SubwayContextKey = iota
	InteractionKey
	InteractionCommandKey
	ArgumentsKey
	RawOptionsKey
	CommandBranchKey
	CommandTreeKey
	IdentifierKey
	ComponentListenerKey
	URLKey
)

// URL context handler.
func AddURLToContext(ctx context.Context, v url.URL) context.Context {
	return context.WithValue(ctx, URLKey, v)
}

func GetURLFromContext(ctx context.Context) url.URL {
	value, ok := ctx.Value(URLKey).(url.URL)
	if !ok {
		panic("GetURLFromContext(): failed to get URL from context")
	}

	return value
}

// InteractionCommand context handler.
func AddInteractionCommandToContext(ctx context.Context, v *InteractionCommandable) context.Context {
	return context.WithValue(ctx, InteractionCommandKey, v)
}

func GetInteractionCommandFromContext(ctx context.Context) *InteractionCommandable {
	value, ok := ctx.Value(InteractionCommandKey).(*InteractionCommandable)
	if !ok {
		panic("GetInteractionCommandFromContext(): failed to get InteractionCommand from context")
	}

	return value
}

// Arguments context handler.
func AddArgumentsToContext(ctx context.Context, v map[string]*Argument) context.Context {
	return context.WithValue(ctx, ArgumentsKey, v)
}

func GetArgumentsFromContext(ctx context.Context) map[string]*Argument {
	value, ok := ctx.Value(ArgumentsKey).(map[string]*Argument)
	if !ok {
		panic("GetArgumentsFromContext(): failed to get Arguments from context")
	}

	return value
}

// RawOptions context handler.
func AddRawOptionsToContext(ctx context.Context, v map[string]discord.InteractionDataOption) context.Context {
	return context.WithValue(ctx, RawOptionsKey, v)
}

func GetRawOptionsFromContext(ctx context.Context) map[string]discord.InteractionDataOption {
	value, ok := ctx.Value(RawOptionsKey).(map[string]discord.InteractionDataOption)
	if !ok {
		panic("GetRawOptionsFromContext(): failed to get RawOptions from context")
	}

	return value
}

// CommandBranch context handler.
func AddCommandBranchToContext(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, CommandBranchKey, v)
}

func GetCommandBranchFromContext(ctx context.Context) []string {
	value, ok := ctx.Value(CommandBranchKey).([]string)
	if !ok {
		panic("GetCommandBranchFromContext(): failed to get CommandBranch from context")
	}

	return value
}

// CommandTree context handler.
func AddCommandTreeToContext(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, CommandTreeKey, v)
}

func GetCommandTreeFromContext(ctx context.Context) []string {
	value, ok := ctx.Value(CommandTreeKey).([]string)
	if !ok {
		panic("GetCommandTreeFromContext(): failed to get CommandTree from context")
	}

	return value
}

// Identifier context handler.
func AddIdentifierToContext(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, IdentifierKey, v)
}

func GetIdentifierFromContext(ctx context.Context) string {
	value, ok := ctx.Value(IdentifierKey).(string)
	if !ok {
		panic("GetIdentifierFromContext(): failed to get Identifier from context")
	}

	return value
}

// ComponentListener context handler.
func AddComponentListenerToContext(ctx context.Context, v *ComponentListener) context.Context {
	return context.WithValue(ctx, ComponentListenerKey, v)
}

func GetComponentListenerFromContext(ctx context.Context) *ComponentListener {
	value, ok := ctx.Value(ComponentListenerKey).(*ComponentListener)
	if !ok {
		panic("GetComponentListenerFromContext(): failed to get ComponentListener from context")
	}

	return value
}
