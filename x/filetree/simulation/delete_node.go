package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"atlas/x/filetree/keeper"
	"atlas/x/filetree/types"
)

func SimulateMsgDeleteNode(
	ak types.AuthKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
	txGen client.TxConfig,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDeleteNode{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handle the DeleteNode simulation

		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "DeleteNode simulation not implemented"), nil, nil
	}
}
