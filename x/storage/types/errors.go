package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/storage module sentinel errors
var (
	ErrInvalidSigner               = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrProviderNotActive           = errors.Register(ModuleName, 2100, "provider not active")
	ErrProviderAlreadyExists       = errors.Register(ModuleName, 2101, "provider already exists")
	ErrProviderDepositInsufficient = errors.Register(ModuleName, 2102, "insufficient provider deposit")
	ErrProviderNotFound            = errors.Register(ModuleName, 2104, "provider not found")
	ErrProviderFull                = errors.Register(ModuleName, 2105, "insufficient provider storage. cannot exceed space available.")

	ErrFileAlreadyExists = errors.Register(ModuleName, 2201, "file already exists")
	ErrFileNotFound      = errors.Register(ModuleName, 2204, "file not found")

	ErrSubscriptionExpired            = errors.Register(ModuleName, 2301, "subscription is expired")
	ErrSubscriptionFull               = errors.Register(ModuleName, 2302, "subscription has no remaining storage")
	ErrSubscriptionInsufficientCredit = errors.Register(ModuleName, 2303, "subscription has insufficient credit")
	ErrSubscriptionNotFound           = errors.Register(ModuleName, 2304, "subscription not found")

	ErrProofInvalid         = errors.Register(ModuleName, 2401, "invalid file proof")
	ErrProofFileBadProvider = errors.Register(ModuleName, 2402, "proof rejected because this file does not belong to you")

	ErrRewardPoolEmpty = errors.Register(ModuleName, 2501, "storage credit reward pool is empty (this should not happen)")
)
