package keeper

import (
	"context"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Providers(ctx context.Context, req *types.QueryProvidersRequest) (*types.QueryProvidersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	providers, err := q.k.GetAllProviders(ctx)

	return &types.QueryProvidersResponse{
		Providers: providers,
	}, err
}
