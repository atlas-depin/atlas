package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/storage module sentinel errors
var (
	ErrInvalidSigner         = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrProviderNotFound      = errors.Register(ModuleName, 2100, "provider not found")
	ErrProviderNotActive     = errors.Register(ModuleName, 2101, "provider not active")
	ErrProviderAlreadyExists = errors.Register(ModuleName, 2102, "provider already exists")
	ErrInsufficientDeposit   = errors.Register(ModuleName, 2103, "insufficient provider deposit")

	ErrFileAlreadyExists = errors.Register(ModuleName, 2301, "file already exists")
	ErrFileNotFound      = errors.Register(ModuleName, 2304, "file not found")

	ErrSubscriptionExpired            = errors.Register(ModuleName, 2301, "subscription is expired")
	ErrSubscriptionFull               = errors.Register(ModuleName, 2302, "subscription has no remaining storage")
	ErrSubscriptionInsufficientCredit = errors.Register(ModuleName, 2303, "subscription has insufficient credit")
	ErrSubscriptionNotFound           = errors.Register(ModuleName, 2304, "subscription not found")
)
