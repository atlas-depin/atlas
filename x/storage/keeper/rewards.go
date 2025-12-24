package keeper

import (
	"nebulix/x/storage/types"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// [TODO]: zero checks, especially for div-zero
func (k Keeper) RewardProviders(ctx sdk.Context) error {
	/// --- compute reward period
	t0, err := k.LastRewardsTime.Get(ctx)
	if err != nil {
		return err
	}
	t1 := ctx.BlockTime().Truncate(time.Second)
	t_delta := math.NewInt(int64(t1.Sub(t0).Seconds()))

	/// --- get current storage tokens and storage credits kept by the module
	storageTokensInPool := k.bankKeeper.GetBalance(ctx, types.ModuleAddress, "uatl")
	storageCreditsInPool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return err
	}

	/// --- setup update collector (can't safely update while iterating)
	type providerUpdate struct {
		address string
		credits math.Int
	}
	var updates []providerUpdate

	// --- first iteration: compute credits to distribute
	iter, err := k.ProviderStore.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		provider, err := iter.Value()
		if err != nil {
			return err
		}

		credits := t_delta.Mul(math.NewInt(provider.SpaceUsed)).Add(math.NewInt(provider.CreditDelta))

		updates = append(updates, providerUpdate{
			address: provider.Address,
			credits: credits,
		})
	}

	// second iteration: distribute rewards and update credit deltas
	var totalCreditsDeducted math.Int
	for _, update := range updates {
		totalCreditsDeducted = totalCreditsDeducted.Add(update.credits)

		tokens := sdk.NewCoin(
			storageTokensInPool.Denom,
			storageTokensInPool.Amount.Mul(update.credits).Quo(storageCreditsInPool),
		)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			sdk.AccAddress(update.address),
			sdk.NewCoins(tokens),
		)
		if err != nil {
			return err
		}

		provider, err := k.ProviderStore.Get(ctx, update.address)
		if err != nil {
			return err
		}
		provider.CreditDelta = 0
		if err := k.ProviderStore.Set(ctx, update.address, provider); err != nil {
			return err
		}
	}

	// --- burn rewarded credits from the module's pool
	newCreditsInPool := storageCreditsInPool.Sub(totalCreditsDeducted)
	err = k.StorageCreditPool.Set(ctx, newCreditsInPool)
	if err != nil {
		return err
	}
	return k.LastRewardsTime.Set(ctx, t1)
}
