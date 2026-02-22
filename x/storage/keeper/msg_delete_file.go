package keeper

import (
	"context"
	"strings"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) DeleteFile(goCtx context.Context, msg *types.MsgDeleteFile) (*types.MsgDeleteFileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	file, err := k.GetFile(ctx, msg.FileId)
	if err != nil {
		// [TODO]: error - file not found
		return nil, err
	}

	/// --- Update provider storage used
	for _, address := range file.Providers {
		err = k.DecrementProviderSpaceUsed(ctx, address, file.FileSize)
		if err != nil && err != types.ErrProviderNotFound {
			return nil, err
		}
	}

	/// --- Update subscription
	_, err = k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, file.Subscription, -file.FileSize, file.Replicas)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "%s", err.Error())
	}

	/// --- Remove file
	err = k.RemoveFile(ctx, msg.FileId)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeleteFile,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyFile, msg.FileId),
			sdk.NewAttribute(types.AttributeKeyProviders, strings.Join(file.Providers, ",")),
		),
	)

	return &types.MsgDeleteFileResponse{}, nil
}
