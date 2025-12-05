package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	modtestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"nebulix/testutils"
	"nebulix/x/storage/keeper"
	module "nebulix/x/storage/module"
	"nebulix/x/storage/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx          sdk.Context
	keeper       keeper.Keeper
	bankKeeper   *testutils.MockBankKeeper
	authKeeper   *testutils.MockAuthKeeper
	addressCodec address.Codec
	storeKey     *storetypes.KVStoreKey
	ctrl         *gomock.Controller

	helpers *testutils.MockHelpers
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	encCfg := modtestutil.MakeTestEncodingConfig(module.AppModule{})

	suite.storeKey = storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(suite.storeKey)
	authority := authtypes.NewModuleAddress(types.GovModuleName)

	suite.helpers = testutils.NewMockHelpers()
	suite.bankKeeper = testutils.NewMockBankKeeper(suite.ctrl)
	suite.authKeeper = testutils.NewMockAuthKeeper(suite.ctrl)

	suite.addressCodec = addresscodec.NewBech32Codec("cosmos")
	suite.keeper = keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		suite.addressCodec,
		authority,
		suite.bankKeeper,
		suite.authKeeper,
	)

	suite.ctx = sdktestutil.DefaultContextWithDB(suite.T(), suite.storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	// Initialize params
	if err := suite.keeper.Params.Set(suite.ctx, types.DefaultParams()); err != nil {
		suite.T().Fatalf("failed to set params: %v", err)
	}
}

func (suite *KeeperTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *KeeperTestSuite) setupMsgServerWithParams(params types.Params) (types.MsgServer, sdk.Context) {
	// Initialize genesis state
	types.DefaultGenesis()

	// Set params
	err := suite.keeper.Params.Set(suite.ctx, params)
	suite.Require().NoError(err)

	// Return the message server and context
	return keeper.NewMsgServerImpl(suite.keeper), suite.ctx
}

// setupMsgServer is the helper function that sets up everything needed for message server tests
func (suite *KeeperTestSuite) setupMsgServer() (types.MsgServer, sdk.Context) {
	return suite.setupMsgServerWithParams(types.DefaultParams())
}

func (suite *KeeperTestSuite) CreateTestAddress(index int) string {
	addr := sdk.AccAddress([]byte{byte(index)})
	addrStr, err := suite.addressCodec.BytesToString(addr)
	suite.Require().NoError(err)
	return addrStr
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
