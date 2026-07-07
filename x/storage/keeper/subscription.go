package keeper

import (
	"atlas/x/storage/types"
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//
// === Subscription CRUD =====
//

/*
Set a new storage subscription record for the specified address.
This method generates and sets the subscription id before commiting the subscription record.
*/
func (k Keeper) SetNewStorageSubscription(ctx sdk.Context, address string, subscription types.StorageSubscription) (string, error) {
	seq, err := k.SubscriptionSequence.Next(ctx)
	if err != nil {
		return "", err
	}
	subscription.Id = fmt.Sprintf("sub_%d", seq)

	if err := k.AddExpiryIndexEntry(ctx, subscription.Id, address, subscription.End.Unix()); err != nil {
		return "", err
	}

	return subscription.Id, k.SetStorageSubscription(ctx, address, subscription)
}

/*
Set a storage subscription record at the key specified by its owner address and subscription id.
*/
func (k Keeper) SetStorageSubscription(ctx context.Context, address string, subscription types.StorageSubscription) error {
	key := collections.Join(address, subscription.Id)
	return k.SubscriptionStore.Set(ctx, key, subscription)
}

/*
Get a storage subscription by owner address and subscription id.
A blank subscription id prompts this function to get the address' default subscription, if any.
*/
func (k Keeper) GetStorageSubscription(ctx context.Context, address string, subscriptionId string) (subscription types.StorageSubscription, err error) {
	if subscriptionId == "" {
		subscriptionId, err = k.GetDefaultStorageSubscriptionId(ctx, address)
		if err != nil {
			ranger := collections.NewPrefixedPairRange[string, string](address)
			err = k.IterateStorageSubscriptions(ctx, ranger, func(key collections.Pair[string, string], value types.StorageSubscription) bool {
				if time.Now().Before(value.End) {
					subscription = value
					return true
				} else {
					return false
				}
			})

			if subscription.Id == "" {
				return subscription, types.ErrSubscriptionNotFound
			} else {
				return
			}
		}
	}

	key := collections.Join(address, subscriptionId)
	subscription, err = k.SubscriptionStore.Get(ctx, key)
	if err == collections.ErrNotFound {
		err = types.ErrSubscriptionNotFound
	}
	return
}

/*
Iterate over all subscriptions and run the provided function argument for each record.
*/
func (k Keeper) IterateAllStorageSubscriptions(ctx context.Context, fn func(key collections.Pair[string, string], subscription types.StorageSubscription) (stop bool)) error {
	return k.IterateStorageSubscriptions(ctx, &collections.PairRange[string, string]{}, fn)
}

/*
Iterate over a range of subscriptions and run the provided function argument for each record.
*/
func (k Keeper) IterateStorageSubscriptions(ctx context.Context, ranger *collections.PairRange[string, string], fn func(key collections.Pair[string, string], sub types.StorageSubscription) bool) error {
	iter, err := k.SubscriptionStore.Iterate(ctx, ranger)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		if fn(kv.Key, kv.Value) {
			break
		}
	}
	return nil
}

/*
Remove a storage subscription by owner address and subscription id.
If the subscription is the address's default subscription, the default subscription record is erased.
*/
func (k Keeper) ExpireStorageSubscription(ctx context.Context, address string, subscriptionId string) error {
	key := collections.Join(address, subscriptionId)

	// Fetch the sub to get its End and owner for removing the expiry index entry.
	sub, err := k.SubscriptionStore.Get(ctx, key)
	if err == nil {
		_ = k.RemoveExpiryIndexEntry(ctx, subscriptionId, address, sub.End.Unix())
	}

	defaultId, _ := k.GetDefaultStorageSubscriptionId(ctx, address)
	if defaultId == subscriptionId {
		_ = k.ClearDefaultStorageSubscriptionId(ctx, address)
	}

	return k.SubscriptionStore.Remove(ctx, key)
}

//
// === Subscription Utility Methods =====
//

/*
Update the usage of a storage subscription, incremented by the amount of bytes multiplied by the number of replicas.
*/
func (k Keeper) UpdateStorageSubscriptionUsage(ctx context.Context, address string, subscriptionId string, bytes int64, replicas int32) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	subscription, err := k.GetStorageSubscription(ctx, address, subscriptionId)
	if err != nil {
		return subscriptionId, err
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

/*
Increase the specified storage subscription's allocated capacity and/or duration.
*/
func (k Keeper) ExpandStorageSubscription(ctx context.Context, address string, subscriptionId string, size int64, duration int64) error {
	// TODO: implement storage susbcription expansion
	// TODO: verify that there is enough provider capacity to accomodate the new subscription
	return nil
}

//
// === Default Subscription Index CRUD =====
//

/*
Get the subscription id of the specified address' default subscription.
*/
func (k Keeper) GetDefaultStorageSubscriptionId(ctx context.Context, address string) (string, error) {
	return k.DefaultSubscriptionIndex.Get(ctx, address)
}

/*
Set the subscription id of the specified address' default subscription.
*/
func (k Keeper) SetDefaultStorageSubscriptionId(ctx context.Context, address string, subscriptionId string) error {
	return k.DefaultSubscriptionIndex.Set(ctx, address, subscriptionId)
}

/*
Remove the default subscription id of the specified address.
*/
func (k Keeper) ClearDefaultStorageSubscriptionId(ctx context.Context, address string) error {
	return k.DefaultSubscriptionIndex.Remove(ctx, address)
}

//
// === Subscription Expiry Index CRUD =====
//

/*
AddExpiryIndexEntry writes the subscription's expiry time and owner into
the expiry index so the EndBlocker can discover it without scanning
all subscriptions.
*/
func (k Keeper) AddExpiryIndexEntry(ctx context.Context, subId string, owner string, endUnix int64) error {
	return k.SubscriptionExpiryIndex.Set(ctx, collections.Join3(endUnix, subId, owner))
}

/*
RemoveExpiryIndexEntry deletes a subscription from the expiry index.
*/
func (k Keeper) RemoveExpiryIndexEntry(ctx context.Context, subId string, owner string, endUnix int64) error {
	return k.SubscriptionExpiryIndex.Remove(ctx, collections.Join3(endUnix, subId, owner))
}
