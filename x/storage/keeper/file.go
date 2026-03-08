// x/storage/keeper/file.go
package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
)

/*
Creates a new file record keyed by its Creator, Subscription, and Fid.
Returns an error if the file key already exists.
*/
func (k Keeper) SetNewFile(ctx context.Context, file types.File) error {
	key := collections.Join3(file.Creator, file.Subscription, file.Fid)

	exists, err := k.FileStore.Has(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrFileAlreadyExists
	}

	return k.FileStore.Set(ctx, key, file)
}

/*
Sets a file record keyed by its Creator, Subscription, and Fid.
*/
func (k Keeper) SetFile(ctx context.Context, file types.File) error {
	key := collections.Join3(file.Creator, file.Subscription, file.Fid)
	return k.FileStore.Set(ctx, key, file)
}

/*
Gets a file record by fid.
*/
func (k Keeper) GetFile(ctx context.Context, fileId string) (file types.File, err error) {
	key, err := k.FileStore.Indexes.ByFileId.MatchExact(ctx, fileId)
	// TODO: investigate better error handling here
	if err != nil || key == collections.Join3("", "", "") {
		return file, types.ErrFileNotFound
	}

	file, err = k.FileStore.Get(ctx, key)
	if err == collections.ErrNotFound {
		err = types.ErrFileNotFound
	}
	return
}

/*
Remove a file record by fid.
*/
func (k Keeper) RemoveFile(ctx context.Context, fileId string) error {
	key, err := k.FileStore.Indexes.ByFileId.MatchExact(ctx, fileId)
	if err != nil {
		return err
	}

	return k.FileStore.Remove(ctx, key)
}

/*
Iterate over all files and run the provided function argument for each record.
The provided function must have a `types.File` argument.
*/
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

/*
Get total file count using a FileStore iterator.
*/
// func (k Keeper) GetFileCount(ctx context.Context) (uint64, error) {
// 	iter, err := k.FileStore.Iterate(ctx, nil)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer iter.Close()

// 	keys, err := iter.Keys()
// 	if err != nil {
// 		return 0, err
// 	}

// 	return uint64(len(keys)), nil
// }
