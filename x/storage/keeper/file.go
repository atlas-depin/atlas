// x/storage/keeper/file.go
package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
)

func (k Keeper) SetFile(ctx context.Context, file types.File) error {
	key := collections.Join3(file.Creator, file.Subscription, file.Fid)
	return k.FileStore.Set(ctx, key, file)
}

// StoreFile stores a new file with a provider
func (k Keeper) SetNewFile(ctx context.Context, file types.File) error {
	// Check if file already exists
	key := collections.Join3(file.Creator, file.Subscription, file.Fid)
	exists, err := k.FileStore.Has(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrFileAlreadyExists
	}

	// Save file
	if err := k.FileStore.Set(ctx, key, file); err != nil {
		return err
	}
	return nil
}

// GetFile retrieves a file by ID
func (k Keeper) GetFile(ctx context.Context, fileId string) (types.File, error) {
	key, err := k.FileStore.Indexes.ByFileId.MatchExact(ctx, fileId)
	// [TODO]: if key is "", there is an error
	if err != nil {
		return types.File{}, err
	}
	return k.FileStore.Get(ctx, key)
}

func (k Keeper) GetFiles(ctx context.Context) ([]types.File, error) {
	iter, err := k.FileStore.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	return iter.Values()
}

// \\ DeleteFile marks a file as deleted
func (k Keeper) RemoveFile(ctx context.Context, fileId string) error {
	key, err := k.FileStore.Indexes.ByFileId.MatchExact(ctx, fileId)
	if err != nil {
		return err
	}

	return k.FileStore.Remove(ctx, key)
}

func (k Keeper) IterateAllFiles(ctx context.Context, fn func(file types.File) (stop bool)) error {
	iter, err := k.FileStore.Iterate(ctx, nil)
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

func (k Keeper) GetFileCount(ctx context.Context) (uint64, error) {
	iter, err := k.FileStore.Iterate(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	keys, err := iter.Keys()
	if err != nil {
		return 0, err
	}

	return uint64(len(keys)), nil
}
