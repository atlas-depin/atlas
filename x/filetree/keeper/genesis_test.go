package keeper_test

import (
	"atlas/x/filetree/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	err := suite.keeper.InitGenesis(suite.ctx, genesisState)
	suite.Require().NoError(err)

	got, err := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NotNil(got)
	suite.Require().EqualExportedValues(genesisState.Params, got.Params)
}

func (suite *KeeperTestSuite) TestGenesis_InitState() {
	// SetupTest already ran InitGenesis with the default genesis state.
	params, err := suite.keeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultParams(), params)
}
