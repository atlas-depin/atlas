package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestFileQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	file := suite.createTestFile("file-1", owner, 1024, 1, []string{suite.CreateTestAddress(1)})

	resp, err := qs.File(suite.ctx, &types.QueryFileRequest{Fid: "file-1"})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.File)
	suite.Equal(file, *resp.File)
}

func (suite *KeeperTestSuite) TestFileQuery_NotFound() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.File(suite.ctx, &types.QueryFileRequest{Fid: "does-not-exist"})
	suite.Require().Error(err)
	suite.Require().Equal(codes.NotFound, status.Code(err))
	suite.Require().ErrorContains(err, "file not found")
	suite.Require().Nil(resp)
}

func (suite *KeeperTestSuite) TestFileQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.File(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
