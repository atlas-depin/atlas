package keeper

import (
	"context"
	"strings"

	"nebulix/x/filetree/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileTreePaths(ctx context.Context, req *types.QueryFileTreePathsRequest) (*types.QueryFileTreePathsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var paths []string
	basepath := strings.TrimSuffix(req.Basepath, "/") + "/"

	_, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.FileTree,
		req.Pagination,
		func(key collections.Triple[string, string, string], _ types.FileTreeNode) (string, error) {
			if key.K2() == basepath {
				paths = append(paths, key.K3())
			}
			return "", nil
		},
		WithCollectionPaginationTriplePrefix[string, string, string](req.Owner, basepath),
	)

	return &types.QueryFileTreePathsResponse{Paths: paths, Pagination: pageRes}, err
}

func WithCollectionPaginationTriplePrefix[K1, K2, K3 any](k1 K1, k2 K2) func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
	return func(o *query.CollectionsPaginateOptions[collections.Triple[K1, K2, K3]]) {
		prefix := collections.TripleSuperPrefix[K1, K2, K3](k1, k2)
		o.Prefix = &prefix
	}
}
