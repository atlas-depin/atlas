package keeper

import (
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Helper: Get a deterministic uint64 stream from your VRF
// Replace with your actual VRF logic
func (k Keeper) getVRFDeterministicUint64(ctx sdk.Context, nonce uint64) uint64 {
	// Example using previous block hash + nonce for stream
	prevHash := ctx.BlockHeader().LastBlockId.Hash
	data := append(prevHash, byte(nonce)) // or use subwindow index as nonce
	hash := sha256.Sum256(data)

	// Take first 8 bytes as uint64 (little endian)
	return binary.LittleEndian.Uint64(hash[:8])
}

// SelectNRandomIndices picks exactly n distinct indices in [0, total-1]
func (k Keeper) SelectRandomChallengeIndices(ctx sdk.Context, n uint64, subwindowIdx uint64) ([]uint64, error) {
	total, err := k.ChallengeFileCount.Get(ctx)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}
	if n > total {
		n = total
	}

	var indices []uint64
	seen := make(map[uint64]struct{}, n)

	// Use subwindowIdx (or block height, etc.) to make each subwindow different
	nonceBase := subwindowIdx

	for uint64(len(indices)) < n {
		nonceBase++
		idx := k.getVRFDeterministicUint64(ctx, nonceBase) % total

		if _, exists := seen[idx]; exists {
			continue
		}

		seen[idx] = struct{}{}
		indices = append(indices, idx)
	}

	return indices, nil
}
