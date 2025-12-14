package keeper

import (
	"context"

	"nebulix/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) GetSubscriptionInfo(ctx context.Context, req *types.QueryGetSubscriptionInfoRequest) (*types.QueryGetSubscriptionInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	subscription, err := q.k.GetStorageSubscription(ctx, req.SubscriberAddress, req.SubscriptionId)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetSubscriptionInfoResponse{Subscription: &subscription}, nil
}
