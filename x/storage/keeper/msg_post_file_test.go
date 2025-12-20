package keeper_test

import (
	"encoding/hex"
	"fmt"
	"nebulix/x/storage/types"
	"strconv"
	"time"

	"cosmossdk.io/math"
	"github.com/cespare/xxhash"
)

func (suite *KeeperTestSuite) TestPostFile_Success_DefaultReplicas() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	merkleHash := []byte("test-merkle-hash-1234567890")
	fileSize := int64(1_000_000_000) // 1GB
	subscriptionId := "sub_0"

	// Create a valid subscription
	endTime := suite.ctx.BlockTime().AddDate(0, 1, 0) // 1 year from now
	suite.createTestSubscription(creator, subscriptionId, 10_000_000_000, 0, endTime)

	msg := &types.MsgPostFile{
		Creator:      creator,
		Merkle:       merkleHash,
		FileSize:     fileSize,
		Subscription: subscriptionId,
		Replicas:     0,
	}
	fidString := hex.EncodeToString(merkleHash) + creator + strconv.FormatInt(ctx.BlockHeight(), 10)
	fid := xxhash.Sum64([]byte(fidString))

	response, err := msgServer.PostFile(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.Require().Equal(ctx.BlockHeight(), response.StartBlock)

	// Verify file was stored
	storedFile, err := suite.keeper.GetFile(ctx, fmt.Sprintf("%x", fid))
	suite.Require().NoError(err)
	suite.Require().Equal(creator, storedFile.Creator)
	suite.Require().Equal(fileSize, storedFile.FileSize)
	suite.Require().Equal(int64(3), storedFile.Replicas)
	suite.Require().Equal([]string{}, storedFile.Providers)
	suite.Require().Equal(subscriptionId, storedFile.Subscription)
	suite.Require().Equal(ctx.BlockHeight(), storedFile.Start)

	// Verify subscription was updated
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, subscriptionId)
	suite.Require().NoError(err)
	suite.Require().Equal(fileSize*3, updatedSub.SpaceUsed)        // 1GB * 3 replicas
	suite.Require().Equal(fileSize*2, updatedSub.ReplicaSpaceUsed) // 1GB * (3-1) replicas

	// Verify credits were transferred to pool
	poolBalance, err := suite.keeper.StorageCreditPool.Get(ctx)
	expectedSpentCredits := math.NewInt(fileSize * 3).Mul(math.NewInt(int64(updatedSub.End.Sub(ctx.BlockTime()) / time.Second)))
	suite.Require().NoError(err)
	suite.Require().True(poolBalance.GT(math.NewInt(0)), "credits should be added to pool")
	suite.Require().Equal(poolBalance, expectedSpentCredits)
}

func (suite *KeeperTestSuite) TestPostFile_Success_CustomReplicas() {
	msgServer, ctx := suite.setupMsgServer()

	creator := suite.CreateTestAddress(0)
	merkleHash := []byte("test-merkle-hash-abcdef")
	fileSize := int64(500_000_000) // 500MB
	subscriptionId := "custom-sub"
	replicas := int64(5)

	// Create subscription with enough space
	endTime := suite.ctx.BlockTime().AddDate(0, 1, 0)
	suite.createTestSubscription(creator, subscriptionId, 10_000_000_000, 0, endTime)

	msg := &types.MsgPostFile{
		Creator:      creator,
		Merkle:       merkleHash,
		FileSize:     fileSize,
		Subscription: subscriptionId,
		Replicas:     replicas,
	}

	response, err := msgServer.PostFile(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	// Verify file
	storedFile, err := suite.keeper.GetFile(ctx, hex.EncodeToString(merkleHash))
	suite.Require().NoError(err)
	suite.Require().Equal(replicas, storedFile.Replicas)

	// Verify subscription update
	updatedSub, err := suite.keeper.GetStorageSubscription(ctx, creator, subscriptionId)
	suite.Require().NoError(err)
	suite.Require().Equal(fileSize*replicas, updatedSub.SpaceUsed)
	suite.Require().Equal(fileSize*(replicas-1), updatedSub.ReplicaSpaceUsed)
}
