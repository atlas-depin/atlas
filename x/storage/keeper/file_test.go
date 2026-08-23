package keeper_test

import (
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestGetFile_NotFound() {
	_, err := suite.keeper.GetFile(suite.ctx, "does-not-exist")
	suite.Require().ErrorIs(err, types.ErrFileNotFound)
}

func (suite *KeeperTestSuite) TestRemoveFile() {
	owner := suite.CreateTestAddress(0)
	suite.createTestFile("to-remove", owner, 1024, 1, nil)

	suite.Require().NoError(suite.keeper.RemoveFile(suite.ctx, "to-remove"))

	_, err := suite.keeper.GetFile(suite.ctx, "to-remove")
	suite.Require().ErrorIs(err, types.ErrFileNotFound)
}

func (suite *KeeperTestSuite) TestStabilizeProvider_RemovesFromUnstable() {
	owner := suite.CreateTestAddress(0)
	p1 := suite.CreateTestAddress(1)
	p2 := suite.CreateTestAddress(2)

	file := suite.createTestFile("stabilize-file", owner, 1024, 2, []string{p1})
	file.UnstableProviders = []string{p1, p2}
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	suite.Require().NoError(suite.keeper.StabilizeProvider(suite.ctx, p1, file))

	stored, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Require().Equal([]string{p2}, stored.UnstableProviders)
}

func (suite *KeeperTestSuite) TestStabilizeProvider_NotUnstable_Noop() {
	owner := suite.CreateTestAddress(0)
	p1 := suite.CreateTestAddress(1)

	file := suite.createTestFile("stabilize-noop", owner, 1024, 1, nil)

	suite.Require().NoError(suite.keeper.StabilizeProvider(suite.ctx, p1, file))

	stored, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Require().Empty(stored.UnstableProviders)
}

func (suite *KeeperTestSuite) TestDestabilizeProvider() {
	owner := suite.CreateTestAddress(0)
	p1 := suite.CreateTestAddress(1)

	file := suite.createTestFile("destabilize-file", owner, 1024, 1, nil)

	suite.Require().NoError(suite.keeper.DestabilizeProvider(suite.ctx, p1, file))

	stored, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)
	suite.Require().Equal([]string{p1}, stored.UnstableProviders)
}

func (suite *KeeperTestSuite) TestIterateAllFiles() {
	owner := suite.CreateTestAddress(0)
	suite.createTestFile("file-a", owner, 1024, 1, nil)
	suite.createTestFile("file-b", owner, 2048, 1, nil)

	var fids []string
	err := suite.keeper.IterateAllFiles(suite.ctx, func(file types.File) bool {
		fids = append(fids, file.Fid)
		return false
	})
	suite.Require().NoError(err)
	suite.ElementsMatch([]string{"file-a", "file-b"}, fids)
}

func (suite *KeeperTestSuite) TestIncrementDecrementFileCount() {
	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), count)

	suite.Require().NoError(suite.keeper.IncrementFileCount(suite.ctx))
	suite.Require().NoError(suite.keeper.IncrementFileCount(suite.ctx))
	count, err = suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2), count)

	suite.Require().NoError(suite.keeper.DecrementFileCount(suite.ctx))
	count, err = suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), count)
}
