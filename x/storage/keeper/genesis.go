package keeper

import (
	"context"
	"time"

	"nebulix/x/storage/types"

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

	var t time.Time
	if err := k.LastRewardsTime.Set(ctx, t); err != nil {
		panic(err)
	}

	return k.Params.Set(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
