package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Provider(ctx context.Context, req *types.QueryProviderRequest) (*types.QueryProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query

	return &types.QueryProviderResponse{}, nil
}
