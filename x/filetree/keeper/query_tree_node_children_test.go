package keeper_test

import (
	"atlas/x/filetree/keeper"
	"atlas/x/filetree/types"
)

// seedTree builds the following structure for owner:
//
//	a
//	a/b
//	a/b/c
//	a/d
func (suite *KeeperTestSuite) seedTree(owner string) {
	for _, path := range []string{"a", "a/b", "a/b/c", "a/d"} {
		suite.Require().NoError(suite.keeper.SetNode(suite.ctx, owner, path, types.TreeNode{
			NodeType: "file",
			Contents: path,
		}))
	}
}

func (suite *KeeperTestSuite) TestTreeNodeChildrenQuery_DirectChildren() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	other := suite.CreateTestAddress(1)
	suite.seedTree(owner)
	suite.seedTree(other)

	// Direct children of "a": nodes whose dir == "a". Deeper nodes and the
	// root itself must be excluded, and other owners' trees must not leak.
	resp, err := qs.TreeNodeChildren(suite.ctx, &types.QueryTreeNodeChildrenRequest{Owner: owner, Path: "a"})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Nodes, 2)

	var contents []string
	for _, node := range resp.Nodes {
		contents = append(contents, node.Contents)
	}
	suite.ElementsMatch([]string{"a/b", "a/d"}, contents)
}

func (suite *KeeperTestSuite) TestTreeNodeChildrenQuery_EmptyPath_ReturnsRoots() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.seedTree(owner)

	resp, err := qs.TreeNodeChildren(suite.ctx, &types.QueryTreeNodeChildrenRequest{Owner: owner, Path: ""})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Nodes, 1)
	suite.Require().Equal("a", resp.Nodes[0].Contents)
}

func (suite *KeeperTestSuite) TestTreeNodeChildrenQuery_DeepPath() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.seedTree(owner)

	resp, err := qs.TreeNodeChildren(suite.ctx, &types.QueryTreeNodeChildrenRequest{Owner: owner, Path: "a/b"})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Nodes, 1)
	suite.Require().Equal("a/b/c", resp.Nodes[0].Contents)
}

func (suite *KeeperTestSuite) TestTreeNodeChildrenQuery_NoChildren() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.seedTree(owner)

	resp, err := qs.TreeNodeChildren(suite.ctx, &types.QueryTreeNodeChildrenRequest{Owner: owner, Path: "missing"})
	suite.Require().NoError(err)
	suite.Require().Empty(resp.Nodes)
}

func (suite *KeeperTestSuite) TestTreeNodeChildrenQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.TreeNodeChildren(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
