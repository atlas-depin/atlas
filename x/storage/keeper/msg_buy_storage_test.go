package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"nebulix/x/storage/types"
)

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success_SelfReceiver() {
	msgServer, ctx := suite.setupMsgServer()

	// Create test address using helper
	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists using helper
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Calculate cost and mock payment using helper
	cost := math.NewInt(12000000)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "unblx", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: "",
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	response, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	// Verify subscription
	suite.assertSubscription(creator, response.SubscriptionId, msg.Bytes, msg.Duration)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success_NewAccount() {
	msgServer, ctx := suite.setupMsgServer()

	// Create test addresses
	creator := suite.CreateTestAddress(0)
	receiver := suite.CreateTestAddress(1)

	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)
	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	// Mock: Receiver doesn't exist, will be created
	suite.helpers.Account.MockAccountDoesNotExist(suite.authKeeper, receiverAddr)

	// Create a mock account object for NewAccountWithAddress
	mockAccount := authtypes.NewBaseAccountWithAddress(receiverAddr)
	suite.helpers.Account.MockNewAccount(suite.authKeeper, receiverAddr, mockAccount)

	// Calculate and mock payment
	cost := math.NewInt(12000000)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "unblx", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: receiver,
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	response, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	// Verify subscription
	suite.assertSubscription(receiver, response.SubscriptionId, msg.Bytes, msg.Duration)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success_MakeDefault() {
	msgServer, ctx := suite.setupMsgServer()

	// Create test address using helper
	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists using helper
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Calculate cost and mock payment using helper
	cost := math.NewInt(12000000)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "unblx", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: "",
		Duration: 30,
		Bytes:    1_000_000_000_000,
		Default:  true,
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

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Calculate and mock payment for large storage
	cost := suite.keeper.GetStorageCost(suite.ctx, 100000, 730) // 100TB for 2 years
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "unblx", cost, types.ModuleName)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: "",
		Duration: 730,                 // 2 years
		Bytes:    100_000_000_000_000, // 100TB
	}

	response, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	suite.assertSubscription(creator, response.SubscriptionId, msg.Bytes, msg.Duration)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InvalidCreatorAddress() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgBuyStorage{
		Creator:  "invalid_bech32_address",
		Receiver: "",
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot parse creator address")
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InvalidReceiverAddress() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: "invalid_bech32_address",
		Duration: 30,
		Bytes:    1_000_000_000_000,
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "for address is not a proper bech32")
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InsufficientFunds() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Mock failed payment
	cost := suite.keeper.GetStorageCost(suite.ctx, 10, 365)
	suite.helpers.Bank.MockFailedPayment(suite.bankKeeper, creatorAddr, "unblx", cost, types.ModuleName, "insufficient funds")

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: "",
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
		Receiver: "",
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
		Receiver: "",
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
		Credits:          math.NewInt(100_000_000_000), // Large initial credits
	}

	// Store the subscription
	err := suite.keeper.SetStorageSubscription(suite.ctx, address, subscription)
	suite.Require().NoError(err)

	return subscription
}

//////////////////////
