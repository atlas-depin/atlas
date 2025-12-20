package keeper

import (
	"context"

	"nebulix/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ProveFile(ctx context.Context, msg *types.MsgProveFile) (*types.MsgProveFileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}
	if _, err := k.GetProvider(ctx, msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid provider address")
	}

	proofInput := types.Proof{
		Data:       msg.Data,
		Hashes:     msg.Hashes,
		ChunkIndex: msg.Chunk,
	}

	if msg.ChallengeId == "" {
		file, err := k.GetFile(ctx, msg.FileId)
		if err != nil {
			return nil, err
		}
		if int64(len(file.Providers)) >= file.Replicas {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "this file has sufficient providers")
		}
		if k.Prove(ctx, []byte(file.Merkle), msg.Chunk, proofInput) {
			file.Providers = append(file.Providers, msg.Creator)
			file.Status = "ACTIVE"
			// [TODO]: eliminate stray from stays index if sufficient providers reached
			if err = k.SetFile(ctx, file); err != nil {
				return nil, err
			}
		} else {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid file proof")
		}
	} else if err := k.AttemptChallenge(ctx, msg.Creator, proofInput); err != nil {
		// [TODO]: clearing out provider?
		return nil, err
	}

	return &types.MsgProveFileResponse{}, nil
}
