package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

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

	if msg.Replicas < -1 || msg.Replicas == 1 {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "must have 2 or more replicas")
	} else if msg.Replicas == 0 {
		msg.Replicas = 3
	}

	/// --- Compile file information
	file := types.File{
		Merkle:   hex.EncodeToString(msg.Merkle),
		Owner:    msg.Creator,
		Start:    ctx.BlockHeight(),
		FileSize: msg.FileSize,
		Replicas: msg.Replicas,
	}
	fidString := file.Merkle + file.Owner + strconv.FormatInt(msg.Ddi, 10) + strconv.FormatInt(msg.FileSize, 10)
	fid := fmt.Sprintf("%x", sha256.Sum256([]byte(fidString)))

	/// --- Ensure file is staged & get providers
	stagedFile, err := k.GetStagedFile(ctx, fid)
	if err != nil {
		return nil, err
	}
	if msg.Creator != stagedFile.Creator || file.FileSize != stagedFile.FileSize {
		return nil, sdkerrors.ErrInvalidRequest
	}
	file.Providers = stagedFile.Providers[:file.Replicas]
	k.RemoveStagedFile(ctx, fid) // idc if there's an error, it'll get cleaned up after expiry anyways

	/// --- Update subscription & commit file
	subscriptionId, err := k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, msg.Subscription, msg.FileSize, msg.Replicas)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err.Error())
	} else {
		file.Subscription = subscriptionId
	}

	k.SetNewFile(ctx, fid, file)

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
