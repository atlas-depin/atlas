package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// Check for expired subscriptions — uses the expiry index so
	// this is bounded to actually-expiring subs, not all subscriptions.
	if err := k.ExpireSubscriptions(ctx); err != nil {
		ctx.Logger().Error("Failed to process expired subscriptions", "err", err)
	}

	params, _ := k.Params.Get(ctx)
	height := uint64(ctx.BlockHeight())

	if height%params.ProofWindowBlocks == 0 {
		k.StartChallengeWindow(ctx)
	}

	if height%params.ProofRoundBlocks == 0 {
		ctx.Logger().Info("Processing provider rewards", "height", height)
		k.HandleMissedChallenges(ctx)
		err := k.RewardProviders(ctx)
		if err != nil {
			ctx.Logger().Error("Failed to distribute rewards...")
		}
		k.RewardStrays(ctx)

		roundIdx := ((height % params.ProofWindowBlocks) / params.ProofRoundBlocks)
		k.StartChallengeRound(ctx, uint64(roundIdx))
	}

	return nil
}
