package keeper

import (
	"context"
	"strings"

	"nebulix/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
DeleteFile deletes a file on-chain and signals its deletion on storage providers.
*/
func (k msgServer) DeleteFile(goCtx context.Context, msg *types.MsgDeleteFile) (*types.MsgDeleteFileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	/// --- Message validation
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	/// --- Get file details
	file, err := k.GetFile(ctx, msg.FileId)
	if err != nil {
		return nil, err
	}

	/// --- Update subscription & provider storage used
	_, err = k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, file.Subscription, -file.FileSize, file.Replicas)
	if err != nil {
		return nil, err
	}

	// TODO: credit delta from rewards
	// TODO: stray manager
	// -------------------------------
	// var rewardsTimeDelta int64
	// blockTime := ctx.BlockTime().Unix()
	// rewardStart, err := k.GetRewardWindowStartTime()
	// if rewardStart > file.Start {
	// 	rewardsTimeDelta = blockTime - rewardStart
	// } else {
	// 	rewardsTimeDelta = blockTime - file.Start
	// }

	for _, address := range file.Providers {
		err = k.DecrementProviderSpaceUsed(ctx, address, file.FileSize, 0 /*rewardsTimeDelta*/)
		if err != nil && err != types.ErrProviderNotFound {
			return nil, err
		}
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
