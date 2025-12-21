package keeper

import (
	"fmt"
	"nebulix/x/storage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// Execute every 300 blocks

	if ctx.BlockHeight() > 0 && ctx.BlockHeight()%300 == 0 {
		ctx.Logger().Info("Processing provider rewards", "height", ctx.BlockHeight())

		// if err := k.RewardProviders(ctx); err != nil {
		// 	// Log error but don't panic
		// 	ctx.Logger().Error("Failed to reward providers", "error", err)

		// 	// Emit event for monitoring
		// 	ctx.EventManager().EmitEvent(
		// 		sdk.NewEvent("module_rewards_failed",
		// 			sdk.NewAttribute("module", types.ModuleName),
		// 			sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
		// 			sdk.NewAttribute("error", err.Error()),
		// 		),
		// 	)
		// 	return nil // Don't return error to avoid chain halt
		// }

		// Success event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("module_rewards_distributed",
				sdk.NewAttribute("module", types.ModuleName),
				sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
			),
		)
	}

	if ctx.BlockHeight() > 0 && (ctx.BlockHeight()+1)%2400 == 0 {
		// [TODO]: cleanup challenges & apply penalties
	}

	if ctx.BlockHeight() > 0 && ctx.BlockHeight()%2400 == 0 {
		k.ChallengeFiles(ctx)
	}

	return nil
}
