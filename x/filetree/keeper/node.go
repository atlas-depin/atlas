package keeper

import (
	"context"
	"path/filepath"

	"nebulix/x/filetree/types"

	"cosmossdk.io/collections"
)

func (k Keeper) SetNode(ctx context.Context, path string, file types.FileTreeNode) error {
	key := collections.Join(filepath.Dir(path), filepath.Base(path))
	return k.FileTree.Set(ctx, key, file)
}

func (k Keeper) GetNode(ctx context.Context, path string) (types.FileTreeNode, error) {
	key := collections.Join(filepath.Dir(path), filepath.Base(path))
	return k.FileTree.Get(ctx, key)
}

func (k Keeper) DeleteNode(ctx context.Context, path string) error {
	var keys []collections.Pair[string, string]

	dir := filepath.Dir(path) + "/"

	// Prefix range: iterate all nodes whose parent directory starts with 'basepath'
	rng := new(collections.Range[collections.Pair[string, string]]).Prefix(
		collections.Join(dir, ""),
	)

	err := k.IterateNodes(ctx, rng, func(key collections.Pair[string, string], _ types.FileTreeNode) bool {
		keys = append(keys, key)
		return false
	})
	if err != nil {
		return err
	}

	// Delete all collected keys
	for _, key := range keys {
		if err := k.FileTree.Remove(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) IterateNodes(
	ctx context.Context,
	rng collections.Ranger[collections.Pair[string, string]],
	fn func(key collections.Pair[string, string], value types.FileTreeNode) (stop bool),
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

		if fn(kv.Key, kv.Value) {
			break
		}
	}

	return nil
}
