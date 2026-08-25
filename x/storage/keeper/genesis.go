package keeper

import (
	"context"

	"atlas/x/storage/types"

	"cosmossdk.io/math"
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

	// Seed the $ATL price; the first MsgUpdatePrice overwrites it.
	if err := k.Price.Set(ctx, types.DefaultPrice); err != nil {
		return err
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params: params,
	}, nil
}
