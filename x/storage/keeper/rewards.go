package keeper

import (
	"fmt"
	"nebulix/x/storage/types"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) RewardProviders(ctx sdk.Context) error {
	t_0, err := k.LastRewardsTime.Get(ctx)
	if err != nil {
		return err
	}
	t_1 := ctx.BlockTime().Truncate(time.Second)
	d_t := math.NewInt(int64(t_1.Sub(t_0).Seconds()))

	/// --- get current storage tokens and storage credits kept by the module
	storageTokensInPool := k.bankKeeper.GetBalance(ctx, types.ModuleAddress, "uatl")
	storageCreditsInPool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return err
	}
	if storageCreditsInPool.LTE(math.ZeroInt()) {
		return types.ErrRewardPoolEmpty
	}

	// --- iterate through all providers, distributing their rewards and resetting their credit delta
	var totalCreditsDeducted math.Int
	k.IterateAllProviders(ctx, func(key string, provider types.Provider) (stop bool) {
		credits := d_t.Mul(math.NewInt(provider.SpaceUsed)).Add(math.NewInt(provider.CreditDelta))
		if credits.IsNegative() { // forward remaining negative credits to next proof round delta
			provider.CreditDelta = credits.Int64()
			return false
		}

		totalCreditsDeducted = totalCreditsDeducted.Add(credits)
		tokens := sdk.NewCoin(
			storageTokensInPool.Denom,
			storageTokensInPool.Amount.Mul(credits).Quo(storageCreditsInPool),
		)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			sdk.AccAddress(provider.Address),
			sdk.NewCoins(tokens),
		)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Failed to distribute rewards to %s.", provider.Address))
		}

		provider.CreditDelta = 0
		if err := k.ProviderStore.Set(ctx, provider.Address, provider); err != nil {
			ctx.Logger().Error(fmt.Sprintf("Failed to reset credit delta for %s.", provider.Address))
		}

		return false
	})

	// --- burn rewarded credits from the module's pool
	newCreditsInPool := storageCreditsInPool.Sub(totalCreditsDeducted)
	err = k.StorageCreditPool.Set(ctx, newCreditsInPool)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to burn %d storage credits", newCreditsInPool.Int64()))
		return nil
	}
	return k.LastRewardsTime.Set(ctx, t_1)
}

func (k Keeper) RewardStrays(ctx sdk.Context) error {
	t_r, err := k.LastRewardsTime.Get(ctx)
	if err != nil {
		return err
	}
	t_0 := t_r.Unix()
	t_1 := ctx.BlockTime().Unix()

	var d_c math.Int
	iter, _ := k.FileStore.Indexes.ByStatus.MatchExact(ctx, "STRAY")
	for ; iter.Valid(); iter.Next() {
		pk, _ := iter.PrimaryKey()
		file, err := k.GetFile(ctx, pk.K3())
		if err != nil {
			continue
		}

		if t_0 < file.Start {
			d_c = math.NewInt((t_1 - file.Start) * file.FileSize)
		} else {
			d_c = math.NewInt((t_1 - t_0) * file.FileSize)
		}

		err = k.DistributeCreditReward(ctx, d_c, types.ProtocolOwnedLiquidityAccountName)
		if err != nil {
			ctx.Logger().Error("Failed to distribute reward to POL for stray %s", file.Fid)
		}
	}
	return nil
}
