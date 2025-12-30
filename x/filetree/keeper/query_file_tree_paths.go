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

	paths, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.FileTree,
		req.Pagination,
		func(key collections.Pair[string, string], _ types.FileTreeNode) (string, error) {
			return key.K1(), nil
		},
		query.WithCollectionPaginationPairPrefix[string, string](strings.TrimSuffix(req.Basepath, "/")+"/"),
	)

	return &types.QueryFileTreePathsResponse{Paths: paths, Pagination: pageRes}, err
}
