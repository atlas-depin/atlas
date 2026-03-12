package keeper

import (
	"context"
	"nebulix/x/storage/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) DistributeCreditReward(ctx context.Context, credits math.Int, receiver string) error {
	tpool := k.bankKeeper.GetBalance(ctx, types.ModuleAddress, "uatl")
	cpool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return err
	}
	tokens := sdk.NewCoin(
		tpool.Denom,
		tpool.Amount.Mul(credits).Quo(cpool),
	)

	cpool = cpool.Sub(credits)
	if err = k.StorageCreditPool.Set(ctx, cpool); err != nil {
		// ctx.Logger().Error(fmt.Sprintf("Failed to burn %d storage credits", cpool.Int64()))
		return err
	}

	if receiver == types.ProtocolOwnedLiquidityAccountName {
		if err = k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			types.ProtocolOwnedLiquidityAccountName,
			sdk.NewCoins(tokens),
		); err != nil {
			// ctx.Logger().Error(fmt.Sprintf("Failed to distribute rewards to POL"))
			return err
		}
	} else {
		if err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			sdk.AccAddress(receiver),
			sdk.NewCoins(tokens),
		); err != nil {
			// ctx.Logger().Error(fmt.Sprintf("Failed to distribute rewards to %s.", receiver))
			return err
		}
	}

	return nil
}
