package internal

type ArgumentType uint16

const (
	ArgumentTypeSnowflake ArgumentType = iota + 1
	ArgumentTypeMember
	ArgumentTypeUser
	_
	_
	ArgumentTypeTextChannel
	ArgumentTypeGuild
	ArgumentTypeRole
	ArgumentTypeColour
	ArgumentTypeVoiceChannel
	ArgumentTypeStageChannel
	ArgumentTypeEmoji
	ArgumentTypePartialEmoji
	ArgumentTypeCategoryChannel
	ArgumentTypeStoreChannel
	ArgumentTypeThread
	ArgumentTypeGuildChannel
	_
	ArgumentTypeString
	ArgumentTypeBool
	ArgumentTypeInt
	ArgumentTypeFloat
)
