package keeper

import (
	"context"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateFeed updates the data of a feed. Only the feed owner may update it.
func (k msgServer) UpdateFeed(goCtx context.Context, msg *types.MsgUpdateFeed) (*types.MsgUpdateFeedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	feed, err := k.GetFeed(ctx, msg.Name)
	if err != nil {
		return nil, types.ErrFeedNotFound
	}

	if feed.Owner != msg.Creator {
		return nil, types.ErrFeedNotOwner
	}

	feed.Data = msg.Data
	feed.LastUpdate = ctx.BlockTime()

	if err := k.SetFeed(ctx, feed); err != nil {
		return nil, err
	}

	return &types.MsgUpdateFeedResponse{}, nil
}
