package keeper

import (
	"fmt"
	"nebulix/x/storage/types"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func (k Keeper) SetStorageSubscription(ctx sdk.Context, address string, subscription types.StorageSubscription) error {
	key := collections.Join(address, subscription.Id)
	// [TODO]: add things that ensure the
	return k.SubscriptionStore.Set(ctx, key, subscription)
}

func (k Keeper) GetStorageSubscription(ctx sdk.Context, address string, subscriptionId string) (types.StorageSubscription, error) {
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

func (k Keeper) GetDefaultStorageSubscriptionId(ctx sdk.Context, address string) (string, error) {
	return k.DefaultSubscriptionIndex.Get(ctx, address)
}

func (k Keeper) SetDefaultStorageSubscriptionId(ctx sdk.Context, address string, subscriptionId string) error {
	return k.DefaultSubscriptionIndex.Set(ctx, address, subscriptionId)
}

// helpers
func (k Keeper) UpdateStorageSubscriptionUsage(ctx sdk.Context, address string, subscriptionId string, bytes int64, replicas int64) (string, error) {
	subscription, err := k.GetStorageSubscription(ctx, address, subscriptionId)
	if err != nil {
		return "", err
	}
	if subscription.End.Before(ctx.BlockTime()) {
		// [TBD]: can I handle storage account expiry here? appstate remains unchanged on error so, idk how
		// [TODO]: better error handling
		return "", sdkerrors.ErrUnauthorized
	}

	subscription.SpaceUsed += bytes * replicas
	subscription.ReplicaSpaceUsed += bytes * (replicas - 1)

	if subscription.SpaceUsed > subscription.SpaceAvailable {
		return "", sdkerrors.ErrInvalidRequest
		// [TODO]: better error handling
		// return nil, errors.Wrapf(sdkerrors.ErrUnauthorized, "storage account does not have enough space available %d > %d", subscription.SpaceUsed, subscription.SpaceAvailable)
	}

	currentCreditPool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return "", err
	}
	credits := math.NewInt(bytes * replicas).Mul(math.NewInt(int64(subscription.End.Sub(ctx.BlockTime()) / time.Second)))

	subscription.Credits = subscription.Credits.Sub(credits)
	k.StorageCreditPool.Set(ctx, currentCreditPool.Add(credits))

	return subscription.Id, k.SetStorageSubscription(ctx, address, subscription)
}
