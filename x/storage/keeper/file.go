// x/storage/keeper/file.go
package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
)

// StoreFile stores a new file with a provider
func (k Keeper) SetNewFile(ctx context.Context, fileId string, file types.File) error {
	// Check if file already exists
	exists, err := k.FileStore.Has(ctx, fileId)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrFileAlreadyExists
	}

	// Save file
	if err := k.FileStore.Set(ctx, fileId, file); err != nil {
		return err
	}

	return nil
}

// \\ GetFile retrieves a file by ID
func (k Keeper) GetFile(ctx context.Context, fileId string) (types.File, error) {
	return k.FileStore.Get(ctx, fileId)
}

func (k Keeper) GetAllFiles(ctx context.Context) ([]collections.KeyValue[string, types.File], error) {
	iter, err := k.FileStore.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	return iter.KeyValues()
}

// \\ DeleteFile marks a file as deleted
func (k Keeper) DeleteFile(ctx context.Context, fileId string, owner string) error {
	file, err := k.FileStore.Get(ctx, fileId)
	if err != nil {
		return err
	}

	// TODO: Verify ownership?

	// Soft delete
	// file.IsDeleted = true
	return k.FileStore.Set(ctx, fileId, file)
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
