package internal

import (
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

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
		Prefix:                nil,
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
