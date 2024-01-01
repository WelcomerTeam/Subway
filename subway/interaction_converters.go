package internal

import (
	"context"
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"sync"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var (
	IDRegex             = regexp.MustCompile("([0-9]{15,20})$")
	GenericMentionRegex = regexp.MustCompile("<(?:@(?:!|&)?|#)([0-9]{15,20})>$")
	UserMentionRegex    = regexp.MustCompile("<@!?([0-9]{15,20})>$")
	ChannelMentionRegex = regexp.MustCompile("<#([0-9]{15,20})>")
	RoleMentionRegex    = regexp.MustCompile("<@&([0-9]{15,20})>$")
	EmojiRegex          = regexp.MustCompile("<a?:[a-zA-Z0-9_]{1,32}:([0-9]{15,20})>$")
	PartialEmojiRegex   = regexp.MustCompile("<(a?):([a-zA-Z0-9_]{1,32}):([0-9]{15,20})>$")
)

type ArgumentParameter struct {
	Required                 bool
	ArgumentType             ArgumentType
	Name                     string
	Description              string
	NameLocalizations        map[string]string
	DescriptionLocalizations map[string]string
	Choices                  []*discord.ApplicationCommandOptionChoice

	ChannelTypes []*discord.ChannelType
	MinValue     *int32
	MaxValue     *int32
	MinLength    *int32
	MaxLength    *int32
	Autocomplete *bool
}

type Argument struct {
	ArgumentType ArgumentType
	value        interface{}
}

type InteractionArgumentConverterType func(ctx context.Context, sub *Subway, interaction discord.Interaction, argument *discord.InteractionDataOption) (out interface{}, err error)

type InteractionConverters struct {
	convertersMu sync.RWMutex
	Converters   map[ArgumentType]*InteractionConverter
}

type InteractionConverter struct {
	converterType InteractionArgumentConverterType
	data          interface{}
}

// RegisterConverter adds a new converter. If there is already a
// converter registered with its name, it will be overridden.
func (co *InteractionConverters) RegisterConverter(converterName ArgumentType, converter InteractionArgumentConverterType, defaultValue interface{}) {
	co.convertersMu.Lock()
	defer co.convertersMu.Unlock()

	co.Converters[converterName] = &InteractionConverter{
		converterType: converter,
		data:          defaultValue,
	}
}

// GetConverter returns a converter based on the converterType passed.
func (co *InteractionConverters) GetConverter(converterType ArgumentType) *InteractionConverter {
	co.convertersMu.RLock()
	defer co.convertersMu.RUnlock()

	return co.Converters[converterType]
}

// HandleInteractionArgumentTypeSnowflake handles converting from a string
// argument into a Snowflake type. Use .Snowflake() within a command
// to get the proper type.
func HandleInteractionArgumentTypeSnowflake(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	match := IDRegex.FindString(argument)
	if match == "" {
		matches := GenericMentionRegex.FindStringSubmatch(argument)
		if len(matches) > 1 {
			match = matches[1]
		}
	}

	var result *discord.Snowflake

	if match == "" {
		return nil, ErrSnowflakeNotFound
	}

	snowflakeID, _ := strconv.ParseInt(match, 10, 64)
	result = (*discord.Snowflake)(&snowflakeID)

	return result, nil
}

// HandleInteractionArgumentTypeMember handles converting from a string
// argument into a Member type. Use .Member() within a command
// to get the proper type.
func HandleInteractionArgumentTypeMember(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	snowflakeID, _ := strconv.ParseInt(argument, 10, 64)

	snowflake := discord.Snowflake(snowflakeID)

	result := interaction.Data.Resolved.Members[snowflake]

	if result == nil {
		return nil, ErrMemberNotFound
	}

	userResult := interaction.Data.Resolved.Users[snowflake]
	if userResult != nil {
		result.User = userResult
	} else {
		sub.Logger.Warn().Int64("id", snowflakeID).Msg("Member present in interaction resolved, but no User is present")
	}

	return result, nil
}

// HandleInteractionArgumentTypeUser handles converting from a string
// argument into a User type. Use .User() within a command
// to get the proper type.
func HandleInteractionArgumentTypeUser(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	snowflakeID, _ := strconv.ParseInt(argument, 10, 64)

	result := interaction.Data.Resolved.Users[discord.Snowflake(snowflakeID)]

	if result == nil {
		return nil, ErrUserNotFound
	}

	return result, nil
}

// HandleInteractionArgumentTypeGuildChannel handles converting from a string
// argument into a TextChannel type. Use .Channel() within a command
// to get the proper type.
func HandleInteractionArgumentTypeGuildChannel(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	snowflakeID, _ := strconv.ParseInt(argument, 10, 64)

	result := interaction.Data.Resolved.Channels[discord.Snowflake(snowflakeID)]

	if result == nil {
		return nil, ErrChannelNotFound
	}

	return result, nil
}

// HandleInteractionArgumentTypeGuild handles converting from a string
// argument into a Guild type. Use .Guild() within a command
// to get the proper type.
func HandleInteractionArgumentTypeGuild(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	match := IDRegex.FindString(argument)

	var result *discord.Guild

	if match == "" {
		guilds, err := sub.GRPCInterface.FetchGuildsByName(sub.NewGRPCContext(ctx), argument)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guild: %w", err)
		}

		if len(guilds) > 0 {
			result = guilds[0]
		}
	} else {
		guildID, _ := strconv.ParseInt(match, 10, 64)

		result = sandwich.NewGuild(discord.Snowflake(guildID))

		result, err = sandwich.FetchGuild(sub.NewGRPCContext(ctx), result)
		if err != nil && !errors.Is(err, ErrGuildNotFound) {
			return nil, fmt.Errorf("failed to fetch guild: %w", err)
		}
	}

	if result == nil {
		return nil, ErrGuildNotFound
	}

	return result, nil
}

// HandleInteractionArgumentTypeRole handles converting from a string
// argument into a Role type. Use .Role() within a command
// to get the proper type.
func HandleInteractionArgumentTypeRole(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	snowflakeID, _ := strconv.ParseInt(argument, 10, 64)

	result := interaction.Data.Resolved.Roles[discord.Snowflake(snowflakeID)]

	if result == nil {
		return nil, ErrRoleNotFound
	}

	return result, nil
}

// HandleInteractionArgumentTypeColour handles converting from a string
// argument into a Colour type. Use .Colour() within a command
// to get the proper type.
func HandleInteractionArgumentTypeColour(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	var result *color.RGBA

	switch {
	case argument[0] == '#':
		hexNum, err := parseHexNumber(argument[1:])
		if err != nil {
			return nil, err
		}

		result = intToColour(hexNum)
	case argument[0:2] == "0x":
		hexNum, err := parseHexNumber(argument[2:])
		if err != nil {
			return nil, err
		}

		result = intToColour(hexNum)
	default:
		hexNum, err := parseHexNumber(argument)
		if err == nil {
			result = intToColour(hexNum)
		}
	}

	if result == nil {
		return nil, ErrBadColourArgument
	}

	return result, nil
}

// HandleInteractionArgumentTypeEmoji handles converting from a string
// argument into a Emoji type. Use .Emoji() within a command
// to get the proper type.
func HandleInteractionArgumentTypeEmoji(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	var result *discord.Emoji

	match := IDRegex.FindString(argument)
	if match == "" {
		matches := PartialEmojiRegex.FindStringSubmatch(argument)

		if len(matches) >= 4 {
			animated := matches[1] != ""
			id, _ := strconv.ParseInt(matches[3], 10, 64)

			result = &discord.Emoji{
				Animated: animated,
				Name:     matches[2],
				ID:       discord.Snowflake(id),
			}
		}
	}

	if result == nil {
		if match == "" {
			if interaction.GuildID != nil {
				emojis, err := sub.GRPCInterface.FetchEmojisByName(sub.NewGRPCContext(ctx), *interaction.GuildID, argument)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch emoji: %w", err)
				}

				if len(emojis) > 0 {
					result = emojis[0]
				}
			}
		} else {
			emojiID, _ := strconv.ParseInt(match, 10, 64)

			result = sandwich.NewEmoji(interaction.GuildID, discord.Snowflake(emojiID))
		}
	}

	result, err = sandwich.FetchEmoji(sub.NewGRPCContext(ctx), result)
	if err != nil && !errors.Is(err, ErrEmojiNotFound) && !errors.Is(err, ErrFetchMissingGuild) {
		return nil, fmt.Errorf("failed to fetch emoji: %w", err)
	}

	return result, nil
}

// HandleInteractionArgumentTypePartialEmoji handles converting from a string
// argument into a Emoji type. Use .Emoji() within a command
// to get the proper type.
func HandleInteractionArgumentTypePartialEmoji(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	matches := PartialEmojiRegex.FindStringSubmatch(argument)

	var result *discord.Emoji

	if len(matches) >= 4 {
		animated := matches[1] != ""
		id, _ := strconv.ParseInt(matches[3], 10, 64)

		result = &discord.Emoji{
			Animated: animated,
			Name:     matches[2],
			ID:       discord.Snowflake(id),
		}
	}

	return result, nil
}

// HandleInteractionArgumentTypeString handles converting from a string
// argument into a String type. Use .String() within a command
// to get the proper type.
func HandleInteractionArgumentTypeString(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	return argument, nil
}

// HandleInteractionArgumentTypeBool handles converting from a string
// argument into a Bool type. Use .Bool() within a command
// to get the proper type.
func HandleInteractionArgumentTypeBool(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument bool

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	return argument, nil
}

// HandleInteractionArgumentTypeInt handles converting from a string
// argument into a Int type. Use .Int64() within a command
// to get the proper type.
func HandleInteractionArgumentTypeInt(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument int64

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	return argument, nil
}

// HandleInteractionArgumentTypeFloat handles converting from a string
// argument into a Float type. Use .Float64() within a command
// to get the proper type.
func HandleInteractionArgumentTypeFloat(ctx context.Context, sub *Subway, interaction discord.Interaction, option *discord.InteractionDataOption) (out interface{}, err error) {
	var argument string

	err = jsoniter.Unmarshal(option.Value, &argument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal option value: %w", err)
	}

	result, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		return nil, ErrBadFloatArgument
	}

	return result, nil
}

func NewInteractionConverters() *InteractionConverters {
	converters := &InteractionConverters{
		convertersMu: sync.RWMutex{},
		Converters:   make(map[ArgumentType]*InteractionConverter),
	}

	converters.RegisterConverter(ArgumentTypeSnowflake, HandleInteractionArgumentTypeSnowflake, nil)
	converters.RegisterConverter(ArgumentTypeMember, HandleInteractionArgumentTypeMember, nil)
	converters.RegisterConverter(ArgumentTypeUser, HandleInteractionArgumentTypeUser, nil)
	converters.RegisterConverter(ArgumentTypeTextChannel, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeGuild, HandleInteractionArgumentTypeGuild, nil)
	converters.RegisterConverter(ArgumentTypeRole, HandleInteractionArgumentTypeRole, nil)
	converters.RegisterConverter(ArgumentTypeColour, HandleInteractionArgumentTypeColour, nil)
	converters.RegisterConverter(ArgumentTypeVoiceChannel, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeStageChannel, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeEmoji, HandleInteractionArgumentTypeEmoji, nil)
	converters.RegisterConverter(ArgumentTypePartialEmoji, HandleInteractionArgumentTypePartialEmoji, nil)
	converters.RegisterConverter(ArgumentTypeCategoryChannel, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeStoreChannel, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeThread, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeGuildChannel, HandleInteractionArgumentTypeGuildChannel, nil)
	converters.RegisterConverter(ArgumentTypeString, HandleInteractionArgumentTypeString, "")
	converters.RegisterConverter(ArgumentTypeBool, HandleInteractionArgumentTypeBool, false)
	converters.RegisterConverter(ArgumentTypeInt, HandleInteractionArgumentTypeInt, int64(0))
	converters.RegisterConverter(ArgumentTypeFloat, HandleInteractionArgumentTypeFloat, float64(0))

	return converters
}
