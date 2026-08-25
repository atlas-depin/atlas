package types

import (
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params instance.
func NewParams() Params {
	return Params{}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		PricePerGbDay:     400, // $0.0004 * 10^6
		ProviderDeposit:   10000,
		ProofWindowBlocks: 50,
		ProofRoundBlocks:  10,
		PriceUpdater:      "atl1234567890",
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.ProofRoundBlocks == 0 {
		return errors.Wrap(ErrProofRoundBlocksZero, "proof_round_blocks must be > 0")
	}
	if p.ProofWindowBlocks == 0 {
		return errors.Wrap(ErrProofWindowBlocksZero, "proof_window_blocks must be > 0")
	}
	if p.ProofRoundBlocks > p.ProofWindowBlocks {
		return errors.Wrapf(ErrProofRoundExceedsWindow,
			"proof_round_blocks (%d) must be <= proof_window_blocks (%d)",
			p.ProofRoundBlocks, p.ProofWindowBlocks,
		)
	}
	if p.ProofWindowBlocks%p.ProofRoundBlocks != 0 {
		return errors.Wrapf(ErrProofWindowUnaligned,
			"proof_window_blocks (%d) must be divisible by proof_round_blocks (%d)",
			p.ProofWindowBlocks, p.ProofRoundBlocks,
		)
	}
	if p.PricePerGbDay < 0 {
		return fmt.Errorf("price_per_gb_day cannot be negative: got %d", p.PricePerGbDay)
	}
	if p.ProviderDeposit < 0 {
		return fmt.Errorf("provider_deposit cannot be negative: got %d", p.ProviderDeposit)
	}
	if p.PriceUpdater != "" {
		if _, err := sdk.AccAddressFromBech32(p.PriceUpdater); err != nil {
			return fmt.Errorf("price_updater is not a valid address: %w", err)
		}
	}

	return nil
}
