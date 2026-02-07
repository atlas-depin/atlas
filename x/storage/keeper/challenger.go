package keeper

import (
	"bytes"
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

func (k Keeper) StartChallengeWindow(ctx sdk.Context) {
	// Example placeholder logic:
	// - Iterate over active storage providers or deals.
	// - Compute a hash or Merkle root of the state.
	// - Store it in KVStore with a key like "proof_snapshot:<window_start_height>".
	// - Emit an event for traceability.
	_ = k.ChallengeCursor.Set(ctx, []byte{})
	ctx.Logger().Info("Taking proof window snapshot", "height", ctx.BlockHeight())

	// TODO: Your snapshot logic here, e.g.:
	// snapshot := k.ComputeStorageStateSnapshot(ctx) // Hypothetical method
	// store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("proof_snapshots"))
	// store.Set(sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())), snapshot.Bytes())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewChallengeWindow,
			// sdk.NewAttribute(, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

func (k Keeper) StartChallengeRound(ctx sdk.Context, roundIdx uint64) {
	params, _ := k.Params.Get(ctx)
	windowIndex := uint64(ctx.BlockHeight() / int64(params.ProofWindowBlocks))
	isAscending := windowIndex%2 == 0

	numThisRound := uint64(300) // replace with your g(x) formula + snapshot N later

	cursorBz, _ := k.ChallengeCursor.Get(ctx)

	var startTriple collections.Triple[string, string, string]
	var hasStart bool
	if len(cursorBz) > 0 {
		n, decoded, err := k.FileStore.KeyCodec().Decode(cursorBz)
		if err == nil && n == len(cursorBz) {
			startTriple = decoded
			hasStart = true
		} else {
			ctx.Logger().Error("Invalid cursor decode, starting from beginning", "err", err)
		}
	}

	crange := &collections.Range[collections.Triple[string, string, string]]{}

	if isAscending {
		if hasStart {
			crange.StartInclusive(startTriple)
		}
	} else {
		// Descending
		if hasStart {
			crange.StartInclusive(startTriple) // resume from previous cursor
		} else {
			// Fresh descending window: start from sentinel
			crange.StartInclusive(types.MaxTripleSentinel)
		}
		crange.Descending()
	}

	iter, err := k.FileStore.Iterate(ctx, crange)
	if err != nil {
		ctx.Logger().Error("Failed to iterate FileStore", "err", err, "ascending", isAscending)
		return
	}
	defer iter.Close()

	var selectedCount uint64
	var lastKeyBz []byte

	for iter.Valid() && selectedCount < numThisRound {
		triple, keyErr := iter.Key()
		if keyErr != nil {
			ctx.Logger().Error("Failed to decode key", "err", keyErr)
			iter.Next()
			continue
		}

		file, valErr := iter.Value()
		if valErr != nil {
			ctx.Logger().Error("Failed to decode value", "err", valErr)
			iter.Next()
			continue
		}

		// Encode triple correctly
		size := k.FileStore.KeyCodec().Size(triple)
		keyBz := make([]byte, size)
		written, encErr := k.FileStore.KeyCodec().Encode(keyBz, triple)
		if encErr != nil {
			ctx.Logger().Error("Encode failed", "err", encErr)
			iter.Next()
			continue
		}
		keyBz = keyBz[:written] // safe slice (normally written == size)

		// Skip cursor key
		if hasStart && bytes.Equal(keyBz, cursorBz) {
			iter.Next()
			continue
		}

		// if !k.IsFileEligibleForChallenge(ctx, file) { ... }

		k.IssueChallengeForFile(ctx, file, roundIdx)
		selectedCount++
		lastKeyBz = keyBz // now safe to assign

		iter.Next()
	}

	// Update cursor to last successfully processed key
	if len(lastKeyBz) > 0 {
		if setErr := k.ChallengeCursor.Set(ctx, lastKeyBz); setErr != nil {
			ctx.Logger().Error("Failed to update challenge cursor", "err", setErr)
		}
	} else if selectedCount == 0 && hasStart {
		// Exhausted in this direction → reset cursor for next round/window
		if setErr := k.ChallengeCursor.Set(ctx, []byte{}); setErr != nil {
			ctx.Logger().Error("Failed to reset challenge cursor", "err", setErr)
		}
	}

	// Logging for observability
	ctx.Logger().Info("Challenge round executed",
		"round", roundIdx,
		"window_index", windowIndex,
		"direction_ascending", isAscending,
		"challenged", selectedCount,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewChallengeRound,
			sdk.NewAttribute("round", fmt.Sprintf("%d", roundIdx)),
			sdk.NewAttribute("window_index", fmt.Sprintf("%d", windowIndex)),
			sdk.NewAttribute("direction_ascending", fmt.Sprintf("%t", isAscending)),
			sdk.NewAttribute("challenged_count", fmt.Sprintf("%d", selectedCount)),
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
