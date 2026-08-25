package keeper

import (
	"context"

	"atlas/x/storage/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	// Write it to the store
	if err := k.StorageCreditPool.Set(ctx, math.ZeroInt()); err != nil {
		panic(err)
	}

	if err := k.FileCount.Set(ctx, 0); err != nil {
		panic(err)
	}

	if err := k.LastRewardsTime.Set(ctx, 0); err != nil {
		panic(err)
	}

	if err := k.Params.Set(ctx, genState.Params); err != nil {
		panic(err)
	}

	for _, feed := range genState.Feeds {
		if err := k.SetFeed(ctx, feed); err != nil {
			return err
		}
	}

	// The built-in $ATL feed always exists. If genesis did not define it,
	// seed it with the module authority (x/gov) as owner so its price can be
	// updated via governance.
	if _, err := k.GetFeed(ctx, types.BuiltinFeedName); err != nil {
		owner, err := k.addressCodec.BytesToString(k.authority)
		if err != nil {
			return err
		}
		builtin := types.Feed{
			Name:       types.BuiltinFeedName,
			Owner:      owner,
			Data:       types.BuiltinFeedInitialData,
			LastUpdate: sdk.UnwrapSDKContext(ctx).BlockTime(),
		}
		if err := k.SetFeed(ctx, builtin); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	feeds, err := k.GetAllFeeds(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params: params,
		Feeds:  feeds,
	}, nil
}
