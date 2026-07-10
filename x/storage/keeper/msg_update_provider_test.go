package keeper_test

import (
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestUpdateProvider_Success_Hostname() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	suite.createTestProvider(creator, 1_000_000_000_000)

	msg := &types.MsgUpdateProvider{
		Creator:  creator,
		Hostname: "new-hostname.example.com",
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().NoError(err)

	provider, err := suite.keeper.GetProvider(ctx, creator)
	suite.Require().NoError(err)
	suite.Equal("new-hostname.example.com", provider.Hostname)
	suite.Equal(int64(1_000_000_000_000), provider.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestUpdateProvider_Success_Capacity() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	suite.createTestProvider(creator, 1_000_000_000_000)

	msg := &types.MsgUpdateProvider{
		Creator:  creator,
		Capacity: 2_000_000_000_000,
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().NoError(err)

	provider, err := suite.keeper.GetProvider(ctx, creator)
	suite.Require().NoError(err)
	suite.Equal(int64(2_000_000_000_000), provider.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestUpdateProvider_Success_Both() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	suite.createTestProvider(creator, 1_000_000_000_000)

	msg := &types.MsgUpdateProvider{
		Creator:  creator,
		Hostname: "new-host.example.com",
		Capacity: 5_000_000_000_000,
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().NoError(err)

	provider, err := suite.keeper.GetProvider(ctx, creator)
	suite.Require().NoError(err)
	suite.Equal("new-host.example.com", provider.Hostname)
	suite.Equal(int64(5_000_000_000_000), provider.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestUpdateProvider_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgUpdateProvider{
		Creator: "invalid_bech32",
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid authority address")
}

func (suite *KeeperTestSuite) TestUpdateProvider_NotFound() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	msg := &types.MsgUpdateProvider{
		Creator:  creator,
		Hostname: "anything",
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "provider not found")
}

func (suite *KeeperTestSuite) TestUpdateProvider_CapacityTooLow() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	// Create provider with 100GB capacity and 80GB already in use
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        creator,
		Hostname:       "test",
		SpaceAvailable: 100_000_000_000,
		SpaceUsed:      80_000_000_000,
	}))

	msg := &types.MsgUpdateProvider{
		Creator:  creator,
		Capacity: 50_000_000_000, // below current usage
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().ErrorIs(err, types.ErrProviderCapacityTooLow)
}

func (suite *KeeperTestSuite) TestUpdateProvider_Success_NoChanges() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	suite.createTestProvider(creator, 1_000_000_000_000)

	msg := &types.MsgUpdateProvider{
		Creator:  creator,
		Hostname: "",
		Capacity: 0,
	}
	_, err := msgServer.UpdateProvider(ctx, msg)
	suite.Require().NoError(err)

	// Provider should be unchanged
	provider, err := suite.keeper.GetProvider(ctx, creator)
	suite.Require().NoError(err)
	suite.Equal(int64(1_000_000_000_000), provider.SpaceAvailable)
	suite.Equal("test-provider", provider.Hostname) // default from createTestProvider
}
