package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"nebulix/x/storage/types"
)

// [NOTE]: I know I'm gonna forget why I did this, so the BIN_DEPTH is 65,536 leaves per bin and the MAX_BINS is set to support 100 million files  (100000000/65536)
const SMT_BIN_DEPTH = 16
const MAX_BINS = 1526

// [IMPORTANT] Bin overflow can occur, in this case respond with an error forcing the storage provider to add a byte or two
// to change the merkle hash of the file. Chain upgrade can increase bin limit if there is an issue.
// Versioned bins will be necessary for this kind of upgrade, otherwise the algo gets skewed and you'll no longer be able to deterministically
// find previously uploaded files' bin.

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	// Collections
	StorageCreditPool collections.Item[math.Int]
	LastRewardsTime   collections.Item[time.Time]

	SubscriptionStore        collections.Map[collections.Pair[string, string], types.StorageSubscription]
	DefaultSubscriptionIndex collections.Map[string, string]
	//// [TBD]: I added this to enable having multiple subscriptions created for the same account in the same block, but idk if its the best way.
	// Can't use block height or timestamp because then you would have a conflict of subscription IDs.
	SubscriptionSequence collections.Sequence
	////
	FileStagingStore collections.Map[string, types.StagedFile]
	FileStore        *collections.IndexedMap[string, types.File, fileIndexes]

	BinStore       collections.Map[string, types.Bin]      // file_id -> File
	ProviderStore  collections.Map[string, types.Provider] // address -> Provider
	ChallengeStore collections.Map[collections.Pair[string, string], types.StorageChallenge]

	// Indicies
	FilesByProvider collections.KeySet[string] // provider_address|file_id -> {}

	bankKeeper types.BankKeeper
	authKeeper types.AuthKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	bankKeeper types.BankKeeper,
	authKeeper types.AuthKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		bankKeeper: bankKeeper,
		authKeeper: authKeeper,
		Params:     collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),

		// TODO: use keys/prefix instead of NewPrefix(<int>)
		// [TODO]: how do I set this value to 0 at genesis?
		StorageCreditPool: collections.NewItem(
			sb, collections.NewPrefix(0), "storage_credit_pool",
			sdk.IntValue,
		),

		SubscriptionStore: collections.NewMap(
			sb, collections.NewPrefix(1), "subscriptions",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.StorageSubscription](cdc),
		),
		DefaultSubscriptionIndex: collections.NewMap(
			sb, collections.NewPrefix(2), "default_subscriptions",
			collections.StringKey,
			collections.StringValue,
		),
		SubscriptionSequence: collections.NewSequence(sb, collections.NewPrefix(3), "subscription_seq"),

		FileStagingStore: collections.NewMap(
			sb, collections.NewPrefix(3), "files_staged",
			collections.StringKey,
			codec.CollValue[types.StagedFile](cdc),
		),
		FileStore: collections.NewIndexedMap(
			sb,
			collections.NewPrefix(4), "files",
			collections.StringKey,
			codec.CollValue[types.File](cdc),
			NewFileIndexes(sb),
		),

		ProviderStore: collections.NewMap(
			sb, collections.NewPrefix(5), "providers",
			collections.StringKey,
			codec.CollValue[types.Provider](cdc),
		),

		ChallengeStore: collections.NewMap(
			sb, collections.NewPrefix(1), "challenges",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.StorageChallenge](cdc),
		),

		// FilesByProvider: collections.NewKeySet(
		// 	sb, collections.NewPrefix(3), "files_by_provider",
		// 	collections.PairKeyCodec(collections.StringKey, collections.StringKey),
		// ),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

func (k Keeper) GetPreviousBlockHash(ctx context.Context) []byte {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	header := sdkCtx.BlockHeader()
	return header.LastBlockId.Hash
}
