package keeper

import (
	"context"
	"strings"

	"atlas/x/filetree/types"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) TreeNodeChildren(ctx context.Context, req *types.QueryTreeNodeChildrenRequest) (*types.QueryTreeNodeChildrenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var nodes []*types.TreeNode
	basepath := strings.ToLower(strings.Trim(req.Path, "/"))

	var rng *collections.Range[collections.Triple[string, string, string]]

	rng = new(collections.Range[collections.Triple[string, string, string]]).Prefix(
		collections.TripleSuperPrefix[string, string, string](req.Owner+"_"+req.Subscription, basepath),
	)

	err := q.k.IterateNodes(ctx, rng, func(key collections.Triple[string, string, string], node types.TreeNode) (_ bool, _ error) {
		nodes = append(nodes, &node)
		return false, nil
	})

	return &types.QueryTreeNodeChildrenResponse{Nodes: nodes}, err
}
