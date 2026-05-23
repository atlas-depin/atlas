package keeper

import (
	"context"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Provider(ctx context.Context, req *types.QueryProviderRequest) (*types.QueryProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	provider, err := q.k.GetProvider(ctx, req.Address)

	if err == types.ErrProviderNotFound {
		return &types.QueryProviderResponse{
			Provider: nil,
		}, nil
	} else {
		return &types.QueryProviderResponse{
			Provider: &provider,
		}, err
	}
}
