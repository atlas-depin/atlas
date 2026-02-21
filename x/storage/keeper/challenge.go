package keeper

import (
	"context"
	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
)

func (k Keeper) SetChallenge(ctx context.Context, challenge types.StorageChallenge) error {
	// Save challenge
	key := collections.Join(challenge.Provider, challenge.ChallengeId)
	if err := k.ChallengeStore.Set(ctx, key, challenge); err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetChallenge(ctx context.Context, provider string, challengeId string) (types.StorageChallenge, error) {
	// Save challenge
	key := collections.Join(provider, challengeId)
	return k.ChallengeStore.Get(ctx, key)
}

func (k Keeper) RemoveChallenge(ctx context.Context, provider string, challengeId string) error {
	// Save challenge
	key := collections.Join(provider, challengeId)
	return k.ChallengeStore.Remove(ctx, key)
}

func (k Keeper) IterateChallenges(ctx context.Context, fn func(file types.StorageChallenge) (stop bool)) error {
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
