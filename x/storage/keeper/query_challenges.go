package keeper

import (
	"context"

	"atlas/x/storage/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Challenges(ctx context.Context, req *types.QueryChallengesRequest) (*types.QueryChallengesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Set default pagination if not provided
	if req.Pagination == nil {
		req.Pagination = &query.PageRequest{
			Limit: query.DefaultLimit,
		}
	}

	var queryOpts []func(opt *query.CollectionsPaginateOptions[collections.Pair[string, string]])
	if req.Provider != "" {
		if _, err := q.k.addressCodec.StringToBytes(req.Provider); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid provider address: %v", err)
		}
		queryOpts = append(queryOpts, query.WithCollectionPaginationPairPrefix[string, string](req.Provider))
	}

	challenges, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.ChallengeStore,
		req.Pagination,
		func(key collections.Pair[string, string], value types.StorageChallenge) (*types.StorageChallenge, error) {
			return &value, nil
		},
		queryOpts...,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryChallengesResponse{
		Challenges: challenges,
		Pagination: pageRes,
	}, nil
}
