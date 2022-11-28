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

// Subway context handler.
func AddSubwayToContext(ctx context.Context, v *Subway) context.Context {
	return context.WithValue(ctx, SubwayKey, v)
}

func GetSubwayFromContext(ctx context.Context) *Subway {
	return ctx.Value(SubwayKey).(*Subway)
}

// Interaction context handler.
func AddInteractionToContext(ctx context.Context, v *discord.Interaction) context.Context {
	return context.WithValue(ctx, InteractionKey, v)
}

func GetInteractionFromContext(ctx context.Context) *discord.Interaction {
	return ctx.Value(InteractionKey).(*discord.Interaction)
}

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
