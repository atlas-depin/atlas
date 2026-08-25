package types_test

import (
	"testing"

	"atlas/x/storage/types"

	"github.com/stretchr/testify/require"
)

func TestParams_Validate_Default(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())
}

func TestParams_Validate_Errors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*types.Params)
		err    error
		msg    string
	}{
		{
			name:   "zero proof round blocks",
			mutate: func(p *types.Params) { p.ProofRoundBlocks = 0 },
			err:    types.ErrProofRoundBlocksZero,
			msg:    "proof_round_blocks must be > 0",
		},
		{
			name:   "zero proof window blocks",
			mutate: func(p *types.Params) { p.ProofWindowBlocks = 0 },
			err:    types.ErrProofWindowBlocksZero,
			msg:    "proof_window_blocks must be > 0",
		},
		{
			name:   "round exceeds window",
			mutate: func(p *types.Params) { p.ProofRoundBlocks = 100 },
			err:    types.ErrProofRoundExceedsWindow,
			msg:    "proof_round_blocks (100) must be <= proof_window_blocks",
		},
		{
			name:   "window not divisible by round",
			mutate: func(p *types.Params) { p.ProofWindowBlocks = 55 },
			err:    types.ErrProofWindowUnaligned,
			msg:    "proof_window_blocks (55) must be divisible by proof_round_blocks",
		},
		{
			name:   "negative price per gb day",
			mutate: func(p *types.Params) { p.PricePerGbDay = -1 },
			msg:    "price_per_gb_day cannot be negative",
		},
		{
			name:   "negative provider deposit",
			mutate: func(p *types.Params) { p.ProviderDeposit = -1 },
			msg:    "provider_deposit cannot be negative",
		},
		{
			name:   "invalid price updater address",
			mutate: func(p *types.Params) { p.PriceUpdater = "not-an-address" },
			msg:    "price_updater is not a valid address",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.mutate(&params)

			err := params.Validate()
			require.Error(t, err)
			require.ErrorContains(t, err, tc.msg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			}
		})
	}
}
