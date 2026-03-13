package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"nebulix/x/storage/types"
	"strconv"

	"cosmossdk.io/collections"
	"github.com/cespare/xxhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"
)

// TODO: delete failed challenge on end of round?
// TODO: fix issue where idk something bad happens every other window, maybe due to desc?

var json = jsoniter.ConfigFastest

/*
Initiates a challenge window by taking a snapshot of the total file count and resetting the iteration cursor.
*/
func (k Keeper) StartChallengeWindow(ctx sdk.Context) {
	params, _ := k.Params.Get(ctx)
	windowIndex := ctx.BlockHeight() / int64(params.ProofWindowBlocks)

	currentFileCount, _ := k.FileCount.Get(ctx)
	_ = k.ChallengeFilesCount.Set(ctx, currentFileCount)
	_ = k.ChallengeFilesCursor.Set(ctx, []byte{})

	ctx.Logger().Info("Challenge window started", "height", ctx.BlockHeight(), "snapshot_eligible", currentFileCount)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewChallengeWindow,
			sdk.NewAttribute("module", types.ModuleName),
			sdk.NewAttribute("window", strconv.FormatInt(windowIndex, 10)),
		),
	)
}

/*
Initiates a challenge round.
Computes the # of files to challenge this round, iterates through # more files and generates challenges for each file+provider.
*/
func (k Keeper) StartChallengeRound(ctx sdk.Context, roundIdx uint64) {
	params, _ := k.Params.Get(ctx)
	fileCount, _ := k.ChallengeFilesCount.Get(ctx)

	windowIndex := uint64(ctx.BlockHeight() / int64(params.ProofWindowBlocks))
	isAscending := windowIndex%2 == 0

	numThisRound := k.computeProofsThisRound(ctx, fileCount, roundIdx)

	cursorBz, _ := k.ChallengeFilesCursor.Get(ctx)
	ctx.Logger().Warn("<StartChallengeRound> ---------------")
	ctx.Logger().Warn(fmt.Sprintf("<StartChallengeRound> challenge cursor: %x", cursorBz))
	ctx.Logger().Warn(fmt.Sprintf("<StartChallengeRound> files this round: %d", numThisRound))

	// compute iteration cursor
	var resume bool
	var cursor collections.Triple[string, string, string]
	if len(cursorBz) > 0 {
		_, decoded, err := k.FileStore.KeyCodec().Decode(cursorBz)
		if err == nil {
			cursor = decoded
			resume = true
		} else {
			ctx.Logger().Error("Null cursor decode, starting from beginning", "err", err)
		}
	} else if roundIdx > 0 {
		ctx.Logger().Warn("No more files remainin for proof round")
		return
	}

	// set iteration cursor & direction
	crange := &collections.Range[collections.Triple[string, string, string]]{}
	if isAscending {
		if resume {
			crange.StartInclusive(cursor)
		}
	} else {
		if resume {
			crange.EndInclusive(cursor)
		}
		crange.Descending()
	}

	// iterate
	var counter uint64
	var lastKeyBz []byte = cursorBz
	iter, err := k.FileStore.Iterate(ctx, crange)
	if err != nil {
		ctx.Logger().Error("Failed to iterate FileStore", "err", err, "ascending", isAscending)
		return
	}
	defer iter.Close()
	if iter.Valid() && resume {
		iter.Next()
	}

	for iter.Valid() && counter < numThisRound {
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

		if file.Status == "STAGED" {
			iter.Next()
			continue
		}

		k.IssueChallengeForFile(ctx, file, roundIdx)
		counter++
		if counter == numThisRound {
			size := k.FileStore.KeyCodec().Size(triple)
			keyBz := make([]byte, size)
			written, encErr := k.FileStore.KeyCodec().Encode(keyBz, triple)
			if encErr != nil {
				ctx.Logger().Error("Encode failed", "err", encErr)
				iter.Next()
				continue
			}
			lastKeyBz = keyBz[:written] // safe slice
		}

		iter.Next()
	}

	// update cursor to last successfully processed key
	if setErr := k.ChallengeFilesCursor.Set(ctx, lastKeyBz); setErr != nil {
		ctx.Logger().Error("Failed to update challenge cursor", "err", setErr)
	}

	ctx.Logger().Warn("Challenge round executed",
		"round", roundIdx,
		"window", windowIndex,
		"direction_ascending", isAscending,
		"challenged", counter,
	)
	ctx.Logger().Warn("<StartChallengeRound> ---------------")

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewChallengeRound,
			sdk.NewAttribute("module", types.ModuleName),
			sdk.NewAttribute("round", strconv.FormatUint(roundIdx, 10)),
			sdk.NewAttribute("window", strconv.FormatUint(windowIndex, 10)),
			sdk.NewAttribute("challenged_count", strconv.FormatUint(counter, 10)),
		),
	)
}

/*
Generate and issue proof challenges for a file.
*/
func (k Keeper) IssueChallengeForFile(ctx context.Context, file types.File, round uint64) ([]types.StorageChallenge, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, _ := k.Params.Get(ctx)

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
	prevHash := sdkCtx.BlockHeader().LastBlockId.Hash
	height := uint64(sdkCtx.BlockHeight())

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
		challengeID := types.GenerateChallengeID(height, file.Fid, providerAddr)
		challenge := types.StorageChallenge{
			ChallengeId:    challengeID,
			FileId:         file.Fid,
			FileMerkle:     file.Merkle,
			Provider:       providerAddr,
			ChunkIndex:     chunkIndex,
			CreatedHeight:  height,
			DeadlineHeight: height + params.ProofRoundBlocks,
		}

		if err := k.SetChallenge(ctx, challenge); err != nil {
			sdkCtx.Logger().Error("failed to store challenge", "err", err, "fid", file.Fid, "provider", providerAddr)
			continue
		}
	}

	return nil, nil
}

/*
HandleMissedChallenges scans all active challenges and handles timeouts
*/
func (k Keeper) HandleMissedChallenges(ctx sdk.Context) {
	currentHeight := uint64(ctx.BlockHeight()) + 1

	// Iterate over ALL challenges (assuming you have a prefix map for them)
	iter, err := k.ChallengeStore.Iterate(ctx, nil) // full range; adjust if you have a different prefix
	if err != nil {
		ctx.Logger().Error("Failed to iterate challenges for missed detection", "err", err)
		return
	}
	defer iter.Close()

	var missedCount uint64

	for iter.Valid() {
		challenge, err := iter.Value()
		if err != nil {
			ctx.Logger().Error("Failed to decode challenge", "err", err)
			iter.Next()
			continue
		}

		// Check deadline
		if currentHeight >= uint64(challenge.DeadlineHeight) {

			// Missed proof - apply penalty
			if k.HandleMissedChallenge(ctx, challenge) {
				missedCount++
			}

			// Remove it
			if err := k.RemoveChallenge(ctx, challenge.Provider, challenge.ChallengeId); err != nil {
				ctx.Logger().Error("Failed to remove missed challenge", "id", challenge.ChallengeId, "err", err)
			}

			ctx.Logger().Warn("Detected and penalized missed challenge",
				"challenge_id", challenge.ChallengeId,
				"provider", challenge.Provider,
				"file_id", challenge.FileId,
				"deadline", challenge.DeadlineHeight,
				"current_height", currentHeight,
			)
		}

		iter.Next()
	}

	if missedCount > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMissedProofsDetected,
				sdk.NewAttribute("missed_count", fmt.Sprintf("%d", missedCount)),
				sdk.NewAttribute("height", fmt.Sprintf("%d", currentHeight)),
			),
		)
	}
}

/*
Slash provider rewards and remove the provider from the file.
*/
func (k Keeper) HandleMissedChallenge(ctx sdk.Context, challenge types.StorageChallenge) bool {
	file, err := k.GetFile(ctx, challenge.FileId)
	if err != nil {
		return false
	}

	t_0, err := k.LastRewardsTime.Get(ctx)
	if err != nil {
		return false
	}
	t_1 := ctx.BlockTime().Unix()
	d_t := t_0.Unix() - t_1

	k.DecrementProviderSpaceUsed(ctx, challenge.Provider, file.FileSize, d_t*2)

	for i, v := range file.Providers {
		if v == challenge.Provider {
			file.Providers = append(file.Providers[:i], file.Providers[i+1:]...)
			break
		}
	}
	k.SetFile(ctx, file)
	return true
}

func (k Keeper) computeProofsThisRound(ctx sdk.Context, n uint64, roundIdx uint64) uint64 {
	params, _ := k.Params.Get(ctx)
	kBlocks := params.ProofRoundBlocks
	gBlocks := params.ProofWindowBlocks

	block := roundIdx * kBlocks

	// g(x) = ceil((n * roundIdx + kBlocks) / gBlocks) - ceil((roundIdx * n) / gBlocks)
	// ceil(a/b) = (a + b - 1) / b
	// THIS EXPRESSION IS SIMPLIFIED:
	term1 := (n*(block+kBlocks) + gBlocks - 1) / gBlocks
	term2 := (block*n + gBlocks - 1) / gBlocks
	return term1 - term2
}
