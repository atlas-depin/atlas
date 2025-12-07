package keeper

import (
	"context"
	"nebulix/x/storage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetStagedFile(ctx context.Context, fid string) (types.StagedFile, error) {
	return k.FileStagingStore.Get(ctx, fid)
}

func (k Keeper) SetStagedFile(ctx context.Context, fid string, stagedFile types.StagedFile) error {
	return k.FileStagingStore.Set(ctx, fid, stagedFile)
}

func (k Keeper) RemoveStagedFile(ctx context.Context, fid string) error {
	return k.FileStagingStore.Remove(ctx, fid)
}

func (k Keeper) CleanupStagedFiles(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentHeight := sdkCtx.BlockHeight()

	iter, err := k.FileStagingStore.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	keysToDelete := []string{}

	// collect keys to delete
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		// Check if older than 50 blocks
		if currentHeight > int64(kv.Value.Start)+50 {
			keysToDelete = append(keysToDelete, kv.Key)
		}
	}

	// delete collected keys
	for _, key := range keysToDelete {
		if err := k.FileStagingStore.Remove(ctx, key); err != nil {
			return err
		}
	}

	return nil
}
