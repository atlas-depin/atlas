package keeper

import (
	"context"

	"nebulix/x/storage/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeleteFile(ctx context.Context, msg *types.MsgDeleteFile) (*types.MsgDeleteFileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgDeleteFileResponse{}, nil
}
