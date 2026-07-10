package keeper

import (
	"context"

	"atlas/x/storage/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) StorageStats(ctx context.Context, req *types.QueryStorageStatsRequest) (*types.QueryStorageStatsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var active int32 = 0
	var providers int32 = 0
	var spacePurchased int64 = 0
	var spaceUsed int64 = 0
	var spaceReplicas int64 = 0
	var spaceAvailable int64 = 0

	if err := q.k.IterateAllStorageSubscriptions(ctx, func(_ collections.Pair[string, string], subscription types.StorageSubscription) bool {
		if sdkCtx.BlockTime().Before(subscription.End) {
			active++
			spacePurchased += subscription.SpaceAvailable
			spaceReplicas += subscription.ReplicaSpaceUsed
		}
		return false
	}); err != nil {
		return nil, err
	}

	if err := q.k.IterateAllProviders(ctx, func(_ string, provider types.Provider) bool {
		providers++
		spaceAvailable += provider.SpaceAvailable
		spaceUsed += provider.SpaceUsed
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
