package keeper_test

import (
	"time"

	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestSubscriptionsQuery_All() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner1 := suite.CreateTestAddress(0)
	owner2 := suite.CreateTestAddress(1)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	suite.createTestSubscription(owner1, "sub_1", 1_000_000, 0, end)
	suite.createTestSubscription(owner1, "sub_2", 2_000_000, 0, end)
	suite.createTestSubscription(owner2, "sub_3", 3_000_000, 0, end)

	resp, err := qs.Subscriptions(suite.ctx, &types.QuerySubscriptionsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Subscriptions, 3)

	var gotIds []string
	for _, sub := range resp.Subscriptions {
		gotIds = append(gotIds, sub.Id)
	}
	suite.ElementsMatch([]string{"sub_1", "sub_2", "sub_3"}, gotIds)
}

func (suite *KeeperTestSuite) TestSubscriptionsQuery_BySubscriber() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner1 := suite.CreateTestAddress(0)
	owner2 := suite.CreateTestAddress(1)
	end := suite.ctx.BlockTime().Add(24 * time.Hour)

	suite.createTestSubscription(owner1, "sub_1", 1_000_000, 0, end)
	suite.createTestSubscription(owner1, "sub_2", 2_000_000, 0, end)
	suite.createTestSubscription(owner2, "sub_3", 3_000_000, 0, end)

	resp, err := qs.Subscriptions(suite.ctx, &types.QuerySubscriptionsRequest{
		SubscriberAddress: owner1,
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Subscriptions, 2)

	var gotIds []string
	for _, sub := range resp.Subscriptions {
		gotIds = append(gotIds, sub.Id)
	}
	suite.ElementsMatch([]string{"sub_1", "sub_2"}, gotIds)
}

func (suite *KeeperTestSuite) TestSubscriptionsQuery_InvalidSubscriber() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Subscriptions(suite.ctx, &types.QuerySubscriptionsRequest{
		SubscriberAddress: "not-a-valid-bech32",
	})
	suite.Require().ErrorContains(err, "bad subscriber address")
	suite.Require().Equal(codes.InvalidArgument, status.Code(err))
}

func (suite *KeeperTestSuite) TestSubscriptionsQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Subscriptions(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
