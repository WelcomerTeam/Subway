package internal

import "errors"

var (
	ErrSubwayAlreadyExists = errors.New("subway already created")

	ErrInvalidRequestSignature = errors.New("invalid request signature")

	ErrReadConfigurationFailure = errors.New("failed to read configuration")
	ErrLoadConfigurationFailure = errors.New("failed to load configuration")
)
