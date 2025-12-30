package keeper

import (
	"context"

    "nebulix/x/filetree/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileTreePaths(ctx context.Context,  req *types.QueryFileTreePathsRequest) (*types.QueryFileTreePathsResponse, error) {
	if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    // TODO: Process the query

	return &types.QueryFileTreePathsResponse{}, nil
}
