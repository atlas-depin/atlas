package keeper

import (
	"context"
	"encoding/binary"

	"atlas/x/storage/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) Strays(ctx context.Context, req *types.QueryStraysRequest) (*types.QueryStraysResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Pagination == nil {
		req.Pagination = &query.PageRequest{
			Limit: query.DefaultLimit,
		}
	}

	limit := req.Pagination.Limit
	if limit <= 0 {
		limit = query.DefaultLimit
	}

	// Iterate over stray files.
	iter, err := q.k.FileStore.Indexes.ByStatus.MatchExact(ctx, "STRAY")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	// Decode the offset from the pagination cursor.
	// A nil/empty cursor means start from the beginning.
	var offset uint64
	if len(req.Pagination.Key) == 8 {
		offset = binary.BigEndian.Uint64(req.Pagination.Key)
	}

	var files []*types.File
	var idx uint64

	pageRes := &query.PageResponse{}

	for ; iter.Valid(); iter.Next() {
		if idx < offset {
			idx++
			continue
		}

		pk, err := iter.PrimaryKey()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		file, err := q.k.FileStore.Get(ctx, pk)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		files = append(files, &file)
		idx++

		if uint64(len(files)) >= limit {
			// Only emit a cursor when more entries remain after this page,
			// so clients stop fetching once the last page is returned.
			iter.Next()
			if iter.Valid() {
				nextKey := make([]byte, 8)
				binary.BigEndian.PutUint64(nextKey, offset+uint64(len(files)))
				pageRes.NextKey = nextKey
			}
			break
		}
	}

	return &types.QueryStraysResponse{
		Files:      files,
		Pagination: pageRes,
	}, nil
}
