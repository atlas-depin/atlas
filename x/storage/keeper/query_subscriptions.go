package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Subscriptions(ctx context.Context, req *types.QuerySubscriptionsRequest) (*types.QuerySubscriptionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var queryOpts []func(opt *query.CollectionsPaginateOptions[collections.Pair[string, string]])
	if req.SubscriberAddress != "" {
		if _, err := q.k.addressCodec.StringToBytes(req.SubscriberAddress); err != nil {
			return nil, status.Error(codes.InvalidArgument, "bad subscriber address")
		}
		queryOpts = append(queryOpts, query.WithCollectionPaginationPairPrefix[string, string](req.SubscriberAddress))
	}

	subscriptions, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.SubscriptionStore,
		req.Pagination,
		func(key collections.Pair[string, string], value types.StorageSubscription) (*types.StorageSubscription, error) {
			return &value, nil
		},
		queryOpts...,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QuerySubscriptionsResponse{
		Subscriptions: subscriptions,
		Pagination:    pageRes,
	}, nil
}
