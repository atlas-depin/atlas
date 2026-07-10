package keeper_test

import (
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success_SelfReceiver() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Add a provider with enough capacity
	suite.createTestProvider(creator, 10_000_000_000_000_000) // 10× more than the max purchase

	cost := suite.keeper.GetStorageCost(suite.ctx, 1_000, 30) // 1TB / 1GB = 1000 GB
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	response, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	suite.assertSubscription(creator, response.SubscriptionId, msg.Bytes, msg.Duration)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success_MakeDefault() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	suite.createTestProvider(creator, 10_000_000_000_000_000) // 10× more than the max purchase
	cost := suite.keeper.GetStorageCost(suite.ctx, 1_000, 30)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:   creator,
		Duration:  30,
		Bytes:     1_000_000_000_000,
		IsDefault: true,
	}

	response, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	// Verify subscription
	suite.assertSubscription(creator, response.SubscriptionId, msg.Bytes, msg.Duration)

	subsId, _ := suite.keeper.GetDefaultStorageSubscriptionId(ctx, creator)
	suite.Require().Equal(response.SubscriptionId, subsId)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success_LargeValues() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	suite.createTestProvider(creator, 10_000_000_000_000_000)   // 10× more than the max purchase
	cost := suite.keeper.GetStorageCost(suite.ctx, 100000, 730) // 100TB for 2 years
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: 730,                 // 2 years
		Bytes:    100_000_000_000_000, // 100TB
	}

	response, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	suite.assertSubscription(creator, response.SubscriptionId, msg.Bytes, msg.Duration)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InsufficientStorage() {
	msgServer, ctx := suite.setupMsgServer()

	// No provider created — pool capacity is 0
	creator := suite.CreateTestAddress(0)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().ErrorIs(err, types.ErrInsufficientStorage)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InvalidCreatorAddress() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgBuyStorage{
		Creator:  "invalid_bech32_address",
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid authority address")
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InsufficientFunds() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Provider must exist so validation passes before payment
	suite.createTestProvider(creator, 10_000_000_000_000_000) // 10× more than the max purchase
	cost := suite.keeper.GetStorageCost(suite.ctx, 10, 365)
	suite.helpers.Bank.MockFailedPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ModuleName, "insufficient funds")

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: 365,
		Bytes:    10_000_000_000,
	}

	_, err = msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot send tokens")
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_ValidateBuyError_DurationTooShort() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	// Test with invalid duration
	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: 0, // Invalid: less than 1 day
		Bytes:    1_000_000_000_000,
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to validate buy request")
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_ValidateBuyError_BytesTooSmall() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	// Test with bytes less than 1GB
	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: 30,
		Bytes:    500_000_000, // 0.5GB
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to validate buy request")
}

//// ------

func (suite *KeeperTestSuite) assertSubscription(address string, subsId string, bytes int64, duration int64) {
	subscription, err := suite.keeper.GetStorageSubscription(suite.ctx, address, subsId)
	suite.Require().True(err == nil, "subscription should be created for %s", address)

	expectedAvailable := bytes * 3
	expectedCredits := math.NewInt(expectedAvailable).Mul(math.NewInt(duration * 24 * 60 * 60))
	expectedEnd := suite.ctx.BlockTime().Truncate(time.Second).Add(time.Duration(duration) * 24 * time.Hour)

	suite.Require().Equal(expectedAvailable, subscription.SpaceAvailable)
	suite.Require().Equal(int64(0), subscription.SpaceUsed)
	// suite.Require().Equal(int64(0), subscription.ExcessUsed)
	suite.Require().Equal(expectedCredits, subscription.Credits)
	suite.Require().Equal(suite.ctx.BlockTime().Truncate(time.Second), subscription.Start)
	suite.Require().Equal(expectedEnd, subscription.End)
}

func (suite *KeeperTestSuite) createTestSubscription(address string, subscriptionId string, spaceAvailable int64, spaceUsed int64, endTime time.Time) types.StorageSubscription {
	subscription := types.StorageSubscription{
		Id:               subscriptionId,
		Start:            suite.ctx.BlockTime().AddDate(0, -1, 0),
		End:              endTime,
		SpaceAvailable:   spaceAvailable,
		SpaceUsed:        spaceUsed,
		ReplicaSpaceUsed: 0,
		Credits:          math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)), // 1e24 — enough for any test
	}

	// Store the subscription
	err := suite.keeper.SetStorageSubscription(suite.ctx, address, subscription)
	suite.Require().NoError(err)

	return subscription
}
