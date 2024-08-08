package internal

import (
	"errors"
	"fmt"
)

var (
	ErrSubwayAlreadyExists = errors.New("subway already created")

	ErrInvalidRequestSignature = errors.New("invalid request signature")
	ErrInvalidPublicKey        = errors.New("invalid public key. this must be a 64 characters long and hex encoded")

	ErrReadConfigurationFailure = errors.New("failed to read configuration")
	ErrLoadConfigurationFailure = errors.New("failed to load configuration")

	ErrFetchMissingGuild     = errors.New("object requires guild ID to fetch")
	ErrFetchMissingSnowflake = errors.New("object requires snowflake to fetch")

	ErrCogAlreadyRegistered     = errors.New("cog with this name already exists")
	ErrCommandAlreadyRegistered = errors.New("command with this name already exists")
	ErrInvalidArgumentType      = errors.New("argument value is not correct type for converter used")
	ErrConversionError          = errors.New("failed to convert argument to desired type")

	ErrCommandNotFound             = errors.New("command with this name was not found")
	ErrCommandAutoCompleteNotFound = errors.New("autocomplete for command with this name was not found")
	ErrComponentListenerNotFound   = errors.New("component listener with this name was not found or has expired")

	ErrCheckFailure            = errors.New("command failed built-in checks")
	ErrMissingRequiredArgument = errors.New("command missing required arguments")
	ErrArgumentNotFound        = errors.New("command argument was not found")
	ErrConverterNotFound       = errors.New("command converter is not setup")

	// Converter errors.

	ErrSnowflakeNotFound = errors.New("id does not follow a valid id or mention format")
	ErrMemberNotFound    = errors.New("member provided was not found")
	ErrUserNotFound      = errors.New("user provided was not found")
	ErrChannelNotFound   = errors.New("channel provided was not found")
	ErrGuildNotFound     = errors.New("guild provided was not found")
	ErrRoleNotFound      = errors.New("role provided was not found")
	ErrEmojiNotFound     = errors.New("emoji provided was not found")

	ErrBadInviteArgument  = errors.New("invite provided was invalid or expired")
	ErrBadColourArgument  = errors.New("colour provided was not in valid format")
	ErrBadBoolArgument    = errors.New("bool provided was not in valid format")
	ErrBadIntArgument     = errors.New("int provided was not in valid format")
	ErrBadFloatArgument   = errors.New("float provided was not in valid format")
	ErrBadWebhookArgument = errors.New("webhook url provided was not in valid format")
)

type PanicError struct {
	Recover interface{}
}

func (cp PanicError) Error() string {
	return fmt.Sprintf("command panicked with error: %v", cp.Recover)
}
