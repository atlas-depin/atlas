package keeper

import (
	"context"
	"encoding/hex"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) PostFile(goCtx context.Context, msg *types.MsgPostFile) (*types.MsgPostFileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errors.Wrap(err, "invalid authority address")
	}

	// [TODO]: bad file msg checks

	/// --- Index the file
	file := types.File{
		Merkle:    hex.EncodeToString(msg.Merkle),
		Owner:     msg.Creator,
		Start:     ctx.BlockHeight(),
		FileSize:  msg.FileSize,
		Providers: msg.Providers,
		Replicas:  msg.Replicas,
	}
	k.SetFile(ctx, file)

	// /// --- Handle storage usage
	// if subscription, found := k.GetStoragePaymentInfo(ctx, msg.Creator); found {
	// 	if subscription.End.Before(ctx.BlockTime()) {
	// 		// [TBD]: can I handle storage account expiry here? appstate remains unchanged on error so, idk how
	// 		return nil, errors.Wrapf(sdkerrors.ErrUnauthorized, "storage account is expired")
	// 	}

	// 	credits := math.NewInt(msg.FileSize).Mul(math.NewInt(int64(subscription.End.Sub(ctx.BlockTime()) / time.Second)))

	// 	subscription.SpaceUsed += msg.FileSize
	// 	k.TransferStorageCredits(ctx, msg.Creator, types.ModuleAddress.String(), credits)

	// 	// sub credits

	// 	if subscription.SpaceUsed > subscription.SpaceAvailable {
	// 		return nil, errors.Wrapf(sdkerrors.ErrUnauthorized, "storage account does not have enough space available %d > %d", subscription.SpaceUsed, subscription.SpaceAvailable)
	// 	}

	// 	k.SetStoragePaymentInfo(ctx, subscription)
	// } else {
	// 	return nil, errors.Wrapf(sdkerrors.ErrKeyNotFound, "storage account does not exist")
	// }

	// /// --- Emit events
	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		sdk.EventTypeMessage,
	// 		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
	// 	),
	// )

	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		types.EventTypeSignContract,
	// 		sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
	// 		sdk.NewAttribute(types.AttributeKeyContract, hex.EncodeToString(msg.Merkle)),
	// 		sdk.NewAttribute(types.AttributeKeyStart, fmt.Sprintf("%d", ctx.BlockHeight())),
	// 	),
	// )

	return &types.MsgPostFileResponse{StartBlock: ctx.BlockHeight()}, nil
}
