package keeper

import (
	"errors"
	"nebulix/x/storage/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ResetStorageCredits(ctx sdk.Context, address string, credits math.Int) error {
	// [TBD]: send existing credits to dump here?
	return nil
	account, err := k.StorageAccountStore.Get(ctx, address)
	if err != nil {
		account = types.StorageAccountInfo{Credits: math.ZeroInt()}
	}
	account.Credits = credits

	return k.StorageAccountStore.Set(ctx, address, account)
}

// [TODO]: add storage credits when upgrading subscription
func (k Keeper) AddStorageCredits(ctx sdk.Context, address string, accountInfo types.StorageAccountInfo) error {
	return errors.New("Not Implemented")
}

func (k Keeper) TransferStorageCredits(ctx sdk.Context, sender string, receiver string, amount math.Int) error {
	accSender, err := k.StorageAccountStore.Get(ctx, sender)
	if err != nil {
		return err
	}

	accReceiver, err := k.StorageAccountStore.Get(ctx, receiver)
	if err != nil {
		return err
	}

	accSender.Credits = accSender.Credits.Sub(amount)
	accReceiver.Credits, err = accReceiver.Credits.SafeAdd(amount)
	if err != nil {
		return err
	}

	return nil
}
