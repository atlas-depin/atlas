package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

func SimulateMsgPostFile(
	ak types.AuthKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
	txGen client.TxConfig,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgPostFile{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handle the PostFile simulation

		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "PostFile simulation not implemented"), nil, nil
	}
}
