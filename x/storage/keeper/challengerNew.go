package keeper

// import (
// 	"context"
// 	"fmt"

// 	"atlas/x/storage/types"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// // BeginChallengeRoundEfficient starts a new challenge round with statistical sampling
// func (k Keeper) BeginChallengeRoundEfficient(ctx context.Context) error {
// 	sdkCtx := sdk.UnwrapSDKContext(ctx)
// 	currentHeight := uint64(sdkCtx.BlockHeight())

// 	// 1. Check if it's time for a challenge round
// 	params, err := k.Params.Get(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// lastRound, err := k.GetLastChallengeHeight(ctx)
// 	// if err != nil && !k.IsNotFound(err) {
// 	// 	return err
// 	// }

// 	// if currentHeight-lastRound < params.ChallengeInterval {
// 	// 	return nil // Not time yet
// 	// }

// 	// 2. Get verifiable randomness from previous block
// 	randomSeed, err := types.GetVerifiableRandomSeed(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to get randomness: %w", err)
// 	}

// 	// 3. Select files using statistical sampling
// 	selectedFiles, err := k.SelectFilesStatistical(
// 		ctx,
// 		randomSeed,
// 		params,
// 		currentHeight,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to select files: %w", err)
// 	}

// 	if len(selectedFiles) == 0 {
// 		// No eligible files, just update round
// 		k.SetLastChallengeHeight(ctx, currentHeight)
// 		return nil
// 	}

// 	// 4. Generate challenges for selected files
// 	allChallenges, err := k.GenerateChallengesForFiles(
// 		ctx,
// 		selectedFiles,
// 		randomSeed,
// 		currentHeight,
// 		params.ProofTimeout,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to generate challenges: %w", err)
// 	}

// 	// 5. Store challenges efficiently
// 	createdCount := k.StoreChallengesBatch(ctx, allChallenges)

// 	// 6. Update file timestamps (batch if possible)
// 	k.UpdateFileTimestampsBatch(ctx, selectedFiles, currentHeight)

// 	// 7. Update round state
// 	k.SetLastChallengeHeight(ctx, currentHeight)
// 	k.IncrementChallengeRound(ctx)

// 	// 8. Store random seed for verification
// 	k.SetChallengeRoundSeed(ctx, currentHeight, randomSeed)

// 	// 9. Emit event
// 	_ = sdkCtx.EventManager().EmitTypedEvent(&types.EventChallengeRound{
// 		RoundNumber:       k.GetCurrentRound(ctx),
// 		BlockHeight:       currentHeight,
// 		FilesSelected:     uint64(len(selectedFiles)),
// 		ChallengesCreated: uint64(createdCount),
// 		RandomSeed:        randomSeed[:16], // First 16 bytes for event
// 		SamplingMethod:    "statistical",
// 	})

// 	sdkCtx.Logger().Info("challenge round completed",
// 		"height", currentHeight,
// 		"files", len(selectedFiles),
// 		"challenges", createdCount,
// 		"targetFiles", params.TargetFilesPerRound,
// 	)

// 	return nil
// }
