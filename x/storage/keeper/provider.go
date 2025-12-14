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

func (k Keeper) GetProvider(ctx context.Context, address string) (types.Provider, error) {
	return k.ProviderStore.Get(ctx, address)
}

func (k Keeper) GetAllProviders(ctx context.Context) ([]*types.Provider, error) {
	values := []*types.Provider{}

	iter, err := k.ProviderStore.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}

	for ; iter.Valid(); iter.Next() {
		provider, err := iter.Value()
		if err != nil {
			return values, err // might as well get some values + an error instead of none?
		}

		values = append(values, &provider)
	}
	return values, err
}
