package testutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
)

// AuthMockHelpers provides helper functions for mocking account keeper
type AuthMockHelpers struct{}

func NewAuthMockHelpers() *AuthMockHelpers {
	return &AuthMockHelpers{}
}

// MockAccountExists mocks that an account exists
func (h *AuthMockHelpers) MockAccountExists(mock *MockAuthKeeper, addr []byte) *gomock.Call {
	return mock.EXPECT().
		HasAccount(gomock.Any(), addr).
		Return(true)
}

// MockAccountDoesNotExist mocks that an account doesn't exist
func (h *AuthMockHelpers) MockAccountDoesNotExist(mock *MockAuthKeeper, addr []byte) *gomock.Call {
	return mock.EXPECT().
		HasAccount(gomock.Any(), addr).
		Return(false)
}

// MockNewAccount mocks creation of a new account
func (h *AuthMockHelpers) MockNewAccount(mock *MockAuthKeeper, addr []byte, newAccount sdk.AccountI) {
	mock.EXPECT().
		NewAccountWithAddress(gomock.Any(), addr).
		Return(newAccount)

	mock.EXPECT().
		SetAccount(gomock.Any(), gomock.Any())
}
