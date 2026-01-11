package keeper

import (
	"context"
	"path/filepath"

	"nebulix/x/filetree/types"

	"cosmossdk.io/collections"
)

func (k Keeper) SetNode(ctx context.Context, owner string, path string, file types.FileNode) error {
	key := collections.Join3(owner, filepath.Dir(path), filepath.Base(path))
	return k.FileTree.Set(ctx, key, file)
}

func (k Keeper) GetNode(ctx context.Context, owner string, path string) (types.FileNode, error) {
	key := collections.Join3(owner, filepath.Dir(path), filepath.Base(path))
	return k.FileTree.Get(ctx, key)
}

// [NOTE]: to be stress-tested, may have to default to regular DeleteNode
func (k Keeper) RemoveNode(ctx context.Context, owner string, path string) error {
	rng := new(collections.Range[collections.Triple[string, string, string]]).Prefix(
		collections.Join3(owner, path, ""),
	)

	// this is "safe"
	return k.IterateNodes(ctx, rng, func(key collections.Triple[string, string, string], _ types.FileNode) (bool, error) {
		if err := k.FileTree.Remove(ctx, key); err != nil {
			return true, err
		}
		return false, nil
	})
}

func (k Keeper) IterateNodes(
	ctx context.Context,
	rng collections.Ranger[collections.Triple[string, string, string]],
	fn func(key collections.Triple[string, string, string], value types.FileNode) (stop bool, err error),
) error {
	iter, err := k.FileTree.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		end, err := fn(kv.Key, kv.Value)
		if err != nil {
			return err
		}
		if end {
			break
		}
	}

	return nil
}
