package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Subscription(ctx context.Context, req *types.QuerySubscriptionRequest) (*types.QuerySubscriptionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	subscription, err := q.k.GetStorageSubscription(ctx, req.SubscriberAddress, req.SubscriptionId)
	if err != nil {
		return nil, err
	}

	return &types.QuerySubscriptionResponse{Subscription: &subscription}, nil
}
