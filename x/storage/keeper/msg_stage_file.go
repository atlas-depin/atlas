package keeper

import (
	"context"

	"nebulix/x/storage/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) StageFile(ctx context.Context, msg *types.MsgStageFile) (*types.MsgStageFileResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	stagedFile, err := k.GetStagedFile(ctx, msg.Fid)
	if err == sdkerrors.ErrNotFound {
		stagedFile = types.StagedFile{
			Providers: []string{msg.Provider},
			Start:     sdkCtx.BlockHeight(),
		}
	} else if err != nil {
		return nil, err
	} else {
		stagedFile.Providers = append(stagedFile.Providers, msg.Provider)
	}

	k.SetStagedFile(ctx, msg.Fid, stagedFile)

	return &types.MsgStageFileResponse{}, nil
}
