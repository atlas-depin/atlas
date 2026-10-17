package keeper

import (
	"context"
	"strings"

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

	totalCostDec := math.LegacyNewDecFromBigInt(totalCost.BigInt()).Quo(math.LegacyNewDec(1000000))
	atlPrice := k.getATLPrice(ctx)

	atlCost := totalCostDec.Quo(atlPrice)

	uatlUnit := math.LegacyNewDec(1_000_000)
	uatlCost := atlCost.Mul(uatlUnit)

	return uatlCost.Ceil().TruncateInt()
}

// getATLPrice returns the USD price of one ATL from the price feed, e.g.
// "0.0004" means 1 ATL = $0.0004. canine-chain style "price:marketCap"
// strings are tolerated by only reading the part before the colon.
//
// Falls back to $1 until a valid price is set, keeping the chain functional
// before the first MsgUpdatePrice. The genesis seed is types.DefaultPrice
// ("1.000000").
func (k Keeper) getATLPrice(ctx context.Context) math.LegacyDec {
	raw, err := k.Price.Get(ctx)
	if err != nil {
		return math.LegacyNewDec(1)
	}

	priceStr := raw
	if idx := strings.IndexByte(raw, ':'); idx != -1 {
		priceStr = raw[:idx]
	}

	price, err := math.LegacyNewDecFromStr(priceStr)
	if err != nil || !price.IsPositive() {
		return math.LegacyNewDec(1)
	}

	return price
}

// func (k Keeper) GetStorageExpansionCost(ctx sdk.Context, gbs int64, days int64, delta_s int64) math.Int {
// 	params, _ := k.Params.Get(ctx)
// 	pricePerGbDay := params.PricePerGbDay

// 	expandCost := float64(pricePerGbDay*gbs) * float64(delta_s) / 86400
// 	newCost := float64(pricePerGbDay * gbs * days)

// 	// [TODO]: oracle for price fetching
// 	totalCost := (newCost + expandCost) / 1000000
// 	atlPrice := 1.0
// 	// ^ temporary price for testing

// 	atlCost := totalCost / atlPrice

// 	uatlUnit := 1_000_000.0
// 	uatlCost := atlCost * uatlUnit

// 	return math.NewInt(int64(gomath.Ceil(uatlCost)))
// }
