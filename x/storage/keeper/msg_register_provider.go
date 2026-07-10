package keeper

import (
	"context"
	"fmt"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
RegisterProvider creates a new provider on-chain and escrows the required deposit.
*/
func (k msgServer) RegisterProvider(goCtx context.Context, msg *types.MsgRegisterProvider) (*types.MsgRegisterProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	/// --- Message validation
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	/// --- Escrow uatl deposit tokens
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	coins := sdk.NewCoins(sdk.NewInt64Coin("uatl", params.ProviderDeposit))

	account, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ProviderDepositAccountName, coins)
	if err != nil {
		return nil, types.ErrProviderDepositInsufficient
	}

	/// --- Build provider record
	provider := types.Provider{
		Address:        msg.Creator,
		Hostname:       msg.Hostname,
		SpaceAvailable: msg.Capacity,
		SpaceUsed:      0,
		CreatedAt:      ctx.BlockHeight(),
		CreditDelta:    0,
	}

	/// --- Commit provider
	if err := k.SetNewProvider(ctx, provider); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register provider")
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
			types.EventTypeNewProvider,
			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyProvider, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyStart, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return &types.MsgRegisterProviderResponse{}, nil
}

// func (k msgServer) ShutdownProvider(goCtx context.Context, msg *types.MsgShutdownProvider) (*types.MsgShutdownProviderResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	_, found := k.GetProviders(ctx, msg.Creator)
// 	if !found {
// 		return nil, types.ErrProviderNotFound
// 	}

// 	collateral, found := k.GetCollateral(ctx, msg.Creator)
// 	if found {
// 		coin := sdk.NewInt64Coin("uatl", collateral.Amount)
// 		coins := sdk.NewCoins(coin)

// 		account, err := sdk.AccAddressFromBech32(msg.Creator)
// 		if err != nil {
// 			return nil, err
// 		}

// 		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.CollateralCollectorName, account, coins)
// 		if err != nil {
// 			return nil, err
// 		}

// 		k.RemoveCollateral(ctx, msg.Creator)
// 	}

// 	k.RemoveProviders(ctx, msg.Creator)

// 	ctx.EventManager().EmitEvent(
// 		sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
// 		),
// 	)

// 	ctx.EventManager().EmitEvent(
// 		sdk.NewEvent(
// 			types.EventTypeJackalMessage,
// 			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
// 		),
// 	)

// 	return &types.MsgShutdownProviderResponse{}, nil
// }
