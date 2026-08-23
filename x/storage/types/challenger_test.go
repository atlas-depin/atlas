package types_test

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/cespare/xxhash"

	"atlas/x/storage/types"

	"github.com/stretchr/testify/require"
)

func TestGenerateChallengeID(t *testing.T) {
	id := types.GenerateChallengeID(42, "file-1", "cosmos1provider")
	expected := "42-file-1-" + strconv.FormatUint(xxhash.Sum64String("cosmos1provider"), 16)
	require.Equal(t, expected, id)
	require.Equal(t, id, types.GenerateChallengeID(42, "file-1", "cosmos1provider"), "deterministic")

	other := types.GenerateChallengeID(42, "file-1", "cosmos1other")
	require.NotEqual(t, id, other, "provider must be part of the id")

	otherHeight := types.GenerateChallengeID(43, "file-1", "cosmos1provider")
	require.NotEqual(t, id, otherHeight, "height must be part of the id")
}

func TestComputeChunkCount(t *testing.T) {
	cases := []struct {
		fileSize int64
		expected uint64
	}{
		{fileSize: 0, expected: 0},
		{fileSize: 1, expected: 1},
		{fileSize: 1024, expected: 1},
		{fileSize: 1025, expected: 2},
		{fileSize: 2048, expected: 2},
		{fileSize: 2049, expected: 3},
		{fileSize: 10_000_000, expected: 9766}, // ceil(10_000_000 / 1024)
	}
	for _, tc := range cases {
		require.Equal(t, tc.expected, types.ComputeChunkCount(tc.fileSize), "fileSize=%d", tc.fileSize)
	}
}

func TestCreateProviderChallengeSeed(t *testing.T) {
	seed := types.CreateProviderChallengeSeed([]byte("base"), 7, "cosmos1provider", 2)
	require.Len(t, seed, 32, "sha256 output")

	require.Equal(t, seed, types.CreateProviderChallengeSeed([]byte("base"), 7, "cosmos1provider", 2), "deterministic")
	require.NotEqual(t, seed, types.CreateProviderChallengeSeed([]byte("other"), 7, "cosmos1provider", 2))
	require.NotEqual(t, seed, types.CreateProviderChallengeSeed([]byte("base"), 8, "cosmos1provider", 2))
	require.NotEqual(t, seed, types.CreateProviderChallengeSeed([]byte("base"), 7, "cosmos1other", 2))
	require.NotEqual(t, seed, types.CreateProviderChallengeSeed([]byte("base"), 7, "cosmos1provider", 3))
}

func TestGenerateRandomChallengeChunkIndex(t *testing.T) {
	seed := types.CreateProviderChallengeSeed([]byte("base"), 7, "cosmos1provider", 2)

	// Zero chunks must error
	_, err := types.GenerateRandomChallengeChunkIndex(seed, 0, "file-1")
	require.Error(t, err)

	// Results are bounded and deterministic
	idx, err := types.GenerateRandomChallengeChunkIndex(seed, 100, "file-1")
	require.NoError(t, err)
	require.Less(t, idx, uint64(100))

	idxAgain, err := types.GenerateRandomChallengeChunkIndex(seed, 100, "file-1")
	require.NoError(t, err)
	require.Equal(t, idx, idxAgain, "deterministic for the same inputs")

	// Different file id mixes the result
	idxOther, err := types.GenerateRandomChallengeChunkIndex(seed, 100, "file-2")
	require.NoError(t, err)
	require.NotEqual(t, idx, idxOther)
}

func TestVerifiableRandomness(t *testing.T) {
	// max=0 returns 0
	vr := types.NewVerifiableRandomness([]byte("seed"))
	require.Equal(t, uint64(0), vr.NextRandom(0))

	// bounded draws
	vr = types.NewVerifiableRandomness([]byte("seed"))
	for i := 0; i < 10; i++ {
		require.Less(t, vr.NextRandom(50), uint64(50))
	}

	// same seed, same sequence
	a := types.NewVerifiableRandomness([]byte("seed"))
	b := types.NewVerifiableRandomness([]byte("seed"))
	for i := 0; i < 5; i++ {
		require.Equal(t, a.NextRandom(1000), b.NextRandom(1000))
	}

	// different seeds diverge (almost surely)
	c := types.NewVerifiableRandomness([]byte("other"))
	require.NotEqual(t, a.NextRandom(1000), c.NextRandom(1000))
}

func TestVerifiableRandomness_StreamsAcrossBufferRefills(t *testing.T) {
	// A single 32-byte sha256 buffer covers 4 uint64 draws; force a refill by
	// drawing more and confirm the stream stays bounded and deterministic.
	a := types.NewVerifiableRandomness(bytes.Repeat([]byte{0xAB}, 32))
	b := types.NewVerifiableRandomness(bytes.Repeat([]byte{0xAB}, 32))
	for i := 0; i < 12; i++ {
		require.Equal(t, a.NextRandom(1<<40), b.NextRandom(1<<40), "draw %d", i)
	}
}
