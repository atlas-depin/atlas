package keeper

import (
	"context"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeleteProvider(ctx context.Context, msg *types.MsgDeleteProvider) (*types.MsgDeleteProviderResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgDeleteProviderResponse{}, nil
}
