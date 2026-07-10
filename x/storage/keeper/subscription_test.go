package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestSetNewStorageSubscription() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000),
	}

	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)
	suite.Contains(subId, "sub_", "auto-generated ID")

	// Verify stored
	stored, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().NoError(err)
	suite.Equal(sub.SpaceAvailable, stored.SpaceAvailable)
	suite.Equal(sub.Credits, stored.Credits)

	// Verify expiry index was written
	suite.True(suite.keeper.SubscriptionExpiryIndex.Has(suite.ctx, collections.Join3(end.Unix(), subId, owner)))
}

func (suite *KeeperTestSuite) TestGetStorageSubscription_ByID() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 500_000_000,
		Credits:        math.NewInt(500_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	got, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().NoError(err)
	suite.Equal(subId, got.Id)
	suite.Equal(int64(500_000_000), got.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestGetStorageSubscription_NotFound() {
	owner := suite.CreateTestAddress(0)
	_, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, "nonexistent")
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)
}

func (suite *KeeperTestSuite) TestGetStorageSubscription_BlankID_UsesDefault() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Set as default
	suite.Require().NoError(suite.keeper.SetDefaultStorageSubscriptionId(suite.ctx, owner, subId))

	// Lookup with blank ID
	got, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, "")
	suite.Require().NoError(err)
	suite.Equal(subId, got.Id)
}

func (suite *KeeperTestSuite) TestGetStorageSubscription_BlankID_NoDefault_FindsActive() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 2_000_000,
		Credits:        math.NewInt(2_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Blank ID with no default set should find an active subscription
	got, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, "")
	suite.Require().NoError(err)
	suite.Equal(subId, got.Id)
}

func (suite *KeeperTestSuite) TestGetStorageSubscription_BlankID_AllExpired() {
	owner := suite.CreateTestAddress(0)
	// Subscription that already expired
	end := suite.ctx.BlockTime().Add(-1 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime().Add(-24 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000),
	}
	_, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Blank ID with no default, only expired subs
	_, err = suite.keeper.GetStorageSubscription(suite.ctx, owner, "")
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)
}

func (suite *KeeperTestSuite) TestExpireStorageSubscription() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Verify index entry exists
	key := collections.Join3(end.Unix(), subId, owner)
	suite.True(suite.keeper.SubscriptionExpiryIndex.Has(suite.ctx, key))

	// Expire it
	err = suite.keeper.ExpireStorageSubscription(suite.ctx, owner, subId)
	suite.Require().NoError(err)

	// Verify subscription is gone
	_, err = suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)

	// Verify index entry cleaned up
	suite.False(suite.keeper.SubscriptionExpiryIndex.Has(suite.ctx, key))
}

func (suite *KeeperTestSuite) TestExpireStorageSubscription_ClearsDefault() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000,
		Credits:        math.NewInt(1_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Set as default
	suite.Require().NoError(suite.keeper.SetDefaultStorageSubscriptionId(suite.ctx, owner, subId))

	// Expire
	err = suite.keeper.ExpireStorageSubscription(suite.ctx, owner, subId)
	suite.Require().NoError(err)

	// Verify default was cleared
	_, err = suite.keeper.GetDefaultStorageSubscriptionId(suite.ctx, owner)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestUpdateStorageSubscriptionUsage_Normal() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)
	fileSize := int64(1_000_000)
	replicas := int32(3)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000), // plenty
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, fileSize, replicas)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().NoError(err)
	suite.Equal(fileSize*int64(replicas), stored.SpaceUsed)
	suite.True(stored.Credits.LT(sub.Credits), "credits should decrease")
}

func (suite *KeeperTestSuite) TestUpdateStorageSubscriptionUsage_Full() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)
	fileSize := int64(1_000_000)
	replicas := int32(3)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: fileSize * int64(replicas), // exactly enough for one file
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// First file fits
	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, fileSize, replicas)
	suite.Require().NoError(err)

	// Second file exceeds
	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, fileSize, replicas)
	suite.Require().ErrorIs(err, types.ErrSubscriptionFull)
}

func (suite *KeeperTestSuite) TestUpdateStorageSubscriptionUsage_Expired() {
	owner := suite.CreateTestAddress(0)
	// Past end time
	end := suite.ctx.BlockTime().Add(-1 * time.Hour)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime().Add(-24 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, 1000, 1)
	suite.Require().ErrorIs(err, types.ErrSubscriptionExpired)

	// Verify subscription was cleaned up
	_, err = suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)
}

func (suite *KeeperTestSuite) TestUpdateStorageSubscriptionUsage_InsufficientCredit() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)
	fileSize := int64(1_000_000_000) // large file
	replicas := int32(3)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000_000,
		Credits:        math.NewInt(1), // not enough
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, fileSize, replicas)
	suite.Require().ErrorIs(err, types.ErrSubscriptionInsufficientCredit)
}

func (suite *KeeperTestSuite) TestUpdateStorageSubscriptionUsage_Decrement() {
	owner := suite.CreateTestAddress(0)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)
	fileSize := int64(1_000_000)
	replicas := int32(3)

	sub := types.StorageSubscription{
		Start:          suite.ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)
	suite.Require().NoError(err)

	// Add file
	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, fileSize, replicas)
	suite.Require().NoError(err)

	// Remove file (negative bytes)
	_, err = suite.keeper.UpdateStorageSubscriptionUsage(suite.ctx, owner, subId, -fileSize, replicas)
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetStorageSubscription(suite.ctx, owner, subId)
	suite.Require().NoError(err)
	suite.Equal(int64(0), stored.SpaceUsed)
}
