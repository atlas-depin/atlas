package keeper

import (
	"context"

	"atlas/x/filetree/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeleteNode(ctx context.Context, msg *types.MsgDeleteNode) (*types.MsgDeleteNodeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	err := k.RemoveNode(ctx, msg.Creator+"_"+msg.Subscription, msg.Path, true)

	return &types.MsgDeleteNodeResponse{}, err
}
