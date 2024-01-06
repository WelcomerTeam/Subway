package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/WelcomerTeam/Discord/discord"
)

// Func Type used for command checks.
type InteractionCheckFuncType func(ctx context.Context, sub *Subway, interaction discord.Interaction) (canRun bool, err error)

type (
	InteractionHandler             func(ctx context.Context, sub *Subway, interaction discord.Interaction) (*discord.InteractionResponse, error)
	InteractionAutocompleteHandler func(ctx context.Context, sub *Subway, interaction discord.Interaction) ([]*discord.ApplicationCommandOptionChoice, error)
	InteractionErrorHandler        func(ctx context.Context, sub *Subway, interaction discord.Interaction, err error) (*discord.InteractionResponse, error)
)

type (
	InteractionRequestHandler  func(ctx context.Context, sub *Subway, interaction discord.Interaction) error
	InteractionResponseHandler func(ctx context.Context, sub *Subway, interaction discord.Interaction, resp *discord.InteractionResponse, err error) error
)

type InteractionCommandableType uint8

const (
	InteractionCommandableTypeCommand InteractionCommandableType = iota
	InteractionCommandableTypeSubcommandGroup
	InteractionCommandableTypeSubcommand
)

type InteractionCommandable struct {
	Name        string
	Description string

	NameLocalizations        map[string]string
	DescriptionLocalizations map[string]string

	Type        InteractionCommandableType
	CommandType *discord.ApplicationCommandType

	Checks            []InteractionCheckFuncType
	ArgumentParameter []ArgumentParameter

	Handler      InteractionHandler
	ErrorHandler InteractionErrorHandler

	DefaultMemberPermission *discord.Int64
	DMPermission            *bool

	AutocompleteHandler InteractionAutocompleteHandler

	commands map[string]*InteractionCommandable
	parent   *InteractionCommandable
}

func (ic *InteractionCommandable) MapApplicationCommands() []*discord.ApplicationCommand {
	applicationCommands := make([]*discord.ApplicationCommand, 0, len(ic.commands))

	applicationCommandType := discord.ApplicationCommandTypeChatInput

	var applicationType *discord.ApplicationCommandType

	for _, interactionCommandable := range ic.commands {
		if interactionCommandable.CommandType != nil {
			applicationType = interactionCommandable.CommandType
		} else {
			applicationType = &applicationCommandType
		}

		applicationCommands = append(applicationCommands, &discord.ApplicationCommand{
			Name:                     interactionCommandable.Name,
			NameLocalizations:        interactionCommandable.NameLocalizations,
			Description:              interactionCommandable.Description,
			DescriptionLocalizations: interactionCommandable.DescriptionLocalizations,
			Options:                  interactionCommandable.MapApplicationOptions(),
			DefaultMemberPermission:  interactionCommandable.DefaultMemberPermission,
			DMPermission:             interactionCommandable.DMPermission,
			Type:                     applicationType,
		})
	}

	return applicationCommands
}

func (ic *InteractionCommandable) MapApplicationOptions() (applicationOptions []*discord.ApplicationCommandOption) {
	applicationOptions = make([]*discord.ApplicationCommandOption, 0)

	var applicationOptionType discord.ApplicationCommandOptionType

	// Map subgroups/subcommands.
	for _, command := range ic.commands {
		switch command.Type {
		case InteractionCommandableTypeCommand:
			applicationOptionType = discord.ApplicationCommandOptionTypeSubCommand
		case InteractionCommandableTypeSubcommand:
			applicationOptionType = discord.ApplicationCommandOptionTypeSubCommand
		case InteractionCommandableTypeSubcommandGroup:
			applicationOptionType = discord.ApplicationCommandOptionTypeSubCommandGroup
		}

		applicationOptions = append(applicationOptions, &discord.ApplicationCommandOption{
			Name:                     command.Name,
			Description:              command.Description,
			NameLocalizations:        command.NameLocalizations,
			DescriptionLocalizations: command.DescriptionLocalizations,
			Options:                  command.MapApplicationOptions(),
			Type:                     applicationOptionType,
		})
	}

	var channelType discord.ChannelType

	// Map arguments.
	for _, argument := range ic.ArgumentParameter {
		channelType = -1

		switch argument.ArgumentType {
		case ArgumentTypeSnowflake:
			applicationOptionType = discord.ApplicationCommandOptionTypeString
		case ArgumentTypeMember, ArgumentTypeUser:
			applicationOptionType = discord.ApplicationCommandOptionTypeUser
		case ArgumentTypeTextChannel:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
			channelType = discord.ChannelTypeGuildText
		case ArgumentTypeVoiceChannel:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
			channelType = discord.ChannelTypeGuildVoice
		case ArgumentTypeStageChannel:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
			channelType = discord.ChannelTypeGuildStageVoice
		case ArgumentTypeCategoryChannel:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
			channelType = discord.ChannelTypeGuildCategory
		case ArgumentTypeStoreChannel:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
			channelType = discord.ChannelTypeGuildStore
		case ArgumentTypeThread:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
			channelType = discord.ChannelTypeGuildPublicThread
		case ArgumentTypeGuildChannel:
			applicationOptionType = discord.ApplicationCommandOptionTypeChannel
		case ArgumentTypeGuild:
			applicationOptionType = discord.ApplicationCommandOptionTypeString
		case ArgumentTypeRole:
			applicationOptionType = discord.ApplicationCommandOptionTypeRole
		case ArgumentTypeColour, ArgumentTypeEmoji, ArgumentTypePartialEmoji, ArgumentTypeString:
			applicationOptionType = discord.ApplicationCommandOptionTypeString
		case ArgumentTypeBool:
			applicationOptionType = discord.ApplicationCommandOptionTypeBoolean
		case ArgumentTypeFloat:
			applicationOptionType = discord.ApplicationCommandOptionTypeString
		case ArgumentTypeInt:
			applicationOptionType = discord.ApplicationCommandOptionTypeInteger
		}

		commandOption := &discord.ApplicationCommandOption{
			Type:                     applicationOptionType,
			Name:                     argument.Name,
			Description:              argument.Description,
			NameLocalizations:        argument.NameLocalizations,
			DescriptionLocalizations: argument.DescriptionLocalizations,
			Required:                 argument.Required,
			Choices:                  argument.Choices,
			ChannelTypes:             argument.ChannelTypes,
			MinValue:                 argument.MinValue,
			MaxValue:                 argument.MaxValue,
			MinLength:                argument.MinLength,
			MaxLength:                argument.MaxLength,
			Autocomplete:             argument.Autocomplete,
		}

		if channelType != -1 {
			commandOption.ChannelTypes = []discord.ChannelType{channelType}
		}

		applicationOptions = append(applicationOptions, commandOption)
	}

	return applicationOptions
}

func (ic *InteractionCommandable) MustAddInteractionCommand(interactionCommandable *InteractionCommandable) (icc *InteractionCommandable) {
	icc, err := ic.AddInteractionCommand(interactionCommandable)
	if err != nil {
		panic(fmt.Sprintf(`sandwich: AddInteractionCommand(%v): %v`, interactionCommandable, err.Error()))
	}

	return icc
}

func (ic *InteractionCommandable) AddInteractionCommand(interactionCommandable *InteractionCommandable) (icc *InteractionCommandable, err error) {
	// If this command is not a base command, turn it into a subcommand
	if ic.Type == InteractionCommandableTypeCommand && ic.parent != nil {
		ic.Type = InteractionCommandableTypeSubcommand
	}

	// Convert interactionCommandable parent to SubcommandGroup if it is a subcommand.
	// Convert interactionCommandable to SubcommandGroup if it is not a Command.
	if ic.Type != InteractionCommandableTypeCommand {
		if ic.parent != nil {
			if ic.parent.Type == InteractionCommandableTypeSubcommand {
				ic.parent.Type = InteractionCommandableTypeSubcommandGroup
			}

			ic.Type = InteractionCommandableTypeSubcommandGroup
		}
	}

	commandName := strings.ToLower(interactionCommandable.Name)
	if _, ok := ic.getCommand(commandName); ok {
		err = ErrCommandAlreadyRegistered

		return nil, err
	}

	interactionCommandable = SetupInteractionCommandable(interactionCommandable)

	icc = interactionCommandable

	if ic.Type == InteractionCommandableTypeSubcommandGroup {
		icc.Type = InteractionCommandableTypeSubcommand
	} else {
		icc.Type = InteractionCommandableTypeCommand
	}

	icc.parent = ic

	ic.setCommand(commandName, icc)

	return icc, nil
}

func (ic *InteractionCommandable) RemoveCommand(name string) *InteractionCommandable {
	command, ok := ic.getCommand(name)

	if !ok {
		return nil
	}

	ic.deleteCommand(name)

	return command
}

func (ic *InteractionCommandable) RecursivelyRemoveAllCommands() {
	for _, command := range ic.commands {
		if command.IsGroup() {
			command.RecursivelyRemoveAllCommands()
		}

		ic.RemoveCommand(command.Name)
	}
}

// GetAllCommands returns all commands.
func (ic *InteractionCommandable) GetAllCommands() []*InteractionCommandable {
	interactionCommandables := make([]*InteractionCommandable, 0)

	for _, commandable := range ic.commands {
		interactionCommandables = append(interactionCommandables, commandable)
	}

	return interactionCommandables
}

func (ic *InteractionCommandable) GetCommand(name string) *InteractionCommandable {
	if !strings.Contains(name, " ") {
		command, _ := ic.getCommand(name)

		return command
	}

	names := strings.Split(name, " ")
	if len(names) == 0 {
		return nil
	}

	command := ic.GetCommand(names[0])
	if !command.IsGroup() {
		return command
	}

	var ok bool

	for _, name := range names[1:] {
		command, ok = command.getCommand(name)
		if !ok {
			return nil
		}
	}

	return command
}

// IsGroup returns true if the command contains other commands.
func (ic *InteractionCommandable) IsGroup() bool {
	return ic.Type == InteractionCommandableTypeCommand || ic.Type == InteractionCommandableTypeSubcommandGroup
}

// Invoke handles the execution of a command or a group.
func (ic *InteractionCommandable) Invoke(ctx context.Context, sub *Subway, interaction discord.Interaction) (*discord.InteractionResponse, error) {
	commandBranch := GetCommandBranchFromContext(ctx)

	if len(commandBranch) > 0 {
		if ic.IsGroup() {
			branch := commandBranch[0]
			commandContext := AddCommandBranchToContext(ctx, commandBranch[1:])

			commandable := ic.GetCommand(branch)

			if commandable == nil {
				return nil, ErrCommandNotFound
			}

			return commandable.Invoke(commandContext, sub, interaction)
		}

		sub.Logger.Warn().
			Str("command", ic.Name).
			Str("branch", commandBranch[0]).
			Msg("Encountered non-group whilst traversing command tree.")
	}

	ctx, err := ic.prepare(ctx, sub, interaction)
	if err != nil {
		return nil, err
	}

	defer func() {
		errorValue := recover()
		if errorValue != nil {
			sub.Logger.Error().Interface("errorValue", errorValue).Msg("Recovered panic on event dispatch")

			ic.propagateError(ctx, sub, interaction, PanicError{errorValue})
		}
	}()

	var resp *discord.InteractionResponse

	switch interaction.Type {
	case discord.InteractionTypeApplicationCommand,
		discord.InteractionTypeMessageComponent,
		discord.InteractionTypeModalSubmit:
		if ic.Handler != nil {
			resp, err = ic.Handler(ctx, sub, interaction)
			if err != nil {
				return ic.propagateError(ctx, sub, interaction, err), err
			}
		} else {
			return ic.propagateError(ctx, sub, interaction, ErrCommandNotFound), ErrCommandNotFound
		}
	case discord.InteractionTypeApplicationCommandAutocomplete:
		if ic.AutocompleteHandler != nil {
			choices, err := ic.AutocompleteHandler(ctx, sub, interaction)
			if err != nil {
				return ic.propagateError(ctx, sub, interaction, err), err
			}

			resp = &discord.InteractionResponse{
				Type: discord.InteractionCallbackTypeAutocompleteResult,
				Data: &discord.InteractionCallbackData{
					Choices: choices,
				},
			}
		} else {
			return ic.propagateError(ctx, sub, interaction, ErrCommandAutoCompleteNotFound), ErrCommandAutoCompleteNotFound
		}
	}

	return resp, nil
}

// propagateError propagates an error to the current command or parent. It will execute the root parent first,
// then go up from there. It will return the highest up error handler in the chain that returns a interaction response.
// If the command and root error handler returns an interaction response, the command error handler response will be
// returned. If the root returns an interaction response and the command does not, the root response is returned.
func (ic *InteractionCommandable) propagateError(ctx context.Context, sub *Subway, interaction discord.Interaction, err error) (interactionResponse *discord.InteractionResponse) {
	if ic.parent != nil {
		rootInteractionResponse := ic.parent.propagateError(ctx, sub, interaction, err)
		if rootInteractionResponse != nil {
			interactionResponse = rootInteractionResponse
		}
	}

	if ic.ErrorHandler != nil {
		commandInteractionResponse, _ := ic.ErrorHandler(ctx, sub, interaction, err)
		if commandInteractionResponse != nil {
			interactionResponse = commandInteractionResponse
		}
	} else if ic.parent == nil {
		// If parent is nil we can assume we are the root. If an error
		// handler is not provided, we will just use defaultErrorPropagator.
		commandInteractionResponse, _ := defaultErrorPropagator(ctx, sub, interaction, err)
		if commandInteractionResponse != nil {
			interactionResponse = commandInteractionResponse
		}
	}

	return interactionResponse
}

// Default error propagator. This will just log an exception.
func defaultErrorPropagator(ctx context.Context, sub *Subway, interaction discord.Interaction, err error) (*discord.InteractionResponse, error) {
	sub.Logger.Error().Err(err).Msg("Exception executing interaction")

	return nil, err
}

// CanRun checks interactionCommandable checks and returns if the interaction passes them all.
// If an error occurs, the message will be treated as not being able to run.
func (ic *InteractionCommandable) CanRun(ctx context.Context, sub *Subway, interaction discord.Interaction) (bool, error) {
	for _, check := range ic.Checks {
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

func (ic *InteractionCommandable) prepare(ctx context.Context, sub *Subway, interaction discord.Interaction) (context.Context, error) {
	ok, err := ic.CanRun(ctx, sub, interaction)

	switch {
	case !ok:
		return ctx, ErrCheckFailure
	case err != nil:
		return ctx, err
	}

	ctx, err = ic.parseArguments(ctx, sub, interaction)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// parseArguments generates the arguments for a command.
func (ic *InteractionCommandable) parseArguments(ctx context.Context, sub *Subway, interaction discord.Interaction) (context.Context, error) {
	arguments := map[string]*Argument{}

	for _, argumentParameter := range ic.ArgumentParameter {
		transformed, err := ic.transform(ctx, sub, interaction, argumentParameter)
		if err != nil {
			return ctx, err
		}

		arguments[argumentParameter.Name] = &Argument{
			ArgumentType: argumentParameter.ArgumentType,
			value:        transformed,
		}
	}

	ctx = AddArgumentsToContext(ctx, arguments)

	return ctx, nil
}

// transform returns a output value based on the argument parameter passed in.
func (ic *InteractionCommandable) transform(ctx context.Context, sub *Subway, interaction discord.Interaction, argumentParameter ArgumentParameter) (out interface{}, err error) {
	converter := sub.Converters.GetConverter(argumentParameter.ArgumentType)
	if converter == nil {
		return nil, ErrConverterNotFound
	}

	rawOptions := GetRawOptionsFromContext(ctx)

	rawOption, ok := rawOptions[argumentParameter.Name]
	if !ok || rawOption == nil {
		if argumentParameter.Required {
			return nil, ErrMissingRequiredArgument
		}

		return nil, nil
	}

	return converter.converterType(ctx, sub, interaction, rawOption)
}

// MustGetArgument returns an argument based on its name. Panics on error.
func MustGetArgument(ctx context.Context, name string) *Argument {
	arg, err := GetArgument(ctx, name)
	if err != nil {
		panic(fmt.Sprintf(`ctx: GetArgument(%s): %v`, name, err.Error()))
	}

	return arg
}

// GetArgument returns an argument based on its name.
func GetArgument(ctx context.Context, name string) (*Argument, error) {
	arguments := GetArgumentsFromContext(ctx)

	arg, ok := arguments[name]
	if !ok {
		return nil, ErrArgumentNotFound
	}

	return arg, nil
}

// SetupInteractionCommandable ensures all nullable variables are properly constructed.
func SetupInteractionCommandable(commandable *InteractionCommandable) *InteractionCommandable {
	if commandable == nil {
		commandable = &InteractionCommandable{}
	}

	if commandable.commands == nil {
		commandable.commands = make(map[string]*InteractionCommandable)
	}

	if commandable.Checks == nil {
		commandable.Checks = make([]InteractionCheckFuncType, 0)
	}

	return commandable
}

func (ic *InteractionCommandable) getCommand(name string) (*InteractionCommandable, bool) {
	commandable, ok := ic.commands[strings.ToLower(name)]

	return commandable, ok
}

func (ic *InteractionCommandable) deleteCommand(name string) {
	delete(ic.commands, strings.ToLower(name))
}

func (ic *InteractionCommandable) setCommand(name string, commandable *InteractionCommandable) {
	ic.commands[strings.ToLower(name)] = commandable
}

func extractOptions(options []*discord.InteractionDataOption, optionsMap map[string]*discord.InteractionDataOption) (newOptionsMap map[string]*discord.InteractionDataOption) {
	for _, dataOption := range options {
		optionsMap[dataOption.Name] = dataOption

		if len(dataOption.Options) > 0 {
			optionsMap = extractOptions(dataOption.Options, optionsMap)
		}
	}

	return optionsMap
}
