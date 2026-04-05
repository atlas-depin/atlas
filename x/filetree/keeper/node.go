package keeper

import (
	"context"
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
	key := collections.Join3(owner, pathDir(path), pathBase(path))
	return k.FileTree.Set(ctx, key, file)
}

func (k Keeper) GetNode(ctx context.Context, owner string, path string) (types.FileNode, error) {
	var key collections.Triple[string, string, string]
	if !strings.Contains(path, "/") {
		key = collections.Join3(owner, "", path)
	} else {
		key = collections.Join3(owner, pathDir(path), pathBase(path))
	}
	return k.FileTree.Get(ctx, key)
}

// [NOTE]: to be stress-tested, may have to default to regular DeleteNode
func (k Keeper) RemoveNode(ctx context.Context, owner string, path string, recursive bool) error {
	// Delete node itself
	var selfKey collections.Triple[string, string, string]
	if !strings.Contains(path, "/") {
		selfKey = collections.Join3(owner, "", path)
	} else {
		selfKey = collections.Join3(owner, pathDir(path), pathBase(path))
	}
	if err := k.FileTree.Remove(ctx, selfKey); err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	rng := new(collections.Range[collections.Triple[string, string, string]]).
		Prefix(collections.Join3(owner, path, ""))

	if err := k.IterateNodes(ctx, rng, func(key collections.Triple[string, string, string], _ types.FileNode) (bool, error) {
		return false, k.FileTree.Remove(ctx, key)
	}); err != nil {
		return err
	}

	if recursive {
		start := collections.Join3(owner, path+"/", "")
		end := collections.Join3(owner, path+"/"+"\xff", "")
		recusiveRng := new(collections.Range[collections.Triple[string, string, string]])
		recusiveRng.StartInclusive(start)
		recusiveRng.EndInclusive(end)

		return k.IterateNodes(ctx, recusiveRng, func(key collections.Triple[string, string, string], _ types.FileNode) (bool, error) {
			return false, k.FileTree.Remove(ctx, key)
		})
	} else {
		return nil
	}
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

func pathDir(path string) string {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return ""
	}
	return path[:lastSlash]
}

func pathBase(path string) string {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return path
	}
	return path[lastSlash+1:]
}
