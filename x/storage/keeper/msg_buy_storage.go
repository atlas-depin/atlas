package keeper

import (
	"context"
	"fmt"
	"log"
	"time"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"nebulix/x/storage/types"
)

const (
	gb_bytes int64 = 1_000_000_000
)

func validateBuy(days int64, bytes int64, denom string) (duration time.Duration, gb int64, err error) {
	if denom != "unblx" {
		err = errors.Wrapf(sdkerrors.ErrInvalidCoins, "cannot pay with anything other than ujkl")
		return
	}

	duration = time.Duration(days) * time.Hour * 24
	if duration < time.Hour*24 {
		err = fmt.Errorf("duration can't be less than 1 day")
		return
	}

	gb = bytes / gb_bytes
	if gb <= 0 {
		err = fmt.Errorf("cannot buy less than a gb")
		return
	}

	return
}

func (k msgServer) BuyStorage(goCtx context.Context, msg *types.MsgBuyStorage) (*types.MsgBuyStorageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	/// --- Validate request
	if msg.Receiver == "" {
		msg.Receiver = msg.Creator
	}

	if msg.Replicas < 2 {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "at least two file replicas are required")
	}

	spender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse creator address %s", msg.Creator)
	}

	recipient, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, errors.Wrapf(err, "for address is not a proper bech32")
	}

	_, found := k.GetStoragePaymentInfo(ctx, recipient.String())
	if found {
		return nil, errors.Wrapf(err, "account has an active storage subscription, consider upgrading storage")
	}

	// [TBD]: do I need to return duration?
	duration, gbs, err := validateBuy(msg.Duration, msg.Bytes, "unblx")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate buy request")
	}

	/// --- compute storage cost --- ///
	storageCost := k.GetStorageCost(ctx, gbs, msg.Duration)
	toPay := sdk.NewCoin("unblx", storageCost)

	/// --- create new storage subscription --- ///
	// create recipient account if it doesn't exist
	if !k.authKeeper.HasAccount(ctx, recipient) {
		defer telemetry.IncrCounter(1, "new", "account") // [TBD]: do I want/need this?
		k.authKeeper.SetAccount(ctx, k.authKeeper.NewAccountWithAddress(ctx, recipient))
	}

	// create subscription object
	seconds := math.NewInt(msg.Duration).Mul(math.NewInt(60 * 60))
	credits := math.NewInt(msg.Bytes).Mul(seconds)
	subscription := types.StorageSubscription{
		Start:          ctx.BlockTime().Truncate(time.Second),
		End:            ctx.BlockTime().Truncate(time.Second).Add(duration),
		SpaceAvailable: msg.Bytes * 3,
		SpaceUsed:      0,
		Credits:        credits,
	}

	// transfer tokens to pay
	test := toPay.Amount.Int64()
	log.Println(test)
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, spender, types.ModuleName, sdk.NewCoins(toPay))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot send tokens from %s", msg.Creator)
	}

	// commit subscription to kvstore
	k.SetNewStorageSubscription(ctx, recipient.String(), subscription)

	/// --- Emit events --- ///
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
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
			sdk.NewAttribute(types.AttributeKeyBytesBought, fmt.Sprintf("%d", msg.Bytes)),
			sdk.NewAttribute(types.AttributeKeyTimeBought, fmt.Sprintf("%d", msg.Duration)),
		),
	)

	return &types.MsgBuyStorageResponse{}, nil
}

// TODO: upgrade storage duration
// TODO: upgrade storage quantity
