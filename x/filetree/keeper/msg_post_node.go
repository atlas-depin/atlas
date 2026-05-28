package keeper

import (
	"context"
	"strings"

	"atlas/x/filetree/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) PostNode(ctx context.Context, msg *types.MsgPostNode) (*types.MsgPostNodeResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// [TODO]: handle illegal path characters
	path := strings.ToLower(strings.Trim(msg.Path, "/"))
	ftn := types.TreeNode{
		NodeType: msg.NodeType,
		Contents: msg.Contents,
		Viewers:  msg.Viewers,
		Editors:  msg.Editors,
	}

	var err error
	if strings.Contains(path, "/") {
		err = k.SetNode(ctx, msg.Creator+"_"+msg.Subscription, path, ftn)
	} else {
		err = k.SetRootNode(ctx, msg.Creator+"_"+msg.Subscription, path, ftn)
	}

	return &types.MsgPostNodeResponse{}, err
}
