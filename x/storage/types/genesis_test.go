package types_test

import (
	"testing"

	"atlas/x/storage/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "custom valid params",
			genState: &types.GenesisState{Params: types.Params{
				PricePerGbDay:     1,
				ProviderDeposit:   1,
				ProofWindowBlocks: 30,
				ProofRoundBlocks:  10,
			}},
			valid: true,
		},
		{
			desc:     "empty params are invalid",
			genState: &types.GenesisState{},
			valid:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
