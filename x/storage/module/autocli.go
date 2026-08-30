package storage

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"atlas/x/storage/types"
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
					RpcMethod:      "Provider",
					Use:            "provider [address]",
					Short:          "Query provider",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},

				{
					RpcMethod:      "Providers",
					Use:            "providers ",
					Short:          "Query providers",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "File",
					Use:            "file [fid]",
					Short:          "Query file",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "fid"}},
				},

				{
					RpcMethod:      "Files",
					Use:            "files ",
					Short:          "Query files",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "Subscription",
					Use:            "subscription ",
					Short:          "Query subscription",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "Subscriptions",
					Use:            "subscriptions ",
					Short:          "Query subscriptions",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "Challenges",
					Use:            "challenges ",
					Short:          "Query challenges",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "StorageStats",
					Use:            "storage-stats ",
					Short:          "Query storage-stats",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "FileStats",
					Use:            "file-stats [stray-files] [dead-files]",
					Short:          "Query file-stats",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "Strays",
					Use:            "strays ",
					Short:          "Query strays",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
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
				// {
				// 	RpcMethod:      "InitProvider",
				// 	Use:            "init-provider [address] [hostname] [total-space]",
				// 	Short:          "Send a init-provider tx",
				// 	PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}, {ProtoField: "hostname"}, {ProtoField: "total_space"}},
				// },
				{
					RpcMethod:      "PostFile",
					Use:            "post-file [merkle] [file-size] [replicas] [subscription]",
					Short:          "Send a post-file tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "merkle"}, {ProtoField: "file_size"}, {ProtoField: "replicas"}, {ProtoField: "subscription"}},
				},
				{
					RpcMethod:      "BuyStorage",
					Use:            "buy-storage [duration] [bytes] [is-default]",
					Short:          "Send a buy-storage tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "duration"}, {ProtoField: "bytes"}, {ProtoField: "is_default"}},
				},
				{
					RpcMethod:      "ProveFile",
					Use:            "prove-file [challenge-id] [file-id] [data] [hashes] [chunk]",
					Short:          "Send a prove-file tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "challenge_id"}, {ProtoField: "fid"}, {ProtoField: "data"}, {ProtoField: "hashes"}, {ProtoField: "chunk"}},
				},
				{
					RpcMethod:      "DeleteFile",
					Use:            "delete-file [file-id]",
					Short:          "Send a delete-file tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "file_id"}},
				},
				{
					RpcMethod:      "ExpandStorage",
					Use:            "expand-storage [subscription-id] [duration] [bytes]",
					Short:          "Send a expand-storage tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subscription_id"}, {ProtoField: "duration"}, {ProtoField: "bytes"}},
				},
				{
					RpcMethod:      "UpdateProvider",
					Use:            "update-provider [hostname] [capacity]",
					Short:          "Send a update-provider tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "hostname"}, {ProtoField: "capacity"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
				{
					RpcMethod:      "DeleteProvider",
					Use:            "delete-provider ",
					Short:          "Send a delete-provider tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
			},
		},
	}
}
