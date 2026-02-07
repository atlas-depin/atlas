package keeper

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	merkletree "github.com/ahm23/go-merkletree-xxh"
	"github.com/cespare/xxhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"
	"github.com/zeebo/blake3"
)

var json = jsoniter.ConfigFastest

func (k Keeper) SetChallenge(ctx context.Context, challenge types.StorageChallenge) error {
	// Save challenge
	key := collections.Join(challenge.Provider, challenge.ChallengeId)
	if err := k.ChallengeStore.Set(ctx, key, challenge); err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetChallenge(ctx context.Context, provider string, challengeId string) (types.StorageChallenge, error) {
	// Save challenge
	key := collections.Join(provider, challengeId)
	return k.ChallengeStore.Get(ctx, key)
}

func (k Keeper) RemoveChallenge(ctx context.Context, provider string, challengeId string) error {
	// Save challenge
	key := collections.Join(provider, challengeId)
	return k.ChallengeStore.Remove(ctx, key)
}

func (k Keeper) StartChallengeRound(ctx sdk.Context, roundIdx uint64) {
	// You already computed this (even distribution)
	// numThisSubwindow := k.calculateFilesPerSubwindow(ctx, roundIdx) // your existing function
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	numThisSubwindow := uint64(300)

	indices, err := k.SelectRandomChallengeIndices(ctx, numThisSubwindow, roundIdx)
	if err != nil || len(indices) == 0 {
		// log and return or fallback
		return
	}

	for _, idx := range indices {
		fileId, err := k.ChallengeSnapshot.Get(ctx, idx)
		if err != nil {
			continue // shouldn't happen
		}

		// Check if file still exists (handles deletes)
		file, err := k.GetFile(ctx, fileId) // your helper
		if err != nil {
			// file was deleted mid-window → skip
			continue
		}

		// Issue the actual challenge
		k.IssueChallengeForFile(ctx, file, roundIdx)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewChallengeRound,
			sdk.NewAttribute("round", fmt.Sprintf("%d", roundIdx)),
		),
	)
}

func (k Keeper) IssueChallengeForFile(
	ctx context.Context,
	file types.File,
	round uint64,
) ([]types.StorageChallenge, error) {
	// Validate file has providers
	if len(file.Providers) == 0 {
		return nil, fmt.Errorf("file has no providers: %s", file.Fid)
	}
	// Validate file size
	if file.FileSize == 0 {
		return nil, fmt.Errorf("file has zero size: %s", file.Fid)
	}
	// Compute total chunks
	chunkCount := types.ComputeChunkCount(file.FileSize)
	if chunkCount == 0 {
		return nil, fmt.Errorf("invalid chunk count for file size %d", file.FileSize)
	}

	// Get deterministic seed source: previous block hash
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	prevHash := sdkCtx.BlockHeader().LastBlockId.Hash
	currentHeight := sdkCtx.BlockHeight()

	// Generate unique chunk indices for each provider
	chunkIndices := make(map[string]uint64) // provider -> chunkIndex

	// Global seed base (deterministic per file + round)
	globalSeed := append([]byte{}, prevHash...) // copy
	globalSeed = append(globalSeed, []byte(file.Fid)...)
	globalSeed = binary.BigEndian.AppendUint64(globalSeed, round)

	for i, providerAddr := range file.Providers {
		// Provider-specific seed: globalSeed + position + nonce
		seed := append(globalSeed, byte(i)) // position as byte (assuming <256 providers)
		// XXHash for randomness
		hashVal := xxhash.Sum64(seed)
		// Modulo to get chunk index
		chunkIndices[providerAddr] = hashVal % chunkCount
	}

	// Now create challenges
	for _, providerAddr := range file.Providers {
		chunkIndex := chunkIndices[providerAddr]

		// Create challenge
		challenge := CreateChallenge(
			file,
			providerAddr,
			chunkIndex,
			currentHeight,
		)

		if err := k.SetChallenge(ctx, challenge); err != nil {
			sdkCtx.Logger().Error("failed to store challenge", "err", err, "fid", file.Fid, "provider", providerAddr)
			continue
		}
	}

	return nil, nil
}

func CreateChallenge(
	file types.File,
	providerAddr string,
	chunkIndex uint64,
	currentHeight int64,
) types.StorageChallenge {
	challengeID := types.GenerateChallengeID(currentHeight, file.Fid, providerAddr)

	return types.StorageChallenge{
		ChallengeId:   challengeID,
		FileId:        file.Fid,
		FileMerkle:    file.Merkle,
		Provider:      providerAddr,
		ChunkIndex:    chunkIndex,
		CreatedHeight: currentHeight,
		Completed:     false,
	}
}

func (k Keeper) AttemptChallenge(ctx context.Context, provider string, proofInput types.Proof) error {
	challenge, err := k.GetChallenge(ctx, provider, proofInput.ChallengeId)
	// [TODO]: allow no challenge
	if err != nil {
		return nil
	}

	file, err := k.GetFile(ctx, challenge.FileId)
	if err != nil {
		// [TODO]: remove challenge
		return nil
	}

	if containsProver(file, provider) {
		root, _ := hex.DecodeString(file.Merkle)
		success := k.Prove(ctx, root, challenge.ChunkIndex, proofInput)
		if success {
			k.RemoveChallenge(ctx, provider, proofInput.ChallengeId)
		}
		return nil
	} else if len(file.Providers) == int(file.Replicas) {
		// [TODO]: not your file
		return nil
	} else {
		return nil
	}
}

func containsProver(file types.File, prover string) bool {
	for _, provider := range file.Providers {
		if provider == prover {
			return true
		}
	}
	return false
}

func (k Keeper) Prove(ctx context.Context, root []byte, index uint64, input types.Proof) bool {
	hasher := blake3.New()
	hasher.Write(input.Data)
	hashName := hasher.Sum(nil)

	var proof merkletree.Proof // unmarshal proof
	proof.Siblings = input.Hashes
	proof.Index = uint64(input.ChunkIndex)
	// err := json.Unmarshal(input.HashList, &proof)
	// if err != nil {
	// 	return false
	// }

	if index != proof.Index {
		return false
	}

	verified, err := merkletree.Verify(hashName, root, &proof, &merkletree.Config{XXH128: true})
	if err != nil {
		return false
	}

	return verified
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
