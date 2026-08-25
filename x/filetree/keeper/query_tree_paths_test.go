package keeper_test

import (
	"atlas/x/filetree/keeper"
	"atlas/x/filetree/types"
)

func (suite *KeeperTestSuite) TestFileTreePathsQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	other := suite.CreateTestAddress(1)
	suite.seedTree(owner)
	suite.seedTree(other)

	resp, err := qs.FileTreePaths(suite.ctx, &types.QueryFileTreePathsRequest{Owner: owner, Basepath: "a"})
	suite.Require().NoError(err)
	suite.ElementsMatch([]string{"b", "d"}, resp.Paths)
}

func (suite *KeeperTestSuite) TestFileTreePathsQuery_EmptyBasepath_ReturnsRoots() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.seedTree(owner)

	resp, err := qs.FileTreePaths(suite.ctx, &types.QueryFileTreePathsRequest{Owner: owner, Basepath: ""})
	suite.Require().NoError(err)
	suite.ElementsMatch([]string{"a"}, resp.Paths)
}

func (suite *KeeperTestSuite) TestFileTreePathsQuery_MissingBasepath() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.seedTree(owner)

	resp, err := qs.FileTreePaths(suite.ctx, &types.QueryFileTreePathsRequest{Owner: owner, Basepath: "zzz"})
	suite.Require().NoError(err)
	suite.Require().Empty(resp.Paths)
}

func (suite *KeeperTestSuite) TestFileTreePathsQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.FileTreePaths(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
