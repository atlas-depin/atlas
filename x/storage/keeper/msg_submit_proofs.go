package keeper

import (
	"context"
	"atlas/x/storage/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgPostProofs struct {
	Proofs   []types.Proof
	Provider string
}

type MsgPostProofsResponse struct {
}

func (k msgServer) SubmitProofs(goCtx context.Context, msg *MsgPostProofs) (*MsgPostProofsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	/// JUNK TO BE REUSED ELSEWHERE
	// ranger := collections.NewPrefixedPairRange[string, string](msg.Provider)
	// iter, err := k.ChallengeStore.Iterate(ctx, ranger)
	// if err != nil {
	// 	return nil, err
	// }

	var passed []string = []string{}
	var failed []string = []string{}

	for _, proof := range msg.Proofs {
		err := k.AttemptChallenge(ctx, msg.Provider, proof)
		if err != nil {
			failed = append(failed, proof.ChallengeId)
		} else {
			passed = append(passed, proof.ChallengeId)
		}
	}

	return &MsgPostProofsResponse{}, nil
}
