package keeper

import (
	"context"
	"path/filepath"
	"strings"

	"errors"

	"nebulix/x/filetree/types"

	"cosmossdk.io/collections"
)

func (k Keeper) SetRootNode(ctx context.Context, owner string, path string, file types.FileNode) error {
	key := collections.Join3(owner, "", path)
	return k.FileTree.Set(ctx, key, file)
}

func (k Keeper) SetNode(ctx context.Context, owner string, path string, file types.FileNode) error {
	key := collections.Join3(owner, filepath.Dir(path)+"/", filepath.Base(path))
	return k.FileTree.Set(ctx, key, file)
}

func (k Keeper) GetNode(ctx context.Context, owner string, path string) (types.FileNode, error) {
	var key collections.Triple[string, string, string]
	if !strings.Contains(path, "/") {
		key = collections.Join3(owner, "", path)
	} else {
		key = collections.Join3(owner, filepath.Dir(path)+"/", filepath.Base(path))
	}
	return k.FileTree.Get(ctx, key)
}

// [NOTE]: to be stress-tested, may have to default to regular DeleteNode
func (k Keeper) RemoveNode(ctx context.Context, owner string, path string) error {
	// Ensure path has trailing slash for consistent storage
	storagePath := path
	if storagePath != "" && !strings.HasSuffix(storagePath, "/") {
		storagePath = storagePath + "/"
	}

	// Delete the node itself
	var selfKey collections.Triple[string, string, string]
	if storagePath == "" {
		selfKey = collections.Join3(owner, "", path) // root file
	} else {
		selfKey = collections.Join3(owner, filepath.Dir(storagePath), filepath.Base(storagePath))
	}

	if err := k.FileTree.Remove(ctx, selfKey); err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	// Delete ALL children with ONE range - they all have second component starting with storagePath
	rng := new(collections.Range[collections.Triple[string, string, string]]).
		Prefix(collections.Join3(owner, storagePath, ""))

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
