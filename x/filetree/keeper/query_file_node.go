package keeper

import (
	"context"
	"strings"

	"nebulix/x/filetree/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) TreeNode(ctx context.Context, req *types.QueryTreeNodeRequest) (*types.QueryTreeNodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	path := strings.ToLower(strings.Trim(req.Path, "/"))
	node, err := q.k.GetNode(ctx, req.Owner, path)

	return &types.QueryTreeNodeResponse{Node: &node}, err
}
