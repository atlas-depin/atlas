package keeper

// import (
// 	"context"
// 	"fmt"

// 	"nebulix/x/storage/types"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
// )

// func (k msgServer) RegisterProvider(goCtx context.Context, msg *types.MsgInitProvider) (*types.MsgInitProviderResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	// Check if provider already exists
// 	exists, err := k.ProviderStore.Has(ctx, msg.Creator)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if exists {
// 		return nil, types.ErrProviderAlreadyExists
// 	}

// 	// Verify minimum deposit
// 	params, err := k.Params.Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	coin := sdk.NewInt64Coin("unblx", params.ProviderDeposit)
// 	coins := sdk.NewCoins(coin)

// 	account, err := sdk.AccAddressFromBech32(msg.Creator)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// [TODO]: make sure this provider collector module account is created at genesis
// 	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ProviderDepositCollectorName, coins)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(err, "%s does not have %s", account, coin.String())
// 	}

// 	provider := types.Providers{
// 		Address:         msg.Creator,
// 		Ip:              msg.Ip,
// 		Totalspace:      fmt.Sprintf("%d", msg.TotalSpace),
// 		Creator:         msg.Creator,
// 		BurnedContracts: "0",
// 		KeybaseIdentity: msg.Keybase,
// 		AuthClaimers:    []string{},
// 		Deposit:         params.ProviderDeposit,
// 	}

// 	k.SetProviders(ctx, provider)

// 	ctx.EventManager().EmitEvent(
// 		sdk.NewEvent(
// 			types2.EventTypeJackalMessage,
// 			sdk.NewAttribute(types.AttributeKeySigner, msg.Creator),
// 		),
// 	)

// 	return &types.MsgInitProviderResponse{}, nil
// }

// func (k msgServer) ShutdownProvider(goCtx context.Context, msg *types.MsgShutdownProvider) (*types.MsgShutdownProviderResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	_, found := k.GetProviders(ctx, msg.Creator)
// 	if !found {
// 		return nil, types.ErrProviderNotFound
// 	}

// 	collateral, found := k.GetCollateral(ctx, msg.Creator)
// 	if found {
// 		coin := sdk.NewInt64Coin("ujkl", collateral.Amount)
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
