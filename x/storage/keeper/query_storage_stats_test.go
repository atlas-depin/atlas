package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestStorageStatsQuery_Empty() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.StorageStats(suite.ctx, &types.QueryStorageStatsRequest{})
	suite.Require().NoError(err)
	suite.Equal(&types.QueryStorageStatsResponse{}, resp)
}

func (suite *KeeperTestSuite) TestStorageStatsQuery_WithData() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)

	// Active subscription
	_, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, types.StorageSubscription{
		Start:            suite.ctx.BlockTime(),
		End:              suite.ctx.BlockTime().Add(24 * time.Hour),
		SpaceAvailable:   300,
		ReplicaSpaceUsed: 60,
		Credits:          math.NewInt(1_000),
	})
	suite.Require().NoError(err)

	// Expired subscription — should not be counted
	_, err = suite.keeper.SetNewStorageSubscription(suite.ctx, owner, types.StorageSubscription{
		Start:          suite.ctx.BlockTime().Add(-48 * time.Hour),
		End:            suite.ctx.BlockTime().Add(-24 * time.Hour),
		SpaceAvailable: 999,
		Credits:        math.NewInt(1_000),
	})
	suite.Require().NoError(err)

	// Providers
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        suite.CreateTestAddress(1),
		Hostname:       "provider-1",
		SpaceAvailable: 1_000,
		SpaceUsed:      90,
	}))
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:        suite.CreateTestAddress(2),
		Hostname:       "provider-2",
		SpaceAvailable: 2_000,
		SpaceUsed:      60,
	}))

	resp, err := qs.StorageStats(suite.ctx, &types.QueryStorageStatsRequest{})
	suite.Require().NoError(err)
	suite.Equal(int32(1), resp.SubscriptionCount)
	suite.Equal(int64(100), resp.SpacePurchased) // 300 / 3
	suite.Equal(int64(50), resp.SpaceUsed)       // (90 + 60) / 3
	suite.Equal(int64(60), resp.SpaceReplicas)
	suite.Equal(int32(2), resp.ProviderCount)
	suite.Equal(int64(3_000), resp.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestStorageStatsQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.StorageStats(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
