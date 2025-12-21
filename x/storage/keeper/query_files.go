package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Files(ctx context.Context, req *types.QueryFilesRequest) (*types.QueryFilesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Set default pagination if not provided
	if req.Pagination == nil {
		req.Pagination = &query.PageRequest{
			Limit: query.DefaultLimit,
		}
	}

	files, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.FileStore,
		req.Pagination,
		func(key collections.Triple[string, string, string], value types.File) (*types.File, error) {
			return &value, nil
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryFilesResponse{
		Files:      files,
		Pagination: pageRes,
	}, nil
}
