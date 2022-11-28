package internal

import (
	"context"
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
)

// ProcessInteraction processes the interaction that has been registered to the bot.
func (subway *Subway) ProcessInteraction(ctx context.Context, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	commandTree := constructCommandTree(interaction.Data.Options, make([]string, 0))
	command := subway.Commands.GetCommand(interaction.Data.Name)

	// Create interaction context
	interactionContext := AddInteractionToContext(ctx, &interaction)
	interactionContext = AddInteractionCommandToContext(interactionContext, command)
	interactionContext = AddArgumentsToContext(interactionContext, make(map[string]*Argument))
	interactionContext = AddRawOptionsToContext(interactionContext, extractOptions(interaction.Data.Options, make(map[string]*discord.InteractionDataOption)))
	interactionContext = AddCommandBranchToContext(interactionContext, commandTree)
	interactionContext = AddCommandTreeToContext(interactionContext, commandTree)

	if command == nil {
		return subway.Commands.propagateError(interactionContext, ErrCommandNotFound), ErrCommandNotFound
	}

	if subway.OnBeforeInteraction != nil {
		err := subway.OnBeforeInteraction(interactionContext)
		if err != nil {
			return subway.Commands.propagateError(interactionContext, err), err
		}
	}

	response, err := command.Invoke(ctx)

	if subway.OnAfterInteraction != nil {
		err = subway.OnAfterInteraction(interactionContext, response, err)
		if err != nil {
			return subway.Commands.propagateError(interactionContext, err), err
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
func (subway *Subway) CanRun(ctx context.Context) (bool, error) {
	for _, check := range subway.Commands.Checks {
		canRun, err := check(ctx)
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

func (subway *Subway) MustRegisterCog(cog Cog) {
	if err := subway.RegisterCog(cog); err != nil {
		panic(fmt.Sprintf(`sandwich: RegisterCog(%v): %v`, cog, err.Error()))
	}
}

func (subway *Subway) RegisterCog(cog Cog) error {
	cogInfo := cog.CogInfo()

	if _, ok := subway.Cogs[cogInfo.Name]; ok {
		return ErrCogAlreadyRegistered
	}

	if err := cog.RegisterCog(subway); err != nil {
		subway.Logger.Panic().Str("cog", cogInfo.Name).Err(err).Msg("Failed to register Cog")

		return fmt.Errorf("failed to register cog: %w", err)
	}

	subway.Cogs[cogInfo.Name] = cog

	subway.Logger.Info().Str("cog", cogInfo.Name).Msg("Loaded Cog")

	if cast, ok := cog.(CogWithBotLoad); ok {
		subway.Logger.Info().Str("cog", cogInfo.Name).Msg("Cog has BotLoad")

		cast.BotLoad(subway)
	}

	if cast, ok := cog.(CogWithInteractionCommands); ok {
		interactionCommandable := cast.GetInteractionCommandable()

		subway.Logger.Info().
			Str("cog", cogInfo.Name).
			Int("commands", len(interactionCommandable.GetAllCommands())).
			Msg("Cog has interaction commands")

		subway.RegisterCogInteractionCommandable(cog, interactionCommandable)
	}

	return nil
}

func (subway *Subway) RegisterCogInteractionCommandable(cog Cog, interactionCommandable *InteractionCommandable) {
	for _, command := range interactionCommandable.GetAllCommands() {
		// Add Cog checks to all commands.
		command.Checks = append(interactionCommandable.Checks, command.Checks...)

		subway.Logger.Debug().Str("name", command.Name).Msg("Registering interaction command")

		subway.Commands.MustAddInteractionCommand(command)
	}
}
