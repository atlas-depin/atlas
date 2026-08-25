package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestFeedQuery_Found() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.Feed(suite.ctx, &types.QueryFeedRequest{Name: types.BuiltinFeedName})
	suite.Require().NoError(err)
	suite.Require().Equal(types.BuiltinFeedName, resp.Feed.Name)
	suite.Require().Equal(types.BuiltinFeedInitialData, resp.Feed.Data)
	suite.Require().Equal(suite.authorityAddress(), resp.Feed.Owner)
}

func (suite *KeeperTestSuite) TestFeedQuery_NotFound() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.Feed(suite.ctx, &types.QueryFeedRequest{Name: "DOGE"})
	suite.Require().Error(err)
	suite.Require().Equal(codes.NotFound, status.Code(err))
	suite.Require().ErrorContains(err, "feed not found")
	suite.Require().Nil(resp)
}

func (suite *KeeperTestSuite) TestFeedQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Feed(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}

func (suite *KeeperTestSuite) TestAllFeedsQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	// Seed an extra feed alongside the built-in one.
	suite.Require().NoError(suite.keeper.SetFeed(suite.ctx, types.Feed{
		Name:  "BTC",
		Owner: "somebody",
		Data:  "1.000000",
	}))

	resp, err := qs.AllFeeds(suite.ctx, &types.QueryAllFeedsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Feeds, 2)

	names := []string{resp.Feeds[0].Name, resp.Feeds[1].Name}
	suite.ElementsMatch([]string{types.BuiltinFeedName, "BTC"}, names)
}

func (suite *KeeperTestSuite) TestAllFeedsQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.AllFeeds(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
