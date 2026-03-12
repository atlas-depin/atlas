package keeper

import (
	"context"
	"encoding/hex"
	"nebulix/x/storage/types"

	"cosmossdk.io/errors"
)

func (k Keeper) AttemptAddProvider(ctx context.Context, file types.File, address string, proof types.Proof) error {
	/// --- Attempt initial proof of the file
	root, _ := hex.DecodeString(file.Merkle)
	if k.Prove(ctx, root, proof.ChunkIndex, proof) {
		/// --- Add provider to the file after a successful proof
		file.Providers = append(file.Providers, address)
		if int32(len(file.Providers)) >= file.Replicas {
			file.Status = "ACTIVE"
		} else {
			file.Status = "STRAY"
		}

		/// --- Commit file update with new list of providers & increment total file count
		if err := k.SetFile(ctx, file); err != nil {
			return errors.Wrap(err, "failed to update the file's provider list")
		}
		oldCount, err := k.FileCount.Get(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get total file count")
		}
		if err := k.FileCount.Set(ctx, oldCount+1); err != nil {
			return errors.Wrap(err, "failed to increment total file count")
		}
	} else {
		return types.ErrProofInvalid
	}

	return nil
}
