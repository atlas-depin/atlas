package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestParamsQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)
	params := types.DefaultParams()

	// Set params in the keeper
	err := suite.keeper.Params.Set(suite.ctx, params)
	suite.Require().NoError(err)

	// Query params
	response, err := qs.Params(suite.ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}
