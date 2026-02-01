package keeper

import (
	"context"
	"strings"

	"nebulix/x/filetree/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileNode(ctx context.Context, req *types.QueryFileNodeRequest) (*types.QueryFileNodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	path := strings.ToLower(strings.TrimSuffix(req.Path, "/"))
	node, err := q.k.GetNode(ctx, req.Owner, path)

	return &types.QueryFileNodeResponse{Node: &node}, err
}
