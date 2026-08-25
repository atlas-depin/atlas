package keeper

import (
	"context"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdatePrice updates the $ATL price. Only the designated price updater
// (params.PriceUpdater) may do so.
func (k msgServer) UpdatePrice(goCtx context.Context, msg *types.MsgUpdatePrice) (*types.MsgUpdatePriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.PriceUpdater == "" || msg.Creator != params.PriceUpdater {
		return nil, types.ErrInvalidPriceUpdater
	}

	if err := k.Price.Set(ctx, msg.Price); err != nil {
		return nil, err
	}

	return &types.MsgUpdatePriceResponse{}, nil
}
