package types

import (
	"cosmossdk.io/collections"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "storage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"

	ProviderDepositAccountName        = "provider_deposit"
	ProtocolOwnedLiquidityAccountName = "pol"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_storage")

// FeedsKey is the prefix to retrieve all price feeds
var FeedsKey = collections.NewPrefix("f_storage")

var ModuleAddress = authtypes.NewModuleAddress(ModuleName)
var ProtocolOwnedLiquidityAddress = authtypes.NewModuleAddress(ProtocolOwnedLiquidityAccountName)
var ProviderDepositAddress = authtypes.NewModuleAddress(ProviderDepositAccountName)
