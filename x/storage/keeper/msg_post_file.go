package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) PostFile(goCtx context.Context, msg *types.MsgPostFile) (*types.MsgPostFileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errors.Wrap(err, "invalid authority address")
	}

	if msg.Replicas == 0 {
		msg.Replicas = 3
	} else if msg.Replicas < 2 {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "must have 2 or more replicas")
	}

	/// --- Build file index element & metadata
	fileIndex, err := k.FileCount.Get(ctx)
	if err != nil {
		return nil, err
	}
	fileIndex++
	fileDeDupeIndex := fileIndex + ctx.BlockHeight()

	fid := sha256.Sum256(append(append(msg.Merkle, []byte(msg.Creator)...), byte(fileDeDupeIndex)))
	file := types.File{
		FileId:      fmt.Sprintf("%x", fid),
		Merkle:      hex.EncodeToString(msg.Merkle),
		Creator:     msg.Creator,
		Status:      "STAGED",
		Start:       ctx.BlockHeight(),
		FileSize:    msg.FileSize,
		Replicas:    msg.Replicas,
		DedupeIndex: fileDeDupeIndex,
	}

	/// --- Update subscription
	subscriptionId, err := k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, msg.Subscription, msg.FileSize, msg.Replicas)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err.Error())
	} else {
		file.Subscription = subscriptionId
	}

	/// --- Commit file
	if err = k.FileCount.Set(ctx, fileIndex); err != nil {
		return nil, err
	}
	if err = k.SetNewFile(ctx, file); err != nil {
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
			types.EventTypeSignContract,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyContract, hex.EncodeToString(msg.Merkle)),
			sdk.NewAttribute(types.AttributeKeyStart, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &types.MsgPostFileResponse{StartBlock: ctx.BlockHeight()}, nil
}
