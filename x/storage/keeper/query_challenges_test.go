package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestChallengesQuery_All() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	p1 := suite.CreateTestAddress(0)
	p2 := suite.CreateTestAddress(1)

	c1 := types.StorageChallenge{ChallengeId: "challenge-1", Provider: p1, FileId: "file-1"}
	c2 := types.StorageChallenge{ChallengeId: "challenge-2", Provider: p1, FileId: "file-2"}
	c3 := types.StorageChallenge{ChallengeId: "challenge-3", Provider: p2, FileId: "file-3"}

	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, c1))
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, c2))
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, c3))

	resp, err := qs.Challenges(suite.ctx, &types.QueryChallengesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Challenges, 3)
	suite.ElementsMatch([]*types.StorageChallenge{&c1, &c2, &c3}, resp.Challenges)
}

func (suite *KeeperTestSuite) TestChallengesQuery_ByProvider() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	p1 := suite.CreateTestAddress(0)
	p2 := suite.CreateTestAddress(1)

	c1 := types.StorageChallenge{ChallengeId: "challenge-1", Provider: p1, FileId: "file-1"}
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, c1))
	suite.Require().NoError(suite.keeper.SetChallenge(suite.ctx, types.StorageChallenge{
		ChallengeId: "challenge-2",
		Provider:    p2,
		FileId:      "file-2",
	}))

	resp, err := qs.Challenges(suite.ctx, &types.QueryChallengesRequest{Provider: p1})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Challenges, 1)
	suite.Equal(c1, *resp.Challenges[0])
}

func (suite *KeeperTestSuite) TestChallengesQuery_InvalidProvider() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Challenges(suite.ctx, &types.QueryChallengesRequest{Provider: "not-a-valid-bech32"})
	suite.Require().ErrorContains(err, "invalid provider address")
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
}

func (suite *KeeperTestSuite) TestChallengesQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Challenges(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
