package testutils

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
)

// BankMockHelpers provides helper functions for mocking bank keeper
type BankMockHelpers struct{}

func NewBankMockHelpers() *BankMockHelpers {
	return &BankMockHelpers{}
}

// MockSuccessfulPayment mocks a successful bank transfer to module
func (h *BankMockHelpers) MockSuccessfulPayment(mock *MockBankKeeper, fromAddr []byte, denom string, amount math.Int, moduleName string) *gomock.Call {
	expectedPayment := sdk.NewCoin(denom, amount)
	return mock.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), fromAddr, moduleName, sdk.NewCoins(expectedPayment)).
		Return(nil)
}

// MockFailedPayment mocks a failed bank transfer
func (h *BankMockHelpers) MockFailedPayment(mock *MockBankKeeper, fromAddr []byte, denom string, amount math.Int, moduleName, errMsg string) *gomock.Call {
	expectedPayment := sdk.NewCoin(denom, amount)
	return mock.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), fromAddr, moduleName, sdk.NewCoins(expectedPayment)).
		Return(fmt.Errorf("%s", errMsg))
}

// // MockSendCoins mocks sending coins between accounts
// func (h *BankMockHelpers) MockSendCoins(mock *MockBankKeeper, fromAddr, toAddr []byte, coins sdk.Coins) *gomock.Call {
// 	return mock.EXPECT().
// 		SendCoins(gomock.Any(), fromAddr, toAddr, coins).
// 		Return(nil)
// }

// // MockGetBalance mocks getting an account's balance
// func (h *BankMockHelpers) MockGetBalance(mock *MockBankKeeper, addr []byte, denom string, amount math.Int) *gomock.Call {
// 	return mock.EXPECT().
// 		GetBalance(gomock.Any(), addr, denom).
// 		Return(sdk.NewCoin(denom, amount))
// }

// // MockGetAllBalances mocks getting all balances for an account
// func (h *BankMockHelpers) MockGetAllBalances(mock *MockBankKeeper, addr []byte, coins sdk.Coins) *gomock.Call {
// 	return mock.EXPECT().
// 		GetAllBalances(gomock.Any(), addr).
// 		Return(coins)
// }
