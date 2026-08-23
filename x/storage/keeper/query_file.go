package keeper

import (
	"context"
	"errors"

	"atlas/x/storage/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) File(ctx context.Context, req *types.QueryFileRequest) (*types.QueryFileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	file, err := q.k.GetFile(ctx, req.Fid)
	if err != nil {
		if errors.Is(err, types.ErrFileNotFound) {
			return nil, status.Error(codes.NotFound, "file not found")
		}
		return nil, err
	}

	return &types.QueryFileResponse{File: &file}, nil
}
