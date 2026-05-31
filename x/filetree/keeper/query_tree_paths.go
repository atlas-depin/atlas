package keeper

import (
	"context"
	"strings"

	"atlas/x/filetree/types"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileTreePaths(ctx context.Context, req *types.QueryFileTreePathsRequest) (*types.QueryFileTreePathsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var paths []string
	basepath := strings.ToLower(strings.Trim(req.Basepath, "/"))

	rng := new(collections.Range[collections.Triple[string, string, string]]).Prefix(
		collections.Join3(req.Owner, basepath, ""),
	)

	err := q.k.IterateNodes(ctx, rng, func(key collections.Triple[string, string, string], node types.TreeNode) (_ bool, _ error) {
		paths = append(paths, key.K3())
		return false, nil
	})

	return &types.QueryFileTreePathsResponse{Paths: paths}, err
}
