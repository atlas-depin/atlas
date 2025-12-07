package keeper

import (
	"context"

	"nebulix/x/storage/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) StageFile(ctx context.Context, msg *types.MsgStageFile) (*types.MsgStageFileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgStageFileResponse{}, nil
}
