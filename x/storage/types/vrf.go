package types

import (
	"context"
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var GetVerifiableRandomSeed = getVerifiableRandomSeed

// GetVerifiableRandomSeed creates seed from block hash
func getVerifiableRandomSeed(ctx context.Context) ([]byte, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Use previous block's hash for randomness
	header := sdkCtx.BlockHeader()

	// Combine block hash with app hash for more entropy
	blockHash := header.LastBlockId.Hash
	appHash := header.AppHash

	hasher := sha256.New()
	hasher.Write(blockHash)
	hasher.Write(appHash)

	// Add height to ensure uniqueness per round
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(sdkCtx.BlockHeight()))
	hasher.Write(heightBytes)

	return hasher.Sum(nil), nil
}

// VerifiableRandomness for generating random numbers
type VerifiableRandomness struct {
	seed    []byte
	counter uint64
	buffer  []byte
	pos     int
}

func NewVerifiableRandomness(seed []byte) *VerifiableRandomness {
	return &VerifiableRandomness{
		seed:    seed,
		counter: 0,
		buffer:  make([]byte, 0),
		pos:     0,
	}
}

// NextRandom generates random number in [0, max)
func (vr *VerifiableRandomness) NextRandom(max uint64) uint64 {
	if max == 0 {
		return 0
	}

	// Need bytes for randomness
	bytesNeeded := 8 // Use 8 bytes for 64-bit number

	// Generate random bytes
	randomBytes := vr.nextBytes(bytesNeeded)

	// Convert to uint64
	var value uint64
	for i := 0; i < 8; i++ {
		value = (value << 8) | uint64(randomBytes[i])
	}

	// Simple modulo (for now - can improve with rejection sampling)
	return value % max
}

// nextBytes gets n random bytes from seed stream
func (vr *VerifiableRandomness) nextBytes(n int) []byte {
	result := make([]byte, n)

	for i := 0; i < n; i++ {
		if vr.pos >= len(vr.buffer) {
			vr.refillBuffer()
		}
		result[i] = vr.buffer[vr.pos]
		vr.pos++
	}

	return result
}

// refillBuffer generates more random bytes
func (vr *VerifiableRandomness) refillBuffer() {
	hasher := sha256.New()

	// Write seed
	hasher.Write(vr.seed)

	// Write counter
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, vr.counter)
	hasher.Write(counterBytes)

	vr.buffer = hasher.Sum(nil)
	vr.pos = 0
	vr.counter++
}
