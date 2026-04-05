package keeper

import (
	"context"
	"fmt"

	"nebulix/x/storage/types"

	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ExpandStorage(goCtx context.Context, msg *types.MsgExpandStorage) (*types.MsgExpandStorageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// Get existing subscription
	subscription, err := k.GetStorageSubscription(ctx, msg.Creator, msg.SubscriptionId)
	if err != nil {
		return nil, errors.Wrapf(err, "subscription %d not found for %s", msg.SubscriptionId, msg.Creator)
	}

	// Check subscription hasn't expired
	if ctx.BlockTime().After(subscription.End) {
		return nil, types.ErrSubscriptionExpired
	}

	// Validate the expansion request
	duration, gbs, err := validateBuy(msg.Duration, msg.Bytes, "uatl")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate expansion request")
	}

	// Calculate cost for the additional resources
	storageCost := k.GetStorageCost(ctx, gbs, msg.Duration)
	toPay := sdk.NewCoin("uatl", storageCost)

	// --- Transfer payment ---
	spender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse creator address %s", msg.Creator)
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, spender, types.ModuleName, sdk.NewCoins(toPay))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send payment for subscription expansion")
	}

	// --- Update subscription ---
	// Extend end time
	newEnd := subscription.End.Add(duration)
	subscription.End = newEnd

	// Increase space available
	subscription.SpaceAvailable += msg.Bytes * 3 // Multiply by replicas

	// Recalculate credits
	seconds := math.NewInt(msg.Duration).Mul(math.NewInt(24 * 60 * 60))
	additionalCredits := math.NewInt(msg.Bytes * 3).Mul(seconds)
	subscription.Credits = subscription.Credits.Add(additionalCredits)

	// Save updated subscription
	err = k.SetStorageSubscription(ctx, msg.Creator, subscription)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to save updated subscription")
	}

	// --- Emit events ---
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExpandSubscription,
			sdk.NewAttribute(types.AttributeKeyBuyer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeySubscriptionId, fmt.Sprintf("%d", msg.SubscriptionId)),
			sdk.NewAttribute(types.AttributeKeyTimeBought, fmt.Sprintf("%d", msg.Duration)),
			sdk.NewAttribute(types.AttributeKeyBytesBought, fmt.Sprintf("%d", msg.Bytes)),
			sdk.NewAttribute(types.AttributeKeyAmountPaid, toPay.String()),
			sdk.NewAttribute(types.AttributeKeyExpiration, subscription.End.String()),
			sdk.NewAttribute(types.AttributeKeySpaceAvailable, fmt.Sprintf("%d", subscription.SpaceAvailable)),
		),
	)

	return &types.MsgExpandSubscriptionResponse{
		SubscriptionId:    msg.SubscriptionId,
		NewExpiration:     subscription.End,
		NewSpaceAvailable: subscription.SpaceAvailable,
	}, nil
}
