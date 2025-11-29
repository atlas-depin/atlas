package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"nebulix/x/storage/types"
)

func (suite *KeeperTestSuite) TestMsgBuyStorage_Success() {
	msgServer, ctx := suite.setupMsgServer()

	// Create test addresses
	testAddresses := suite.createTestAddresses(2)
	creator := testAddresses[0]
	receiver := testAddresses[1]

	// [NOTE]: not `sdk.AccAddressFromBech32`
	// This []bytes argument forces usage of the dependency-injected address codec from keeper
	creatorAddr, err := suite.addressCodec.StringToBytes(creator)
	suite.Require().NoError(err)

	// Configure loose mock expectations
	suite.authKeeper.EXPECT().
		NewAccountWithAddress(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Configure mock expectations
	suite.authKeeper.EXPECT().
		HasAccount(gomock.Any(), gomock.Any()).
		Return(false).
		Times(1)

	suite.authKeeper.EXPECT().
		SetAccount(gomock.Any(), gomock.Any()).
		Times(1)

	expectedPayment := sdk.NewCoin("unblx", math.NewInt(300000000))
	suite.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), creatorAddr, types.ModuleName, sdk.NewCoins(expectedPayment)).
		Return(nil).
		Times(1)

	// Execute
	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: receiver,
		Duration: 30,
		Bytes:    1_000_000_000,
	}

	response, err := msgServer.BuyStorage(ctx, msg)

	// Verify
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	// Check state
	payInfo, err := suite.keeper.PaymentStore.Get(suite.ctx, receiver)
	suite.Require().NoError(err)
	suite.Require().Equal(msg.Bytes*3, payInfo.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_InvalidAddress() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgBuyStorage{
		Creator:  "invalid_address",
		Receiver: "invalid_address",
		Duration: 30,
		Bytes:    1_000_000_000,
	}

	_, err := msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "decoding bech32 failed")
}

func (suite *KeeperTestSuite) TestMsgBuyStorage_ExistingSubscription() {
	msgServer, ctx := suite.setupMsgServer()

	testAddresses := suite.createTestAddresses(1)
	creator := testAddresses[0]

	// Set up existing storage
	payInfo := types.StoragePaymentInfo{
		Address:        creator,
		Start:          suite.ctx.BlockTime(),
		End:            suite.ctx.BlockTime().AddDate(0, 1, 0),
		SpaceAvailable: 3_000_000_000,
		SpaceUsed:      0,
	}
	err := suite.keeper.PaymentStore.Set(suite.ctx, creator, payInfo)
	suite.Require().NoError(err)

	// Mock expectations
	suite.authKeeper.EXPECT().
		HasAccount(gomock.Any(), gomock.Any()).
		Return(true).
		Times(1)

	msg := &types.MsgBuyStorage{
		Creator:  creator,
		Receiver: creator,
		Duration: 30,
		Bytes:    1_000_000_000,
	}

	_, err = msgServer.BuyStorage(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "account has an active storage subscription")
}
