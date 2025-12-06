package keeper

import (
	"context"
	"errors"
	"nebulix/x/storage/types"
)

// \\ RegisterProvider registers a new storage provider
func (k Keeper) SetProvider(ctx context.Context, provider types.Provider) error {
	exists, err := k.ProviderStore.Has(ctx, provider.Address)
	if err != nil {
		return err
	}
	if exists {
		// [TODO]: better error handling
		return errors.New("provider already exists")
	}

	return k.ProviderStore.Set(ctx, provider.Address, provider)
}
