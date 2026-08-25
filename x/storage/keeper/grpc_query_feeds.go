package keeper

import (
	"context"
	"errors"

	"atlas/x/storage/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Feed queries a single feed by name.
func (q queryServer) Feed(ctx context.Context, req *types.QueryFeedRequest) (*types.QueryFeedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	feed, err := q.k.GetFeed(ctx, req.Name)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "feed not found")
		}
		return nil, err
	}

	return &types.QueryFeedResponse{Feed: feed}, nil
}

// AllFeeds queries all feeds with pagination.
func (q queryServer) AllFeeds(ctx context.Context, req *types.QueryAllFeedsRequest) (*types.QueryAllFeedsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	feeds, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Feeds,
		req.Pagination,
		func(_ string, feed types.Feed) (types.Feed, error) {
			return feed, nil
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryAllFeedsResponse{
		Feeds:      feeds,
		Pagination: pageRes,
	}, nil
}
