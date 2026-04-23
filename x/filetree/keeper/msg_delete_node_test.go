package keeper_test

import (
	"atlas/x/filetree/types"

	"cosmossdk.io/collections"
)

func (suite *KeeperTestSuite) TestRemoveNodeRecursively() {
	ctx := suite.ctx
	owner := "user1"

	// Create nested structure:
	// s/Test1
	// s/Test1/Test2
	// s/Test1/Test3
	// s/Test1/Test2/Test4
	testPaths := []string{
		"s/Test1",
		"s/Test1/Test2",
		"s/Test1/Test3",
		"s/Test1/Test2/Test4",
	}
	for _, path := range testPaths {
		err := suite.keeper.SetNode(ctx, owner, path, types.TreeNode{ /* some data */ })
		suite.Require().NoError(err)
	}

	// Also create an unrelated key to ensure it's not deleted
	unrelatedPath := "other/path"
	err := suite.keeper.SetNode(ctx, owner, unrelatedPath, types.TreeNode{})
	suite.Require().NoError(err)

	// Delete s/Test1 recursively
	err = suite.keeper.RemoveNode(ctx, owner, "s/Test1", true)
	suite.Require().NoError(err)

	// Verify all s/Test1 descendants are gone
	for _, path := range testPaths {
		_, err := suite.keeper.GetNode(ctx, owner, path)
		suite.Require().ErrorIs(err, collections.ErrNotFound, "path %q should be deleted", path)
	}

	// Verify unrelated key still exists
	_, err = suite.keeper.GetNode(ctx, owner, unrelatedPath)
	suite.Require().NoError(err, "unrelated path should not be deleted")
}
