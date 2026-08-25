package keeper_test

import (
	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	ctx, k := suite.ctx, suite.keeper

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	err := k.InitGenesis(ctx, genesisState)
	suite.Require().NoError(err)
	got, err := k.ExportGenesis(ctx)
	suite.Require().NoError(err)
	suite.Require().NotNil(got)

	suite.Require().EqualExportedValues(genesisState.Params, got.Params)
}

func (suite *KeeperTestSuite) TestGenesis_InitState() {
	// SetupTest already ran InitGenesis with the default genesis state.
	// All counters and the credit pool must start empty.
	pool, err := suite.keeper.StorageCreditPool.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(math.ZeroInt(), pool)

	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), count)

	lastRewards, err := suite.keeper.LastRewardsTime.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), lastRewards)

	params, err := suite.keeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultParams(), params)
}

func (suite *KeeperTestSuite) TestGenesis_SeedsDefaultPrice() {
	// SetupTest already ran InitGenesis with the default genesis state.
	price, err := suite.keeper.Price.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultPrice, price)
}
