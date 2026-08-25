package keeper

import (
	"context"
	"errors"
	"strings"

	"atlas/x/filetree/types"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) TreeNode(ctx context.Context, req *types.QueryTreeNodeRequest) (*types.QueryTreeNodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	path := strings.ToLower(strings.Trim(req.Path, "/"))
	node, err := q.k.GetNode(ctx, req.Owner, path)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "node not found")
		}
		return nil, err
	}

	return &types.QueryTreeNodeResponse{Node: &node}, nil
}
