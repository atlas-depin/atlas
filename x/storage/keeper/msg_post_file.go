package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

/*
PostFile stages a new file and updates the file's subscription's usage,
*/
func (k msgServer) PostFile(goCtx context.Context, msg *types.MsgPostFile) (*types.MsgPostFileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	height := ctx.BlockHeight()

	/// --- Message validation
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errors.Wrap(err, "invalid authority address")
	}

	if msg.Replicas == 0 {
		msg.Replicas = 3
	} else if msg.Replicas < 2 {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "must have 2 or more replicas")
	}

	/// --- Build file record
	file := types.File{
		Fid:      msg.Fid,
		Merkle:   hex.EncodeToString(msg.Merkle),
		Creator:  msg.Creator,
		Status:   "STAGED",
		Start:    height,
		FileSize: msg.FileSize,
		Replicas: msg.Replicas,
	}

	/// --- Update subscription
	subscriptionId, err := k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, msg.Subscription, msg.FileSize, msg.Replicas)
	if err != nil {
		return nil, errors.Wrapf(err, "Subscription: %s", subscriptionId)
	} else {
		file.Subscription = subscriptionId
	}

	/// --- Commit file
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
			types.EventTypeNewFile,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyFile, hex.EncodeToString(msg.Merkle)),
			sdk.NewAttribute(types.AttributeKeyStart, fmt.Sprintf("%d", height)),
		),
	)

	return &types.MsgPostFileResponse{StartBlock: height}, nil
}
