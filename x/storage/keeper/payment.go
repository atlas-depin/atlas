package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NOTE: I know I can pull this from the bankkeeper but this is just way more efficient and we should never be changing this anyways....
const (
	uatlUnit int64 = 1_000_000
)

func (k Keeper) GetStorageCost(ctx sdk.Context, gbs int64, days int64) math.Int {
	params, _ := k.Params.Get(ctx)
	pricePerGbDay := math.NewInt(params.PricePerGbDay)

	totalCost := pricePerGbDay.Mul(math.NewInt(gbs)).Mul(math.NewInt(days))

	// [TODO]: oracle for price fetching
	totalCostDec := math.LegacyNewDecFromBigInt(totalCost.BigInt()).Quo(math.LegacyNewDec(1000000))
	atlPrice, _ := math.LegacyNewDecFromStr("1")
	// ^ temporary price for testing

	atlCost := totalCostDec.Quo(atlPrice)

	uatlUnit := math.LegacyNewDec(1_000_000)
	uatlCost := atlCost.Mul(uatlUnit)

	return uatlCost.Ceil().TruncateInt()
}
