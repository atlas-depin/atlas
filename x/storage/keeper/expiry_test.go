package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestExpireFile_Dead_Noop() {
	owner := suite.CreateTestAddress(0)
	file := suite.createTestFile("dead-file", owner, 1024, 1, nil)

	// Simulate the challenger killing the file.
	file.Status = "DEAD"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))
	suite.Require().NoError(suite.keeper.DecrementFileCount(suite.ctx))

	suite.Require().NoError(suite.keeper.ExpireFile(suite.ctx, file))

	// Dead files are left alone entirely.
	_, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err)

	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), count)
}

func (suite *KeeperTestSuite) TestExpireFile_Staged_RemovesWithoutDecrement() {
	owner := suite.CreateTestAddress(0)
	file := types.File{
		Fid:      "staged-expire",
		Creator:  owner,
		Status:   "STAGED",
		FileSize: 1024,
		Replicas: 3,
	}
	suite.Require().NoError(suite.keeper.SetNewFile(suite.ctx, file))

	// Staged files were never counted (count increments on first proof).
	suite.Require().NoError(suite.keeper.ExpireFile(suite.ctx, file))

	_, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().ErrorIs(err, types.ErrFileNotFound)

	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), count)
}

func (suite *KeeperTestSuite) TestExpireFile_NoSubscription() {
	owner := suite.CreateTestAddress(0)
	file := suite.createTestFile("no-sub-expire", owner, 1024, 1, nil)

	suite.Require().NoError(suite.keeper.ExpireFile(suite.ctx, file))

	_, err := suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().ErrorIs(err, types.ErrFileNotFound)

	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), count)
}

func (suite *KeeperTestSuite) TestExpireFile_Normal() {
	owner := suite.CreateTestAddress(0)
	providerAddr := suite.CreateTestAddress(1)

	end := suite.ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Provider holds the file's 1024 bytes.
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:   providerAddr,
		Hostname:  "provider-1",
		SpaceUsed: 1024,
	}))

	file := suite.createTestFileWithSub("expire-file", owner, subId, 1024, 3, []string{providerAddr})

	subProvKey := collections.Join(subId, providerAddr)
	suite.Require().NoError(suite.keeper.SubscriptionProviderBytes.Set(suite.ctx, subProvKey, 1024))

	suite.Require().NoError(suite.keeper.ExpireFile(suite.ctx, file))

	// Provider space freed and credit delta accrued since genesis rewards time (0).
	delta := end.Unix() - 0
	provider, err := suite.keeper.GetProvider(suite.ctx, providerAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), provider.SpaceUsed)
	suite.Require().Equal(int64(1024)*delta, provider.CreditDelta)

	// Per-subscription tracker cleaned up.
	_, err = suite.keeper.SubscriptionProviderBytes.Get(suite.ctx, subProvKey)
	suite.Require().Error(err)

	// File record removed and the global count decremented.
	_, err = suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().ErrorIs(err, types.ErrFileNotFound)

	count, err := suite.keeper.FileCount.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), count)

	// Expiry event emitted.
	events := suite.ctx.EventManager().Events()
	suite.Require().NotEmpty(events)
	suite.Require().Equal(types.EventTypeExpireFile, events[len(events)-1].Type)
}

func (suite *KeeperTestSuite) TestExpireSubscriptions() {
	owner := suite.CreateTestAddress(0)
	providerAddr := suite.CreateTestAddress(1)

	// Expired subscription: End is an hour in the past.
	end := suite.ctx.BlockTime().Add(-1 * time.Hour)
	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime().Add(-48 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.keeper.SetDefaultStorageSubscriptionId(suite.ctx, owner, subId))

	// Provider holds 500 bytes under this subscription.
	suite.Require().NoError(suite.keeper.SetProvider(suite.ctx, types.Provider{
		Address:   providerAddr,
		Hostname:  "provider-1",
		SpaceUsed: 500,
	}))

	subProvKey := collections.Join(subId, providerAddr)
	suite.Require().NoError(suite.keeper.SubscriptionProviderBytes.Set(suite.ctx, subProvKey, 500))

	suite.Require().NoError(suite.keeper.ExpireSubscriptions(suite.ctx))

	// Provider space freed via the per-subscription tracker.
	delta := end.Unix() - 0
	provider, err := suite.keeper.GetProvider(suite.ctx, providerAddr)
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), provider.SpaceUsed)
	suite.Require().Equal(int64(500)*delta, provider.CreditDelta)

	// Tracker entry removed.
	_, err = suite.keeper.SubscriptionProviderBytes.Get(suite.ctx, subProvKey)
	suite.Require().Error(err)

	// Subscription record gone and the default id cleared.
	_, err = suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)

	_, err = suite.keeper.GetDefaultStorageSubscriptionId(suite.ctx, owner)
	suite.Require().Error(err)

	// Expiry index entry cleaned up.
	suite.Require().False(suite.keeper.SubscriptionExpiryIndex.Has(suite.ctx, collections.Join3(end.Unix(), subId, owner)))
}
