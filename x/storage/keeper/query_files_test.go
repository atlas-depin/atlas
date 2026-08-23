package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestFilesQuery_NoFilter() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	other := suite.CreateTestAddress(1)

	f1 := suite.createTestFile("file-1", owner, 1024, 1, nil)
	f2 := suite.createTestFile("file-2", owner, 2048, 2, nil)
	f3 := suite.createTestFile("file-3", other, 4096, 3, nil)

	resp, err := qs.Files(suite.ctx, &types.QueryFilesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Files, 3)
	suite.ElementsMatch([]*types.File{&f1, &f2, &f3}, resp.Files)
}

func (suite *KeeperTestSuite) TestFilesQuery_ByCreator() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	other := suite.CreateTestAddress(1)

	f1 := suite.createTestFile("file-1", owner, 1024, 1, nil)
	f2 := suite.createTestFile("file-2", owner, 2048, 2, nil)
	suite.createTestFile("file-3", other, 4096, 3, nil)

	resp, err := qs.Files(suite.ctx, &types.QueryFilesRequest{Creator: owner})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Files, 2)
	suite.ElementsMatch([]*types.File{&f1, &f2}, resp.Files)
}

func (suite *KeeperTestSuite) TestFilesQuery_ByCreatorAndSubscription() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.createTestFile("file-1", owner, 1024, 1, nil)
	fSub := suite.createTestFileWithSub("file-sub", owner, "sub_1", 2048, 2, nil)

	resp, err := qs.Files(suite.ctx, &types.QueryFilesRequest{Creator: owner, Subscription: "sub_1"})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Files, 1)
	suite.Equal(fSub, *resp.Files[0])
}

func (suite *KeeperTestSuite) TestFilesQuery_SubscriptionWithoutCreator() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Files(suite.ctx, &types.QueryFilesRequest{Subscription: "sub_1"})
	suite.Require().ErrorContains(err, "need to provide creator to query by subscription")
}

func (suite *KeeperTestSuite) TestFilesQuery_InvalidCreator() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Files(suite.ctx, &types.QueryFilesRequest{Creator: "not-a-valid-bech32"})
	suite.Require().ErrorContains(err, "invalid creator address")
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
}

func (suite *KeeperTestSuite) TestFilesQuery_Pagination() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.createTestFile("file-001", owner, 1024, 1, nil)
	suite.createTestFile("file-002", owner, 1024, 1, nil)
	suite.createTestFile("file-003", owner, 1024, 1, nil)
	suite.createTestFile("file-004", owner, 1024, 1, nil)
	suite.createTestFile("file-005", owner, 1024, 1, nil)

	resp, err := qs.Files(suite.ctx, &types.QueryFilesRequest{
		Creator:    owner,
		Pagination: &query.PageRequest{Limit: 2},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Files, 2)
	suite.Equal("file-001", resp.Files[0].Fid)
	suite.Equal("file-002", resp.Files[1].Fid)
	suite.Require().NotEmpty(resp.Pagination.NextKey)

	resp2, err := qs.Files(suite.ctx, &types.QueryFilesRequest{
		Creator:    owner,
		Pagination: &query.PageRequest{Key: resp.Pagination.NextKey, Limit: 2},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp2.Files, 2)
	suite.Equal("file-003", resp2.Files[0].Fid)
	suite.Equal("file-004", resp2.Files[1].Fid)
}

func (suite *KeeperTestSuite) TestFilesQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Files(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
