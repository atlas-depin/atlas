package keeper

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
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
	PaymentStore        collections.Map[string, types.StoragePaymentInfo]
	StorageAccountStore collections.Map[string, types.StorageAccountInfo]
	FileStore           collections.Map[string, types.File]      // file_id -> File
	BinStore            collections.Map[string, types.Bin]       // file_id -> File
	ProviderStore       collections.Map[string, types.Provider]  // address -> Provider
	ChallengeStore      collections.Map[string, types.Challenge] // challenge_id -> Challenge

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
		// PaymentStore: collections.NewMap(
		// 	sb, collections.NewPrefix(0), "subscriptions",
		// 	collections.StringKey,
		// 	codec.CollValue[types.StoragePaymentInfo](cdc),
		// ),

		FileStore: collections.NewMap(
			sb, collections.NewPrefix(0), "files",
			collections.StringKey,
			codec.CollValue[types.File](cdc),
		),

		ProviderStore: collections.NewMap(
			sb, collections.NewPrefix(1), "providers",
			collections.StringKey,
			codec.CollValue[types.Provider](cdc),
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

// GetBinRoot retrieves the current Merkle Root of a specific SMT bin.
func (k Keeper) GetBinRoot(ctx sdk.Context, binId uint64) []byte {

	bin, err := k.BinStore.Get(ctx, string(types.GetBinRootKey(uint16(binId))))
	if err != nil {
		// [TBD]: is this what I should do?
		return make([]byte, 32)
	}
	return bin.Root
}

// SetBinRoot sets new merkle root for a specific SMT bin.
func (k Keeper) SetBinRoot(ctx sdk.Context, binId uint64, root []byte) {

	binKey := string(types.GetBinRootKey(uint16(binId)))
	bin, err := k.BinStore.Get(ctx, binKey)
	if err != nil {
		// [TODO]: error - bin does not exist
		return
	}

	bin.Root = root
	k.BinStore.Set(ctx, binKey, bin)
}

// CalculateBinID deterministically maps a file ID to a bin ID.
func CalculateBinID(fileID string) uint16 {
	fileHash := sha256.Sum256([]byte(fileID))
	fileHashInt := new(big.Int).SetBytes(fileHash[:])

	return uint16(fileHashInt.Mod(fileHashInt, big.NewInt(MAX_BINS)).Uint64())
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}
