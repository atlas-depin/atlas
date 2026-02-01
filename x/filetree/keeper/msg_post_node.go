package keeper

import (
	"context"
	"strings"

	"nebulix/x/filetree/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) PostNode(ctx context.Context, msg *types.MsgPostNode) (*types.MsgPostNodeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// [TODO]: handle illegal path characters
	path := strings.TrimSuffix(msg.Path, "/")
	ftn := types.FileNode{
		NodeType: msg.NodeType,
		Contents: msg.Contents,
	}
	err := k.SetNode(ctx, msg.Creator, path, ftn)

	return &types.MsgPostNodeResponse{}, err
}
