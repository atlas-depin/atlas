package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/mock/gomock"

	"atlas/x/storage/types"
)

// RewardProviders distributes storage tokens to providers proportional to the
// credits they accrued since the last reward round, then burns those credits.
func (suite *KeeperTestSuite) TestRewardProviders_FirstCall_InitializesTime() {
	// Genesis sets LastRewardsTime to 0, so the first call only records the time.
	lastRewardsTime, err := suite.keeper.LastRewardsTime.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), lastRewardsTime)

	err = suite.keeper.RewardProviders(suite.ctx)
	suite.Require().NoError(err)

	// No bank calls should have been made (nothing mocked, so any call would fail).
	lastRewardsTime, err = suite.keeper.LastRewardsTime.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.ctx.BlockTime().Truncate(time.Second).Unix(), lastRewardsTime)
}

func (suite *KeeperTestSuite) TestRewardProviders_DistributesTokens() {
	providerAddr := suite.CreateTestAddress(1)
	providerAcc, err := sdk.AccAddressFromBech32(providerAddr)
	suite.Require().NoError(err)

	// Provider holds 100 bytes for the full 3600s since the last round.
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:   providerAddr,
		Hostname:  "provider-1",
		SpaceUsed: 100,
	}))

	t1 := suite.ctx.BlockTime().Truncate(time.Second).Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-3600))
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	// credits = 3600 * 100 = 360_000; tokens = 10M * 360_000 / 1M = 3.6M
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(10_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, providerAcc,
			sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(3_600_000)))).
		Return(nil)

	err = suite.keeper.RewardProviders(suite.ctx)
	suite.Require().NoError(err)

	// Credit delta reset, credits burned, round time advanced.
	provider, err := suite.keeper.GetProvider(suite.ctx, providerAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), provider.CreditDelta)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(640_000), pool)

	lastRewardsTime, err := suite.keeper.LastRewardsTime.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(t1, lastRewardsTime)
}

func (suite *KeeperTestSuite) TestRewardProviders_NegativeDelta_CarriedForward() {
	providerAddr := suite.CreateTestAddress(1)
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:     providerAddr,
		Hostname:    "provider-1",
		CreditDelta: -500,
	}))

	t1 := suite.ctx.BlockTime().Truncate(time.Second).Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-3600))
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(10_000_000)))

	err := suite.keeper.RewardProviders(suite.ctx)
	suite.Require().NoError(err)

	// Negative credits are not paid out; they persist for the next round.
	provider, err := suite.keeper.GetProvider(suite.ctx, providerAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(-500), provider.CreditDelta)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1_000_000), pool, "no credits should be burned")
}

func (suite *KeeperTestSuite) TestRewardProviders_EmptyPool() {
	t1 := suite.ctx.BlockTime().Truncate(time.Second).Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-3600))

	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(10_000_000)))

	err := suite.keeper.RewardProviders(suite.ctx)
	suite.Require().ErrorIs(err, types.ErrRewardPoolEmpty)
}

func (suite *KeeperTestSuite) TestRewardProviders_SendFailure_Tolerated() {
	providerAddr := suite.CreateTestAddress(1)
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:   providerAddr,
		Hostname:  "provider-1",
		SpaceUsed: 100,
	}))

	t1 := suite.ctx.BlockTime().Truncate(time.Second).Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-3600))
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))

	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(10_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, gomock.Any(), gomock.Any()).
		Return(sdkerrors.ErrInsufficientFunds)

	err := suite.keeper.RewardProviders(suite.ctx)
	suite.Require().NoError(err, "failed sends are logged, not propagated")

	// Nothing was deducted: the provider's delta was not reset and the pool is intact.
	provider, err := suite.keeper.GetProvider(suite.ctx, providerAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), provider.CreditDelta)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(1_000_000), pool)
}

// RewardStrays distributes credits accrued by STRAY files to the POL account.
func (suite *KeeperTestSuite) TestRewardStrays_RecentStray() {
	// Stray file started 50s ago; last rewards time is 100s ago -> t0 < start
	// d_c = (t1 - start) * size = 50 * 100 = 5000
	t1 := suite.ctx.BlockTime().Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-100))
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(10_000_000)))

	owner := suite.CreateTestAddress(0)
	stray := suite.createTestFile("stray-recent", owner, 100, 1, nil)
	stray.Start = t1 - 50
	stray.Status = "STRAY"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, stray))

	// tokens = 10M * 5000 / 10M = 5000
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(10_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), types.ModuleName, types.ProtocolOwnedLiquidityAccountName,
			sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(5_000)))).
		Return(nil)

	err := suite.keeper.RewardStrays(suite.ctx)
	suite.Require().NoError(err)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(9_995_000), pool)
}

func (suite *KeeperTestSuite) TestRewardStrays_OldStray() {
	// Stray file started before the last rewards time -> d_c = (t1 - t0) * size
	// = 100 * 200 = 20000
	t1 := suite.ctx.BlockTime().Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-100))
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(10_000_000)))

	owner := suite.CreateTestAddress(0)
	stray := suite.createTestFile("stray-old", owner, 200, 1, nil)
	stray.Start = t1 - 150
	stray.Status = "STRAY"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, stray))

	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(10_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), types.ModuleName, types.ProtocolOwnedLiquidityAccountName,
			sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(20_000)))).
		Return(nil)

	err := suite.keeper.RewardStrays(suite.ctx)
	suite.Require().NoError(err)

	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(9_980_000), pool)
}

func (suite *KeeperTestSuite) TestRewardStrays_NoStrays() {
	t1 := suite.ctx.BlockTime().Unix()
	suite.Require().NoError(suite.keeper.LastRewardsTime.Set(suite.ctx, t1-100))
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(10_000_000)))

	// No bank calls expected; any call would fail the mock.
	err := suite.keeper.RewardStrays(suite.ctx)
	suite.Require().NoError(err)
}
