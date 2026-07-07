package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"atlas/x/storage/types"
)

// BuyStorage creates a new storage subscription for the sender.
// Tokens are transferred from the sender to the module account and
// storage credits are minted into the rewards pool proportional to
// the purchased bytes * duration * 3 replicas.
func (k msgServer) BuyStorage(goCtx context.Context, msg *types.MsgBuyStorage) (*types.MsgBuyStorageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// === validate request =====
	spender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse creator address %s", msg.Creator)
	}

	duration, gbs, err := validateStorageBuy(msg.Duration, msg.Bytes, "uatl")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate buy request")
	}

	valid, err := k.validateSufficientStorage(ctx, msg.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate buy request")
	}
	if !valid {
		return nil, types.ErrInsufficientStorage
	}

	// === compute storage cost =====
	storageCost := k.GetStorageCost(ctx, gbs, msg.Duration)
	toPay := sdk.NewCoin("uatl", storageCost)

	// === transfer payment to storage module =====
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, spender, types.ModuleName, sdk.NewCoins(toPay))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot send tokens from %s", msg.Creator)
	}

	// === mint storage credits into rewards pool =====
	seconds := math.NewInt(msg.Duration).Mul(math.NewInt(24 * 60 * 60))
	credits := math.NewInt(msg.Bytes * 3).Mul(seconds)
	cpool, err := k.StorageCreditPool.Get(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update storage credit pool")
	}
	err = k.StorageCreditPool.Set(ctx, cpool.Add(credits))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update storage credit pool")
	}

	// === create & commit storage subscription (with expiry record) =====
	subscription := types.StorageSubscription{
		Start:            ctx.BlockTime().Truncate(time.Second),
		End:              ctx.BlockTime().Truncate(time.Second).Add(duration),
		SpaceAvailable:   msg.Bytes * 3, // [TODO]: defaultReplicas keeper parameter?
		SpaceUsed:        0,
		ReplicaSpaceUsed: 0,
		Credits:          credits,
	}

	subscriptionId, err := k.SetNewStorageSubscription(ctx, msg.Creator, subscription)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to commit new storage subscription")
	}

	if msg.IsDefault {
		k.SetDefaultStorageSubscriptionId(ctx, msg.Creator, subscriptionId)
	}

	// === Emit events =====
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBuyStorage,
			sdk.NewAttribute(types.AttributeKeyBuyer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBytesBought, fmt.Sprintf("%d", msg.Bytes)),
			sdk.NewAttribute(types.AttributeKeyTimeBought, fmt.Sprintf("%d", msg.Duration)),
		),
	)

	return &types.MsgBuyStorageResponse{SubscriptionId: subscriptionId}, nil
}

// validateStorageBuy validates the duration (≥1 day), size (≥1 GB),
// and denom before computing a storage purchase.
func validateStorageBuy(days int64, bytes int64, denom string) (duration time.Duration, gb int64, err error) {
	if denom != "uatl" {
		err = errors.Wrapf(sdkerrors.ErrInvalidCoins, "cannot pay with anything other than uatl")
		return
	}
	if days < 0 || bytes < 0 {
		return time.Duration(0), 0, types.ErrSubscriptionInvalid
	}

	duration = time.Duration(days) * time.Hour * 24
	if duration < time.Hour*24 {
		err = fmt.Errorf("duration can't be less than 1 day")
		return
	}

	gb = bytes / 1_000_000_000
	if gb <= 0 {
		err = fmt.Errorf("cannot buy less than a gb")
		return
	}

	return
}

// validateSufficientStorage checks that the network has enough
// total provider capacity to accommodate the requested bytes.
func (k Keeper) validateSufficientStorage(ctx sdk.Context, bytes int64) (bool, error) {
	var spacePurchased int64 = 0
	var spaceAvailable int64 = 0

	if err := k.IterateAllStorageSubscriptions(ctx, func(_ collections.Pair[string, string], subscription types.StorageSubscription) bool {
		if ctx.BlockTime().Before(subscription.End) {
			spacePurchased += subscription.SpaceAvailable
		}
		return false
	}); err != nil {
		return false, err
	}

	if err := k.IterateAllProviders(ctx, func(_ string, provider types.Provider) bool {
		spaceAvailable += provider.SpaceAvailable
		return false
	}); err != nil {
		return false, err
	}

	return spacePurchased+bytes < spaceAvailable, nil
}
