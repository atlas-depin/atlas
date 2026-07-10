package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestMsgDeleteFile_Success() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	// Set up a subscription and provider
	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.keeper.SetDefaultStorageSubscriptionId(suite.ctx, creator, subId))

	// Create provider with space
	suite.createTestProvider(provider, 1_000_000_000)

	// Create a file with subscription set
	file := suite.createTestFileWithSub("delete-test-file", creator, subId, 1024*1024, 3, []string{provider})

	// Track subscription provider bytes
	subProvKey := collections.Join(subId, provider)
	suite.Require().NoError(suite.keeper.SubscriptionProviderBytes.Set(suite.ctx, subProvKey, file.FileSize))

	// Delete the file
	msg := &types.MsgDeleteFile{
		Creator: creator,
		FileId:  file.Fid,
	}
	_, err = msgServer.DeleteFile(ctx, msg)
	suite.Require().NoError(err)

	// Verify file is gone
	_, err = suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().ErrorIs(err, types.ErrFileNotFound)

	// Verify subscription provider bytes removed
	_, err = suite.keeper.SubscriptionProviderBytes.Get(suite.ctx, subProvKey)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestMsgDeleteFile_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgDeleteFile{
		Creator: "invalid_bech32",
		FileId:  "some-file",
	}
	_, err := msgServer.DeleteFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid authority address")
}

func (suite *KeeperTestSuite) TestMsgDeleteFile_FileNotFound() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	msg := &types.MsgDeleteFile{
		Creator: creator,
		FileId:  "nonexistent-file",
	}
	_, err := msgServer.DeleteFile(ctx, msg)
	suite.Require().ErrorIs(err, types.ErrFileNotFound)
}

func (suite *KeeperTestSuite) TestMsgDeleteFile_Unauthorized() {
	msgServer, ctx := suite.setupMsgServer()

	owner := suite.CreateTestAddress(0)
	attacker := suite.CreateTestAddress(1)

	file := suite.createTestFile("unauth-file", owner, 1024*1024, 3, []string{})

	msg := &types.MsgDeleteFile{
		Creator: attacker,
		FileId:  file.Fid,
	}
	_, err := msgServer.DeleteFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "unauthorized")
}

func (suite *KeeperTestSuite) TestMsgDeleteFile_SubscriptionExpired() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)

	// Past-end subscription
	end := ctx.BlockTime().Add(-1 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime().Add(-24 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)
	suite.Require().NoError(err)

	file := suite.createTestFileWithSub("expired-sub-file", creator, subId, 1024*1024, 3, []string{})

	msg := &types.MsgDeleteFile{
		Creator: creator,
		FileId:  file.Fid,
	}
	_, err = msgServer.DeleteFile(ctx, msg)
	suite.Require().NoError(err) // expired sub returns nil, nil (silent no-op)

	// File is NOT removed by DeleteFile when subscription is expired —
	// it's handled lazily by ExpireFile on next prove/challenge.
	_, err = suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().NoError(err, "file should still exist (expired sub cleanup is lazy)")
}

func (suite *KeeperTestSuite) TestMsgDeleteFile_StagedFile() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewInt(1_000_000_000_000),
	}
	subId, err := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)

	suite.Require().NoError(err)
	suite.Require().NoError(suite.keeper.SetDefaultStorageSubscriptionId(suite.ctx, creator, subId))

	suite.createTestProvider(provider, 1_000_000_000)

	// Create a STAGED file (no providers, subscription set)
	file := types.File{
		Fid:          "staged-file",
		Creator:      creator,
		Subscription: subId,
		Status:       "STAGED",
		FileSize:     1024 * 1024,
		Replicas:     3,
	}
	suite.Require().NoError(suite.keeper.SetNewFile(suite.ctx, file))
	suite.Require().NoError(suite.keeper.IncrementFileCount(suite.ctx))

	msg := &types.MsgDeleteFile{
		Creator: creator,
		FileId:  file.Fid,
	}
	_, err = msgServer.DeleteFile(ctx, msg)
	suite.Require().NoError(err)

	// File should be removed
	_, err = suite.keeper.GetFile(suite.ctx, file.Fid)
	suite.Require().Error(err)
}
