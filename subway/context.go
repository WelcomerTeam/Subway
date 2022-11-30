package internal

import (
	"context"

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
)

// InteractionCommand context handler.
func AddInteractionCommandToContext(ctx context.Context, v *InteractionCommandable) context.Context {
	return context.WithValue(ctx, InteractionCommandKey, v)
}

func GetInteractionCommandFromContext(ctx context.Context) *InteractionCommandable {
	value, _ := ctx.Value(InteractionCommandKey).(*InteractionCommandable)

	return value
}

// Arguments context handler.
func AddArgumentsToContext(ctx context.Context, v map[string]*Argument) context.Context {
	return context.WithValue(ctx, ArgumentsKey, v)
}

func GetArgumentsFromContext(ctx context.Context) map[string]*Argument {
	value, _ := ctx.Value(ArgumentsKey).(map[string]*Argument)

	return value
}

// RawOptions context handler.
func AddRawOptionsToContext(ctx context.Context, v map[string]*discord.InteractionDataOption) context.Context {
	return context.WithValue(ctx, RawOptionsKey, v)
}

func GetRawOptionsFromContext(ctx context.Context) map[string]*discord.InteractionDataOption {
	value, _ := ctx.Value(RawOptionsKey).(map[string]*discord.InteractionDataOption)

	return value
}

// CommandBranch context handler.
func AddCommandBranchToContext(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, CommandBranchKey, v)
}

func GetCommandBranchFromContext(ctx context.Context) []string {
	value, _ := ctx.Value(CommandBranchKey).([]string)

	return value
}

// CommandTree context handler.
func AddCommandTreeToContext(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, CommandTreeKey, v)
}

func GetCommandTreeFromContext(ctx context.Context) []string {
	value, _ := ctx.Value(CommandTreeKey).([]string)

	return value
}

// Identifier context handler.
func AddIdentifierToContext(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, IdentifierKey, v)
}

func GetIdentifierFromContext(ctx context.Context) string {
	value, _ := ctx.Value(IdentifierKey).(string)

	return value
}

// ComponentListener context handler.
func AddComponentListenerToContext(ctx context.Context, v *ComponentListener) context.Context {
	return context.WithValue(ctx, ComponentListenerKey, v)
}

func GetComponentListenerFromContext(ctx context.Context) *ComponentListener {
	value, _ := ctx.Value(ComponentListenerKey).(*ComponentListener)

	return value
}
