package types

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/cespare/xxhash"
)

func GenerateChallengeID(height uint64, fileID string, providerAddr string) string {
	hStr := strconv.FormatUint(height, 10)
	pStr := strconv.FormatUint(xxhash.Sum64String(providerAddr), 16)
	return hStr + "-" + fileID + "-" + pStr
}

func ComputeChunkCount(fileSize int64) uint64 {
	if fileSize == 0 {
		return 0
	}

	chunks := fileSize / 1024
	if fileSize%1024 != 0 {
		chunks++
	}
	return uint64(chunks)
}

func CreateProviderChallengeSeed(baseSeed []byte, seedOffset uint64, providerAddr string, providerIndex uint64) []byte {
	hasher := sha256.New()

	// Base randomness (from block)
	hasher.Write(baseSeed)

	// File-specific offset
	offsetBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(offsetBytes, seedOffset)
	hasher.Write(offsetBytes)

	// Provider-specific components
	hasher.Write([]byte(providerAddr))

	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, providerIndex)
	hasher.Write(indexBytes)

	return hasher.Sum(nil)
}

func GenerateRandomChallengeChunkIndex(providerSeed []byte, chunkCount uint64, fileID string) (uint64, error) {
	if chunkCount == 0 {
		return 0, fmt.Errorf("chunkCount cannot be zero")
	}

	// Mix file ID for extra uniqueness
	hasher := sha256.New()
	hasher.Write(providerSeed)
	hasher.Write([]byte(fileID))

	finalSeed := hasher.Sum(nil)
	vr := NewVerifiableRandomness(finalSeed)

	// Generate random index (0 to chunkCount-1)
	return vr.NextRandom(chunkCount), nil
}
