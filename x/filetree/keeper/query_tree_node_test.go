package keeper_test

import (
	"atlas/x/filetree/keeper"
	"atlas/x/filetree/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestTreeNodeQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetNode(suite.ctx, owner, "docs/notes", types.TreeNode{
		NodeType: "file",
		Contents: "hello",
	}))

	resp, err := qs.TreeNode(suite.ctx, &types.QueryTreeNodeRequest{Owner: owner, Path: "docs/notes"})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Node)
	suite.Require().Equal("file", resp.Node.NodeType)
	suite.Require().Equal("hello", resp.Node.Contents)
}

func (suite *KeeperTestSuite) TestTreeNodeQuery_NotFound() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.TreeNode(suite.ctx, &types.QueryTreeNodeRequest{
		Owner: suite.CreateTestAddress(0),
		Path:  "does-not-exist",
	})
	suite.Require().Error(err)
	suite.Require().Equal(codes.NotFound, status.Code(err))
	suite.Require().ErrorContains(err, "node not found")
	suite.Require().Nil(resp)
}

func (suite *KeeperTestSuite) TestTreeNodeQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.TreeNode(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
