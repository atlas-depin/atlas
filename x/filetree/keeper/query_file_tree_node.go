package keeper

import (
	"context"
	"strings"

	"nebulix/x/filetree/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileTreeNode(ctx context.Context, req *types.QueryFileTreeNodeRequest) (*types.QueryFileTreeNodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	node, err := q.k.GetNode(ctx, strings.TrimSuffix(req.Path, "/"))

	return &types.QueryFileTreeNodeResponse{Node: &node}, err
}
