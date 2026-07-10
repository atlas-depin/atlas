package keeper_test

import (
	"math/big"
	"time"

	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestProveFile_InvalidCreator() {
	msgServer, ctx := suite.setupMsgServer()

	msg := &types.MsgProveFile{
		Creator: "invalid_bech32",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid address")
}

func (suite *KeeperTestSuite) TestProveFile_InvalidProvider() {
	msgServer, ctx := suite.setupMsgServer()

	// Creator is valid but not registered as a provider
	creator := suite.CreateTestAddress(0)

	msg := &types.MsgProveFile{
		Creator: creator,
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid provider address")
}

func (suite *KeeperTestSuite) TestProveFile_StrayClaim_FileNotFound() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	suite.createTestProvider(creator, 1_000_000_000_000)

	msg := &types.MsgProveFile{
		Creator: creator,
		Fid:     "nonexistent-file",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "file not found")
}

func (suite *KeeperTestSuite) TestProveFile_StrayClaim_FileAlreadyActive() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	suite.createTestProvider(provider, 1_000_000_000_000)

	// Create a subscription and a file with ACTIVE status (already has all providers)
	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}
	subId, _ := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)

	file := suite.createTestFileWithSub("active-file", creator, subId, 1024*1024, 1, []string{provider})
	file.Status = "ACTIVE"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	msg := &types.MsgProveFile{
		Creator: provider,
		Fid:     "active-file",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not a stray")
}

func (suite *KeeperTestSuite) TestProveFile_StrayClaim_ProviderAlreadyAssigned() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	suite.createTestProvider(provider, 1_000_000_000_000)
	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}
	subId, _ := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)

	// File already has this provider — create as STRAY so the "already assigned" check fires
	// (not enough replicas met, so status won't be ACTIVE)
	file := suite.createTestFileWithSub("dup-file", creator, subId, 1024*1024, 3, []string{provider})
	file.Status = "STRAY"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))

	msg := &types.MsgProveFile{
		Creator: provider,
		Fid:     "dup-file",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "already assigned")
}

func (suite *KeeperTestSuite) TestProveFile_StrayClaim_SubscriptionExpired() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	suite.createTestProvider(provider, 1_000_000_000_000)
	// Expired subscription
	end := ctx.BlockTime().Add(-1 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime().Add(-24 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}
	subId, _ := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)

	_ = suite.createTestFileWithSub("expired-claim-file", creator, subId, 1024*1024, 3, []string{})

	msg := &types.MsgProveFile{
		Creator: provider,
		Fid:     "expired-claim-file",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().NoError(err, "expired sub should silently handle expiry")
}

func (suite *KeeperTestSuite) TestProveFile_Challenge_ExpiredSubscription() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)

	suite.createTestProvider(provider, 1_000_000_000_000)

	// Expired sub
	end := ctx.BlockTime().Add(-1 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime().Add(-24 * time.Hour),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}
	subId, _ := suite.keeper.SetNewStorageSubscription(suite.ctx, creator, sub)

	file := suite.createTestFileWithSub("challenge-expired-file", creator, subId, 1024*1024, 3, []string{provider})

	// Create a challenge
	challenge := types.StorageChallenge{
		ChallengeId:    "test-challenge-expired",
		FileId:         file.Fid,
		FileMerkle:     file.Merkle,
		Provider:       provider,
		ChunkIndex:     0,
		CreatedHeight:  100,
		DeadlineHeight: 200,
	}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	msg := &types.MsgProveFile{
		Creator:     provider,
		Fid:         file.Fid,
		ChallengeId: "test-challenge-expired",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().NoError(err, "expired sub should handle and return success")

	// Challenge should be removed
	_, err = suite.keeper.GetChallenge(suite.ctx, provider, "test-challenge-expired")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestProveFile_Challenge_NotFound() {
	msgServer, ctx := suite.setupMsgServer()

	provider := suite.CreateTestAddress(0)
	suite.createTestProvider(provider, 1_000_000_000_000)

	msg := &types.MsgProveFile{
		Creator:     provider,
		Fid:         "some-file",
		ChallengeId: "nonexistent-challenge",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().NoError(err, "missing challenge is handled silently")
}

func (suite *KeeperTestSuite) TestProveFile_Challenge_BadProvider() {
	msgServer, ctx := suite.setupMsgServer()

	owner := suite.CreateTestAddress(0)
	provider := suite.CreateTestAddress(1)
	attacker := suite.CreateTestAddress(2)

	suite.createTestProvider(attacker, 1_000_000_000_000)

	end := ctx.BlockTime().Add(24 * time.Hour)
	sub := types.StorageSubscription{
		Start:          ctx.BlockTime(),
		End:            end,
		SpaceAvailable: 1_000_000_000,
		Credits:        math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}
	subId, _ := suite.keeper.SetNewStorageSubscription(suite.ctx, owner, sub)

	file := suite.createTestFileWithSub("bad-prov-file", owner, subId, 1024*1024, 3, []string{provider})

	// Challenge for a DIFFERENT provider than the file's actual provider
	challenge := types.StorageChallenge{
		ChallengeId:    "bad-prov-challenge",
		FileId:         file.Fid,
		FileMerkle:     file.Merkle,
		Provider:       attacker,
		ChunkIndex:     0,
		CreatedHeight:  100,
		DeadlineHeight: 200,
	}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	msg := &types.MsgProveFile{
		Creator:     attacker,
		Fid:         file.Fid,
		ChallengeId: "bad-prov-challenge",
	}
	_, err := msgServer.ProveFile(ctx, msg)
	suite.Require().Error(err)
}

// TestProveFile_StrayClaim_Success and TestProveFile_Challenge_Success require
// constructing a valid Merkle proof via the tree builder. Those tests need a
// helper that creates a file with a known Merkle root and generates valid
// proof data (Data, Hashes, Path, Chunk) for a leaf in that tree.
