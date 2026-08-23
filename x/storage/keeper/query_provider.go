package keeper

import (
	"context"
	"errors"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Provider(ctx context.Context, req *types.QueryProviderRequest) (*types.QueryProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	provider, err := q.k.GetProvider(ctx, req.Address)
	if errors.Is(err, types.ErrProviderNotFound) {
		return nil, status.Error(codes.NotFound, "provider not found")
	} else if err != nil {
		return nil, err
	}

	return &types.QueryProviderResponse{
		Provider: &provider,
	}, nil
}
