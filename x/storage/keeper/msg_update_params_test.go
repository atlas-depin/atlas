package keeper_test

import (
	"github.com/stretchr/testify/require"

	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	ms := keeper.NewMsgServerImpl(suite.keeper)

	params := types.DefaultParams()
	require.NoError(suite.T(), suite.keeper.Params.Set(suite.ctx, params))

	authorityStr, err := suite.addressCodec.BytesToString(suite.keeper.GetAuthority())
	suite.Require().NoError(err)

	// default params
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
			name: "empty params are rejected",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params:    types.Params{},
			},
			expErr:    true,
			expErrMsg: "proof_round_blocks must be > 0",
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
