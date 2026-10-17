package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protov2 "google.golang.org/protobuf/proto"

	"atlas/x/storage/ante"
	"atlas/x/storage/types"
)

// fakeFeeTx implements sdk.FeeTx for testing the ProveFileExemptChecker.
type fakeFeeTx struct {
	msgs     []sdk.Msg
	fee      sdk.Coins
	gas      uint64
	feePayer sdk.AccAddress
}

func (t fakeFeeTx) GetMsgs() []sdk.Msg                    { return t.msgs }
func (t fakeFeeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (t fakeFeeTx) GetGas() uint64                        { return t.gas }
func (t fakeFeeTx) GetFee() sdk.Coins                     { return t.fee }
func (t fakeFeeTx) FeePayer() []byte                      { return t.feePayer }
func (t fakeFeeTx) FeeGranter() []byte                    { return nil }

func (suite *KeeperTestSuite) TestFeeChecker_NonFeeTx() {
	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)

	_, _, err := checker.Check(suite.ctx, nil)
	suite.Require().ErrorContains(err, "Tx must be a FeeTx")
}

func (suite *KeeperTestSuite) TestFeeChecker_ProveFile_Exempt() {
	provider := suite.CreateTestAddress(0)
	suite.createTestProvider(provider, 1_000_000)
	providerAcc := sdk.MustAccAddressFromBech32(provider)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgProveFile{Creator: provider}},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000_000))),
		gas:      1_000_000,
		feePayer: providerAcc,
	}

	fee, priority, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Empty(fee, "registered providers pay zero fees for ProveFile")
	suite.Require().Equal(int64(0), priority)
}

func (suite *KeeperTestSuite) TestFeeChecker_RateLimit() {
	provider := suite.CreateTestAddress(0)
	suite.createTestProvider(provider, 1_000_000)
	providerAcc := sdk.MustAccAddressFromBech32(provider)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 1)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgProveFile{Creator: provider}},
		feePayer: providerAcc,
	}

	_, _, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)

	_, _, err = checker.Check(suite.ctx, tx)
	suite.Require().ErrorContains(err, "rate limit: max 1 ProveFile txs per block per provider")
}

func (suite *KeeperTestSuite) TestFeeChecker_RateLimit_ResetsPerBlock() {
	provider := suite.CreateTestAddress(0)
	suite.createTestProvider(provider, 1_000_000)
	providerAcc := sdk.MustAccAddressFromBech32(provider)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 1)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgProveFile{Creator: provider}},
		feePayer: providerAcc,
	}

	_, _, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	_, _, err = checker.Check(suite.ctx, tx)
	suite.Require().NoError(err, "counter resets on a new block height")
}

func (suite *KeeperTestSuite) TestFeeChecker_NotProvider_FallsThrough() {
	// Payer is not a registered provider -> default fee logic applies.
	nonProvider := suite.CreateTestAddress(1)
	nonProviderAcc := sdk.MustAccAddressFromBech32(nonProvider)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgProveFile{Creator: nonProvider}},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000))),
		gas:      100,
		feePayer: nonProviderAcc,
	}

	fee, priority, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Equal(tx.fee, fee)
	suite.Require().Equal(int64(10), priority) // 1000uatl / 100 gas
}

func (suite *KeeperTestSuite) TestFeeChecker_MixedMsgs_FallsThrough() {
	provider := suite.CreateTestAddress(0)
	suite.createTestProvider(provider, 1_000_000)
	providerAcc := sdk.MustAccAddressFromBech32(provider)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs: []sdk.Msg{
			&types.MsgProveFile{Creator: provider},
			&types.MsgDeleteFile{Creator: provider, FileId: "some-file"},
		},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000))),
		gas:      100,
		feePayer: providerAcc,
	}

	fee, _, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Equal(tx.fee, fee, "mixed txs are not exempt")
}

func (suite *KeeperTestSuite) TestFeeChecker_CheckTx_MinGasPrice() {
	provider := suite.CreateTestAddress(0)
	suite.createTestProvider(provider, 1_000_000)
	providerAcc := sdk.MustAccAddressFromBech32(provider)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)

	// Mempool check context with a 1uatl/gas minimum price.
	memCtx := suite.ctx.
		WithIsCheckTx(true).
		WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin("uatl", math.NewInt(1))))

	// Not a ProveFile tx -> default fee check; 5uatl for 1000 gas is below the floor.
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgDeleteFile{Creator: provider, FileId: "f"}},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(5))),
		gas:      1_000,
		feePayer: providerAcc,
	}

	_, _, err := checker.Check(memCtx, tx)
	suite.Require().ErrorContains(err, "insufficient fees")
}

func (suite *KeeperTestSuite) TestFeeChecker_UpdatePrice_Exempt() {
	updater := suite.CreateTestAddress(0)
	params := types.DefaultParams()
	params.PriceUpdater = updater
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))
	updaterAcc := sdk.MustAccAddressFromBech32(updater)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgUpdatePrice{Creator: updater, Price: "1.5"}},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000_000))),
		gas:      1_000_000,
		feePayer: updaterAcc,
	}

	fee, priority, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Empty(fee, "price updater pays zero fees for UpdatePrice")
	suite.Require().Equal(int64(0), priority)
}

func (suite *KeeperTestSuite) TestFeeChecker_UpdatePrice_NotUpdater_FallsThrough() {
	updater := suite.CreateTestAddress(0)
	other := suite.CreateTestAddress(1)
	params := types.DefaultParams()
	params.PriceUpdater = updater
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))
	otherAcc := sdk.MustAccAddressFromBech32(other)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgUpdatePrice{Creator: other, Price: "1.5"}},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000))),
		gas:      100,
		feePayer: otherAcc,
	}

	fee, priority, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Equal(tx.fee, fee, "non-updater is not exempt")
	suite.Require().Equal(int64(10), priority) // 1000uatl / 100 gas
}

func (suite *KeeperTestSuite) TestFeeChecker_UpdatePrice_NoUpdater_FallsThrough() {
	updater := suite.CreateTestAddress(0)
	params := types.DefaultParams()
	params.PriceUpdater = "" // no updater configured
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))
	updaterAcc := sdk.MustAccAddressFromBech32(updater)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs:     []sdk.Msg{&types.MsgUpdatePrice{Creator: updater, Price: "1.5"}},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000))),
		gas:      100,
		feePayer: updaterAcc,
	}

	fee, _, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Equal(tx.fee, fee, "no configured updater -> default fee")
}

func (suite *KeeperTestSuite) TestFeeChecker_UpdatePrice_MixedMsgs_FallsThrough() {
	updater := suite.CreateTestAddress(0)
	params := types.DefaultParams()
	params.PriceUpdater = updater
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))
	updaterAcc := sdk.MustAccAddressFromBech32(updater)

	checker := ante.NewProveFileExemptChecker(&suite.keeper, 100)
	tx := fakeFeeTx{
		msgs: []sdk.Msg{
			&types.MsgUpdatePrice{Creator: updater, Price: "1.5"},
			&types.MsgDeleteFile{Creator: updater, FileId: "f"},
		},
		fee:      sdk.NewCoins(sdk.NewCoin("uatl", math.NewInt(1_000))),
		gas:      100,
		feePayer: updaterAcc,
	}

	fee, _, err := checker.Check(suite.ctx, tx)
	suite.Require().NoError(err)
	suite.Require().Equal(tx.fee, fee, "mixed txs are not exempt")
}
