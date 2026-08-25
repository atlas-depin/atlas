package keeper_test

import (
	"atlas/x/filetree/types"
)

func (suite *KeeperTestSuite) TestPostNode_RootAndNested() {
	msgServer, ctx := suite.setupMsgServer()
	owner := suite.CreateTestAddress(0)

	// Root node
	_, err := msgServer.PostNode(ctx, &types.MsgPostNode{
		Creator:  owner,
		Path:     "docs",
		NodeType: "folder",
	})
	suite.Require().NoError(err)

	node, err := suite.keeper.GetNode(ctx, owner, "docs")
	suite.Require().NoError(err)
	suite.Require().Equal("folder", node.NodeType)

	// Nested node
	_, err = msgServer.PostNode(ctx, &types.MsgPostNode{
		Creator:  owner,
		Path:     "docs/notes/readme",
		NodeType: "file",
		Contents: "hello",
	})
	suite.Require().NoError(err)

	node, err = suite.keeper.GetNode(ctx, owner, "docs/notes/readme")
	suite.Require().NoError(err)
	suite.Require().Equal("file", node.NodeType)
	suite.Require().Equal("hello", node.Contents)
}

func (suite *KeeperTestSuite) TestPostNode_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	_, err := msgServer.PostNode(ctx, &types.MsgPostNode{Creator: "invalid_bech32", Path: "x"})
	suite.Require().ErrorContains(err, "invalid authority address")
}

func (suite *KeeperTestSuite) TestPostNode_OwnerIsolation() {
	msgServer, ctx := suite.setupMsgServer()
	userA := suite.CreateTestAddress(0)
	userB := suite.CreateTestAddress(1)

	// Both users write the same path; namespaces must not collide.
	_, err := msgServer.PostNode(ctx, &types.MsgPostNode{Creator: userA, Path: "x", Contents: "A"})
	suite.Require().NoError(err)
	_, err = msgServer.PostNode(ctx, &types.MsgPostNode{Creator: userB, Path: "x", Contents: "B"})
	suite.Require().NoError(err)

	nodeA, err := suite.keeper.GetNode(ctx, userA, "x")
	suite.Require().NoError(err)
	suite.Require().Equal("A", nodeA.Contents)

	nodeB, err := suite.keeper.GetNode(ctx, userB, "x")
	suite.Require().NoError(err)
	suite.Require().Equal("B", nodeB.Contents)
}
