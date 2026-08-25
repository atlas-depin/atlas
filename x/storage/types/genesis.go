package types

// DefaultPrice is the $ATL price seeded at genesis until the first
// MsgUpdatePrice. The format is defined by whatever reads it (canine-chain
// uses "price:marketCap" style strings).
const DefaultPrice = "1.000000"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
