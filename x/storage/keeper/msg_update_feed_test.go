package keeper_test

import (
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) authorityAddress() string {
	addr, err := suite.addressCodec.BytesToString(suite.keeper.GetAuthority())
	suite.Require().NoError(err)
	return addr
}

func (suite *KeeperTestSuite) TestUpdateFeed_Success() {
	msgServer, ctx := suite.setupMsgServer()

	_, err := msgServer.UpdateFeed(ctx, &types.MsgUpdateFeed{
		Creator: suite.authorityAddress(),
		Name:    types.BuiltinFeedName,
		Data:    "1.234567",
	})
	suite.Require().NoError(err)

	feed, err := suite.keeper.GetFeed(suite.ctx, types.BuiltinFeedName)
	suite.Require().NoError(err)
	suite.Require().Equal("1.234567", feed.Data)
	suite.Require().False(feed.LastUpdate.IsZero(), "last update must be set")
}

func (suite *KeeperTestSuite) TestUpdateFeed_NotOwner() {
	msgServer, ctx := suite.setupMsgServer()

	_, err := msgServer.UpdateFeed(ctx, &types.MsgUpdateFeed{
		Creator: suite.CreateTestAddress(1),
		Name:    types.BuiltinFeedName,
		Data:    "9.9",
	})
	suite.Require().ErrorIs(err, types.ErrFeedNotOwner)

	// Data unchanged.
	feed, err := suite.keeper.GetFeed(suite.ctx, types.BuiltinFeedName)
	suite.Require().NoError(err)
	suite.Require().Equal(types.BuiltinFeedInitialData, feed.Data)
}

func (suite *KeeperTestSuite) TestUpdateFeed_NotFound() {
	msgServer, ctx := suite.setupMsgServer()

	_, err := msgServer.UpdateFeed(ctx, &types.MsgUpdateFeed{
		Creator: suite.authorityAddress(),
		Name:    "DOGE",
		Data:    "1",
	})
	suite.Require().ErrorIs(err, types.ErrFeedNotFound)
}

func (suite *KeeperTestSuite) TestUpdateFeed_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	_, err := msgServer.UpdateFeed(ctx, &types.MsgUpdateFeed{
		Creator: "not-a-bech32-address",
		Name:    types.BuiltinFeedName,
		Data:    "1",
	})
	suite.Require().ErrorContains(err, "invalid creator address")
}
