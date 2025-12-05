package types

// NewParams creates a new Params instance.
func NewParams() Params {
	return Params{}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		PricePerGbDay: 400, // $0.0004 * 10^6
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {

	return nil
}
