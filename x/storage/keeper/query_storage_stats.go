package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) StorageStats(ctx context.Context, req *types.QueryStorageStatsRequest) (*types.QueryStorageStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query

	return &types.QueryStorageStatsResponse{}, nil
}
