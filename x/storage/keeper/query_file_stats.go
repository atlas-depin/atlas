package keeper

import (
	"context"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) FileStats(ctx context.Context, req *types.QueryFileStatsRequest) (*types.QueryFileStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	totalFiles, _ := q.k.FileCount.Get(ctx)

	var strayFiles int64 = 0
	var deadFiles int64 = 0
	iterStray, _ := q.k.FileStore.Indexes.ByStatus.MatchExact(ctx, "STRAY")
	for ; iterStray.Valid(); iterStray.Next() {
		strayFiles++
	}
	iterDead, _ := q.k.FileStore.Indexes.ByStatus.MatchExact(ctx, "DEAD")
	for ; iterDead.Valid(); iterDead.Next() {
		deadFiles++
	}

	return &types.QueryFileStatsResponse{
		TotalFiles: int64(totalFiles),
		StrayFiles: strayFiles,
		DeadFiles:  deadFiles,
	}, nil
}
