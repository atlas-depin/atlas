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

func (suite *KeeperTestSuite) TestGenesis_SeedsBuiltinFeed() {
	// SetupTest already ran InitGenesis with the default genesis state.
	feed, err := suite.keeper.GetFeed(suite.ctx, types.BuiltinFeedName)
	suite.Require().NoError(err)
	suite.Require().Equal(types.BuiltinFeedName, feed.Name)
	suite.Require().Equal(suite.authorityAddress(), feed.Owner)
	suite.Require().Equal(types.BuiltinFeedInitialData, feed.Data)
	suite.Require().False(feed.LastUpdate.IsZero())
}

func (suite *KeeperTestSuite) TestGenesis_ExportRoundtrip_IncludesFeeds() {
	got, err := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NotNil(got)
	suite.Require().Len(got.Feeds, 1)
	suite.Require().Equal(types.BuiltinFeedName, got.Feeds[0].Name)
}

func (suite *KeeperTestSuite) TestGenesis_CustomFeedsPreserved() {
	genState := types.DefaultGenesis()
	genState.Feeds = append(genState.Feeds, types.Feed{Name: "BTC", Owner: "somebody", Data: "1.000000"})

	err := suite.keeper.InitGenesis(suite.ctx, *genState)
	suite.Require().NoError(err)

	feed, err := suite.keeper.GetFeed(suite.ctx, "BTC")
	suite.Require().NoError(err)
	suite.Require().Equal("somebody", feed.Owner)

	// The built-in feed still exists alongside custom feeds.
	_, err = suite.keeper.GetFeed(suite.ctx, types.BuiltinFeedName)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGenesis_BuiltinFeedOwnerOverridable() {
	genState := types.DefaultGenesis()
	genState.Feeds = append(genState.Feeds, types.Feed{
		Name:  types.BuiltinFeedName,
		Owner: "custom-updater",
		Data:  "1.500000",
	})

	err := suite.keeper.InitGenesis(suite.ctx, *genState)
	suite.Require().NoError(err)

	feed, err := suite.keeper.GetFeed(suite.ctx, types.BuiltinFeedName)
	suite.Require().NoError(err)
	suite.Require().Equal("custom-updater", feed.Owner)
	suite.Require().Equal("1.500000", feed.Data)
}
