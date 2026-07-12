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

	if file.Status == "DEAD" {
		return nil
	}

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
	sub, err := k.GetStorageSubscription(ctx, file.Creator, file.Subscription)
	if err == nil {
		// Subscription still exists.. somehow
		delta := sub.End.Unix() - lastRewardsTime
		if delta < 0 {
			delta = 0
		}
		// Free provider space and compensate for the period since last reward.
		for _, provider := range file.Providers {
			if err := k.DecrementProviderSpaceUsed(ctx, provider, file.FileSize, delta); err != nil {
				if err != types.ErrProviderNotFound {
					return err
				}
			}

			// Remove this provider's bytes from the per-subscription tracker.
			subProvKey := collections.Join(file.Subscription, provider)
			currentBytes, getErr := k.SubscriptionProviderBytes.Get(ctx, subProvKey)
			if getErr == nil {
				newBytes := currentBytes - file.FileSize
				if newBytes <= 0 {
					_ = k.SubscriptionProviderBytes.Remove(ctx, subProvKey)
				} else {
					_ = k.SubscriptionProviderBytes.Set(ctx, subProvKey, newBytes)
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

// ExpireSubscriptions scans the expiry index for subscriptions whose
// End time has passed and expires them. Called from EndBlocker every block.
func (k Keeper) ExpireSubscriptions(ctx sdk.Context) error {
	now := ctx.BlockTime().Unix()

	rng := collections.NewPrefixUntilTripleRange[int64, string, string](now)
	iter, err := k.SubscriptionExpiryIndex.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		triple, err := iter.Key()
		if err != nil {
			ctx.Logger().Error("Failed to decode expiry index key", "err", err)
			continue
		}

		owner := triple.K3()
		subID := triple.K2()

		// Compute the time window between the last reward and the subscription expiry
		endUnix := triple.K1()
		lastRewardTime, _ := k.LastRewardsTime.Get(ctx)
		delta := endUnix - lastRewardTime
		if delta < 0 {
			delta = 0
		}

		// Free provider space immediately using the per-subscription byte tracker.
		// O(providers_per_sub) instead of O(files).
		subProvRange := collections.NewPrefixedPairRange[string, string](subID)
		spIter, spErr := k.SubscriptionProviderBytes.Iterate(ctx, subProvRange)
		if spErr == nil {
			for ; spIter.Valid(); spIter.Next() {
				kv, kvErr := spIter.KeyValue()
				if kvErr != nil {
					continue
				}
				provider := kv.Key.K2()
				bytes := kv.Value
				prov, provErr := k.GetProvider(ctx, provider)
				if provErr == nil {
					prov.SpaceUsed -= bytes
					if prov.SpaceUsed < 0 {
						prov.SpaceUsed = 0
					}
					prov.CreditDelta += bytes * delta
					_ = k.SetProvider(ctx, prov)
				}
				_ = k.SubscriptionProviderBytes.Remove(ctx, kv.Key)
			}
			spIter.Close()
		}

		// ExpireStorageSubscription removes the subscription record and
		// cleans up the expiry index entry.
		if err := k.ExpireStorageSubscription(ctx, owner, subID); err != nil {
			ctx.Logger().Error("Failed to expire subscription", "owner", owner, "sub_id", subID, "err", err)
			continue
		}
		ctx.Logger().Info("Subscription expired", "owner", owner, "sub_id", subID)
	}

	return nil
}
