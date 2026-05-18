package keeper

import (
	"context"

	"atlas/x/storage/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) UpdateProvider(ctx context.Context, msg *types.MsgUpdateProvider) (*types.MsgUpdateProviderResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgUpdateProviderResponse{}, nil
}
