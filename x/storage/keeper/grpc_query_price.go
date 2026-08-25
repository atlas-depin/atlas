package keeper

import (
	"context"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Price queries the current $ATL price.
func (q queryServer) Price(ctx context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	price, err := q.k.Price.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryPriceResponse{Price: price}, nil
}
