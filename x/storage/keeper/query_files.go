package keeper

import (
	"context"

	"atlas/x/storage/types"

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
	if req.Creator != "" {
		if _, err := q.k.addressCodec.StringToBytes(req.Creator); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid creator address: %v", err)
		}
	}

	var queryOpts []func(opt *query.CollectionsPaginateOptions[collections.Triple[string, string, string]])
	switch {
	case req.Creator != "" && req.Subscription != "":
		queryOpts = append(queryOpts, withCollectionPaginationTripleSuperPrefix[string, string, string](req.Creator, req.Subscription))
	case req.Creator != "":
		queryOpts = append(queryOpts, withCollectionPaginationTriplePrefix[string, string, string](req.Creator))
	case req.Subscription != "":
		return nil, status.Error(codes.InvalidArgument, "need to provide creator to query by subscription")
	}

	files, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.FileStore,
		req.Pagination,
		func(key collections.Triple[string, string, string], value types.File) (*types.File, error) {
			return &value, nil
		},
		queryOpts...,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryFilesResponse{
		Files:      files,
		Pagination: pageRes,
	}, nil
}

// WithCollectionPaginationPairPrefix applies a prefix to a collection, whose key is a collection.Pair,
// being paginated that needs prefixing.
func withCollectionPaginationTriplePrefix[K1, K2, K3 any](k1 K1) func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
	return func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
		prefix := collections.TriplePrefix[K1, K2, K3](k1)
		o.Prefix = &prefix
	}
}

func withCollectionPaginationTripleSuperPrefix[K1, K2, K3 any](k1 K1, k2 K2) func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
	return func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
		prefix := collections.TripleSuperPrefix[K1, K2, K3](k1, k2)
		o.Prefix = &prefix
	}
}
