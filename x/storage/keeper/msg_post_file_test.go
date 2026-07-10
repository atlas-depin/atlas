package keeper_test

import (
	"atlas/x/storage/types"
	"encoding/hex"
	"strconv"
)

func (suite *KeeperTestSuite) TestPostFile_Success_DefaultReplicas() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	merkleHash := []byte("test-merkle-hash-1234567890")
	fileSize := int64(1_000_000_000) // 1GB
	subscriptionId := "sub_0"
	nonce := int32(1)

	// Create a valid subscription
	endTime := suite.ctx.BlockTime().AddDate(0, 1, 0) // 1 year from now
	initialCredits := suite.createTestSubscription(creator, subscriptionId, 10_000_000_000, 0, endTime).Credits

	fid := hex.EncodeToString(merkleHash) + creator + strconv.FormatInt(int64(nonce), 10)

	msg := &types.MsgPostFile{
		Fid:          fid,
		Creator:      creator,
		Merkle:       merkleHash,
		FileSize:     fileSize,
		Subscription: subscriptionId,
		Replicas:     0,
	}

	response, err := msgServer.PostFile(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.Require().Equal(ctx.BlockHeight(), response.StartBlock)

	// Verify file was stored
	storedFile, err := suite.keeper.GetFile(ctx, fid)
	suite.Require().NoError(err)
	suite.Require().Equal(creator, storedFile.Creator)
	suite.Require().Equal(fileSize, storedFile.FileSize)
	suite.Require().Equal(int32(3), storedFile.Replicas)
	suite.Require().Nil(storedFile.Providers, "STAGED files have no providers assigned")
	suite.Require().Equal(subscriptionId, storedFile.Subscription)
	suite.Require().Equal(ctx.BlockTime().Unix(), storedFile.Start)

	// Verify subscription was updated
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, subscriptionId)
	suite.Require().NoError(err)
	suite.Require().Equal(fileSize*3, updatedSub.SpaceUsed)        // 1GB * 3 replicas
	suite.Require().Equal(fileSize*2, updatedSub.ReplicaSpaceUsed) // 1GB * (3-1) replicas

	// Verify subscription credits were deducted
	suite.Require().True(updatedSub.Credits.LT(initialCredits), "credits should be deducted from subscription")
}

func (suite *KeeperTestSuite) TestPostFile_Success_CustomReplicas() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	merkleHash := []byte("test-merkle-hash-abcdef")
	fileSize := int64(500_000_000) // 500MB
	subscriptionId := "custom-sub"
	replicas := int32(5)
	nonce := int32(1)

	// Create subscription with enough space
	endTime := suite.ctx.BlockTime().AddDate(0, 1, 0)
	suite.createTestSubscription(creator, subscriptionId, 10_000_000_000, 0, endTime)

	fid := hex.EncodeToString(merkleHash) + creator + strconv.FormatInt(int64(nonce), 10)

	msg := &types.MsgPostFile{
		Fid:          fid,
		Creator:      creator,
		Merkle:       merkleHash,
		FileSize:     fileSize,
		Subscription: subscriptionId,
		Replicas:     replicas,
	}

	response, err := msgServer.PostFile(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	// Verify file — lookup by the exact FID used in the message
	storedFile, err := suite.keeper.GetFile(ctx, fid)
	suite.Require().NoError(err)
	suite.Require().Equal(replicas, storedFile.Replicas)
	suite.Require().NotContains(storedFile.Providers, creator, "provider should not be assigned yet")

	// Verify subscription update
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, subscriptionId)
	suite.Require().NoError(err)
	suite.Require().Equal(fileSize*int64(replicas), updatedSub.SpaceUsed)
	suite.Require().Equal(fileSize*(int64(replicas)-1), updatedSub.ReplicaSpaceUsed)
}
