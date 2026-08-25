package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

// Security audit: no handler may write state belonging to another account.
// PostFile and ExpandStorage both resolve the subscription under the signer's
// own address, so an attacker supplying someone else's subscription id must
// fail with no side effects.

func (suite *KeeperTestSuite) TestPostFile_AnotherUsersSubscription() {
	msgServer, ctx := suite.setupMsgServer()

	victim := suite.CreateTestAddress(0)
	attacker := suite.CreateTestAddress(1)

	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, victim, sub)
	suite.Require().NoError(err)

	msg := &types.MsgPostFile{
		Creator:      attacker,
		Fid:          "attacker-file",
		Merkle:       []byte("merkle-root"),
		FileSize:     1024,
		Replicas:     3,
		Subscription: subId,
	}
	_, err = msgServer.PostFile(ctx, msg)
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)

	// No file may have been staged for the attacker.
	_, err = suite.keeper.GetFile(suite.ctx, "attacker-file")
	suite.Require().ErrorIs(err, types.ErrFileNotFound)
}

func (suite *KeeperTestSuite) TestExpandStorage_AnotherUsersSubscription() {
	msgServer, ctx := suite.setupMsgServer()

	victim := suite.CreateTestAddress(0)
	attacker := suite.CreateTestAddress(1)

	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, victim, sub)
	suite.Require().NoError(err)

	msg := &types.MsgExpandStorage{
		Creator:        attacker,
		SubscriptionId: subId,
		Duration:       30,
		Bytes:          1_000_000_000,
	}
	_, err = msgServer.ExpandStorage(ctx, msg)
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)

	// Victim's subscription is untouched.
	stored, err := suite.keeper.GetStorageSubscription(suite.ctx, victim, subId)
	suite.Require().NoError(err)
	suite.Require().Equal(sub.SpaceAvailable, stored.SpaceAvailable)
	suite.Require().Equal(sub.End, stored.End)
	suite.Require().Equal(sub.Credits, stored.Credits)
}
