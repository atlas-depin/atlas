package keeper_test

import (
	"atlas/x/storage/keeper"
	"atlas/x/storage/types"

	"github.com/cosmos/cosmos-sdk/types/query"
)

func (suite *KeeperTestSuite) TestStraysQuery_NoStrays() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.createTestFile("file-1", owner, 1024, 1, nil)

	resp, err := qs.Strays(suite.ctx, &types.QueryStraysRequest{})
	suite.Require().NoError(err)
	suite.Require().Empty(resp.Files)
	suite.Empty(resp.Pagination.NextKey)
}

func (suite *KeeperTestSuite) TestStraysQuery_WithStrays() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	suite.createTestFile("file-active", owner, 1024, 1, nil)

	stray1 := suite.createTestFile("stray-1", owner, 2048, 1, nil)
	stray1.Status = "STRAY"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, stray1))

	stray2 := suite.createTestFile("stray-2", owner, 4096, 1, nil)
	stray2.Status = "STRAY"
	suite.Require().NoError(suite.keeper.SetFile(suite.ctx, stray2))

	resp, err := qs.Strays(suite.ctx, &types.QueryStraysRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Files, 2)

	fids := []string{resp.Files[0].Fid, resp.Files[1].Fid}
	suite.ElementsMatch([]string{"stray-1", "stray-2"}, fids)
}

func (suite *KeeperTestSuite) TestStraysQuery_Pagination() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	owner := suite.CreateTestAddress(0)
	for _, fid := range []string{"stray-a", "stray-b", "stray-c"} {
		file := suite.createTestFile(fid, owner, 1024, 1, nil)
		file.Status = "STRAY"
		suite.Require().NoError(suite.keeper.SetFile(suite.ctx, file))
	}

	// First page: limit 1
	resp, err := qs.Strays(suite.ctx, &types.QueryStraysRequest{
		Pagination: &query.PageRequest{Limit: 1},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp.Files, 1)
	suite.Equal("stray-a", resp.Files[0].Fid)
	suite.Require().NotEmpty(resp.Pagination.NextKey)

	// Second page: resume from cursor
	resp2, err := qs.Strays(suite.ctx, &types.QueryStraysRequest{
		Pagination: &query.PageRequest{Key: resp.Pagination.NextKey, Limit: 1},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp2.Files, 1)
	suite.Equal("stray-b", resp2.Files[0].Fid)

	// Third page: last file — no cursor should be returned
	resp3, err := qs.Strays(suite.ctx, &types.QueryStraysRequest{
		Pagination: &query.PageRequest{Key: resp2.Pagination.NextKey, Limit: 1},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp3.Files, 1)
	suite.Equal("stray-c", resp3.Files[0].Fid)
	suite.Empty(resp3.Pagination.NextKey)
}

func (suite *KeeperTestSuite) TestStraysQuery_NilRequest() {
	qs := keeper.NewQueryServerImpl(suite.keeper)

	_, err := qs.Strays(suite.ctx, nil)
	suite.Require().ErrorContains(err, "invalid request")
}
