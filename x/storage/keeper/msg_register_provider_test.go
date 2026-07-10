package keeper_test

import (
	"atlas/x/storage/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestRegisterProvider_Success() {
	msgServer, ctx := suite.setupMsgServer()
	params, _ := suite.keeper.Params.Get(ctx)

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	cost := math.NewInt(params.ProviderDeposit)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ProviderDepositAccountName)

	msg := &types.MsgRegisterProvider{
		Creator:  creator,
		Hostname: "test-provider.example.com",
		Capacity: 1_000_000_000_000,
	}

	_, err = msgServer.RegisterProvider(ctx, msg)
	suite.Require().NoError(err)

	// Verify provider was stored
	provider, err := suite.keeper.GetProvider(ctx, creator)
	suite.Require().NoError(err)
	suite.Equal(creator, provider.Address)
	suite.Equal("test-provider.example.com", provider.Hostname)
	suite.Equal(int64(1_000_000_000_000), provider.SpaceAvailable)
	suite.Equal(int64(0), provider.SpaceUsed)
	suite.Equal(int64(0), provider.CreditDelta)
}

func (suite *KeeperTestSuite) TestRegisterProvider_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgRegisterProvider{
		Creator: "invalid_bech32",
	}
	_, err := msgServer.RegisterProvider(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid creator address")
}

func (suite *KeeperTestSuite) TestRegisterProvider_InsufficientDeposit() {
	msgServer, ctx := suite.setupMsgServer()
	params, _ := suite.keeper.Params.Get(ctx)

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	cost := math.NewInt(params.ProviderDeposit)
	suite.helpers.Bank.MockFailedPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ProviderDepositAccountName, "insufficient funds")

	msg := &types.MsgRegisterProvider{
		Creator:  creator,
		Hostname: "test-provider.example.com",
		Capacity: 1_000_000_000_000,
	}

	_, err = msgServer.RegisterProvider(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "insufficient")
}

func (suite *KeeperTestSuite) TestRegisterProvider_AlreadyExists() {
	msgServer, ctx := suite.setupMsgServer()
	params, _ := suite.keeper.Params.Get(ctx)

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	cost := math.NewInt(params.ProviderDeposit)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ProviderDepositAccountName)

	// Register once
	msg := &types.MsgRegisterProvider{
		Creator:  creator,
		Hostname: "first-time",
		Capacity: 1_000_000_000_000,
	}
	_, err = msgServer.RegisterProvider(ctx, msg)
	suite.Require().NoError(err)

	// Register again — should fail
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", cost, types.ProviderDepositAccountName)
	msg2 := &types.MsgRegisterProvider{
		Creator:  creator,
		Hostname: "second-time",
		Capacity: 2_000_000_000_000,
	}
	_, err = msgServer.RegisterProvider(ctx, msg2)
	suite.Require().ErrorIs(err, types.ErrProviderAlreadyExists)
}
