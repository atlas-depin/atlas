package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	storage "atlas/x/storage/module"
)

// MakeEncodingConfig creates encoding config for your specific blockchain
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	// Register standard Cosmos SDK types
	std.RegisterLegacyAminoCodec(amino)
	std.RegisterInterfaces(interfaceRegistry)

	// Create module basics
	moduleBasics := []module.AppModuleBasic{
		storage.AppModule{},
	}

	// Register all modules
	for _, mb := range moduleBasics {
		mb.RegisterLegacyAminoCodec(amino)
		mb.RegisterInterfaces(interfaceRegistry)
	}

	// Create tx config
	txConfig := tx.NewTxConfig(marshaler, []signing.SignMode{
		signing.SignMode_SIGN_MODE_DIRECT,
		signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	})

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          txConfig,
		Amino:             amino,
	}
}

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}
