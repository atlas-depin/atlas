package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	params, _ := k.Params.Get(ctx)
	height := uint64(ctx.BlockHeight())

	if height%params.ProofWindowBlocks == 0 {
		k.StartChallengeWindow(ctx)
	}

	if (height+1)%params.ProofRoundBlocks == 0 {
		// [TODO]: apply penalties
		k.HandleMissedChallenges(ctx)
	}

	if height%params.ProofRoundBlocks == 0 {
		ctx.Logger().Info("Processing provider rewards", "height", height)

		// TODO: reward processing
		err := k.RewardProviders(ctx)
		if err != nil {
			ctx.Logger().Warn(fmt.Sprintf("Failed to distribute rewards..."))
		}

		roundIdx := ((height % params.ProofWindowBlocks) / params.ProofRoundBlocks)
		k.StartChallengeRound(ctx, uint64(roundIdx))
	}

	return nil
}

// if (ctx.BlockHeight()+1)%params.ProofWindowBlocks == 0 {
// 	// [TODO]: cleanup challenges & apply penalties
// }
