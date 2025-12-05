package keeper

import (
	"fmt"
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetNewStorageSubscription(ctx sdk.Context, address string, subscription types.StorageSubscription) (string, error) {
	seq, err := k.SubscriptionSequence.Next(ctx)
	if err != nil {
		return "", err
	}

	subscriptionId := fmt.Sprintf("sub_%d", seq)
	key := collections.Join(address, subscriptionId)
	return subscriptionId, k.SubscriptionStore.Set(ctx, key, subscription)
}

func (k Keeper) SetStorageSubscription(ctx sdk.Context, address string, subscriptionId string, subscription types.StorageSubscription) error {
	key := collections.Join(address, subscriptionId)
	// [TODO]: add things that ensure the
	return k.SubscriptionStore.Set(ctx, key, subscription)
}

func (k Keeper) GetStorageSubscription(ctx sdk.Context, address string, subscriptionId string) (types.StorageSubscription, error) {
	key := collections.Join(address, subscriptionId)
	return k.SubscriptionStore.Get(ctx, key)
}
