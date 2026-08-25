package keeper_test

import (
	"atlas/x/filetree/keeper"
	"atlas/x/filetree/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)
	params := types.DefaultParams()

	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	response, err := qs.Params(suite.ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}

func (suite *KeeperTestSuite) TestParamsQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Params(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
