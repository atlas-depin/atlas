package keeper

import (
	"context"
	"fmt"
	"nebulix/x/storage/types"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetNewStorageSubscription(ctx sdk.Context, address string, subscription types.StorageSubscription) (string, error) {
	seq, err := k.SubscriptionSequence.Next(ctx)
	if err != nil {
		return "", err
	}

	subscription.Id = fmt.Sprintf("sub_%d", seq)
	key := collections.Join(address, subscription.Id)
	return subscription.Id, k.SubscriptionStore.Set(ctx, key, subscription)
}

func (k Keeper) SetStorageSubscription(ctx context.Context, address string, subscription types.StorageSubscription) error {
	key := collections.Join(address, subscription.Id)
	// [TODO]: add things that ensure the
	return k.SubscriptionStore.Set(ctx, key, subscription)
}

func (k Keeper) GetStorageSubscription(ctx context.Context, address string, subscriptionId string) (types.StorageSubscription, error) {
	if subscriptionId == "" {
		id, err := k.GetDefaultStorageSubscriptionId(ctx, address)
		if err != nil {
			return types.StorageSubscription{}, err
		} else {
			subscriptionId = id
		}
	}

	key := collections.Join(address, subscriptionId)
	return k.SubscriptionStore.Get(ctx, key)
}

func (k Keeper) ExpireStorageSubscription(ctx context.Context, address string, subscriptionId string) error {
	key := collections.Join(address, subscriptionId)

	defaultId, _ := k.GetDefaultStorageSubscriptionId(ctx, address)
	if defaultId == subscriptionId {
		_ = k.DefaultSubscriptionIndex.Remove(ctx, address)
	}

	return k.SubscriptionStore.Remove(ctx, key)
}

func (k Keeper) GetDefaultStorageSubscriptionId(ctx context.Context, address string) (string, error) {
	return k.DefaultSubscriptionIndex.Get(ctx, address)
}

func (k Keeper) SetDefaultStorageSubscriptionId(ctx context.Context, address string, subscriptionId string) error {
	return k.DefaultSubscriptionIndex.Set(ctx, address, subscriptionId)
}

/*
Update the usage of a storage subscription, incremented by the amount of bytes multiplied by the number of replicas.
*/
func (k Keeper) UpdateStorageSubscriptionUsage(ctx context.Context, address string, subscriptionId string, bytes int64, replicas int32) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	subscription, err := k.GetStorageSubscription(ctx, address, subscriptionId)
	if err != nil {
		if err == collections.ErrNotFound {
			return subscriptionId, types.ErrSubscriptionNotFound
		} else {
			return subscriptionId, err
		}
	}

	if subscription.End.Before(sdkCtx.BlockTime()) {
		k.ExpireStorageSubscription(ctx, address, subscriptionId)
		return subscription.Id, types.ErrSubscriptionExpired
	}

	subscription.SpaceUsed += bytes * int64(replicas)
	subscription.ReplicaSpaceUsed += bytes * (int64(replicas) - 1)
	if subscription.SpaceUsed > subscription.SpaceAvailable {
		return subscription.Id, types.ErrSubscriptionFull
	}

	credits := math.NewInt(bytes * int64(replicas)).Mul(math.NewInt(int64(subscription.End.Sub(sdkCtx.BlockTime()) / time.Second)))
	subscription.Credits = subscription.Credits.Sub(credits)
	if subscription.Credits.LT(math.ZeroInt()) {
		return subscription.Id, types.ErrSubscriptionInsufficientCredit
	}

	return subscription.Id, k.SetStorageSubscription(ctx, address, subscription)
}
