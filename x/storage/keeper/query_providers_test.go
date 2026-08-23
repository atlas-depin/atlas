package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestProvidersQuery_Empty() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.Providers(suite.ctx, &types.QueryProvidersRequest{})
	suite.Require().NoError(err)
	suite.Require().Empty(resp.Providers)
}

func (suite *KeeperTestSuite) TestProvidersQuery_WithProviders() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	addr1 := suite.CreateTestAddress(0)
	addr2 := suite.CreateTestAddress(1)

	suite.createTestProvider(addr1, 1_000)
	suite.createTestProvider(addr2, 2_000)

	resp, err := qs.Providers(suite.ctx, &types.QueryProvidersRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Providers, 2)

	addresses := []string{resp.Providers[0].Address, resp.Providers[1].Address}
	suite.ElementsMatch([]string{addr1, addr2}, addresses)
}

func (suite *KeeperTestSuite) TestProvidersQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Providers(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
