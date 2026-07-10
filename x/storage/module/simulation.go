package storage

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	storagesimulation "atlas/x/storage/simulation"
	"atlas/x/storage/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	storageGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&storageGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgInitProvider          = "op_weight_msg_storage"
		defaultWeightMsgInitProvider int = 100
	)

	var weightMsgInitProvider int
	simState.AppParams.GetOrGenerate(opWeightMsgInitProvider, &weightMsgInitProvider, nil,
		func(_ *rand.Rand) {
			weightMsgInitProvider = defaultWeightMsgInitProvider
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgInitProvider,
		storagesimulation.SimulateMsgRegisterProvider(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	const (
		opWeightMsgPostFile          = "op_weight_msg_storage"
		defaultWeightMsgPostFile int = 100
	)

	var weightMsgPostFile int
	simState.AppParams.GetOrGenerate(opWeightMsgPostFile, &weightMsgPostFile, nil,
		func(_ *rand.Rand) {
			weightMsgPostFile = defaultWeightMsgPostFile
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgPostFile,
		storagesimulation.SimulateMsgPostFile(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgBuyStorage          = "op_weight_msg_storage"
		defaultWeightMsgBuyStorage int = 100
	)

	var weightMsgBuyStorage int
	simState.AppParams.GetOrGenerate(opWeightMsgBuyStorage, &weightMsgBuyStorage, nil,
		func(_ *rand.Rand) {
			weightMsgBuyStorage = defaultWeightMsgBuyStorage
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgBuyStorage,
		storagesimulation.SimulateMsgBuyStorage(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgProveFile          = "op_weight_msg_storage"
		defaultWeightMsgProveFile int = 100
	)

	var weightMsgProveFile int
	simState.AppParams.GetOrGenerate(opWeightMsgProveFile, &weightMsgProveFile, nil,
		func(_ *rand.Rand) {
			weightMsgProveFile = defaultWeightMsgProveFile
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgProveFile,
		storagesimulation.SimulateMsgProveFile(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteFile          = "op_weight_msg_storage"
		defaultWeightMsgDeleteFile int = 100
	)

	var weightMsgDeleteFile int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteFile, &weightMsgDeleteFile, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteFile = defaultWeightMsgDeleteFile
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteFile,
		storagesimulation.SimulateMsgDeleteFile(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgExpandStorage          = "op_weight_msg_storage"
		defaultWeightMsgExpandStorage int = 100
	)

	var weightMsgExpandStorage int
	simState.AppParams.GetOrGenerate(opWeightMsgExpandStorage, &weightMsgExpandStorage, nil,
		func(_ *rand.Rand) {
			weightMsgExpandStorage = defaultWeightMsgExpandStorage
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgExpandStorage,
		storagesimulation.SimulateMsgExpandStorage(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdateProvider          = "op_weight_msg_storage"
		defaultWeightMsgUpdateProvider int = 100
	)

	var weightMsgUpdateProvider int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateProvider, &weightMsgUpdateProvider, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateProvider = defaultWeightMsgUpdateProvider
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateProvider,
		storagesimulation.SimulateMsgUpdateProvider(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgDeleteProvider          = "op_weight_msg_storage"
		defaultWeightMsgDeleteProvider int = 100
	)

	var weightMsgDeleteProvider int
	simState.AppParams.GetOrGenerate(opWeightMsgDeleteProvider, &weightMsgDeleteProvider, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteProvider = defaultWeightMsgDeleteProvider
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteProvider,
		storagesimulation.SimulateMsgDeleteProvider(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
