// x/storage/keeper/file.go
package keeper

import (
	"context"
	"errors"
	"slices"

	"atlas/x/storage/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	if errors.Is(err, collections.ErrNotFound) {
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

func (k Keeper) IncrementFileCount(ctx context.Context) error {
	if currentCount, err := k.FileCount.Get(ctx); err == nil {
		return k.FileCount.Set(ctx, currentCount+1)
	} else {
		return err
	}
}

func (k Keeper) DecrementFileCount(ctx context.Context) error {
	if currentCount, err := k.FileCount.Get(ctx); err == nil {
		return k.FileCount.Set(ctx, currentCount-1)
	} else {
		return err
	}
}

/*
Remove a provider from the list of unstable providers for a given file.
*/
func (k Keeper) StabilizeProvider(ctx sdk.Context, provider string, file types.File) error {
	if slices.Contains(file.UnstableProviders, provider) {
		file.UnstableProviders = slices.DeleteFunc(file.UnstableProviders, func(p string) bool {
			return p == provider
		})

		if err := k.SetFile(ctx, file); err != nil {
			ctx.Logger().Error("Failed to remove provider from unstable providers:",
				"provider", provider,
				"fid", file.Fid,
			)
			return err
		}
	}

	return nil
}

/*
Add a provider to the list of unstable providers for a given file.
*/
func (k Keeper) DestabilizeProvider(ctx sdk.Context, provider string, file types.File) error {
	file.UnstableProviders = append(file.UnstableProviders, provider)
	if err := k.SetFile(ctx, file); err != nil {
		ctx.Logger().Warn("Failed to add provider to unstable providers:",
			"provider", provider,
			"fid", file.Fid,
		)
		return err
	}

	return nil
}
