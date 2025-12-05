package keeper

import (
	"nebulix/x/storage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// [ENABLE] probabilistic auditing
// const ChallengeRate = 10000

func (k Keeper) SetChallenge(ctx sdk.Context, challenge types.Challenge) error {
	// Save challenge
	if err := k.ChallengeStore.Set(ctx, challenge.ChallengeSeed, challenge); err != nil {
		return err
	}
	return nil
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
