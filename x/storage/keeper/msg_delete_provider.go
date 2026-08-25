package keeper

import (
	"context"
	"fmt"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DeleteProvider(goCtx context.Context, msg *types.MsgDeleteProvider) (*types.MsgDeleteProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// Only a registered provider may withdraw their own deposit.
	if _, err := k.GetProvider(ctx, msg.Creator); err != nil {
		return nil, err
	}

	// Refund the deposit from the module escrow account back to the provider.
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	coins := sdk.NewCoins(sdk.NewInt64Coin("uatl", params.ProviderDeposit))
	account, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ProviderDepositAccountName, account, coins); err != nil {
		return nil, err
	}

	// Remove provider record.
	if err := k.RemoveProvider(ctx, msg.Creator); err != nil {
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
			sdk.NewAttribute("action", "deleted"),
			sdk.NewAttribute(types.AttributeKeyStart, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &types.MsgDeleteProviderResponse{}, nil
}
