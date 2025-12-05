package testutils

import (
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperTestSuiteExtensions struct {
	AddressCodec address.Codec
}

func (suite *KeeperTestSuiteExtensions) CreateTestAddress(index int) (string, error) {
	addr := sdk.AccAddress([]byte{byte(index)})
	return suite.AddressCodec.BytesToString(addr)
}
