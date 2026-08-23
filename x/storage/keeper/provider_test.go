package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestSetNewProvider_Success() {
	addr := suite.CreateTestAddress(0)
	provider := types.Provider{Address: addr, Hostname: "p1", SpaceAvailable: 1000}

	suite.Require().NoError(suite.keeper.SetNewProvider(suite.ctx, provider))

	stored, err := suite.keeper.GetProvider(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().Equal(provider, stored)
}

func (suite *KeeperTestSuite) TestSetNewProvider_AlreadyExists() {
	addr := suite.CreateTestAddress(0)
	suite.createTestProvider(addr, 1000)

	err := suite.keeper.SetNewProvider(suite.ctx, types.Provider{Address: addr, Hostname: "p2"})
	suite.Require().ErrorIs(err, types.ErrProviderAlreadyExists)
}

func (suite *KeeperTestSuite) TestRemoveProvider() {
	addr := suite.CreateTestAddress(0)
	suite.createTestProvider(addr, 1000)

	suite.Require().NoError(suite.keeper.RemoveProvider(suite.ctx, addr))

	_, err := suite.keeper.GetProvider(suite.ctx, addr)
	suite.Require().ErrorIs(err, types.ErrProviderNotFound)
}

func (suite *KeeperTestSuite) TestIncrementProviderSpaceUsed() {
	addr := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        addr,
		Hostname:       "p1",
		SpaceAvailable: 1000,
		SpaceUsed:      100,
	}))

	// d_s = 0 means the POL distribution gets zero credits (empty coins send).
	suite.Require().NoError(suite.keeper.StorageCreditPool.Set(suite.ctx, math.NewInt(1_000_000)))
	suite.bankKeeper.EXPECT().
		GetBalance(gomock.Any(), types.ModuleAddress, "uatl").
		Return(sdk.NewCoin("uatl", math.NewInt(5_000_000)))
	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), types.ModuleName, types.ProtocolOwnedLiquidityAccountName,
			sdk.Coins{}).
		Return(nil)

	err := suite.keeper.IncrementProviderSpaceUsed(suite.ctx, addr, 50, 1, 0)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetProvider(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(150), stored.SpaceUsed)
	suite.Require().Equal(int64(-50), stored.CreditDelta) // +bytes * -d_r
}

func (suite *KeeperTestSuite) TestIncrementProviderSpaceUsed_Full() {
	addr := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        addr,
		Hostname:       "p1",
		SpaceAvailable: 1000,
		SpaceUsed:      950,
	}))

	err := suite.keeper.IncrementProviderSpaceUsed(suite.ctx, addr, 100, 1, 0)
	suite.Require().ErrorIs(err, types.ErrProviderFull)
}

func (suite *KeeperTestSuite) TestDecrementProviderSpaceUsed() {
	addr := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:     addr,
		Hostname:    "p1",
		SpaceUsed:   100,
		CreditDelta: 10,
	}))

	err := suite.keeper.DecrementProviderSpaceUsed(suite.ctx, addr, 40, 5)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetProvider(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(60), stored.SpaceUsed)
	suite.Require().Equal(int64(210), stored.CreditDelta) // 10 + 40*5
}

func (suite *KeeperTestSuite) TestDecrementProviderSpaceUsed_ClampsAtZero() {
	addr := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:   addr,
		Hostname:  "p1",
		SpaceUsed: 5,
	}))

	err := suite.keeper.DecrementProviderSpaceUsed(suite.ctx, addr, 10, 3)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetProvider(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), stored.SpaceUsed)
	// Only the actually-freed 5 bytes accrue the delta.
	suite.Require().Equal(int64(15), stored.CreditDelta) // 5 * 3
}
