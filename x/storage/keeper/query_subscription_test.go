package keeper_test

import (
	"time"

	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestSubscriptionQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	sub := suite.createTestSubscription(owner, "sub_test", 1_000_000, 0, suite.ctx.BlockTime().Add(24*time.Hour))

	resp, err := qs.Subscription(suite.ctx, &types.QuerySubscriptionRequest{
		SubscriberAddress: owner,
		SubscriptionId:    sub.Id,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Subscription)
	suite.Equal(sub.Id, resp.Subscription.Id)
	suite.Equal(sub.SpaceAvailable, resp.Subscription.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestSubscriptionQuery_NotFound() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Subscription(suite.ctx, &types.QuerySubscriptionRequest{
		SubscriberAddress: suite.CreateTestAddress(0),
		SubscriptionId:    "sub_missing",
	})
	suite.Require().ErrorIs(err, types.ErrSubscriptionNotFound)
}

func (suite *KeeperTestSuite) TestSubscriptionQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Subscription(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
