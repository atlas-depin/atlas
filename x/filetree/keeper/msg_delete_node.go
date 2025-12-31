package keeper

import (
	"context"

	"nebulix/x/filetree/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeleteNode(ctx context.Context, msg *types.MsgDeleteNode) (*types.MsgDeleteNodeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	err := k.RemoveNode(ctx, msg.Creator, msg.Path)

	return &types.MsgDeleteNodeResponse{}, err
}
