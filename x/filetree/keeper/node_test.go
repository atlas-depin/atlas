package keeper_test

import (
	"atlas/x/filetree/types"

	"cosmossdk.io/collections"
)

func (suite *KeeperTestSuite) TestSetGetNode_RootAndNested() {
	owner := suite.CreateTestAddress(0)

	suite.Require().NoError(suite.keeper.SetNode(suite.ctx, owner, "docs", types.TreeNode{NodeType: "folder"}))
	suite.Require().NoError(suite.keeper.SetNode(suite.ctx, owner, "docs/notes", types.TreeNode{
		NodeType: "file",
		Contents: "hello",
	}))

	root, err := suite.keeper.GetNode(suite.ctx, owner, "docs")
	suite.Require().NoError(err)
	suite.Require().Equal("folder", root.NodeType)

	nested, err := suite.keeper.GetNode(suite.ctx, owner, "docs/notes")
	suite.Require().NoError(err)
	suite.Require().Equal("hello", nested.Contents)
}

func (suite *KeeperTestSuite) TestGetNode_NotFound() {
	_, err := suite.keeper.GetNode(suite.ctx, suite.CreateTestAddress(0), "ghost")
	suite.Require().ErrorIs(err, collections.ErrNotFound)
}

func (suite *KeeperTestSuite) TestRemoveNode_NonRecursive() {
	owner := suite.CreateTestAddress(0)

	for _, path := range []string{"s", "s/t1", "s/t1/x", "s/t2"} {
		suite.Require().NoError(suite.keeper.SetNode(suite.ctx, owner, path, types.TreeNode{}))
	}

	err := suite.keeper.RemoveNode(suite.ctx, owner, "s", false)
	suite.Require().NoError(err)

	// Self and direct children are removed...
	for _, path := range []string{"s", "s/t1", "s/t2"} {
		_, err := suite.keeper.GetNode(suite.ctx, owner, path)
		suite.Require().ErrorIs(err, collections.ErrNotFound, "path %q should be deleted", path)
	}

	// ...but deeper descendants survive a non-recursive delete.
	_, err = suite.keeper.GetNode(suite.ctx, owner, "s/t1/x")
	suite.Require().NoError(err, "descendant should survive non-recursive delete")
}

func (suite *KeeperTestSuite) TestRemoveNode_Recursive() {
	owner := suite.CreateTestAddress(0)

	for _, path := range []string{"s", "s/t1", "s/t1/x", "s/t2"} {
		suite.Require().NoError(suite.keeper.SetNode(suite.ctx, owner, path, types.TreeNode{}))
	}

	err := suite.keeper.RemoveNode(suite.ctx, owner, "s", true)
	suite.Require().NoError(err)

	for _, path := range []string{"s", "s/t1", "s/t1/x", "s/t2"} {
		_, err := suite.keeper.GetNode(suite.ctx, owner, path)
		suite.Require().ErrorIs(err, collections.ErrNotFound, "path %q should be deleted", path)
	}
}

func (suite *KeeperTestSuite) TestRemoveNode_NonExistent_NoError() {
	err := suite.keeper.RemoveNode(suite.ctx, suite.CreateTestAddress(0), "ghost", true)
	suite.Require().NoError(err)
}
