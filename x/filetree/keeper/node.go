package keeper

import (
	"context"

	"nebulix/x/filetree/types"

	"cosmossdk.io/collections"
)

func (k Keeper) SetNode(ctx context.Context, path string, file types.FileTreeNode) error {
	return k.FileTree.Set(ctx, path, file)
}

func (k Keeper) GetNode(ctx context.Context, path string) (types.FileTreeNode, error) {
	return k.FileTree.Get(ctx, path)
}

func (k Keeper) DeleteNode(ctx context.Context, path string) error {
	keys := []string{path}

	rng := new(collections.Range[string]).Prefix(path + "/")
	err := k.IterateNodes(ctx, rng, func(key string, _ types.FileTreeNode) bool {
		keys = append(keys, key)
		return false
	})
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := k.FileTree.Remove(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) IterateNodes(ctx context.Context, rng collections.Ranger[string], fn func(key string, value types.FileTreeNode) (stop bool)) error {
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
