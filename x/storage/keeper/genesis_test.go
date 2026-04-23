package keeper_test

import (
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	ctx, k := suite.ctx, suite.keeper

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	err := k.InitGenesis(ctx, genesisState)
	suite.Require().NoError(err)
	got, err := k.ExportGenesis(ctx)
	suite.Require().NoError(err)
	suite.Require().NotNil(got)

	suite.Require().EqualExportedValues(genesisState.Params, got.Params)
}
