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
