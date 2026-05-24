package ante

import (
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"atlas/x/storage/keeper"
)

// BuildAnteHandler constructs the full ante handler decorator chain,
// replacing only the fee deduction logic with our provider-aware TxFeeChecker.
//
// The chain matches the SDK 0.53 default exactly, with the exception that
// NewDeductFeeDecorator receives our custom checker.
func BuildAnteHandler(
	ak ante.AccountKeeper,
	bk authtypes.BankKeeper,
	sk *keeper.Keeper,
	signModeHandler *txsigning.HandlerMap,
	maxProveFilePerBlock int,
) sdk.AnteHandler {
	checker := NewProveFileExemptChecker(sk, maxProveFilePerBlock)

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(nil),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		// Custom fee deduction with provider ProveFile exemption
		ante.NewDeductFeeDecorator(ak, bk, nil, checker.Check),
		ante.NewSetPubKeyDecorator(ak),
		ante.NewValidateSigCountDecorator(ak),
		ante.NewSigGasConsumeDecorator(ak, ante.DefaultSigVerificationGasConsumer),
		ante.NewSigVerificationDecorator(ak, signModeHandler),
		ante.NewIncrementSequenceDecorator(ak),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
