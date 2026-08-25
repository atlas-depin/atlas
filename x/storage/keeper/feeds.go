package keeper

import (
	"context"

	"atlas/x/storage/types"
)

// SetFeed sets a feed keyed by its name.
func (k Keeper) SetFeed(ctx context.Context, feed types.Feed) error {
	return k.Feeds.Set(ctx, feed.Name, feed)
}

// GetFeed returns a feed by name.
func (k Keeper) GetFeed(ctx context.Context, name string) (types.Feed, error) {
	return k.Feeds.Get(ctx, name)
}

// RemoveFeed removes a feed by name.
func (k Keeper) RemoveFeed(ctx context.Context, name string) error {
	return k.Feeds.Remove(ctx, name)
}

// GetAllFeeds returns all feeds.
func (k Keeper) GetAllFeeds(ctx context.Context) ([]types.Feed, error) {
	feeds := make([]types.Feed, 0)
	iter, err := k.Feeds.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, kv.Value)
	}

	return feeds, nil
}
