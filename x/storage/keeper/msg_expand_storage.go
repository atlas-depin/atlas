package keeper

import (
	"context"
	"fmt"
	"time"

	"atlas/x/storage/types"

	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func validateStorageExpand(days int64, bytes int64, denom string) (duration time.Duration, gb int64, err error) {
	if denom != "uatl" {
		err = errors.Wrapf(sdkerrors.ErrInvalidCoins, "cannot pay with anything other than ujkl")
		return
	}
	if days < 0 || bytes < 0 {
		return time.Duration(0), 0, types.ErrSubscriptionInvalid
	}

	if days > 0 {
		duration = time.Duration(days) * time.Hour * 24
		if duration < time.Hour*24 {
			err = fmt.Errorf("duration can't be less than 1 day")
			return
		}
	}

	if bytes > 0 {
		gb = bytes / gb_bytes
		if gb <= 0 {
			err = fmt.Errorf("cannot buy less than a gb")
			return
		}
	}
	return
}

func (k msgServer) ExpandStorage(goCtx context.Context, msg *types.MsgExpandStorage) (*types.MsgExpandStorageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// Get existing subscription
	subscription, err := k.GetStorageSubscription(ctx, msg.Creator, msg.SubscriptionId)
	if err != nil {
		return nil, errors.Wrapf(err, "subscription %s not found for %s", msg.SubscriptionId, msg.Creator)
	}

	// Check subscription hasn't expired
	if ctx.BlockTime().After(subscription.End) {
		return nil, types.ErrSubscriptionExpired
	}
	oldSeconds := subscription.End.Unix() - ctx.BlockTime().Unix()

	// Validate the expansion request
	duration, gbs, err := validateStorageExpand(msg.Duration, msg.Bytes, "uatl")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate expansion request")
	}

	// Calculate cost for the additional resources
	// storageCost := k.GetStorageCost(ctx, gbs, msg.Duration)
	// storageCost := k.GetStorageExpansionCost(ctx, gbs, msg.Duration, oldSeconds)
	c1 := k.GetStorageCost(ctx, gbs, (oldSeconds+86400-1)/86400)                               // expansion
	c2 := k.GetStorageCost(ctx, subscription.SpaceAvailable/3/1_000_000_000+gbs, msg.Duration) // duration
	toPay := sdk.NewCoin("uatl", c1.Add(c2))

	// --- Update subscription ---
	newSeconds := math.NewInt(msg.Duration).Mul(math.NewInt(24 * 60 * 60))

	// Extend end time & Increase space available
	subscription.End = subscription.End.Add(duration)
	subscription.SpaceAvailable += msg.Bytes * 3 // Multiply by replicas

	// Recalculate credits
	newCredits := math.NewInt(subscription.SpaceAvailable).Mul(newSeconds)
	additionalCredits := math.NewInt(msg.Bytes).MulRaw(3 * oldSeconds).Add(newCredits)
	subscription.Credits = subscription.Credits.Add(additionalCredits)

	// --- Transfer payment ---
	spender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse creator address %s", msg.Creator)
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, spender, types.ModuleName, sdk.NewCoins(toPay))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send payment for subscription expansion")
	}

	cpool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update storage credit pool")
	}
	err = k.StorageCreditPool.Set(ctx, cpool.Add(additionalCredits))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update storage credit pool")
	}

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
			sdk.NewAttribute(types.AttributeKeySubscriptionId, fmt.Sprintf("%s", msg.SubscriptionId)),
			sdk.NewAttribute(types.AttributeKeyTimeBought, fmt.Sprintf("%d", msg.Duration)),
			sdk.NewAttribute(types.AttributeKeyBytesBought, fmt.Sprintf("%d", msg.Bytes)),
			sdk.NewAttribute(types.AttributeKeyAmountPaid, toPay.String()),
			sdk.NewAttribute(types.AttributeKeyExpiration, subscription.End.String()),
			sdk.NewAttribute(types.AttributeKeySpaceAvailable, fmt.Sprintf("%d", subscription.SpaceAvailable)),
		),
	)

	return &types.MsgExpandStorageResponse{
		SubscriptionId:    msg.SubscriptionId,
		NewExpiration:     subscription.End.Unix(),
		NewSpaceAvailable: subscription.SpaceAvailable,
	}, nil
}
