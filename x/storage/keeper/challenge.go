package keeper

import (
	"atlas/x/storage/types"
	"context"

	"cosmossdk.io/collections"
)

/*
Set a new challenge in the ChallengeStore.
*/
func (k Keeper) SetChallenge(ctx context.Context, challenge types.StorageChallenge) error {
	key := collections.Join(challenge.Provider, challenge.ChallengeId)
	if err := k.ChallengeStore.Set(ctx, key, challenge); err != nil {
		return err
	}
	return nil
}

/*
Get a proof challenge by provider & challenge ID from the ChallengeStore.
*/
func (k Keeper) GetChallenge(ctx context.Context, provider string, challengeId string) (types.StorageChallenge, error) {
	key := collections.Join(provider, challengeId)
	return k.ChallengeStore.Get(ctx, key)
}

/*
Remove a proof challenge from the ChallengeStore.
*/
func (k Keeper) RemoveChallenge(ctx context.Context, provider string, challengeId string) error {
	key := collections.Join(provider, challengeId)
	return k.ChallengeStore.Remove(ctx, key)
}

/*
Iterate over all challenges and run the provided function argument for each value record.
The provided function must have a `types.StorageChallenge` argument.
*/
func (k Keeper) IterateChallenges(ctx context.Context, fn func(challenge types.StorageChallenge) (stop bool)) error {
	iter, err := k.ChallengeStore.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if fn(kv.Value) {
			break
		}
	}

	return nil
}
