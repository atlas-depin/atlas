package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	params, _ := k.Params.Get(ctx)
	height := ctx.BlockHeight()

	if ctx.BlockHeight()%params.ProofWindowBlocks == 0 {
		k.StartChallengeWindow(ctx)
	}

	if (ctx.BlockHeight()+1)%params.ProofWindowBlocks == 0 {
		// [TODO]: cleanup challenges & apply penalties
	}

	if ctx.BlockHeight()%params.ProofRoundBlocks == 0 {
		ctx.Logger().Info("Processing provider rewards", "height", ctx.BlockHeight())

		roundIdx := ((height % params.ProofWindowBlocks) / params.ProofRoundBlocks) + 1
		k.StartChallengeRound(ctx, uint64(roundIdx))
	}

	return nil
}
