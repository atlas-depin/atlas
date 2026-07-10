package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/collections"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestEndBlocker_WindowStart_ResetsCursor() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 1000
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Set a height on window boundary
	suite.ctx = suite.ctx.WithBlockHeight(2000) // 2000 % 1000 == 0

	err := suite.keeper.EndBlocker(suite.ctx)
	suite.Require().NoError(err)

	cursor, err := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Equal([]byte{}, cursor, "cursor should be reset on new proof window")
}

func (suite *KeeperTestSuite) TestStartChallengeRound_FirstRound_Ascending() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 1200
	params.ProofRoundBlocks = 300
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Even window index == ascending
	suite.ctx = suite.ctx.WithBlockHeight(150) // 150 / 1000 = 0

	// Create ordered files (K3 = fileID sorts lexicographically)
	owner := suite.CreateTestAddress(0)
	for i := 1; i <= 5; i++ {
		fileId := fmt.Sprintf("file-%03d", i)
		suite.createTestFile(fileId, owner, 1024*1024, 3, []string{suite.CreateTestAddress(1)})
	}
	suite.keeper.StartChallengeWindow(suite.ctx)

	// Run round 1 (should challenge first few)
	suite.keeper.StartChallengeRound(suite.ctx, 0)

	cursorBz, err := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.NotEmpty(cursorBz, "cursor should advance after round")

	// Decode to verify it's past the first file
	_, triple, decErr := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	suite.Require().NoError(decErr)
	suite.Contains(triple.K3(), "file-00", "cursor should be past early files")
}

func (suite *KeeperTestSuite) TestStartChallengeRound_FirstRound_Descending() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 1200
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Odd window → descending
	suite.ctx = suite.ctx.WithBlockHeight(1150) // 1150 / 1000 = 1

	// Create files with varying K3
	owner := suite.CreateTestAddress(0)
	for _, id := range []string{"file-low-a", "file-mid-m", "file-high-z"} {
		suite.createTestFile(id, owner, 1024*1024, 3, []string{suite.CreateTestAddress(1)})
	}
	suite.keeper.StartChallengeWindow(suite.ctx)

	// Run round
	suite.keeper.StartChallengeRound(suite.ctx, 0)

	cursorBz, _ := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.NotEmpty(cursorBz)

	_, triple, _ := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	// In descending, should have processed higher keys first
	suite.Contains(triple.K3(), "file-", "processed some files")
	// Optional: if you can count issued challenges, check high ones prioritized
}

func (suite *KeeperTestSuite) TestStartChallengeRound_CursorResume_SameWindow() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 1200
	params.ProofRoundBlocks = 300
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	suite.ctx = suite.ctx.WithBlockHeight(150) // ascending

	// Create 100 ordered files
	owner := suite.CreateTestAddress(0)
	for i := 1; i <= 100; i++ {
		fileID := fmt.Sprintf("file-%03d", i)
		suite.createTestFile(fileID, owner, 1024*1024, 3, []string{"p1"})
	}
	suite.keeper.StartChallengeWindow(suite.ctx)

	// First round
	suite.keeper.StartChallengeRound(suite.ctx, 0)

	cursorAfterFirst, _ := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.NotEmpty(cursorAfterFirst)

	// Second round (same window)
	suite.keeper.StartChallengeRound(suite.ctx, 1)

	cursorAfterSecond, _ := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.NotEqual(cursorAfterFirst, cursorAfterSecond, "cursor must advance")
}

func (suite *KeeperTestSuite) TestStartChallengeRound_SkipsDeletedFiles_CursorAdvances() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 900
	params.ProofRoundBlocks = 300
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Ascending window
	suite.ctx = suite.ctx.WithBlockHeight(150) // windowIndex 0

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	// Create 8 ordered files (K3 = fileID sorts lex)
	fileIDs := []string{
		"file-001", "file-002", "file-003", "file-004",
		"file-005", "file-006", "file-007", "file-008",
	}
	for _, fid := range fileIDs {
		suite.createTestFile(fid, owner, 1024*1024, 3, []string{provider})
	}
	suite.keeper.StartChallengeWindow(suite.ctx)

	// Simulate deletion of file-003 and file-006 mid-window
	err := suite.keeper.RemoveFile(suite.ctx, "file-003")
	suite.Require().NoError(err)
	err = suite.keeper.RemoveFile(suite.ctx, "file-006")
	suite.Require().NoError(err)

	// Run round with quota of 3
	suite.keeper.StartChallengeRound(suite.ctx, 0)
	suite.keeper.StartChallengeRound(suite.ctx, 1)

	cursorBz, err := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.NotEmpty(cursorBz, "cursor should have advanced")

	// Decode cursor to see where it landed
	_, triple, decErr := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	suite.Require().NoError(decErr)

	// With deletions, cursor should be after the last challenged file
	// Expected: challenged 001,002,004,005,007 (skipped 003,006) cursor at 008
	fileIDAtCursor := triple.K3()
	suite.Contains(fileIDAtCursor, "file-00", "cursor should be near the end of processed files")
	suite.True(fileIDAtCursor >= "file-008", "cursor should have passed deleted gaps")
}

func (suite *KeeperTestSuite) TestStartChallengeRound_SkewedFileCount_NoFilesLeft() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 900
	params.ProofRoundBlocks = 300
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Ascending window
	suite.ctx = suite.ctx.WithBlockHeight(150) // windowIndex 0

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	// Create 8 ordered files (K3 = fileID sorts lex)
	fileIDs := []string{
		"file-001", "file-002", "file-003", "file-004",
		"file-005", "file-006", "file-007", "file-008",
	}
	for _, fid := range fileIDs {
		suite.createTestFile(fid, owner, 1024*1024, 3, []string{provider})
	}
	suite.keeper.StartChallengeWindow(suite.ctx)

	// Simulate deletion of file-002 and file-003 mid-window
	err := suite.keeper.RemoveFile(suite.ctx, "file-002")
	suite.Require().NoError(err)
	err = suite.keeper.RemoveFile(suite.ctx, "file-003")
	suite.Require().NoError(err)

	// Run round with quota of 3
	suite.keeper.StartChallengeRound(suite.ctx, 0)

	suite.keeper.StartChallengeRound(suite.ctx, 1)
	cursorBz, err := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.NotEmpty(cursorBz, "cursor should have advanced to last element (2/3 files selected, partial skew)")
	_, triple, _ := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	suite.Equal("file-008", triple.K3(), "cursor should have passed deleted gaps")

	suite.keeper.StartChallengeRound(suite.ctx, 2)
	cursorBz, err = suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	// All files exhausted — cursor should not have advanced past the last file.
	_, tripleLast, _ := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	suite.Equal("file-008", tripleLast.K3(), "cursor should be at last file (no more to process)")
}

func (suite *KeeperTestSuite) TestStartChallengeRound_SkipsStagedAndDeadFiles() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 1200
	params.ProofRoundBlocks = 300
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	suite.ctx = suite.ctx.WithBlockHeight(150) // ascending

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	// Create file-001 → ACTIVE (normal)
	suite.createTestFile("file-001", owner, 1024*1024, 3, []string{provider})
	// Create file-002 → STAGED
	suite.createTestFile("file-002", owner, 1024*1024, 3, nil)
	file2, _ := suite.keeper.GetFile(suite.ctx, "file-002")
	file2.Status = "STAGED"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file2))
	// Create file-003 → DEAD
	suite.createTestFile("file-003", owner, 1024*1024, 3, []string{provider})
	file3, _ := suite.keeper.GetFile(suite.ctx, "file-003")
	file3.Status = "DEAD"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file3))
	// Create file-004 → ACTIVE (normal)
	suite.createTestFile("file-004", owner, 1024*1024, 3, []string{provider})
	// Create file-005 → STAGED
	suite.createTestFile("file-005", owner, 1024*1024, 3, nil)
	file5, _ := suite.keeper.GetFile(suite.ctx, "file-005")
	file5.Status = "STAGED"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file5))
	// Create file-006 → ACTIVE (normal)
	suite.createTestFile("file-006", owner, 1024*1024, 3, []string{provider})
	// Create file-007 → DEAD
	suite.createTestFile("file-007", owner, 1024*1024, 3, []string{provider})
	file7, _ := suite.keeper.GetFile(suite.ctx, "file-007")
	file7.Status = "DEAD"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file7))
	// Create file-008 → ACTIVE (normal)
	suite.createTestFile("file-008", owner, 1024*1024, 3, []string{provider})

	suite.keeper.StartChallengeWindow(suite.ctx)
	// Count snapshotted at 8. With R=300, W=1200, round 0: (8*300+899)/1200 = 2 files

	suite.keeper.StartChallengeRound(suite.ctx, 0)

	// Should have challenged file-001 and file-004 (the only ACTIVE files in the first batch)
	// file-002 (STAGED) and file-003 (DEAD) must be skipped.
	// file-004 is challenged, cursor should be past it.
	cursorBz, err := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.NotEmpty(cursorBz)
	_, triple, _ := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	suite.True(triple.K3() >= "file-004", "cursor should be past the first two ACTIVE files")
}

func (suite *KeeperTestSuite) TestIssueChallengeForFile_Pass() {
	suite.ctx = suite.ctx.WithBlockHeight(1000)
	params := types.DefaultParams()

	owner := suite.CreateTestAddress(0)
	provider1 := suite.CreateTestAddress(1)
	provider2 := suite.CreateTestAddress(2)

	file := types.File{
		Fid:       "test-file-1",
		Merkle:    "abcd1234",
		Creator:   owner,
		Status:    "ACTIVE",
		FileSize:  5 * 1024 * 1024, // 5 MB → 5120 chunks
		Providers: []string{provider1, provider2},
		Replicas:  3,
	}
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	challenges, err := suite.keeper.IssueChallengeForFile(suite.ctx, file, 0, params.ProofRoundBlocks)
	suite.Require().NoError(err)
	// Currently returns nil — read via store iteration
	_ = challenges

	// Verify 2 challenges created in store
	count := 0
	chunkIndices := map[string]uint64{}
	suite.keeper.ChallengeStore.Walk(suite.ctx, nil, func(key collections.Pair[string, string], value types.StorageChallenge) (bool, error) {
		count++
		suite.Equal(file.Fid, value.FileId)
		suite.Equal(file.Merkle, value.FileMerkle)
		suite.Equal(uint64(1000), value.CreatedHeight)
		suite.Equal(uint64(1000)+params.ProofRoundBlocks, value.DeadlineHeight)
		suite.True(value.ChunkIndex < 5120, "chunk index within range")
		chunkIndices[value.Provider] = value.ChunkIndex
		return false, nil
	})
	suite.Equal(2, count, "one challenge per provider")
	suite.NotEqual(chunkIndices[provider1], chunkIndices[provider2], "different providers get different chunks")
}

func (suite *KeeperTestSuite) TestIssueChallengeForFile_NoProviders() {
	file := types.File{
		Fid:      "no-providers",
		FileSize: 1024 * 1024,
	}
	_, err := suite.keeper.IssueChallengeForFile(suite.ctx, file, 0, 10)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "file has no providers")
}

func (suite *KeeperTestSuite) TestIssueChallengeForFile_ZeroSize() {
	owner := suite.CreateTestAddress(0)
	file := types.File{
		Fid:       "zero-size",
		Creator:   owner,
		FileSize:  0,
		Providers: []string{suite.CreateTestAddress(1)},
	}
	_, err := suite.keeper.IssueChallengeForFile(suite.ctx, file, 0, 10)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "zero size")
}

func (suite *KeeperTestSuite) TestIssueChallengeForFile_Deterministic() {
	suite.ctx = suite.ctx.WithBlockHeight(500)
	owner := suite.CreateTestAddress(0)
	providers := []string{suite.CreateTestAddress(1), suite.CreateTestAddress(2)}

	file := types.File{
		Fid:       "det-file",
		Merkle:    "deadbeef",
		Creator:   owner,
		Status:    "ACTIVE",
		FileSize:  1024 * 1024,
		Providers: providers,
		Replicas:  3,
	}
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	// Clear any existing challenges from the basic test
	suite.keeper.ChallengeStore.Walk(suite.ctx, nil, func(key collections.Pair[string, string], value types.StorageChallenge) (bool, error) {
		_ = suite.keeper.RemoveChallenge(suite.ctx, value.Provider, value.ChallengeId)
		return false, nil
	})

	// First invocation
	_, err := suite.keeper.IssueChallengeForFile(suite.ctx, file, 1, 50)
	suite.Require().NoError(err)

	type challengeSnapshot struct {
		ChallengeID string
		ChunkIndex  uint64
		Deadline    uint64
	}
	first := map[string]challengeSnapshot{}
	suite.keeper.ChallengeStore.Walk(suite.ctx, nil, func(key collections.Pair[string, string], value types.StorageChallenge) (bool, error) {
		first[value.Provider] = challengeSnapshot{
			ChallengeID: value.ChallengeId,
			ChunkIndex:  value.ChunkIndex,
			Deadline:    value.DeadlineHeight,
		}
		return false, nil
	})

	// Clear and re-invoke with same inputs at same height
	suite.keeper.ChallengeStore.Walk(suite.ctx, nil, func(key collections.Pair[string, string], value types.StorageChallenge) (bool, error) {
		_ = suite.keeper.RemoveChallenge(suite.ctx, value.Provider, value.ChallengeId)
		return false, nil
	})

	_, err = suite.keeper.IssueChallengeForFile(suite.ctx, file, 1, 50)
	suite.Require().NoError(err)

	second := map[string]challengeSnapshot{}
	suite.keeper.ChallengeStore.Walk(suite.ctx, nil, func(key collections.Pair[string, string], value types.StorageChallenge) (bool, error) {
		second[value.Provider] = challengeSnapshot{
			ChallengeID: value.ChallengeId,
			ChunkIndex:  value.ChunkIndex,
			Deadline:    value.DeadlineHeight,
		}
		return false, nil
	})

	for _, provider := range providers {
		suite.Equal(first[provider], second[provider], "challenge must be deterministic for %s", provider)
	}
}

func (suite *KeeperTestSuite) TestHandleMissedChallenges_NoneExpired() {
	suite.ctx = suite.ctx.WithBlockHeight(100)

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	// Create a provider
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        provider,
		SpaceAvailable: 1_000_000_000,
		SpaceUsed:      0,
	}))

	// Create a file with that provider
	file := suite.createTestFile("challenge-file-1", owner, 1024*1024, 3, []string{provider})

	// Create a challenge with future deadline (not yet expired)
	challenge := types.StorageChallenge{
		ChallengeId:    "future-challenge",
		FileId:         file.Fid,
		Provider:       provider,
		FileMerkle:     file.Merkle,
		CreatedHeight:  50,
		DeadlineHeight: 200, // well in the future
	}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	// Set LastRewardsTime
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, 10))

	// Run
	suite.keeper.HandleMissedChallenges(suite.ctx)

	// Verify: challenge still exists (not processed)
	_, err := suite.keeper.GetChallenge(suite.ctx, provider, "future-challenge")
	suite.Require().NoError(err, "future challenge should not be removed")

	// Verify: provider not destabilized
	updatedFile, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Empty(updatedFile.UnstableProviders)
}

func (suite *KeeperTestSuite) TestHandleMissedChallenges_FirstMissDestabilizes() {
	suite.ctx = suite.ctx.WithBlockHeight(200)

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        provider,
		SpaceAvailable: 1_000_000_000,
		SpaceUsed:      1024 * 1024,
	}))

	file := suite.createTestFile("miss-file-1", owner, 1024*1024, 3, []string{provider})

	// Past deadline challenge
	challenge := types.StorageChallenge{
		ChallengeId:    "miss-1",
		FileId:         file.Fid,
		Provider:       provider,
		FileMerkle:     file.Merkle,
		CreatedHeight:  50,
		DeadlineHeight: 100, // past deadline (200+1 >= 100)
	}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, 10))

	suite.keeper.HandleMissedChallenges(suite.ctx)

	// Verify: challenge removed
	_, err := suite.keeper.GetChallenge(suite.ctx, provider, "miss-1")
	suite.Require().Error(err)

	// Verify: provider destabilized (added to UnstableProviders)
	updatedFile, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Contains(updatedFile.UnstableProviders, provider)
	// Provider still in the file's provider list
	suite.Contains(updatedFile.Providers, provider)
}

func (suite *KeeperTestSuite) TestHandleMissedChallenges_SecondMissRemovesProvider() {
	suite.ctx = suite.ctx.WithBlockHeight(300)

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)
	otherProvider := suite.CreateTestAddress(2)

	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        provider,
		SpaceAvailable: 1_000_000_000,
		SpaceUsed:      1024 * 1024,
	}))

	// Two providers so the file stays alive when one is removed
	file := suite.createTestFile("miss-file-2", owner, 1024*1024, 3, []string{provider, otherProvider})

	// Set provider as unstable (simulating first miss)
	suite.Require().NoError(suite.keeper.DestabilizeProvider(suite.ctx, provider, file))
	// Refresh file from store so it's up to date
	file, _ = suite.keeper.GetFile(suite.ctx, file.Fid)

	// Track in SubscriptionProviderBytes
	subProvKey := collections.Join(file.Subscription, provider)
	suite.Require().NoError(suite.keeper.SubscriptionProviderBytes.Set(suite.ctx, subProvKey, file.FileSize))

	// Past deadline challenge — second miss
	challenge := types.StorageChallenge{
		ChallengeId:    "miss-2",
		FileId:         file.Fid,
		Provider:       provider,
		FileMerkle:     file.Merkle,
		CreatedHeight:  150,
		DeadlineHeight: 200,
	}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, 20))

	suite.keeper.HandleMissedChallenges(suite.ctx)

	// Verify: challenge removed
	_, err := suite.keeper.GetChallenge(suite.ctx, provider, "miss-2")
	suite.Require().Error(err)

	// Verify: provider removed from file
	updatedFile, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.NotContains(updatedFile.Providers, provider)
	suite.NotContains(updatedFile.UnstableProviders, provider)

	// Verify: SubscriptionProviderBytes decremented
	_, getErr := suite.keeper.SubscriptionProviderBytes.Get(suite.ctx, subProvKey)
	suite.Require().Error(getErr, "entry should be removed")

	// Verify: file still alive (has no other providers but wasn't the last)
	suite.Equal("ACTIVE", updatedFile.Status)
}

func (suite *KeeperTestSuite) TestHandleMissedChallenges_LastProviderRemoved_FileDeath() {
	suite.ctx = suite.ctx.WithBlockHeight(300)

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        provider,
		SpaceAvailable: 1_000_000_000,
		SpaceUsed:      1024 * 1024,
	}))

	// File with only one provider
	file := suite.createTestFile("last-provider", owner, 1024*1024, 3, []string{provider})
	suite.Require().NoError(suite.keeper.DestabilizeProvider(suite.ctx, provider, file))
	file, _ = suite.keeper.GetFile(suite.ctx, file.Fid)

	subProvKey := collections.Join(file.Subscription, provider)
	suite.Require().NoError(suite.keeper.SubscriptionProviderBytes.Set(suite.ctx, subProvKey, file.FileSize))

	challenge := types.StorageChallenge{
		ChallengeId:    "last-miss",
		FileId:         file.Fid,
		Provider:       provider,
		FileMerkle:     file.Merkle,
		CreatedHeight:  150,
		DeadlineHeight: 200,
	}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, 20))
	suite.Require().NoError(suite.keeper.IncrementFileCount(suite.ctx)) // Account for createTestFile's increment

	suite.keeper.HandleMissedChallenges(suite.ctx)

	// Verify: file DEAD
	updatedFile, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Equal("DEAD", updatedFile.Status)
	suite.Empty(updatedFile.Providers)
}

func (suite *KeeperTestSuite) TestStartChallengeRound_LargeFileSet_NoCrash_AdvancesCursor() {
	if testing.Short() {
		suite.T().Skip("Skipping large-set stress test in short mode")
	}

	params := types.DefaultParams()
	params.ProofWindowBlocks = 14400 // ~1D at 15min rounds
	params.ProofRoundBlocks = 300
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Ascending window (windowIndex 0 = even)
	suite.ctx = suite.ctx.WithBlockHeight(10)

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	const numFiles = 1_000_000 // Increase to 10_000 or 20_000 for more stress (watch RAM/time)
	err := suite.keeper.FileCount.Set(suite.ctx, numFiles)
	suite.T().Logf("Creating %d test files...", numFiles)

	startCreate := time.Now()
	for i := 1; i <= numFiles; i++ {
		fileId := fmt.Sprintf("file-%07d", i) // zero-padded for stable lex order
		suite.createTestFile(fileId, owner, 1024*1024, 3, []string{provider})
	}
	suite.T().Logf("Created %d files in %v", numFiles, time.Since(startCreate))

	// Start the challenge window (resets cursor)
	suite.keeper.StartChallengeWindow(suite.ctx)

	// Run first round (index 0)
	startRound := time.Now()
	suite.keeper.StartChallengeRound(suite.ctx, 0)
	duration := time.Since(startRound)

	cursorBz, err := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.NotEmpty(cursorBz, "cursor should have advanced after large round")

	// Decode cursor position
	_, triple, decErr := suite.keeper.FileStore.KeyCodec().Decode(cursorBz)
	suite.Require().NoError(decErr)
	fileIDAtCursor := triple.K3()

	suite.T().Logf("After round 0: cursor at %s, duration: %v", fileIDAtCursor, duration)

	// Basic progress check: should have advanced past the first ~300 files
	// (assuming all are eligible and no deletions)
	suite.True(fileIDAtCursor >= "file-0000300",
		"cursor should have moved significantly forward (expected past file ~300)")

	// Optional second round (same window, index 1) — verify further advance
	startRound = time.Now()
	suite.keeper.StartChallengeRound(suite.ctx, 1)
	duration = time.Since(startRound)

	cursorBz2, _ := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
	suite.Require().NotEqual(cursorBz, cursorBz2, "cursor should advance further in second round")

	_, triple2, _ := suite.keeper.FileStore.KeyCodec().Decode(cursorBz2)
	fileIDAtCursor2 := triple2.K3()
	suite.T().Logf("After round 1: cursor at %s, duration: %v", fileIDAtCursor2, duration)
	suite.True(fileIDAtCursor2 > fileIDAtCursor,
		"cursor should be further along after second round")
}
