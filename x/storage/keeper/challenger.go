package keeper

import (
	"context"
	"fmt"
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// [ENABLE] probabilistic auditing
// const ChallengeRate = 10000

func (k Keeper) SetChallenge(ctx sdk.Context, challenge types.StorageChallenge) error {
	// Save challenge
	key := collections.Join(challenge.Provider, challenge.ChallengeId)
	if err := k.ChallengeStore.Set(ctx, key, challenge); err != nil {
		return err
	}
	return nil
}

func (k Keeper) ChallengeFiles(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	height := uint64(sdkCtx.BlockHeight())
	baseSeed, err := types.GetVerifiableRandomSeed(ctx)
	if err != nil {
		return fmt.Errorf("failed to get randomness: %w", err)
	}

	iter, err := k.FileStore.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	var i uint64 = 0
	for ; iter.Valid(); i++ {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		k.GenerateFileChallenges(ctx, kv.Value, baseSeed, i, height)

		iter.Next()
	}

	return nil
}

// //
// GenerateFileChallenges creates challenges for ALL providers of a file
// //
func (k Keeper) GenerateFileChallenges(
	ctx context.Context,
	file types.File,
	seed []byte,
	seedOffset uint64,
	currentHeight uint64,
) ([]types.StorageChallenge, error) {

	// Validate file has providers
	if len(file.Providers) == 0 {
		return nil, fmt.Errorf("file has no providers: %s", file.FileId)
	}

	// Validate file size
	if file.FileSize == 0 {
		return nil, fmt.Errorf("file has zero size: %s", file.FileId)
	}

	// Compute total chunks
	chunkCount := types.ComputeChunkCount(file.FileSize)
	if chunkCount == 0 {
		return nil, fmt.Errorf("invalid chunk count for file size %d", file.FileSize)
	}

	// Generate challenges for each provider
	challenges := make([]types.StorageChallenge, 0, len(file.Providers))

	for i, providerAddr := range file.Providers {
		// Validate provider address
		// [TBD]: do I really need this? or is it just a waste of CPU time
		if _, err := sdk.AccAddressFromBech32(providerAddr); err != nil {
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			sdkCtx.Logger().Error("invalid provider address",
				"file", file.FileId,
				"provider", providerAddr,
				"error", err,
			)
			continue
		}

		// Generate provider-specific random seed
		providerSeed := types.CreateProviderChallengeSeed(
			seed,
			seedOffset,
			providerAddr,
			uint64(i), // Provider position for seed diversity
		)

		// Generate random chunk index for THIS provider
		chunkIndex, err := types.GenerateRandomChallengeChunkIndex(
			providerSeed,
			chunkCount,
			file.FileId,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate chunk index for provider %s: %w",
				providerAddr, err)
		}

		// Create challenge
		challenge, err := CreateChallenge(
			file,
			providerAddr,
			chunkIndex,
			providerSeed,
			currentHeight,
			chunkCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create challenge for provider %s: %w",
				providerAddr, err)
		}

		challenges = append(challenges, challenge)
	}

	return challenges, nil
}

func CreateChallenge(
	file types.File,
	providerAddr string,
	chunkIndex uint64,
	seed []byte,
	currentHeight uint64,
	chunkCount uint64,
) (types.StorageChallenge, error) {
	challengeID := types.GenerateChallengeID(currentHeight, file.FileId, providerAddr)

	return types.StorageChallenge{
		ChallengeId:   challengeID,
		FileId:        file.FileId,
		FileMerkle:    file.Merkle,
		Provider:      providerAddr,
		ChunkIndex:    chunkIndex,
		CreatedHeight: currentHeight,
		Seed:          seed,
		Completed:     false,
	}, nil
}

// // GenerateChallenges iterates over all files and probabilistically generates a challenge.
// func (k Keeper) GenerateChallenges(ctx sdk.Context) []types.Challenge {

// 	// [ENABLE] probabilistic auditing
// 	// blockHash := ctx.BlockHeader().LastBlockId.Hash
// 	var challenges []types.Challenge

// 	k.IterateAllFiles(ctx, func(file types.File) (stop bool) {

// 		// [ENABLE] probabilistic auditing
// 		// [NOTE]: disabled for now... until everything else works. I don't even know if what I'm doing here works lol!
// 		// 1. Probabilistic Auditing
// 		// -----------------------------------------------------
// 		// seedData := append([]byte(file.Id), blockHash.Bytes()...)
// 		// hash := address.Hash(seedData)
// 		// hashInt := new(big.Int).SetBytes(hash)

// 		// // Check if hashInt % ChallengeRate == 0
// 		// zero := big.NewInt(0)
// 		// rate := big.NewInt(ChallengeRate)
// 		// if hashInt.Mod(hashInt, rate).Cmp(zero) != 0 {
// 		// 	return false // Skip this file
// 		// }

// 		// 2. File Selected: Generate a challenge for each of the 3 responsible providers.
// 		// -----------------------------------------------------
// 		binRoot := k.GetBinRoot(ctx, file.BinId)

// 		challengeSeed := make([]byte, 32)
// 		_, err := rand.Read(challengeSeed)
// 		if err != nil {
// 			ctx.Logger().Error("Failed to generate challenge seed", "error", err)
// 			return false
// 		}

// 		for _, providerAddr := range file.Providers {
// 			challenge := types.Challenge{
// 				FileId:        file.Merkle,
// 				Provider:      providerAddr,
// 				ChallengeSeed: string(challengeSeed),
// 				BinRoot:       binRoot, // CRUCIAL: The Bin Root the SP must prove against
// 			}

// 			challenges = append(challenges, challenge)
// 			k.SetChallenge(ctx, challenge)
// 		}

// 		// [TODO]: file challenge traceability
// 		// file.LatestChallengeBlock = uint64(ctx.BlockHeight())
// 		k.SetFile(ctx, file)

// 		return false
// 	})

// 	ctx.Logger().Info("Generated challenges", "count", len(challenges))
// 	return challenges
// }
