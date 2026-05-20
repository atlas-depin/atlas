package keeper

import (
	"context"

	"atlas/x/storage/types"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) StorageStats(ctx context.Context, req *types.QueryStorageStatsRequest) (*types.QueryStorageStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var active int32 = 0
	var providers int32 = 0
	var spacePurchased int64 = 0
	var spaceUsed int64 = 0
	var spaceReplicas int64 = 0
	var spaceAvailable int64 = 0

	if err := q.k.IterateAllStorageSubscriptions(ctx, nil, func(_ collections.Pair[string, string], subscription types.StorageSubscription) bool {
		if subscription.Status == "ACTIVE" {
			active++
			spacePurchased += subscription.SpaceAvailable
			spaceUsed += subscription.SpaceUsed
			spaceReplicas += subscription.ReplicaSpaceUsed
		}
		return false
	}); err != nil {
		return nil, err
	}

	if err := q.k.IterateAllProviders(ctx, func(_ string, provider types.Provider) bool {
		providers++
		spaceAvailable += provider.SpaceAvailable
		return false
	}); err != nil {
		return nil, err
	}

	return &types.QueryStorageStatsResponse{
		SubscriptionCount: active,
		SpacePurchased:    spacePurchased / 3,
		SpaceUsed:         spaceUsed / 3,
		SpaceReplicas:     spaceReplicas,
		ProviderCount:     providers,
		SpaceAvailable:    spaceAvailable,
	}, nil
}
