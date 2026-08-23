package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (suite *KeeperTestSuite) TestProviderQuery() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	addr := suite.CreateTestAddress(0)
	suite.createTestProvider(addr, 1_000_000)

	resp, err := qs.Provider(suite.ctx, &types.QueryProviderRequest{Address: addr})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Provider)
	suite.Equal(addr, resp.Provider.Address)
	suite.Equal("test-provider", resp.Provider.Hostname)
	suite.Equal(int64(1_000_000), resp.Provider.SpaceAvailable)
}

func (suite *KeeperTestSuite) TestProviderQuery_NotFound() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	resp, err := qs.Provider(suite.ctx, &types.QueryProviderRequest{Address: suite.CreateTestAddress(9)})
	suite.Require().Error(err)
	suite.Require().Equal(codes.NotFound, status.Code(err))
	suite.Require().ErrorContains(err, "provider not found")
	suite.Require().Nil(resp)
}

func (suite *KeeperTestSuite) TestProviderQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Provider(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
