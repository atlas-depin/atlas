package keeper_test

import (
	"fmt"
	"testing"
	"time"

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

func (suite *KeeperTestSuite) TestStartChallengeRound_Ascending_FirstRound() {
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

func (suite *KeeperTestSuite) TestStartChallengeRound_Descending_FirstRound_Sentinel() {
	params := types.DefaultParams()
	params.ProofWindowBlocks = 1200
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// Odd window → descending
	suite.ctx = suite.ctx.WithBlockHeight(1150) // 1150 / 1000 = 1

	// Create files with varying K3
	owner := suite.CreateTestAddress(0)
	suite.createTestFile("file-low-a", owner, 1024*1024, 3, []string{suite.CreateTestAddress(1)})
	suite.createTestFile("file-mid-m", owner, 1024*1024, 3, []string{suite.CreateTestAddress(1)})
	suite.createTestFile("file-high-z", owner, 1024*1024, 3, []string{suite.CreateTestAddress(1)})

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

	// Create 10 ordered files
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

// func (suite *KeeperTestSuite) TestDirectionToggle_EvenOddWindows() {
// 	params := types.DefaultParams()
// 	params.ProofWindowBlocks = 1200
// 	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

// 	// Window 0 (even) → ascending
// 	suite.ctx = suite.ctx.WithBlockHeight(500)
// 	suite.keeper.StartChallengeRound(suite.ctx, 0)
// 	// In test, you'd need to spy on ranger.Order or log to verify ascending

// 	// Window 1 (odd) → descending
// 	suite.ctx = suite.ctx.WithBlockHeight(1500)
// 	suite.keeper.StartChallengeRound(suite.ctx, 0)
// 	// Verify descending (e.g., sentinel used if no cursor)
// 	cursor, _ := suite.keeper.ChallengeFilesCursor.Get(suite.ctx)
// 	suite.NotEmpty(cursor) // sentinel should have been used
// }

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
	suite.Empty(cursorBz, "cursor should be empty")
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
