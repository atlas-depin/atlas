package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"atlas/x/storage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExpireFile removes a file from the chain when its subscription has expired.
//
// It compensates providers for the time between the last reward distribution
// and the subscription expiry, then cleans up the file record.
//
// Called automatically from ProveFile and AttemptChallenge when
// subscription.End < blockTime.
func (k Keeper) ExpireFile(ctx context.Context, file types.File) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Skip staged & dead files
	if file.Status == "STAGED" {
		if err := k.RemoveFile(ctx, file.Fid); err != nil {
			return err
		}
		return nil
	}

	// Get the last reward distribution time for CreditDelta calculation
	lastRewardsTime, err := k.LastRewardsTime.Get(ctx)
	if err != nil {
		lastRewardsTime = 0
	}

	// Compute the time delta: the period from the last reward to the subscription expiry.
	// All files under the same subscription use the same reference point (subscription.End)
	// regardless of which block the batch processor happens to reach them in.
	if file.Status != "DEAD" {
		var delta int64
		sub, _ := k.GetStorageSubscription(ctx, file.Creator, file.Subscription)

		delta = sub.End.Unix() - lastRewardsTime
		if delta < 0 {
			delta = 0
		}
		// Free provider space and compensate for the period since last reward
		for _, provider := range file.Providers {
			if err := k.DecrementProviderSpaceUsed(ctx, provider, file.FileSize, delta); err != nil {
				if err != types.ErrProviderNotFound {
					return err
				}
			}
		}
	}

	// Remove the file record
	if err := k.RemoveFile(ctx, file.Fid); err != nil {
		return err
	}

	// Decrement global file count
	if err := k.DecrementFileCount(ctx); err != nil {
		return err
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExpireFile,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyFile, file.Fid),
			sdk.NewAttribute(types.AttributeKeySubscriptionId, file.Subscription),
		),
	)

	return nil
}

// ExpireSubscription removes all files under a subscription and cleans up
// the subscription record itself.
//
// Because ExpireFile modifies the FileStore (RemoveFile), we first collect
// all file IDs under the subscription, then expire them one by one.
func (k Keeper) ExpireSubscription(ctx context.Context, creator, subscriptionID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// 1. Collect all file IDs under this subscription
	rng := collections.NewSuperPrefixedTripleRange[string, string, string](creator, subscriptionID)
	iter, err := k.FileStore.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()

	var fileIDs []string
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		fileIDs = append(fileIDs, kv.Value.Fid)
	}

	// 2. Expire each file (safe because we're not iterating the store anymore)
	for _, fid := range fileIDs {
		file, err := k.GetFile(ctx, fid)
		if err != nil {
			continue // already removed by a concurrent ProveFile expiry
		}
		if err := k.ExpireFile(ctx, file); err != nil {
			return err
		}
	}

	// 3. Remove the subscription record
	if err := k.ExpireStorageSubscription(ctx, creator, subscriptionID); err != nil {
		// subscription might already be gone — that's fine
		if err != collections.ErrNotFound {
			return err
		}
	}

	// 4. Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExpireFile,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeySubscriptionId, subscriptionID),
		),
	)

	return nil
}
