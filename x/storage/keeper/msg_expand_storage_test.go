package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestMsgExpandStorage_Success_ExtendDurationOnly() {
	msgServer, ctx := suite.setupMsgServer()

	// Create test address
	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create initial subscription (30 days, 1TB)
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	// Get the created subscription to verify initial state
	initialSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)
	initialEnd := initialSub.End

	// Expand by adding 15 more days (duration only)
	additionalDays := int64(15)
	additionalBytes := int64(0)

	// Mock payment for expansion
	expansionCost := suite.keeper.GetStorageCost(suite.ctx, (initialBytes+additionalBytes)/1_000_000_000, additionalDays)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", expansionCost, types.ModuleName)

	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       additionalDays,
		Bytes:          additionalBytes,
	}
	expandResponse, err := msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(expandResponse)

	// Verify subscription was updated correctly
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)

	// End time should be extended by 15 days
	expectedEnd := initialEnd.Add(time.Duration(additionalDays) * 24 * time.Hour)
	suite.Require().Equal(expectedEnd, updatedSub.End)
	// Space available should remain the same
	suite.Require().Equal(initialSub.SpaceAvailable, updatedSub.SpaceAvailable)
	// Credits should increase
	suite.Require().True(updatedSub.Credits.GT(initialSub.Credits))

	// Verify response
	suite.Require().Equal(buyResponse.SubscriptionId, expandResponse.SubscriptionId)
	suite.Require().Equal(expectedEnd.Unix(), expandResponse.NewExpiration)
	suite.Require().Equal(initialSub.SpaceAvailable, expandResponse.NewSpaceAvailable)
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Success_IncreaseSpaceOnly() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create initial subscription (30 days, 1TB)
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	initialSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)

	// Expand by adding 2TB more space
	additionalDays := int64(0)
	additionalBytes := int64(2_000_000_000_000)
	additionalGB := additionalBytes / 1_000_000_000

	// Mock payment for expansion
	expansionCost := suite.keeper.GetStorageCost(suite.ctx, additionalGB, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", expansionCost, types.ModuleName)

	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       additionalDays,
		Bytes:          additionalBytes,
	}
	expandResponse, err := msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().NoError(err)

	// Verify subscription was updated correctly
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)

	// Space available should increase by additionalBytes * 3
	expectedSpaceAvailable := initialSub.SpaceAvailable + (additionalBytes * 3)
	suite.Require().Equal(expectedSpaceAvailable, updatedSub.SpaceAvailable)
	// End time should remain the same
	suite.Require().Equal(initialSub.End, updatedSub.End)
	// Credits should increase
	suite.Require().True(updatedSub.Credits.GT(initialSub.Credits))

	// Verify response
	suite.Require().Equal(expectedSpaceAvailable, expandResponse.NewSpaceAvailable)
	suite.Require().Equal(initialSub.End.Unix(), expandResponse.NewExpiration)
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Success_ExtendBothDurationAndSpace() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create initial subscription (30 days, 1TB)
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	initialSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)

	// Expand by adding 15 days and 2TB space
	additionalDays := int64(15)
	additionalBytes := int64(2_000_000_000_000)
	additionalGB := additionalBytes / 1_000_000_000

	// Mock payment for expansion
	e1 := suite.keeper.GetStorageCost(suite.ctx, additionalGB, initialDuration)                           // expansion
	e2 := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000+additionalGB, additionalDays) // duration
	expansionCost := e1.Add(e2)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", expansionCost, types.ModuleName)

	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       additionalDays,
		Bytes:          additionalBytes,
	}
	expandResponse, err := msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().NoError(err)

	// Verify subscription was updated correctly
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)

	// End time should be extended by 15 days
	expectedEnd := initialSub.End.Add(time.Duration(additionalDays) * 24 * time.Hour)
	suite.Require().Equal(expectedEnd, updatedSub.End)

	// Space available should increase by additionalBytes * 3
	expectedSpaceAvailable := initialSub.SpaceAvailable + (additionalBytes * 3)
	suite.Require().Equal(expectedSpaceAvailable, updatedSub.SpaceAvailable)

	// Credits should increase
	expectedCredits := math.NewInt(initialBytes + additionalBytes).MulRaw(3).MulRaw(initialDuration*86400 + additionalDays*86400)
	suite.Require().Equal(expectedCredits, updatedSub.Credits)

	// Verify response
	suite.Require().Equal(expectedEnd.Unix(), expandResponse.NewExpiration)
	suite.Require().Equal(expectedSpaceAvailable, expandResponse.NewSpaceAvailable)
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_SubscriptionNotFound() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: "nonexistent_id",
		Duration:       10,
		Bytes:          1_000_000_000_000,
	}

	_, err := msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "subscription")
	suite.Require().Contains(err.Error(), "not found")
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_SubscriptionExpired() {
	msgServer, ctx := suite.setupMsgServer()
	creator := suite.CreateTestAddress(0)

	// Create a subscription that expires in 1 day
	expiredEnd := suite.ctx.BlockTime().Add(-24 * time.Hour) // Already expired
	subscriptionId := "test_expired_id"

	// Create expired subscription directly
	expiredSub := types.StorageSubscription{
		Id:             subscriptionId,
		Start:          suite.ctx.BlockTime().AddDate(0, -1, 0),
		End:            expiredEnd,
		SpaceAvailable: 3_000_000_000_000, // 1TB * 3
		SpaceUsed:      0,
		Credits:        math.NewInt(100_000_000_000),
	}
	err := suite.keeper.SetStorageSubscription(ctx, creator, expiredSub)
	suite.Require().NoError(err)

	// Try to expand expired subscription
	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: subscriptionId,
		Duration:       10,
		Bytes:          1_000_000_000_000,
	}

	_, err = msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrSubscriptionExpired, err)
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_InvalidCreatorAddress() {
	msgServer, ctx := suite.setupMsgServer()

	expandMsg := &types.MsgExpandStorage{
		Creator:        "invalid_bech32_address",
		SubscriptionId: "some_id",
		Duration:       10,
		Bytes:          1_000_000_000_000,
	}

	_, err := msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid authority address")
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_ZeroDurationBytes() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create a valid subscription first
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	initialSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)

	// Expand subscription by 0 more seconds and 0 more bytes
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", math.ZeroInt(), types.ModuleName)
	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       0,
		Bytes:          0,
	}
	expandResponse, err := msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(expandResponse)

	// Verify subscription was updated correctly
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, buyResponse.SubscriptionId)
	suite.Require().NoError(err)
	suite.Require().Equal(initialSub.End, updatedSub.End)
	suite.Require().Equal(initialSub.SpaceAvailable, updatedSub.SpaceAvailable)
	suite.Require().True(updatedSub.Credits.Equal(initialSub.Credits))

	// Verify response
	suite.Require().Equal(buyResponse.SubscriptionId, expandResponse.SubscriptionId)
	suite.Require().Equal(initialSub.End.Unix(), expandResponse.NewExpiration)
	suite.Require().Equal(initialSub.SpaceAvailable, expandResponse.NewSpaceAvailable)
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_InvalidDuration() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create a valid subscription first
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	// Try to expand with invalid duration (0 days)
	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       -1,
		Bytes:          0,
	}

	_, err = msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to validate expansion request")
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_InvalidBytes() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create a valid subscription first
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	// Try to expand with invalid duration (0 days)
	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       0,
		Bytes:          -1,
	}

	_, err = msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to validate expansion request")
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_InsufficientFunds() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create initial subscription
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	// Mock failed payment for expansion
	additionalDays := int64(15)
	additionalBytes := int64(1_000_000_000_000)
	additionalGB := additionalBytes / 1_000_000_000

	e1 := suite.keeper.GetStorageCost(suite.ctx, additionalGB, initialDuration)                           // expansion
	e2 := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000+additionalGB, additionalDays) // duration
	expansionCost := e1.Add(e2)
	suite.helpers.Bank.MockFailedPayment(suite.bankKeeper, creatorAddr, "uatl", expansionCost, types.ModuleName, "insufficient funds")

	expandMsg := &types.MsgExpandStorage{
		Creator:        creator,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       additionalDays,
		Bytes:          additionalBytes,
	}

	_, err = msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to send payment")
}

func (suite *KeeperTestSuite) TestMsgExpandStorage_Error_DifferentCreator() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	otherUser := suite.CreateTestAddress(1)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock account exists
	suite.helpers.Account.MockAccountExists(suite.authKeeper, creatorAddr)

	// Create initial subscription for creator
	initialDuration := int64(30)
	initialBytes := int64(1_000_000_000_000)
	initialCost := suite.keeper.GetStorageCost(suite.ctx, initialBytes/1_000_000_000, initialDuration)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", initialCost, types.ModuleName)

	buyMsg := &types.MsgBuyStorage{
		Creator:  creator,
		Duration: initialDuration,
		Bytes:    initialBytes,
	}
	buyResponse, err := msgServer.BuyStorage(ctx, buyMsg)
	suite.Require().NoError(err)

	// Try to expand with different creator (should fail when getting subscription)
	expandMsg := &types.MsgExpandStorage{
		Creator:        otherUser,
		SubscriptionId: buyResponse.SubscriptionId,
		Duration:       10,
		Bytes:          1_000_000_000_000,
	}

	_, err = msgServer.ExpandStorage(ctx, expandMsg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not found")
}
