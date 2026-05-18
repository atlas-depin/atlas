package keeper

import (
	"atlas/x/storage/types"
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/// ---- CRUD Methods ----

/*
Set a new provider record keyed by the provider address.
This method ensures no provider exists with the same address.
*/
func (k Keeper) SetNewProvider(ctx context.Context, provider types.Provider) error {
	exists, err := k.ProviderStore.Has(ctx, provider.Address)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrProviderAlreadyExists
	}

	return k.SetProvider(ctx, provider)
}

/*
Set a provider record keyed by the provider address.
*/
func (k Keeper) SetProvider(ctx context.Context, provider types.Provider) error {
	return k.ProviderStore.Set(ctx, provider.Address, provider)
}

/*
Get a provider record for a given provider address.
*/
func (k Keeper) GetProvider(ctx context.Context, address string) (provider types.Provider, err error) {
	provider, err = k.ProviderStore.Get(ctx, address)
	if err == collections.ErrNotFound {
		err = types.ErrProviderNotFound
	}
	return
}

/*
Get all provider records.
*/
func (k Keeper) GetAllProviders(ctx context.Context) ([]*types.Provider, error) {
	values := []*types.Provider{}

	iter, err := k.ProviderStore.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		provider, err := iter.Value()
		if err != nil {
			return values, err
		}
		values = append(values, &provider)
	}
	return values, nil
}

/*
Iterate over all providers and run the provided function argument for each record.
*/
func (k Keeper) IterateAllProviders(ctx context.Context, fn func(key string, provider types.Provider) (stop bool)) error {
	iter, err := k.ProviderStore.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		if fn(kv.Key, kv.Value) {
			break
		}
	}

	return nil
}

/// ---- Utility Methods ----

/*
Increment the space used of a given provider.
*/
func (k Keeper) IncrementProviderSpaceUsed(ctx context.Context, address string, bytes int64, d_r int64, d_s int64) error {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return err
	}

	provider.SpaceUsed += bytes
	if provider.SpaceUsed > provider.SpaceAvailable {
		return types.ErrProviderFull
	}

	provider.CreditDelta += bytes * -d_r
	if err = k.DistributeCreditReward(ctx, math.NewInt(bytes*d_s), types.ProtocolOwnedLiquidityAccountName); err != nil {
		return err
	}

	return k.SetProvider(ctx, provider)
}

/*
Decrement the space used of a given provider.
*/
func (k Keeper) DecrementProviderSpaceUsed(ctx context.Context, address string, bytes int64, timeDelta int64) error {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return err
	}

	provider.SpaceUsed -= bytes
	if provider.SpaceUsed < 0 {
		bytes = bytes + provider.SpaceUsed
		provider.SpaceUsed = 0
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Warn(fmt.Sprintf("Something is terribly wrong... Tried to decrement SpaceUsed below 0 for provider %s", provider.Address))
	}

	provider.CreditDelta += bytes * timeDelta

	return k.SetProvider(ctx, provider)
}
