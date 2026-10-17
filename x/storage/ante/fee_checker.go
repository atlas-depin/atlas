package ante

import (
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"atlas/x/storage/keeper"
	"atlas/x/storage/types"
)

const DefaultMaxProveFilePerBlock = 100

// ProveFileExemptChecker is a TxFeeChecker that exempts registered providers
// submitting only MsgProveFile transactions, and the designated price updater
// submitting only MsgUpdatePrice transactions, from fees.
type ProveFileExemptChecker struct {
	storageKeeper *keeper.Keeper
	maxPerBlock   int

	// block-level rate limiter state
	blockHeight int64
	proofCounts map[string]int // provider address -> count this block
}

func NewProveFileExemptChecker(sk *keeper.Keeper, maxPerBlock int) *ProveFileExemptChecker {
	if maxPerBlock <= 0 {
		maxPerBlock = DefaultMaxProveFilePerBlock
	}
	return &ProveFileExemptChecker{
		storageKeeper: sk,
		maxPerBlock:   maxPerBlock,
		proofCounts:   make(map[string]int),
	}
}

// Check implements the TxFeeChecker signature.
// For registered providers submitting only MsgProveFile, and for the
// designated price updater submitting only MsgUpdatePrice, it returns zero fee.
// All other transactions fall through to the standard fee check logic.
func (c *ProveFileExemptChecker) Check(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	msgs := feeTx.GetMsgs()

	// Provider exemption: only MsgProveFile txs from a registered provider.
	if allProveFileMsgs(msgs) {
		return c.checkProveFileExempt(ctx, tx, feeTx)
	}

	// Price updater exemption: only MsgUpdatePrice txs from the designated updater.
	if allUpdatePriceMsgs(msgs) {
		return c.checkUpdatePriceExempt(ctx, tx, feeTx)
	}

	return defaultFeeCheck(ctx, tx)
}

// checkProveFileExempt grants zero-fee to registered providers submitting only
// MsgProveFile txs, subject to a per-block rate limit.
func (c *ProveFileExemptChecker) checkProveFileExempt(ctx sdk.Context, tx sdk.Tx, feeTx sdk.FeeTx) (sdk.Coins, int64, error) {
	// fee payer must be a registered provider
	feePayer := feeTx.FeePayer()
	if feePayer == nil {
		return defaultFeeCheck(ctx, tx)
	}

	feePayerStr := sdk.AccAddress(feePayer).String()
	if _, err := c.storageKeeper.GetProvider(ctx, feePayerStr); err != nil {
		return defaultFeeCheck(ctx, tx)
	}

	// rate limit: max ProveFile txs per provider per block
	if ctx.BlockHeight() != c.blockHeight {
		c.blockHeight = ctx.BlockHeight()
		c.proofCounts = make(map[string]int)
	}
	c.proofCounts[feePayerStr]++
	if c.proofCounts[feePayerStr] > c.maxPerBlock {
		return sdk.Coins{}, 0, fmt.Errorf(
			"rate limit: max %d ProveFile txs per block per provider",
			c.maxPerBlock,
		)
	}

	return sdk.Coins{}, 0, nil
}

// checkUpdatePriceExempt grants zero-fee to the designated price updater
// (params.PriceUpdater) when submitting only MsgUpdatePrice txs.
//
// No per-block rate limit is applied: only the single trusted updater account
// can produce valid updates (the msg handler rejects any other creator), so
// there is no spam vector beyond that account itself.
func (c *ProveFileExemptChecker) checkUpdatePriceExempt(ctx sdk.Context, tx sdk.Tx, feeTx sdk.FeeTx) (sdk.Coins, int64, error) {
	feePayer := feeTx.FeePayer()
	if feePayer == nil {
		return defaultFeeCheck(ctx, tx)
	}

	params, err := c.storageKeeper.Params.Get(ctx)
	if err != nil || params.PriceUpdater == "" {
		return defaultFeeCheck(ctx, tx)
	}

	if sdk.AccAddress(feePayer).String() != params.PriceUpdater {
		return defaultFeeCheck(ctx, tx)
	}

	return sdk.Coins{}, 0, nil
}

// allProveFileMsgs returns true only if every message is MsgProveFile.
func allProveFileMsgs(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*types.MsgProveFile); !ok {
			return false
		}
	}
	return true
}

// allUpdatePriceMsgs returns true only if every message is MsgUpdatePrice.
func allUpdatePriceMsgs(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*types.MsgUpdatePrice); !ok {
			return false
		}
	}
	return true
}

// =============================== //
// ---- Copied from CosmosSDK ---- //

// defaultFeeCheck implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, and the tx priority is computed from the gas price.
func defaultFeeCheck(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Mempool check: fees must meet validator min gas prices
	if ctx.IsCheckTx() {
		minGasPrices := ctx.MinGasPrices()
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdkmath.LegacyNewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}

			if !feeCoins.IsAnyGTE(requiredFees) {
				return nil, 0, errorsmod.Wrapf(
					sdkerrors.ErrInsufficientFee,
					"insufficient fees; got: %s required: %s",
					feeCoins, requiredFees,
				)
			}
		}
	}

	priority := getTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritize as expected.
func getTxPriority(fee sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
