package internal

import (
	"context"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
)

// ProcessComponent processes the component that has been sent.
func (sub *Subway) ProcessComponent(ctx context.Context, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	sub.ComponentListenersMu.RLock()
	listener, ok := sub.ComponentListeners[interaction.Data.CustomID]
	sub.ComponentListenersMu.RUnlock()

	ctx = AddComponentListenerToContext(ctx, listener)

	if !ok {
		return nil, ErrComponentListenerNotFound
	}

	if listener.Channel != nil {
		listener.Channel <- &interaction

		return nil, nil
	}

	return listener.Handler(ctx, sub, interaction)
}

// ProcessInteraction processes the interaction that has been sent.
func (sub *Subway) ProcessInteraction(ctx context.Context, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	commandTree := constructCommandTree(interaction.Data.Options, make([]string, 0))
	command := sub.Commands.GetCommand(interaction.Data.Name)

	// Create interaction context
	ctx = AddInteractionCommandToContext(ctx, command)
	ctx = AddArgumentsToContext(ctx, make(map[string]*Argument))
	ctx = AddRawOptionsToContext(ctx, extractOptions(interaction.Data.Options, make(map[string]*discord.InteractionDataOption)))
	ctx = AddCommandBranchToContext(ctx, commandTree)
	ctx = AddCommandTreeToContext(ctx, commandTree)

	if command == nil {
		return sub.Commands.propagateError(ctx, sub, interaction, ErrCommandNotFound), ErrCommandNotFound
	}

	if sub.OnBeforeInteraction != nil {
		err := sub.OnBeforeInteraction(ctx, sub, interaction)
		if err != nil {
			return sub.Commands.propagateError(ctx, sub, interaction, err), err
		}
	}

	response, err := command.Invoke(ctx, sub, interaction)

	if sub.OnAfterInteraction != nil {
		err = sub.OnAfterInteraction(ctx, sub, interaction, response, err)
		if err != nil {
			return sub.Commands.propagateError(ctx, sub, interaction, err), err
		}
	}

	return response, nil
}

func constructCommandTree(options []*discord.InteractionDataOption, tree []string) []string {
	newTree := tree

	for _, option := range options {
		switch option.Type {
		case discord.ApplicationCommandOptionTypeSubCommandGroup:
		case discord.ApplicationCommandOptionTypeSubCommand:
			newTree = append(newTree, option.Name)
			newTree = constructCommandTree(option.Options, newTree)
		default:
		}
	}

	return newTree
}

// CanRun checks all global bot checks and returns if the message passes them all.
// If an error occurs, the message will be treated as not being able to run.
func (sub *Subway) CanRun(ctx context.Context, interaction discord.Interaction) (bool, error) {
	for _, check := range sub.Commands.Checks {
		canRun, err := check(ctx, sub, interaction)
		if err != nil {
			return false, err
		}

		if !canRun {
			return false, nil
		}
	}

	return true, nil
}

// Subway commands

func (sub *Subway) MustRegisterCog(cog Cog) {
	if err := sub.RegisterCog(cog); err != nil {
		panic(fmt.Sprintf(`sandwich: RegisterCog(%v): %v`, cog, err.Error()))
	}
}

func (sub *Subway) RegisterCog(cog Cog) error {
	cogInfo := cog.CogInfo()

	if _, ok := sub.Cogs[cogInfo.Name]; ok {
		return ErrCogAlreadyRegistered
	}

	if err := cog.RegisterCog(sub); err != nil {
		sub.Logger.Panic().Str("cog", cogInfo.Name).Err(err).Msg("Failed to register Cog")

		return fmt.Errorf("failed to register cog: %w", err)
	}

	sub.Cogs[cogInfo.Name] = cog

	sub.Logger.Info().Str("cog", cogInfo.Name).Msg("Loaded Cog")

	if cast, ok := cog.(CogWithBotLoad); ok {
		sub.Logger.Info().Str("cog", cogInfo.Name).Msg("Cog has BotLoad")

		cast.BotLoad(sub)
	}

	if cast, ok := cog.(CogWithInteractionCommands); ok {
		interactionCommandable := cast.GetInteractionCommandable()

		sub.Logger.Info().
			Str("cog", cogInfo.Name).
			Int("commands", len(interactionCommandable.GetAllCommands())).
			Msg("Cog has interaction commands")

		sub.RegisterCogInteractionCommandable(cog, interactionCommandable)
	}

	return nil
}

func (sub *Subway) RegisterCogInteractionCommandable(cog Cog, interactionCommandable *InteractionCommandable) {
	for _, command := range interactionCommandable.GetAllCommands() {
		// Add Cog checks to all commands.
		command.Checks = append(interactionCommandable.Checks, command.Checks...)

		sub.Logger.Debug().Str("name", command.Name).Msg("Registering interaction command")

		sub.Commands.MustAddInteractionCommand(command)
	}
}
