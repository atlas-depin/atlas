package keeper_test

import (
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) authorityAddress() string {
	addr, err := suite.addressCodec.BytesToString(suite.keeper.GetAuthority())
	suite.Require().NoError(err)
	return addr
}

func (suite *KeeperTestSuite) TestUpdatePrice_Success() {
	msgServer, ctx := suite.setupMsgServer()

	updater := suite.CreateTestAddress(0)
	params := types.DefaultParams()
	params.PriceUpdater = updater
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	_, err := msgServer.UpdatePrice(ctx, &types.MsgUpdatePrice{
		Creator: updater,
		Price:   "1.234567",
	})
	suite.Require().NoError(err)

	price, err := suite.keeper.Price.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal("1.234567", price)
}

func (suite *KeeperTestSuite) TestUpdatePrice_NotUpdater() {
	msgServer, ctx := suite.setupMsgServer()

	updater := suite.CreateTestAddress(0)
	params := types.DefaultParams()
	params.PriceUpdater = updater
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	_, err := msgServer.UpdatePrice(ctx, &types.MsgUpdatePrice{
		Creator: suite.CreateTestAddress(1),
		Price:   "9.9",
	})
	suite.Require().ErrorIs(err, types.ErrInvalidPriceUpdater)

	// Price unchanged.
	price, err := suite.keeper.Price.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultPrice, price)
}

func (suite *KeeperTestSuite) TestUpdatePrice_NoUpdaterConfigured() {
	msgServer, ctx := suite.setupMsgServer()

	// Default params: no updater configured, so even the authority cannot update.
	_, err := msgServer.UpdatePrice(ctx, &types.MsgUpdatePrice{
		Creator: suite.authorityAddress(),
		Price:   "1.0",
	})
	suite.Require().ErrorIs(err, types.ErrInvalidPriceUpdater)
}

func (suite *KeeperTestSuite) TestUpdatePrice_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	_, err := msgServer.UpdatePrice(ctx, &types.MsgUpdatePrice{
		Creator: "not-a-bech32-address",
		Price:   "1.0",
	})
	suite.Require().ErrorContains(err, "invalid creator address")
}
