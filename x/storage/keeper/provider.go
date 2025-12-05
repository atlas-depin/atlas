package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"nebulix/x/storage/types"
)

// \\ RegisterProvider registers a new storage provider
func (k Keeper) RegisterProvider(ctx context.Context, providerAddr string, deposit sdk.Coin) error {
	// Check if provider already exists
	exists, err := k.ProviderStore.Has(ctx, providerAddr)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrProviderAlreadyExists
	}

	// Verify minimum deposit
	// params, err := k.Params.Get(ctx)
	// if err != nil {
	// 	return err
	// }
	// if deposit.IsLT(params.MinProviderDeposit) {
	// 	return types.ErrInsufficientDeposit
	// }

	// Transfer deposit to module
	// providerAcc, err := sdk.AccAddressFromBech32(providerAddr)
	// if err != nil {
	// 	return err
	// }

	// if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, providerAcc, types.ModuleName, sdk.NewCoins(deposit)); err != nil {
	// 	return err
	// }

	// Create provider
	provider := types.Provider{
		Address: providerAddr,

		CreatedAt: sdk.UnwrapSDKContext(ctx).BlockHeight(),
	}

	return k.ProviderStore.Set(ctx, providerAddr, provider)
}

// GetProviderFiles returns all files for a provider
// func (k Keeper) GetProviderFiles(ctx context.Context, providerAddr string) ([]string, error) {
// 	var fileIds []string

// 	err := k.FilesByProvider.Walk(ctx, collections.NewPrefixedPairRange[string, string](providerAddr), func(key collections.Pair[string, string]) (stop bool, err error) {
// 		fileIds = append(fileIds, key.K2())
// 		return false, nil
// 	})

// 	return fileIds, err
// }
