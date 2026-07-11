package keeper_test

import (
	"fmt"

	"atlas/x/storage/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
)

func (suite *KeeperTestSuite) TestDeleteProvider_Success() {
	msgServer, ctx := suite.setupMsgServer()
	params, _ := suite.keeper.Params.Get(ctx)

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Register a provider first
	deposit := math.NewInt(params.ProviderDeposit)
	suite.helpers.Bank.MockSuccessfulPayment(suite.bankKeeper, creatorAddr, "uatl", deposit, types.ProviderDepositAccountName)
	regMsg := &types.MsgRegisterProvider{
		Creator:  creator,
		Hostname: "test-provider",
		Capacity: 1_000_000_000_000,
	}
	_, err = msgServer.RegisterProvider(ctx, regMsg)
	suite.Require().NoError(err)

	// Mock the refund from the module back to the provider
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ProviderDepositAccountName,
			creatorAddr,
			sdk.NewCoins(sdk.NewCoin("uatl", deposit)),
		).
		Return(nil)

	// Delete the provider
	delMsg := &types.MsgDeleteProvider{
		Creator: creator,
	}
	_, err = msgServer.DeleteProvider(ctx, delMsg)
	suite.Require().NoError(err)

	// Verify provider is gone
	_, err = suite.keeper.GetProvider(ctx, creator)
	suite.Require().ErrorIs(err, types.ErrProviderNotFound)
}

func (suite *KeeperTestSuite) TestDeleteProvider_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgDeleteProvider{
		Creator: "invalid_bech32",
	}
	_, err := msgServer.DeleteProvider(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid authority address")
}

func (suite *KeeperTestSuite) TestDeleteProvider_NotFound() {
	msgServer, ctx := suite.setupMsgServer()
	params, _ := suite.keeper.Params.Get(ctx)

	creator := suite.CreateTestAddress(0)
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	suite.Require().NoError(err)

	// Mock refund but provider was never registered — bank send fails
	deposit := math.NewInt(params.ProviderDeposit)
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ProviderDepositAccountName,
			creatorAddr,
			sdk.NewCoins(sdk.NewCoin("uatl", deposit)),
		).
		Return(fmt.Errorf("provider not registered"))

	delMsg := &types.MsgDeleteProvider{
		Creator: creator,
	}
	_, err = msgServer.DeleteProvider(ctx, delMsg)
	suite.Require().Error(err)
}
