package keeper

import (
	"context"
	"slices"
	"strconv"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

/*
ProveFile validates a file proof for a given challenge.
If there is no challenge and the file is a stray, the provider is added to the list of the file's providers.
*/
func (k msgServer) ProveFile(goCtx context.Context, msg *types.MsgProveFile) (*types.MsgProveFileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	/// --- Message validation
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid address")
	}
	if _, err := k.GetProvider(ctx, msg.Creator); err != nil {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid provider address")
	}

	/// --- Message validation
	proofInput := types.Proof{
		Data:       msg.Data,
		Hashes:     msg.Hashes,
		ChunkIndex: msg.Chunk,
	}

	if msg.ChallengeId == "" { // if: no challenge provider (new provider proof)
		/// --- Validate file is not ACTIVE and provider not already assigned to this file
		file, err := k.GetFile(ctx, msg.Fid)
		if err != nil {
			return nil, err
		}
		if file.Status == "ACTIVE" {
			return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "this file is not a stray")
		}
		if slices.Contains(file.Providers, msg.Creator) {
			return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "this provider is already assigned to the file")
		}

		// deal with storage credits first to avoid having to compute proof on an invalid request
		t_r0, err := k.LastRewardsTime.Get(ctx)
		if err != nil {
			return nil, err
		}
		t_0 := max(t_r0.Unix(), file.Start)

		delta_r := ctx.BlockTime().Unix() - t_r0.Unix()
		delta_s := ctx.BlockTime().Unix() - t_0
		if err := k.IncrementProviderSpaceUsed(ctx, msg.Creator, file.FileSize, delta_r, delta_s); err != nil {
			return nil, err
		}

		if err := k.AttemptAddProvider(ctx, file, msg.Creator, proofInput); err != nil {
			return nil, err
		}

	} else if err := k.AttemptChallenge(ctx, msg.Creator, proofInput); err != nil { // else: this is a response to a challenge
		return nil, err
	}

	/// --- Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeAttempt,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyFile, msg.Fid),
			sdk.NewAttribute(types.AttributeKeyChallengeId, msg.ChallengeId),
			sdk.NewAttribute(types.AttributeKeyProofChunk, strconv.Itoa(int(msg.Chunk))),
		),
	)

	return &types.MsgProveFileResponse{}, nil
}
