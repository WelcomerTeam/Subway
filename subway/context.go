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
)

// InteractionCommand context handler.
func AddInteractionCommandToContext(ctx context.Context, v *InteractionCommandable) context.Context {
	return context.WithValue(ctx, InteractionCommandKey, v)
}

func GetInteractionCommandFromContext(ctx context.Context) *InteractionCommandable {
	return ctx.Value(InteractionCommandKey).(*InteractionCommandable)
}

// Arguments context handler.
func AddArgumentsToContext(ctx context.Context, v map[string]*Argument) context.Context {
	return context.WithValue(ctx, ArgumentsKey, v)
}

func GetArgumentsFromContext(ctx context.Context) map[string]*Argument {
	return ctx.Value(ArgumentsKey).(map[string]*Argument)
}

// RawOptions context handler.
func AddRawOptionsToContext(ctx context.Context, v map[string]*discord.InteractionDataOption) context.Context {
	return context.WithValue(ctx, RawOptionsKey, v)
}

func GetRawOptionsFromContext(ctx context.Context) map[string]*discord.InteractionDataOption {
	return ctx.Value(RawOptionsKey).(map[string]*discord.InteractionDataOption)
}

// CommandBranch context handler.
func AddCommandBranchToContext(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, CommandBranchKey, v)
}

func GetCommandBranchFromContext(ctx context.Context) []string {
	return ctx.Value(CommandBranchKey).([]string)
}

// CommandTree context handler.
func AddCommandTreeToContext(ctx context.Context, v []string) context.Context {
	return context.WithValue(ctx, CommandTreeKey, v)
}

func GetCommandTreeFromContext(ctx context.Context) []string {
	return ctx.Value(CommandTreeKey).([]string)
}

// Identifier context handler.
func AddIdentifierToContext(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, IdentifierKey, v)
}

func GetIdentifierFromContext(ctx context.Context) string {
	return ctx.Value(IdentifierKey).(string)
}
