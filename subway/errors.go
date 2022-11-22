package internal

import "errors"

var (
	ErrSubwayAlreadyExists = errors.New("subway already created")

	ErrInvalidRequestSignature = errors.New("invalid request signature")
	ErrInvalidPublicKey        = errors.New("invalid public key. this must be a 64 characters long and hex encoded")

	ErrCommandNotFound = errors.New("command with this name was not found")

	ErrReadConfigurationFailure = errors.New("failed to read configuration")
	ErrLoadConfigurationFailure = errors.New("failed to load configuration")
)
