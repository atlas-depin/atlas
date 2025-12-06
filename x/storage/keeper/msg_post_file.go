package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"

	"github.com/cespare/xxhash"
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

	/// --- Index the file
	file := types.File{
		Merkle:    hex.EncodeToString(msg.Merkle),
		Owner:     msg.Creator,
		Start:     ctx.BlockHeight(),
		FileSize:  msg.FileSize,
		Providers: msg.Providers,
		Replicas:  msg.Replicas,
	}
	// [TODO]: fix protoc
	fidString := file.Merkle + file.Owner + strconv.FormatInt(ctx.BlockHeight(), 10)
	fid := xxhash.Sum64([]byte(fidString))

	// update subscription usage stats
	subscriptionId, err := k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, msg.Subscription, msg.FileSize, msg.Replicas)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err.Error())
	} else {
		file.Subscription = subscriptionId
	}
	// commit file to kvstore

	k.SetNewFile(ctx, fmt.Sprintf("%x", fid), file)

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
