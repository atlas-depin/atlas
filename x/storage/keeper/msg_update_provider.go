package keeper

import (
	"context"
	"fmt"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
UpdateProvider updates an existing provider's hostname and/or capacity on-chain.
*/
func (k msgServer) UpdateProvider(goCtx context.Context, msg *types.MsgUpdateProvider) (*types.MsgUpdateProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	provider, err := k.GetProvider(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	if msg.Hostname != "" {
		provider.Hostname = msg.Hostname
	}
	if msg.Capacity > 0 {
		if msg.Capacity < provider.SpaceUsed {
			return nil, types.ErrProviderCapacityTooLow
		}
		provider.SpaceAvailable = msg.Capacity
	}

	if err := k.SetProvider(ctx, provider); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNewProvider,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyProvider, msg.Creator),
			sdk.NewAttribute("hostname", provider.Hostname),
			sdk.NewAttribute("capacity", fmt.Sprintf("%d", provider.SpaceAvailable)),
			sdk.NewAttribute(types.AttributeKeyStart, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &types.MsgUpdateProviderResponse{}, nil
}
