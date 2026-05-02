package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/WelcomerTeam/Discord/discord"
)

// ProcessApplicationCommandInteraction processes the application command that has been received.
func (sub *Subway) ProcessApplicationCommandInteraction(ctx context.Context, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	commandTree := constructCommandTree(interaction.Data.Options, make([]string, 0))
	command := sub.Commands.GetCommand(interaction.Data.Name)

	// Create interaction context
	ctx = AddInteractionCommandToContext(ctx, command)
	ctx = AddArgumentsToContext(ctx, make(map[string]*Argument))
	ctx = AddRawOptionsToContext(ctx, extractOptions(interaction.Data.Options, make(map[string]discord.InteractionDataOption)))
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
		onAfterInteractionErr := sub.OnAfterInteraction(ctx, sub, interaction, response, err)
		if onAfterInteractionErr != nil {
			return sub.Commands.propagateError(ctx, sub, interaction, onAfterInteractionErr), onAfterInteractionErr
		}
	}

	return response, err
}

// Checks if search custom ID matches the key, using the path.Match function
// Keys can be in a pattern such as "button_*" to match any custom ID that starts with "button_".
func keyMatches(pattern, name string) bool {
	matches, err := path.Match(pattern, name)
	if err != nil {
		return false
	}

	return matches
}

// ProcessMessageComponentInteraction processes the message component that has been received.
func (sub *Subway) ProcessMessageComponentInteraction(ctx context.Context, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	var listener *ComponentListener

	var hasListener bool

	sub.ComponentListenersMu.RLock()
	for key, componentListener := range sub.ComponentListeners {
		if key == interaction.Data.CustomID || keyMatches(key, interaction.Data.CustomID) {
			listener = componentListener
			hasListener = true

			break
		}
	}
	sub.ComponentListenersMu.RUnlock()

	if !hasListener {
		return nil, ErrComponentListenerNotFound
	}

	arguments := make(map[string]*Argument)

	var err error

	// Currently message component arguments will only be stored as string or strings.
	// TODO: Add arguments to ComponentListener to allow for type transformation.
	arguments, err = parseComponentData(arguments, interaction.Data)
	if err != nil {
		return nil, err
	}

	ctx = AddComponentListenerToContext(ctx, listener)
	ctx = AddArgumentsToContext(ctx, arguments)

	if listener.Channel != nil {
		listener.Channel <- &interaction

		return nil, nil
	}

	return listener.Handler(ctx, sub, interaction)
}

// parseComponentData generates the arguments for a component interaction.
func parseComponentData(arguments map[string]*Argument, data *discord.InteractionData) (map[string]*Argument, error) {
	for _, component := range data.Components {
		arguments = extractValues(component, arguments)
	}

	return arguments, nil
}

// Infer the type of the value based on its JSON representation.
func inferType(value json.RawMessage) (ArgumentType, any) {
	// Handle strings
	if value[0] == '"' {
		var strValue string
		if err := json.Unmarshal(value, &strValue); err == nil {
			return ArgumentTypeString, strValue
		}
	}

	// Handle booleans
	if value[0] == 't' || value[0] == 'f' {
		var boolValue bool
		if err := json.Unmarshal(value, &boolValue); err == nil {
			return ArgumentTypeBool, boolValue
		}
	}

	// Handle numbers (integers and floats)
	if value[0] == '-' || (value[0] >= '0' && value[0] <= '9') {
		var intValue int
		if err := json.Unmarshal(value, &intValue); err == nil {
			return ArgumentTypeInt, intValue
		}

		var floatValue float64
		if err := json.Unmarshal(value, &floatValue); err == nil {
			return ArgumentTypeFloat, floatValue
		}
	}

	// Handle arrays of strings
	if value[0] == '[' {
		var strSliceValue []string
		if err := json.Unmarshal(value, &strSliceValue); err == nil {
			return ArgumentTypeStrings, strSliceValue
		}
	}

	// If all else fails, treat it as a string (this is a fallback and probably will not be correct).
	return ArgumentTypeString, string(value)
}

func extractValues(component discord.InteractionComponent, arguments map[string]*Argument) map[string]*Argument {
	for _, componentChild := range component.Components {
		arguments = extractValues(componentChild, arguments)
	}

	if component.Component != nil {
		arguments = extractValues(*component.Component, arguments)
	}

	if len(component.Value) > 0 {
		argumentType, value := inferType(component.Value)
		arguments[component.CustomID] = &Argument{
			ArgumentType: argumentType,
			value:        value,
		}
	}

	if len(component.Values) > 0 {
		arguments[component.CustomID] = &Argument{
			ArgumentType: ArgumentTypeStrings,
			value:        component.Values,
		}
	}

	return arguments
}

func constructCommandTree(options []discord.InteractionDataOption, tree []string) []string {
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
		sub.Logger.Error("Failed to register Cog", "cog", cogInfo.Name, "error", err)

		return fmt.Errorf("failed to register cog: %w", err)
	}

	sub.Cogs[cogInfo.Name] = cog

	sub.Logger.Info("Loaded Cog", "cog", cogInfo.Name)

	if cast, ok := cog.(CogWithBotLoad); ok {
		sub.Logger.Info("Cog has BotLoad", "cog", cogInfo.Name)

		cast.BotLoad(sub)
	}

	if cast, ok := cog.(CogWithInteractionCommands); ok {
		interactionCommandable := cast.GetInteractionCommandable()

		sub.Logger.Info("Cog has interaction commands", "cog", cogInfo.Name, "commands", len(interactionCommandable.GetAllCommands()))

		sub.RegisterCogInteractionCommandable(cog, interactionCommandable)
	}

	return nil
}

func (sub *Subway) RegisterCogInteractionCommandable(cog Cog, interactionCommandable *InteractionCommandable) {
	for _, command := range interactionCommandable.GetAllCommands() {
		// Add Cog checks to all commands.
		command.Checks = append(interactionCommandable.Checks, command.Checks...)

		sub.Logger.Debug("Registering interaction command", "name", command.Name)

		sub.Commands.MustAddInteractionCommand(command)
	}
}
