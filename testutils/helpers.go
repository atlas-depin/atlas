package testutils

// MockHelpers aggregates all helpers for convenience
type MockHelpers struct {
	Account *AuthMockHelpers
	Bank    *BankMockHelpers
}

func NewMockHelpers() *MockHelpers {
	return &MockHelpers{
		Account: NewAuthMockHelpers(),
		Bank:    NewBankMockHelpers(),
	}
}

// [TODO]: address helpers?
