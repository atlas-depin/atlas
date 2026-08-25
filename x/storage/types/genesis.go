package types

import "fmt"

const (
	// BuiltinFeedName is the name of the built-in $ATL price feed.
	BuiltinFeedName = "ATL"

	// BuiltinFeedInitialData is the seed data for the built-in feed until the
	// first UpdateFeed. The format is defined by the price source that writes
	// the feed; canine-chain stores "price:marketCap" style strings.
	BuiltinFeedInitialData = "0.000000"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(gs.Feeds))
	for _, feed := range gs.Feeds {
		if feed.Name == "" {
			return fmt.Errorf("feed name cannot be empty")
		}
		if _, ok := seen[feed.Name]; ok {
			return fmt.Errorf("duplicate feed name %q", feed.Name)
		}
		seen[feed.Name] = struct{}{}
	}

	return nil
}
