package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestPriceQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.Price(suite.ctx, &types.QueryPriceRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultPrice, resp.Price)
}

func (suite *KeeperTestSuite) TestPriceQuery_Updated() {
	qs := keeper.NewQueryServerImpl(suite.keeper)
	suite.Require().NoError(suite.keeper.Price.Set(suite.ctx, "2.500000"))

	resp, err := qs.Price(suite.ctx, &types.QueryPriceRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal("2.500000", resp.Price)
}

func (suite *KeeperTestSuite) TestPriceQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Price(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
