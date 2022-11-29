package internal

import (
	"fmt"
	"image/color"

	"github.com/WelcomerTeam/Discord/discord"
)

// Argument fetchers

// Snowflake returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Snowflake() (*discord.Snowflake, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeSnowflake) {
		value, _ := a.value.(*discord.Snowflake)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustSnowflake will attempt to do Snowflake() and will panic if not possible.
func (a *Argument) MustSnowflake() *discord.Snowflake {
	value, err := a.Snowflake()
	if err != nil {
		panic(fmt.Sprintf(`argument: Snowflake(): %v`, err.Error()))
	}

	return value
}

// Member returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Member() (*discord.GuildMember, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeMember) {
		value, _ := a.value.(*discord.GuildMember)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustMember will attempt to do Member() and will panic if not possible.
func (a *Argument) MustMember() *discord.GuildMember {
	value, err := a.Member()
	if err != nil {
		panic(fmt.Sprintf(`argument: Member(): %v`, err.Error()))
	}

	return value
}

// User returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) User() (*discord.User, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeUser) {
		value, _ := a.value.(*discord.User)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustUser will attempt to do User() and will panic if not possible.
func (a *Argument) MustUser() *discord.User {
	value, err := a.User()
	if err != nil {
		panic(fmt.Sprintf(`argument: User(): %v`, err.Error()))
	}

	return value
}

// Channel returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Channel() (*discord.Channel, error) {
	if argumentTypeIs(a.ArgumentType,
		ArgumentTypeTextChannel, ArgumentTypeVoiceChannel, ArgumentTypeStageChannel,
		ArgumentTypeCategoryChannel, ArgumentTypeStoreChannel, ArgumentTypeGuildChannel) {
		value, _ := a.value.(*discord.Channel)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustTextChannel will attempt to do Channel() and will panic if not possible.
func (a *Argument) MustChannel() *discord.Channel {
	value, err := a.Channel()
	if err != nil {
		panic(fmt.Sprintf(`argument: Channel(): %v`, err.Error()))
	}

	return value
}

// Guild returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Guild() (*discord.Guild, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeGuild) {
		value, _ := a.value.(*discord.Guild)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustGuild will attempt to do Guild() and will panic if not possible.
func (a *Argument) MustGuild() *discord.Guild {
	value, err := a.Guild()
	if err != nil {
		panic(fmt.Sprintf(`argument: Guild(): %v`, err.Error()))
	}

	return value
}

// Role returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Role() (*discord.Role, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeRole) {
		value, _ := a.value.(*discord.Role)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustRole will attempt to do Role() and will panic if not possible.
func (a *Argument) MustRole() *discord.Role {
	value, err := a.Role()
	if err != nil {
		panic(fmt.Sprintf(`argument: Role(): %v`, err.Error()))
	}

	return value
}

// Colour returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Colour() (*color.RGBA, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeColour) {
		value, _ := a.value.(*color.RGBA)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustColour will attempt to do Colour() and will panic if not possible.
func (a *Argument) MustColour() *color.RGBA {
	value, err := a.Colour()
	if err != nil {
		panic(fmt.Sprintf(`argument: Colour(): %v`, err.Error()))
	}

	return value
}

// Emoji returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Emoji() (*discord.Emoji, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeEmoji, ArgumentTypePartialEmoji) {
		value, _ := a.value.(*discord.Emoji)

		return value, nil
	}

	return nil, ErrInvalidArgumentType
}

// MustEmoji will attempt to do Emoji() and will panic if not possible.
func (a *Argument) MustEmoji() *discord.Emoji {
	value, err := a.Emoji()
	if err != nil {
		panic(fmt.Sprintf(`argument: Emoji(): %v`, err.Error()))
	}

	return value
}

// String returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) String() (string, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeString) {
		value, _ := a.value.(string)

		return value, nil
	}

	return "", ErrInvalidArgumentType
}

// MustString will attempt to do String() and will panic if not possible.
func (a *Argument) MustString() string {
	value, err := a.String()
	if err != nil {
		panic(fmt.Sprintf(`argument: String(): %v`, err.Error()))
	}

	return value
}

// Bool returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Bool() (bool, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeBool) {
		value, _ := a.value.(bool)

		return value, nil
	}

	return false, ErrInvalidArgumentType
}

// MustBool will attempt to do Bool() and will panic if not possible.
func (a *Argument) MustBool() bool {
	value, err := a.Bool()
	if err != nil {
		panic(fmt.Sprintf(`argument: Bool(): %v`, err.Error()))
	}

	return value
}

// Int returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Int() (int64, error) {
	if argumentTypeIs(a.ArgumentType, ArgumentTypeInt) {
		value, _ := a.value.(int64)

		return value, nil
	}

	return 0, ErrInvalidArgumentType
}

// MustInt will attempt to do Int() and will panic if not possible.
func (a *Argument) MustInt() int64 {
	value, err := a.Int()
	if err != nil {
		panic(fmt.Sprintf(`argument: Int(): %v`, err.Error()))
	}

	return value
}

// Float returns an argument as the specified Type.
// If the argument is not the right type for the converter
// that made the argument, ErrInvalidArgumentType will be returned.
func (a *Argument) Float() (float64, error) {
	v, ok := a.value.(float64)
	if !ok {
		return v, ErrInvalidArgumentType
	}

	return v, nil
}

// MustFloat will attempt to do Float() and will panic if not possible.
func (a *Argument) MustFloat() float64 {
	value, err := a.Float()
	if err != nil {
		panic(fmt.Sprintf(`argument: Float(): %v`, err.Error()))
	}

	return value
}

func argumentTypeIs(argumentType ArgumentType, argumentTypes ...ArgumentType) bool {
	for _, aType := range argumentTypes {
		if argumentType == aType {
			return true
		}
	}

	return false
}
