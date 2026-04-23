package keeper

import (
	"context"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) File(ctx context.Context, req *types.QueryFileRequest) (*types.QueryFileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	file, err := q.k.GetFile(ctx, req.Fid)

	return &types.QueryFileResponse{File: &file}, err
}
