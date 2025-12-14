package storage

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"nebulix/x/storage/types"
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
					RpcMethod:      "GetSubscriptionInfo",
					Use:            "get-subscription-info [subscriber-address] [subscription-id]",
					Short:          "Query get-subscription-info",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subscriber_address"}, {ProtoField: "subscription_id"}},
				},

				{
					RpcMethod:      "Provider",
					Use:            "provider [address]",
					Short:          "Query provider",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
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
					Use:            "post-file [merkle] [index] [file-size] [replicas] [subscription]",
					Short:          "Send a post-file tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "merkle"}, {ProtoField: "ddi"}, {ProtoField: "file_size"}, {ProtoField: "replicas"}, {ProtoField: "subscription"}},
				},
				{
					RpcMethod:      "StageFile",
					Use:            "stage-file [provider] [fid] [file-creator] [file-size]",
					Short:          "Send a stage-file tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "provider"}, {ProtoField: "fid"}, {ProtoField: "file_creator"}, {ProtoField: "file_size"}},
				},
				{
					RpcMethod:      "BuyStorage",
					Use:            "buy-storage [receiver] [duration] [bytes] [is-default]",
					Short:          "Send a buy-storage tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "receiver"}, {ProtoField: "duration"}, {ProtoField: "bytes"}, {ProtoField: "is_default"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
