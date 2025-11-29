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

	"nebulix/x/storage/keeper"
	module "nebulix/x/storage/module"
	mocks "nebulix/x/storage/testutil"
	"nebulix/x/storage/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx          sdk.Context
	keeper       keeper.Keeper
	bankKeeper   *mocks.MockBankKeeper
	authKeeper   *mocks.MockAuthKeeper
	addressCodec address.Codec
	storeKey     *storetypes.KVStoreKey
	ctrl         *gomock.Controller
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.reset()
}

func (suite *KeeperTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *KeeperTestSuite) reset() {
	encCfg := modtestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec("cosmos")
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	ctx := sdktestutil.DefaultContextWithDB(suite.T(), storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	// Initialize gomock keepers
	bankKeeper := mocks.NewMockBankKeeper(suite.ctrl)
	authKeeper := mocks.NewMockAuthKeeper(suite.ctrl)

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
		bankKeeper,
		authKeeper,
	)

	// Initialize params
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		suite.T().Fatalf("failed to set params: %v", err)
	}

	suite.ctx = ctx
	suite.keeper = k
	suite.bankKeeper = bankKeeper
	suite.authKeeper = authKeeper
	suite.addressCodec = addressCodec
	suite.storeKey = storeKey
}

// setupMsgServer is the helper function that sets up everything needed for message server tests
func (suite *KeeperTestSuite) setupMsgServer() (types.MsgServer, sdk.Context) {
	// Initialize genesis state
	types.DefaultGenesis()

	// Set default parameters
	params := types.DefaultParams()
	err := suite.keeper.Params.Set(suite.ctx, params)
	suite.Require().NoError(err)

	// Create test addresses for use in tests
	// testAddresses := suite.createTestAddresses(3)

	// Update params with test addresses if needed
	// params.DepositAccount = testAddresses[0]
	err = suite.keeper.Params.Set(suite.ctx, params)
	suite.Require().NoError(err)

	// Return the message server and context
	return keeper.NewMsgServerImpl(suite.keeper), suite.ctx
}

// [TODO]: make this accessible from testutils module
func (suite *KeeperTestSuite) createTestAddresses(count int) []string {
	addresses := make([]string, count)
	for i := 0; i < count; i++ {
		addr := sdk.AccAddress([]byte{byte(i + 1), byte(i + 1), byte(i + 1), byte(i + 1), byte(i + 1)})
		addrStr, err := suite.addressCodec.BytesToString(addr)
		suite.Require().NoError(err)
		addresses[i] = addrStr
	}
	return addresses
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
