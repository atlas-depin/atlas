package keeper

import (
	"context"
	"strings"

	"nebulix/x/filetree/types"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileNodeChildren(ctx context.Context, req *types.QueryFileNodeChildrenRequest) (*types.QueryFileNodeChildrenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var nodes []*types.FileNode
	basepath := strings.ToLower(strings.TrimSuffix(req.Path, "/"))

	var rng *collections.Range[collections.Triple[string, string, string]]

	rng = new(collections.Range[collections.Triple[string, string, string]]).Prefix(
		collections.TripleSuperPrefix[string, string, string](req.Owner, basepath),
	)

	err := q.k.IterateNodes(ctx, rng, func(key collections.Triple[string, string, string], node types.FileNode) (_ bool, _ error) {
		nodes = append(nodes, &node)
		return false, nil
	})

	return &types.QueryFileNodeChildrenResponse{Nodes: nodes}, err
}
