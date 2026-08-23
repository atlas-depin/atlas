package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestFileStatsQuery_Empty() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.FileStats(suite.ctx, &types.QueryFileStatsRequest{})
	suite.Require().NoError(err)
	suite.Equal(&types.QueryFileStatsResponse{}, resp)
}

func (suite *KeeperTestSuite) TestFileStatsQuery_WithStrayAndDead() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)

	// Active file
	suite.createTestFile("file-active", owner, 1024, 1, nil)

	// Stray file
	stray := suite.createTestFile("file-stray", owner, 2048, 1, nil)
	stray.Status = "STRAY"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, stray))

	// Dead file (in the real flow this also decrements the file count)
	dead := suite.createTestFile("file-dead", owner, 4096, 1, nil)
	dead.Status = "DEAD"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, dead))
	suite.Require().NoError(suite.keeper.DecrementFileCount(suite.ctx))

	resp, err := qs.FileStats(suite.ctx, &types.QueryFileStatsRequest{})
	suite.Require().NoError(err)
	suite.Equal(int64(3), resp.TotalFiles) // 2 counted files + 1 dead
	suite.Equal(int64(1), resp.StrayFiles)
	suite.Equal(int64(1), resp.DeadFiles)
}

func (suite *KeeperTestSuite) TestFileStatsQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.FileStats(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
