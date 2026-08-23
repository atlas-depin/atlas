package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/mock/gomock"

	"atlas/x/storage/types"
)

// DistributeCreditReward converts credits held in the storage credit pool into
// uatl tokens and sends them to either the POL module account or a regular account.
func (suite *KeeperTestSuite) TestDistributeCreditReward_ToPOL() {
	// Seed the credit pool
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	// Module holds 5M uatl; 300k of 1M credits -> 1.5M uatl
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(5_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), types.ModuleName, types.ProtocolOwnedLiquidityAccountName,
			sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_500_000)))).
		Return(nil)

	err := suite.keeper.DistributeCreditReward(suite.ctx, math.NewInt(300_000), types.ProtocolOwnedLiquidityAccountName)
	suite.Require().NoError(err)

	// Credits were burned from the pool
	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(700_000), pool)
}

func (suite *KeeperTestSuite) TestDistributeCreditReward_ToAccount() {
	receiver := suite.CreateTestAddress(1)
	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	// 200k of 1M credits with 5M uatl in the module -> 1M uatl
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(5_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, receiverAddr,
			sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000_000)))).
		Return(nil)

	err = suite.keeper.DistributeCreditReward(suite.ctx, math.NewInt(200_000), receiver)
	suite.Require().NoError(err)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(800_000), pool)
}

func (suite *KeeperTestSuite) TestDistributeCreditReward_EmptyPool() {
	// Genesis leaves the pool at zero
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(5_000_000)))

	err := suite.keeper.DistributeCreditReward(suite.ctx, math.NewInt(100), types.ProtocolOwnedLiquidityAccountName)
	suite.Require().ErrorIs(err, types.ErrRewardPoolEmpty)
}

func (suite *KeeperTestSuite) TestDistributeCreditReward_ZeroCredits() {
	// Zero credits produces an empty coins transfer and leaves the pool untouched.
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(5_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), types.ModuleName, types.ProtocolOwnedLiquidityAccountName,
			sdk.Coins{}).
		Return(nil)

	err := suite.keeper.DistributeCreditReward(suite.ctx, math.ZeroInt(), types.ProtocolOwnedLiquidityAccountName)
	suite.Require().NoError(err)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1_000_000), pool)
}

func (suite *KeeperTestSuite) TestDistributeCreditReward_SendFailure() {
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(5_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), types.ModuleName, types.ProtocolOwnedLiquidityAccountName,
			gomock.Any()).
		Return(sdkerrors.ErrInsufficientFunds)

	err := suite.keeper.DistributeCreditReward(suite.ctx, math.NewInt(300_000), types.ProtocolOwnedLiquidityAccountName)
	suite.Require().Error(err)

	// NOTE: current behavior burns the credits even when the token transfer fails,
	// because the pool is decremented before the send. This test pins that behavior;
	// it is arguably a bug worth revisiting.
	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(700_000), pool)
}
