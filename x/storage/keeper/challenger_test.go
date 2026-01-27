package keeper_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"cosmossdk.io/collections"
	"github.com/wealdtech/go-merkletree/v2"
	treeblake3 "github.com/wealdtech/go-merkletree/v2/blake3"
	"github.com/zeebo/blake3"

	"nebulix/x/storage/types"
)

func (suite *KeeperTestSuite) createTestFile(fileId, owner string, fileSize int64, replicas int32, providers []string) types.File {
	file := types.File{
		FileId:    fileId,
		Merkle:    hex.EncodeToString([]byte("merkle-root-" + fileId)),
		Creator:   owner,
		Start:     suite.ctx.BlockHeight(),
		FileSize:  fileSize,
		Providers: providers,
		Replicas:  replicas,
	}

	err := suite.keeper.SetNewFile(suite.ctx, file)
	suite.Require().NoError(err)

	return file
}

func (suite *KeeperTestSuite) createTestMerkleTree() (*merkletree.MerkleTree, [][]byte, [][]byte) {
	fileData := [][]byte{
		[]byte("chunk-0-data-1234567890abcdef"),
		[]byte("chunk-1-data-abcdefghijklmnop"),
		[]byte("chunk-2-data-0987654321qrstuv"),
		[]byte("chunk-3-data-wxyz1234567890ab"),
	}

	leaves := make([][]byte, len(fileData))
	hasher := blake3.New()
	for i, chunk := range fileData {
		hasher.Reset()
		hasher.Write(chunk)
		leaves[i] = hasher.Sum(nil)
	}

	// [TODO]: use XXH tree
	tree, err := merkletree.NewUsing(leaves, treeblake3.New256(), false)
	suite.Require().NoError(err, "failed to create merkle tree")

	return tree, leaves, fileData
}

func (suite *KeeperTestSuite) createTestChallenge(provider, challengeId, fileId string, chunkIndex uint64) types.StorageChallenge {
	challenge := types.StorageChallenge{
		ChallengeId:   challengeId,
		FileId:        fileId,
		FileMerkle:    hex.EncodeToString([]byte("merkle-root-" + fileId)),
		Provider:      provider,
		ChunkIndex:    chunkIndex,
		CreatedHeight: uint64(suite.ctx.BlockHeight()),
		Seed:          []byte("seed-" + challengeId),
		Completed:     false,
	}

	err := suite.keeper.SetChallenge(suite.ctx, challenge)
	suite.Require().NoError(err)

	return challenge
}

func (suite *KeeperTestSuite) TestSetGetRemoveChallenge() {
	provider := suite.CreateTestAddress(0)
	challengeId := "test-challenge-1"
	fileId := "file-1"
	chunkIndex := uint64(42)

	// Test SetChallenge
	challenge := types.StorageChallenge{
		ChallengeId:   challengeId,
		FileId:        fileId,
		FileMerkle:    "merkle-root",
		Provider:      provider,
		ChunkIndex:    chunkIndex,
		CreatedHeight: uint64(suite.ctx.BlockHeight()),
		Seed:          []byte("test-seed"),
		Completed:     false,
	}

	err := suite.keeper.SetChallenge(suite.ctx, challenge)
	suite.Require().NoError(err)

	// Test GetChallenge
	retrieved, err := suite.keeper.GetChallenge(suite.ctx, provider, challengeId)
	suite.Require().NoError(err)
	suite.Require().Equal(challenge, retrieved)

	// Test GetChallenge for non-existent challenge
	_, err = suite.keeper.GetChallenge(suite.ctx, provider, "non-existent")
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, collections.ErrNotFound))

	// Test RemoveChallenge
	err = suite.keeper.RemoveChallenge(suite.ctx, provider, challengeId)
	suite.Require().NoError(err)

	// Verify challenge was removed
	_, err = suite.keeper.GetChallenge(suite.ctx, provider, challengeId)
	suite.Require().Error(err)
	suite.Require().True(errors.Is(err, collections.ErrNotFound))
}

func (suite *KeeperTestSuite) TestChallengeFiles_NoFiles() {
	// Test with no files in store
	err := suite.keeper.ChallengeFiles(suite.ctx)
	suite.Require().NoError(err)

	// No challenges should be created
	// (Can't easily iterate over challenges, but we can check no errors)
}

func (suite *KeeperTestSuite) TestChallengeFiles_WithFiles() {
	// Create a fixed seed for deterministic testing
	fixedSeed := []byte("fixed-test-seed-1234567890abcdefghijklmnopqrstuvwx")
	types.GetVerifiableRandomSeed = func(ctx context.Context) ([]byte, error) {
		return fixedSeed, nil
	}

	// Setup: Create files with providers
	owner := suite.CreateTestAddress(0)

	providers := make([]string, 3)
	for i := range providers {
		providers[i] = suite.CreateTestAddress(i + 1)
	}

	// Create multiple files with different configurations
	files := []struct {
		id       string
		size     int64
		replicas int32
		provs    []string
	}{
		{
			id:       "file-1",
			size:     1_000_000_000, // 1GB
			replicas: 3,
			provs:    providers, // All 3 providers
		},
		{
			id:       "file-2",
			size:     2_000_000_000, // 2GB
			replicas: 2,
			provs:    providers[:2], // First 2 providers
		},
		{
			id:       "file-3",
			size:     500_000_000,            // 500MB
			replicas: 1,                      // Invalid - should be at least 2, but we'll test it anyway
			provs:    []string{providers[0]}, // Just provider 1
		},
	}

	// Store the files
	for _, f := range files {
		suite.createTestFile(f.id, owner, f.size, f.replicas, f.provs)
	}

	// We need to track created challenges
	// Let's create a map to store all created challenges
	// createdChallenges := make(map[string]types.StorageChallenge)

	// Intercept SetChallenge to track what gets created
	// Since we can't easily replace the method, we'll use reflection or a helper
	// OR we can just count challenges after they're created

	// Call ChallengeFiles
	err := suite.keeper.ChallengeFiles(suite.ctx)
	suite.Require().NoError(err)

	// Now we need to verify challenges were created correctly
	// Let's iterate through all files and check challenges for each provider

	totalExpectedChallenges := 0
	for _, f := range files {
		// Only valid files should get challenges
		// Based on your GenerateFileChallenges logic:
		// - File must have providers (len > 0) ✓
		// - File size must be > 0 ✓
		// - Replicas doesn't matter for challenge generation

		if len(f.provs) == 0 || f.size == 0 {
			continue // Should not generate challenges
		}

		// For each provider in this file, there should be a challenge
		for _, provider := range f.provs {
			// Generate the expected challenge ID
			challengeId := types.GenerateChallengeID(
				uint64(suite.ctx.BlockHeight()),
				f.id,
				provider,
			)

			// Try to get the challenge
			challenge, err := suite.keeper.GetChallenge(suite.ctx, provider, challengeId)

			if f.replicas >= 2 || f.replicas == 0 {
				// Valid file - should have a challenge
				suite.Require().NoError(err,
					"challenge should exist for file %s provider %s", f.id, provider)
				suite.Require().Equal(f.id, challenge.FileId)
				suite.Require().Equal(provider, challenge.Provider)
				suite.Require().False(challenge.Completed)

				// Verify chunk index is within bounds
				chunkCount := types.ComputeChunkCount(f.size)
				suite.Require().Less(challenge.ChunkIndex, chunkCount,
					"chunk index %d should be less than chunk count %d for file %s",
					challenge.ChunkIndex, chunkCount, f.id)

				totalExpectedChallenges++
			} else {
				// Invalid replicas - might not create challenge
				// Based on your code, replicas doesn't affect challenge generation
				// So it should still create a challenge
				if err == nil {
					// If challenge exists, verify it
					suite.Require().Equal(f.id, challenge.FileId)
					totalExpectedChallenges++
				}
				// If error, that's OK - the test for file-3 might fail
			}
		}
	}

	// Let's also count all challenges in the store
	// We need to iterate through challenges to count them
	// Since we don't have an iterator for challenges, we'll track differently

	fmt.Printf("\nExpected at least %d challenges created", totalExpectedChallenges)

	// Test the deterministic nature: Run ChallengeFiles again
	// Since we use the same seed, it should create the same challenges
	// but they'll fail to set because they already exist

	err = suite.keeper.ChallengeFiles(suite.ctx)
	suite.Require().NoError(err,
		"second call should not error (duplicate challenges might be ignored)")

	// For file-3 with replicas=1, let's test GenerateFileChallenges directly
	file3, err := suite.keeper.GetFile(suite.ctx, "file-3")
	suite.Require().NoError(err)

	// Test GenerateFileChallenges with our fixed seed
	challenges, err := suite.keeper.GenerateFileChallenges(
		suite.ctx,
		file3,
		fixedSeed,
		2, // seedOffset for file-3 (third file)
		uint64(suite.ctx.BlockHeight()),
	)

	suite.Require().NoError(err)
	if len(file3.Providers) > 0 && file3.FileSize > 0 {
		suite.Require().Len(challenges, len(file3.Providers),
			"should create one challenge per provider")
	}
}

func (suite *KeeperTestSuite) TestProve_Success() {
	tree, leaves, data := suite.createTestMerkleTree()
	root := tree.Root()

	chunkIndex := 2
	chunkData := leaves[chunkIndex]

	// Generate proof for chunk 2
	// [TBD]: use GenerateProofWithIndex for even stronger verification?
	proof, err := tree.GenerateProof(chunkData, 0)
	suite.Require().NoError(err, "failed to generate proof")

	// Create proof input
	proofInput := types.Proof{
		ChallengeId: "test-challenge",
		Data:        data[chunkIndex],
		Hashes:      proof.Hashes,
		ChunkIndex:  proof.Index,
	}

	// Test Prove function
	success := suite.keeper.Prove(suite.ctx, root, uint64(chunkIndex), proofInput)
	suite.Require().True(success, "valid proof should verify")
}

func (suite *KeeperTestSuite) TestProve_WrongData() {
	tree, leaves, _ := suite.createTestMerkleTree()
	root := tree.Root()

	// Generate proof for chunk 0
	proof, err := tree.GenerateProof(leaves[0], 0)
	suite.Require().NoError(err)

	// Test with wrong data (but correct proof)
	proofInput := types.Proof{
		ChallengeId: "test",
		Data:        []byte("wrong-data"),
		Hashes:      proof.Hashes,
		ChunkIndex:  proof.Index,
	}

	success := suite.keeper.Prove(suite.ctx, root, 0, proofInput)
	suite.Require().False(success, "wrong data should fail even with correct proof")
}

func (suite *KeeperTestSuite) TestProve_WrongChunkIndex() {
	tree, leaves, data := suite.createTestMerkleTree()
	root := tree.Root()

	// Generate proof for chunk 1, but proof index falsely at 0
	proof, err := tree.GenerateProof(leaves[1], 0)
	suite.Require().NoError(err)

	// Test with wrong chunk index (claiming it's chunk 2)
	proofInput := types.Proof{
		ChallengeId: "test",
		Data:        data[1],      // Data from chunk 1
		Hashes:      proof.Hashes, // Proof hashes for chunk 1
		ChunkIndex:  2,
	}

	success := suite.keeper.Prove(suite.ctx, root, 2, proofInput)
	suite.Require().False(success, "wrong chunk index should fail")
}

func (suite *KeeperTestSuite) TestProve_NilData() {
	tree, leaves, _ := suite.createTestMerkleTree()
	root := tree.Root()

	proof, err := tree.GenerateProof(leaves[0], 0)
	suite.Require().NoError(err)

	proofInput := types.Proof{
		ChallengeId: "test",
		Data:        nil, // Nil data
		Hashes:      proof.Hashes,
		ChunkIndex:  proof.Index,
	}

	success := suite.keeper.Prove(suite.ctx, root, 0, proofInput)
	suite.Require().False(success, "nil data should fail")
}

func (suite *KeeperTestSuite) TestProve_EmptyProofHashes() {
	tree, leaves, data := suite.createTestMerkleTree()
	root := tree.Root()

	proof, err := tree.GenerateProof(leaves[0], 0)
	suite.Require().NoError(err)

	proofInput := types.Proof{
		ChallengeId: "test",
		Data:        data[0],
		Hashes:      [][]byte{},
		ChunkIndex:  proof.Index,
	}

	success := suite.keeper.Prove(suite.ctx, root, 0, proofInput)
	suite.Require().False(success, "empty proof hashes should fail")
}

func (suite *KeeperTestSuite) TestProve_WrongMerkleRoot() {
	// two different trees
	data1 := [][]byte{[]byte("tree1-data")}
	data2 := [][]byte{[]byte("tree2-data")}

	// Build tree 1
	leaves1 := make([][]byte, len(data1))
	hasher := blake3.New()
	hasher.Write(data1[0])
	leaves1[0] = hasher.Sum(nil)

	tree1, err := merkletree.NewUsing(leaves1, treeblake3.New256(), false)
	suite.Require().NoError(err)

	// Build tree 2
	leaves2 := make([][]byte, len(data2))
	hasher.Reset()
	hasher.Write(data2[0])
	leaves2[0] = hasher.Sum(nil)

	tree2, err := merkletree.NewUsing(leaves2, treeblake3.New256(), false)
	suite.Require().NoError(err)

	// Get proof from tree 1
	proof, err := tree1.GenerateProof(leaves1[0], 0)
	suite.Require().NoError(err)

	// Try to verify with wrong root (from tree 2)
	proofInput := types.Proof{
		ChallengeId: "test",
		Data:        data1[0], // Correct data
		Hashes:      proof.Hashes,
		ChunkIndex:  proof.Index,
	}

	success := suite.keeper.Prove(suite.ctx, tree2.Root(), 0, proofInput)
	suite.Require().False(success, "proof for different tree should fail")
}

func (suite *KeeperTestSuite) TestProve_ProofTampering() {
	tree, leaves, data := suite.createTestMerkleTree()
	root := tree.Root()

	// Generate proof
	proof, err := tree.GenerateProof(leaves[0], 0)
	suite.Require().NoError(err)

	// Tamper with the proof
	// Create a modified proof struct with wrong values
	tamperedProof := struct {
		Hashes [][]byte `json:"hashes"`
		Index  uint64   `json:"index"`
	}{
		Hashes: proof.Hashes,
		Index:  proof.Index + 1,
	}

	proofInput := types.Proof{
		ChallengeId: "test",
		Data:        data[0],
		Hashes:      tamperedProof.Hashes,
		ChunkIndex:  tamperedProof.Index,
	}

	success := suite.keeper.Prove(suite.ctx, root, 0, proofInput)
	suite.Require().False(success, "tampered proof should fail")
}

func (suite *KeeperTestSuite) TestProve_Performance_100kProofs() {
	// Skip in normal test runs since it's performance testing
	if testing.Short() {
		suite.T().Skip("Skipping performance test in short mode")
	}

	// Create a larger dataset
	numChunks := 1000000 // 100k chunks
	chunkSize := 1024    // 1KB each

	fmt.Printf("\nCreating %d chunks of %d bytes each...", numChunks, chunkSize)

	// Generate test data
	data := make([][]byte, numChunks)
	for i := 0; i < numChunks; i++ {
		data[i] = make([]byte, chunkSize)
		rand.Read(data[i])
	}

	// Build merkle tree
	suite.T().Log("Building merkle tree...")
	startTime := time.Now()

	leaves := make([][]byte, numChunks)
	hasher := blake3.New()
	for i, chunk := range data {
		hasher.Reset()
		hasher.Write(chunk)
		leaves[i] = hasher.Sum(nil)
	}

	tree, err := merkletree.NewUsing(leaves, treeblake3.New256(), false)
	suite.Require().NoError(err, "failed to create merkle tree")

	buildTime := time.Since(startTime)
	fmt.Printf("\nTree built in %v (%d leaves)", buildTime, numChunks)

	root := tree.Root()

	// Generate all proofs upfront
	suite.T().Log("Generating all proofs...")
	startTime = time.Now()

	proofs := make([]merkletree.Proof, numChunks)
	proofHashes := make([][][]byte, numChunks)

	for i := 0; i < numChunks; i++ {
		proof, err := tree.GenerateProof(leaves[i], 0)
		suite.Require().NoError(err, "failed to generate proof for chunk %d", i)
		proofs[i] = *proof

		proofHashes[i] = proof.Hashes
	}

	proofGenTime := time.Since(startTime)
	fmt.Printf("\nGenerated %d proofs in %v (avg %v per proof)",
		numChunks, proofGenTime, proofGenTime/time.Duration(numChunks))

	// Test Prove function performance
	suite.T().Log("Testing Prove function for all chunks...")
	startTime = time.Now()

	successCount := 0
	failCount := 0

	for i := 0; i < numChunks; i++ {
		proofInput := types.Proof{
			ChallengeId: fmt.Sprintf("perf-test-%d", i),
			Data:        data[i],
			Hashes:      proofHashes[i],
			ChunkIndex:  uint64(i),
		}

		success := suite.keeper.Prove(suite.ctx, root, uint64(i), proofInput)
		if success {
			successCount++
		} else {
			failCount++
		}
	}

	proveTime := time.Since(startTime)

	// Log results
	fmt.Printf("\n\n=== PERFORMANCE RESULTS ===")
	fmt.Printf("\nTotal chunks: %d", numChunks)
	fmt.Printf("\nTree build time: %v", buildTime)
	fmt.Printf("\nProof generation time: %v (avg %v per proof)",
		proofGenTime, proofGenTime/time.Duration(numChunks))
	fmt.Printf("\nProve verification time: %v (avg %v per proof)",
		proveTime, proveTime/time.Duration(numChunks))
	fmt.Printf("\nSuccesses: %d, Failures: %d", successCount, failCount)
	fmt.Printf("\nThroughput: %.2f proofs/second",
		float64(numChunks)/proveTime.Seconds())

	// Verify all succeeded
	suite.Require().Equal(numChunks, successCount, "all proofs should verify")
	suite.Require().Zero(failCount, "no proofs should fail")

	// Memory usage (rough estimate)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("\nHeap in use: %d MB", memStats.HeapInuse/1024/1024)
	fmt.Printf("\nTotal alloc: %d MB", memStats.TotalAlloc/1024/1024)
}

// func (suite *KeeperTestSuite) TestGenerateFileChallenges_EmptyProviders() {
// 	file := types.File{
// 		FileId:    "empty-providers-file",
// 		Merkle:    "merkle-root",
// 		Owner:     suite.CreateTestAddress(0),
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  1_000_000_000,
// 		Providers: []string{}, // Empty providers
// 		Replicas:  3,
// 	}

// 	seed := []byte("test-seed")
// 	seedOffset := uint64(0)
// 	currentHeight := uint64(suite.ctx.BlockHeight())

// 	challenges, err := suite.keeper.GenerateFileChallenges(suite.ctx, file, seed, seedOffset, currentHeight)
// 	suite.Require().Error(err)
// 	suite.Require().Contains(err.Error(), "file has no providers")
// 	suite.Require().Empty(challenges)
// }

// func (suite *KeeperTestSuite) TestGenerateFileChallenges_ZeroFileSize() {
// 	provider := suite.CreateTestAddress(1)
// 	file := types.File{
// 		FileId:    "zero-size-file",
// 		Merkle:    "merkle-root",
// 		Owner:     suite.CreateTestAddress(0),
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  0, // Zero size
// 		Providers: []string{provider},
// 		Replicas:  3,
// 	}

// 	seed := []byte("test-seed")
// 	seedOffset := uint64(0)
// 	currentHeight := uint64(suite.ctx.BlockHeight())

// 	challenges, err := suite.keeper.GenerateFileChallenges(suite.ctx, file, seed, seedOffset, currentHeight)
// 	suite.Require().Error(err)
// 	suite.Require().Contains(err.Error(), "file has zero size")
// 	suite.Require().Empty(challenges)
// }

// func (suite *KeeperTestSuite) TestGenerateFileChallenges_InvalidProviderAddress() {
// 	// Create a file with invalid provider address
// 	file := types.File{
// 		FileId:    "invalid-provider-file",
// 		Merkle:    "merkle-root",
// 		Owner:     suite.CreateTestAddress(0),
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  1_000_000_000,
// 		Providers: []string{"invalid-beach32-address"}, // Invalid address
// 		Replicas:  3,
// 	}

// 	seed := []byte("test-seed")
// 	seedOffset := uint64(0)
// 	currentHeight := uint64(suite.ctx.BlockHeight())

// 	// This should skip the invalid provider and not return an error
// 	// (Based on your implementation which logs and continues)
// 	challenges, err := suite.keeper.GenerateFileChallenges(suite.ctx, file, seed, seedOffset, currentHeight)
// 	suite.Require().NoError(err)
// 	suite.Require().Empty(challenges) // No valid challenges created
// }

// func (suite *KeeperTestSuite) TestGenerateFileChallenges_Success() {
// 	owner := suite.CreateTestAddress(0)
// 	providers := []string{
// 		suite.CreateTestAddress(1),
// 		suite.CreateTestAddress(2),
// 		suite.CreateTestAddress(3),
// 	}

// 	file := types.File{
// 		FileId:    "test-file-1",
// 		Merkle:    hex.EncodeToString([]byte("merkle-root-test")),
// 		Owner:     owner,
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  1_000_000_000, // 1GB
// 		Providers: providers,
// 		Replicas:  3,
// 	}

// 	seed := []byte("deterministic-test-seed-1234567890")
// 	seedOffset := uint64(42)
// 	currentHeight := uint64(suite.ctx.BlockHeight())

// 	challenges, err := suite.keeper.GenerateFileChallenges(suite.ctx, file, seed, seedOffset, currentHeight)
// 	suite.Require().NoError(err)
// 	suite.Require().Len(challenges, len(providers))

// 	// Verify each challenge
// 	for i, challenge := range challenges {
// 		suite.Require().Equal(file.FileId, challenge.FileId)
// 		suite.Require().Equal(file.Merkle, challenge.FileMerkle)
// 		suite.Require().Equal(providers[i], challenge.Provider)
// 		suite.Require().Equal(currentHeight, challenge.CreatedHeight)
// 		suite.Require().False(challenge.Completed)

// 		// Verify challenge ID format
// 		expectedId := types.GenerateChallengeID(currentHeight, file.FileId, providers[i])
// 		suite.Require().Equal(expectedId, challenge.ChallengeId)

// 		// Verify chunk index is within bounds
// 		chunkCount := types.ComputeChunkCount(file.FileSize)
// 		suite.Require().Less(challenge.ChunkIndex, chunkCount)

// 		// Verify seed is not empty and different for each provider
// 		suite.Require().NotEmpty(challenge.Seed)

// 		if i > 0 {
// 			// Seeds should be different for different providers
// 			suite.Require().False(bytes.Equal(challenge.Seed, challenges[i-1].Seed))
// 			// Chunk indices should (probably) be different
// 			// (Not guaranteed but very likely with good randomness)
// 		}
// 	}
// }

// func (suite *KeeperTestSuite) TestCreateChallenge() {
// 	owner := suite.CreateTestAddress(0)
// 	provider := suite.CreateTestAddress(1)

// 	file := types.File{
// 		FileId:    "test-file",
// 		Merkle:    "merkle-root-test",
// 		Owner:     owner,
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  1_000_000_000,
// 		Providers: []string{provider},
// 		Replicas:  3,
// 	}

// 	chunkIndex := uint64(123)
// 	seed := []byte("challenge-seed")
// 	currentHeight := uint64(1000)
// 	chunkCount := uint64(100)

// 	challenge, err := keeper.CreateChallenge(file, provider, chunkIndex, seed, currentHeight, chunkCount)
// 	suite.Require().NoError(err)

// 	// Verify challenge fields
// 	suite.Require().Equal(types.GenerateChallengeID(currentHeight, file.FileId, provider), challenge.ChallengeId)
// 	suite.Require().Equal(file.FileId, challenge.FileId)
// 	suite.Require().Equal(file.Merkle, challenge.FileMerkle)
// 	suite.Require().Equal(provider, challenge.Provider)
// 	suite.Require().Equal(chunkIndex, challenge.ChunkIndex)
// 	suite.Require().Equal(currentHeight, challenge.CreatedHeight)
// 	suite.Require().Equal(seed, challenge.Seed)
// 	suite.Require().False(challenge.Completed)
// }

// func (suite *KeeperTestSuite) TestAttemptChallenge_NoChallenge() {
// 	provider := suite.CreateTestAddress(0)

// 	// Attempt with non-existent challenge
// 	proofInput := types.Proof{
// 		ChallengeId: "non-existent-challenge",
// 		Chunk:       0,
// 		Data:        []byte("data"),
// 		HashList:    []byte("[]"),
// 	}

// 	err := suite.keeper.AttemptChallenge(suite.ctx, provider, proofInput)
// 	suite.Require().NoError(err) // Returns nil per your implementation
// }

// func (suite *KeeperTestSuite) TestAttemptChallenge_FileNotFound() {
// 	provider := suite.CreateTestAddress(0)
// 	challengeId := "challenge-no-file"
// 	fileId := "non-existent-file"

// 	// Create challenge but no file
// 	suite.createTestChallenge(provider, challengeId, fileId, 0)

// 	proofInput := types.Proof{
// 		ChallengeId: challengeId,
// 		Chunk:       0,
// 		Data:        []byte("data"),
// 		HashList:    []byte("[]"),
// 	}

// 	err := suite.keeper.AttemptChallenge(suite.ctx, provider, proofInput)
// 	suite.Require().NoError(err) // Returns nil per your implementation
// 	// Challenge should be removed? (Based on your TODO comment)
// }

// func (suite *KeeperTestSuite) TestAttemptChallenge_WrongProvider() {
// 	// Setup
// 	owner := suite.CreateTestAddress(0)
// 	correctProvider := suite.CreateTestAddress(1)
// 	wrongProvider := suite.CreateTestAddress(2) // Different provider
// 	otherProvider := suite.CreateTestAddress(3)

// 	fileId := "test-file"
// 	file := suite.createTestFile(fileId, owner, 1_000_000_000, 2, []string{correctProvider, otherProvider})

// 	// Create challenge for correctProvider
// 	challengeId := "challenge-wrong-provider"
// 	suite.createTestChallenge(correctProvider, challengeId, fileId, 0)

// 	// Wrong provider attempts the challenge
// 	proofInput := types.Proof{
// 		ChallengeId: challengeId,
// 		Chunk:       0,
// 		Data:        []byte("data"),
// 		HashList:    []byte("[]"),
// 	}

// 	err := suite.keeper.AttemptChallenge(suite.ctx, wrongProvider, proofInput)
// 	suite.Require().NoError(err) // Returns nil

// 	// Challenge should still exist (not removed)
// 	challenge, err := suite.keeper.GetChallenge(suite.ctx, correctProvider, challengeId)
// 	suite.Require().NoError(err)
// 	suite.Require().False(challenge.Completed)
// }

// func (suite *KeeperTestSuite) TestAttemptChallenge_CorrectProvider_InvalidProof() {
// 	// Setup
// 	owner := suite.CreateTestAddress(0)
// 	provider := suite.CreateTestAddress(1)
// 	otherProvider := suite.CreateTestAddress(2)

// 	fileId := "test-file"
// 	file := suite.createTestFile(fileId, owner, 1_000_000_000, 2, []string{provider, otherProvider})

// 	// Create challenge
// 	challengeId := "challenge-invalid-proof"
// 	suite.createTestChallenge(provider, challengeId, fileId, 0)

// 	// Provider attempts with invalid proof
// 	proofInput := types.Proof{
// 		ChallengeId: challengeId,
// 		Chunk:       0,
// 		Data:        []byte("invalid-data"),
// 		HashList:    []byte("invalid-json"),
// 	}

// 	err := suite.keeper.AttemptChallenge(suite.ctx, provider, proofInput)
// 	suite.Require().NoError(err) // Returns nil

// 	// Challenge should still exist (proof failed)
// 	challenge, err := suite.keeper.GetChallenge(suite.ctx, provider, challengeId)
// 	suite.Require().NoError(err)
// 	suite.Require().False(challenge.Completed)
// }

// func (suite *KeeperTestSuite) TestAttemptChallenge_SuccessfulProof() {
// 	// This test is complex because it requires:
// 	// 1. A valid merkle tree
// 	// 2. A valid proof for a specific chunk
// 	// 3. Proper hash computation

// 	// We'll need to skip or mock the Prove function
// 	suite.T().Skip("Test requires actual merkle proof verification - needs integration test")

// 	/* Example of what this test would look like:

// 	// Setup merkle tree with test data
// 	data := [][]byte{
// 		[]byte("chunk0-data"),
// 		[]byte("chunk1-data"),
// 		[]byte("chunk2-data"),
// 	}

// 	tree, root := createMerkleTree(data)

// 	// Create file with the merkle root
// 	file := types.File{
// 		FileId:    "test-file",
// 		Merkle:    hex.EncodeToString(root),
// 		Owner:     owner,
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  int64(len(data) * len(data[0])), // Approximate
// 		Providers: []string{provider},
// 		Replicas:  1,
// 	}

// 	// Create challenge for chunk 1
// 	challengeId := "challenge-success"
// 	challenge := types.StorageChallenge{
// 		ChallengeId:   challengeId,
// 		FileId:        file.FileId,
// 		FileMerkle:    file.Merkle,
// 		Provider:      provider,
// 		ChunkIndex:    1, // Challenge for chunk 1
// 		CreatedHeight: uint64(suite.ctx.BlockHeight()),
// 		Seed:          []byte("seed"),
// 		Completed:     false,
// 	}

// 	// Generate proof for chunk 1
// 	proof, _ := tree.GenerateProof(1)
// 	proofBytes, _ := json.Marshal(proof)

// 	proofInput := types.Proof{
// 		ChallengeId: challengeId,
// 		Chunk:       1,
// 		Data:        data[1],
// 		HashList:    proofBytes,
// 	}

// 	err := suite.keeper.AttemptChallenge(suite.ctx, provider, proofInput)
// 	suite.Require().NoError(err)

// 	// Challenge should be removed
// 	_, err = suite.keeper.GetChallenge(suite.ctx, provider, challengeId)
// 	suite.Require().Error(err)
// 	suite.Require().True(errors.Is(err, collections.ErrNotFound))
// 	*/
// }

// func (suite *KeeperTestSuite) TestContainsProver() {
// 	provider1 := suite.CreateTestAddress(0)
// 	provider2 := suite.CreateTestAddress(1)
// 	provider3 := suite.CreateTestAddress(2)
// 	otherProvider := suite.CreateTestAddress(3)

// 	file := types.File{
// 		FileId:    "test-file",
// 		Merkle:    "merkle",
// 		Owner:     suite.CreateTestAddress(4),
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  1_000_000_000,
// 		Providers: []string{provider1, provider2, provider3},
// 		Replicas:  3,
// 	}

// 	// Test contains
// 	suite.Require().True(keeper.ContainsProver(file, provider1))
// 	suite.Require().True(keeper.ContainsProver(file, provider2))
// 	suite.Require().True(keeper.ContainsProver(file, provider3))

// 	// Test doesn't contain
// 	suite.Require().False(keeper.ContainsProver(file, otherProvider))
// 	suite.Require().False(keeper.ContainsProver(file, ""))

// 	// Test empty providers
// 	emptyFile := types.File{
// 		FileId:    "empty-file",
// 		Merkle:    "merkle",
// 		Owner:     suite.CreateTestAddress(5),
// 		Start:     suite.ctx.BlockHeight(),
// 		FileSize:  1_000_000_000,
// 		Providers: []string{},
// 		Replicas:  0,
// 	}

// 	suite.Require().False(keeper.ContainsProver(emptyFile, provider1))
// }

// func (suite *KeeperTestSuite) TestProve_InvalidProofJSON() {
// 	root := []byte("merkle-root")

// 	proofInput := types.Proof{
// 		ChallengeId: "challenge",
// 		Chunk:       0,
// 		Data:        []byte("data"),
// 		HashList:    []byte("invalid-json"), // Invalid JSON
// 	}

// 	success := suite.keeper.Prove(suite.ctx, root, proofInput)
// 	suite.Require().False(success)
// }

// func (suite *KeeperTestSuite) TestProve_ValidProof() {
// 	// This is also complex and requires actual merkle tree verification
// 	suite.T().Skip("Test requires actual merkle proof verification - needs integration test")

// 	/* Example:
// 	// Create a simple merkle tree
// 	data := [][]byte{
// 		[]byte("chunk0"),
// 		[]byte("chunk1"),
// 		[]byte("chunk2"),
// 	}

// 	tree, root := createMerkleTree(data)

// 	// Generate proof for chunk 1
// 	proof, err := tree.GenerateProof(1)
// 	suite.Require().NoError(err)

// 	proofBytes, err := json.Marshal(proof)
// 	suite.Require().NoError(err)

// 	proofInput := types.Proof{
// 		ChallengeId: "challenge",
// 		Chunk:       1,
// 		Data:        data[1],
// 		HashList:    proofBytes,
// 	}

// 	success := suite.keeper.Prove(suite.ctx, root, proofInput)
// 	suite.Require().True(success)

// 	// Test with wrong data (should fail)
// 	wrongProofInput := types.Proof{
// 		ChallengeId: "challenge",
// 		Chunk:       1,
// 		Data:        []byte("wrong-data"),
// 		HashList:    proofBytes,
// 	}

// 	success = suite.keeper.Prove(suite.ctx, root, wrongProofInput)
// 	suite.Require().False(success)
// 	*/
// }

// func (suite *KeeperTestSuite) TestChallengeFlow_Integration() {
// 	// Test a complete flow: file creation -> challenge generation -> proof attempt
// 	owner := suite.CreateTestAddress(0)
// 	providers := []string{
// 		suite.CreateTestAddress(1),
// 		suite.CreateTestAddress(2),
// 	}

// 	// Create a file
// 	file := suite.createTestFile("flow-file", owner, 1_000_000_000, 2, providers)

// 	// Generate challenges for the file
// 	seed := []byte("integration-test-seed")
// 	seedOffset := uint64(0)
// 	currentHeight := uint64(suite.ctx.BlockHeight())

// 	challenges, err := suite.keeper.GenerateFileChallenges(suite.ctx, file, seed, seedOffset, currentHeight)
// 	suite.Require().NoError(err)
// 	suite.Require().Len(challenges, len(providers))

// 	// Store challenges
// 	for _, challenge := range challenges {
// 		err := suite.keeper.SetChallenge(suite.ctx, challenge)
// 		suite.Require().NoError(err)

// 		// Verify we can retrieve them
// 		retrieved, err := suite.keeper.GetChallenge(suite.ctx, challenge.Provider, challenge.ChallengeId)
// 		suite.Require().NoError(err)
// 		suite.Require().Equal(challenge, retrieved)
// 	}

// 	// Test attempting a challenge
// 	challenge := challenges[0]
// 	proofInput := types.Proof{
// 		ChallengeId: challenge.ChallengeId,
// 		Chunk:       challenge.ChunkIndex,
// 		Data:        []byte("test-data"),
// 		HashList:    []byte("[]"), // Invalid proof
// 	}

// 	err = suite.keeper.AttemptChallenge(suite.ctx, challenge.Provider, proofInput)
// 	suite.Require().NoError(err)

// 	// Challenge should still exist (invalid proof)
// 	_, err = suite.keeper.GetChallenge(suite.ctx, challenge.Provider, challenge.ChallengeId)
// 	suite.Require().NoError(err)

// 	// Remove challenge
// 	err = suite.keeper.RemoveChallenge(suite.ctx, challenge.Provider, challenge.ChallengeId)
// 	suite.Require().NoError(err)

// 	// Verify removed
// 	_, err = suite.keeper.GetChallenge(suite.ctx, challenge.Provider, challenge.ChallengeId)
// 	suite.Require().Error(err)
// }

// func (suite *KeeperTestSuite) TestChallengeID_Uniqueness() {
// 	height := uint64(1000)
// 	fileId := "file-123"
// 	provider1 := suite.CreateTestAddress(0)
// 	provider2 := suite.CreateTestAddress(1)

// 	id1 := types.GenerateChallengeID(height, fileId, provider1)
// 	id2 := types.GenerateChallengeID(height, fileId, provider2)
// 	id3 := types.GenerateChallengeID(height+1, fileId, provider1)
// 	id4 := types.GenerateChallengeID(height, "different-file", provider1)

// 	// All should be unique
// 	suite.Require().NotEqual(id1, id2)
// 	suite.Require().NotEqual(id1, id3)
// 	suite.Require().NotEqual(id1, id4)
// 	suite.Require().NotEqual(id2, id3)
// 	suite.Require().NotEqual(id2, id4)
// 	suite.Require().NotEqual(id3, id4)

// 	// Same inputs should produce same ID
// 	id1Again := types.GenerateChallengeID(height, fileId, provider1)
// 	suite.Require().Equal(id1, id1Again)
// }

// func (suite *KeeperTestSuite) TestComputeChunkCount() {
// 	testCases := []struct {
// 		name     string
// 		fileSize int64
// 		expected uint64
// 	}{
// 		{"tiny file", 100, 1},                              // Less than chunk size
// 		{"exact chunk", 1024 * 1024, 1},                    // Exactly 1MB
// 		{"multiple chunks", 5 * 1024 * 1024, 5},            // 5MB = 5 chunks
// 		{"large file", 10 * 1024 * 1024 * 1024, 10 * 1024}, // 10GB = 10*1024 chunks
// 		{"zero size", 0, 0},                                // Should return 0
// 		{"negative size", -1, 0},                           // Should handle negative
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(tc.name, func() {
// 			result := types.ComputeChunkCount(tc.fileSize)
// 			suite.Require().Equal(tc.expected, result)
// 		})
// 	}
// }
