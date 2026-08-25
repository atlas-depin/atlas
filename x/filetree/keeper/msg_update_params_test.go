package keeper_test

import (
	"atlas/x/filetree/keeper"
	"atlas/x/filetree/types"
)

func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	ms := keeper.NewMsgServerImpl(suite.keeper)

	params := types.DefaultParams()
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	authorityStr, err := suite.addressCodec.BytesToString(suite.keeper.GetAuthority())
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := ms.UpdateParams(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
