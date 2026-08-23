package keeper

import (
	"context"
	"strings"

	"cosmossdk.io/collections"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	/// --- Verify caller owns the file
	if msg.Creator != file.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only the file creator can delete the file")
	}

	/// --- Update subscription & provider storage used
	_, err = k.UpdateStorageSubscriptionUsage(ctx, msg.Creator, file.Subscription, -file.FileSize, file.Replicas)
	if err != nil {
		if err == types.ErrSubscriptionExpired {
			return nil, nil
		}
		return nil, err
	}

	// TODO: stray manager
	blockTime := ctx.BlockTime().Unix()
	rewardStart, err := k.LastRewardsTime.Get(ctx)
	rewardsTimeDelta := blockTime - rewardStart

	for _, address := range file.Providers {
		err = k.DecrementProviderSpaceUsed(ctx, address, file.FileSize, rewardsTimeDelta)
		if err != nil && err != types.ErrProviderNotFound {
			return nil, err
		}

		// Remove this provider's bytes from the per-subscription tracker.
		subProvKey := collections.Join(file.Subscription, address)
		currentBytes, getErr := k.SubscriptionProviderBytes.Get(ctx, subProvKey)
		if getErr == nil {
			newBytes := currentBytes - file.FileSize
			if newBytes <= 0 {
				_ = k.SubscriptionProviderBytes.Remove(ctx, subProvKey)
			} else {
				_ = k.SubscriptionProviderBytes.Set(ctx, subProvKey, newBytes)
			}
		}
	}

	/// --- Remove file
	err = k.RemoveFile(ctx, msg.FileId)
	if err != nil {
		return nil, err
	}
	if file.Status != "STAGED" && file.Status != "DEAD" {
		if err = k.DecrementFileCount(ctx); err != nil {
			return nil, err
		}
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
