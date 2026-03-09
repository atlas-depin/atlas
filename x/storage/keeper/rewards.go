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
	t_delta := math.NewInt(int64(t_1.Sub(t_0).Seconds()))

	/// --- get current storage tokens and storage credits kept by the module
	storageTokensInPool := k.bankKeeper.GetBalance(ctx, types.ModuleAddress, "uatl")
	storageCreditsInPool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return err
	}

	if storageCreditsInPool.LTE(math.ZeroInt()) {
		return nil
	}

	// --- iterate through all providers, distributing their rewards and resetting their credit delta
	var totalCreditsDeducted math.Int
	k.IterateAllProviders(ctx, func(key string, provider types.Provider) (stop bool) {
		credits := t_delta.Mul(math.NewInt(provider.SpaceUsed)).Add(math.NewInt(provider.CreditDelta))
		if credits.IsNegative() { // forward negative credits to next proof round delta
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
