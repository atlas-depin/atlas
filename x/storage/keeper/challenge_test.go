package keeper_test

import (
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestSetGetChallenge() {
	provider := suite.CreateTestAddress(0)
	challenge := types.StorageChallenge{
		ChallengeId:    "challenge-1",
		FileId:         "file-1",
		FileMerkle:     "merkle",
		Provider:       provider,
		ChunkIndex:     3,
		CreatedHeight:  100,
		DeadlineHeight: 200,
	}

	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, challenge))

	stored, err := suite.keeper.GetChallenge(suite.ctx, provider, "challenge-1")
	suite.Require().NoError(err)
	suite.Require().Equal(challenge, stored)
}

func (suite *KeeperTestSuite) TestGetChallenge_NotFound() {
	_, err := suite.keeper.GetChallenge(suite.ctx, suite.CreateTestAddress(0), "missing")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestRemoveChallenge() {
	provider := suite.CreateTestAddress(0)
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "challenge-1",
		Provider:    provider,
	}))

	suite.Require().NoError(suite.keeper.RemoveChallenge(suite.ctx, provider, "challenge-1"))

	_, err := suite.keeper.GetChallenge(suite.ctx, provider, "challenge-1")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestIterateChallenges() {
	p1 := suite.CreateTestAddress(0)
	p2 := suite.CreateTestAddress(1)

	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{ChallengeId: "c1", Provider: p1}))
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{ChallengeId: "c2", Provider: p2}))

	var ids []string
	err := suite.keeper.IterateChallenges(suite.ctx, func(challenge types.StorageChallenge) bool {
		ids = append(ids, challenge.ChallengeId)
		return false
	})
	suite.Require().NoError(err)
	suite.ElementsMatch([]string{"c1", "c2"}, ids)
}
