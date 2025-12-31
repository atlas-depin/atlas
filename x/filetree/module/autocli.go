package filetree

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"nebulix/x/filetree/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod:      "FileTreeNode",
					Use:            "file-tree-node [path]",
					Short:          "Query file-tree-node",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "path"}},
				},

				{
					RpcMethod:      "FileTreePaths",
					Use:            "file-tree-paths [basepath]",
					Short:          "Query file-tree-paths",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "basepath"}},
				},

				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "PostNode",
					Use:            "post-node [path]",
					Short:          "Send a post-node tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "path"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
