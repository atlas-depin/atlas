package keeper

import (
	"context"
	"encoding/hex"
	"nebulix/x/storage/types"
	"slices"

	"cosmossdk.io/errors"
	"github.com/ahm23/go-merkletree-xxh"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeebo/blake3"
)

/*
Process a proof from a provider attempting to claim it has a file for the first time.
*/
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

/*
Process a challenge attempt by verifying the provided file proof from a given provider.
*/
func (k Keeper) AttemptChallenge(ctx sdk.Context, provider string, proofInput types.Proof) error {
	challenge, err := k.GetChallenge(ctx, provider, proofInput.ChallengeId)
	if err != nil {
		_ = k.RemoveChallenge(ctx, provider, proofInput.ChallengeId)
		ctx.Logger().Error("Unable to find challenge:",
			"challenge_id", proofInput.ChallengeId,
			"provider", provider,
		)
		return nil
	}

	file, err := k.GetFile(ctx, challenge.FileId)
	if err != nil {
		_ = k.RemoveChallenge(ctx, provider, proofInput.ChallengeId)
		return nil
	}

	if slices.Contains(file.Providers, provider) {
		root, _ := hex.DecodeString(file.Merkle)
		if k.Prove(ctx, root, challenge.ChunkIndex, proofInput) {
			_ = k.RemoveChallenge(ctx, provider, proofInput.ChallengeId)
			_ = k.StabilizeProvider(ctx, provider, file)
		} else {
			return types.ErrProofInvalid
		}
	} else {
		return types.ErrProofFileBadProvider
	}
	return nil
}

/*
Verify a merkle proof.
*/
func (k Keeper) Prove(ctx context.Context, root []byte, index uint64, input types.Proof) bool {
	hasher := blake3.New()
	hasher.Write(input.Data)
	hashName := hasher.Sum(nil)

	if index != input.Index {
		return false
	}

	var proof merkletree.Proof
	proof.Siblings = input.Hashes
	proof.PathBits = input.Path

	verified, err := merkletree.Verify(hashName, root, &proof, &merkletree.Config{XXH128: true})
	if err != nil {
		return false
	}

	return verified
}
