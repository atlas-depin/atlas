package keeper_test

import (
	"cosmossdk.io/math"

	"atlas/x/storage/types"
)

// GetStorageCost computes pricePerGbDay * gbs * days in uatl.
// With the current fixed ATL price of 1, the unit conversion is exact,
// so the result is always pricePerGbDay * gbs * days.
func (suite *KeeperTestSuite) TestGetStorageCost_DefaultPrice() {
	params := types.DefaultParams()
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	cases := []struct {
		gbs      int64
		days     int64
		expected int64
	}{
		{gbs: 1, days: 1, expected: 400},
		{gbs: 1, days: 30, expected: 12_000},
		{gbs: 5, days: 10, expected: 20_000},
		{gbs: 100, days: 365, expected: 14_600_000},
	}

	for _, tc := range cases {
		cost := suite.keeper.GetStorageCost(suite.ctx, tc.gbs, tc.days)
		suite.Require().Equal(math.NewInt(tc.expected), cost,
			"cost mismatch for gbs=%d days=%d", tc.gbs, tc.days)
	}
}

func (suite *KeeperTestSuite) TestGetStorageCost_CustomPrice() {
	params := types.DefaultParams()
	params.PricePerGbDay = 1000
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	cost := suite.keeper.GetStorageCost(suite.ctx, 2, 7)
	suite.Require().Equal(math.NewInt(14_000), cost)
}

func (suite *KeeperTestSuite) TestGetStorageCost_SmallAmounts() {
	params := types.DefaultParams()
	params.PricePerGbDay = 1
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// 1 * 1 * 1 -> the decimal division path must still land on the exact integer
	cost := suite.keeper.GetStorageCost(suite.ctx, 1, 1)
	suite.Require().Equal(math.NewInt(1), cost)

	cost = suite.keeper.GetStorageCost(suite.ctx, 1, 7)
	suite.Require().Equal(math.NewInt(7), cost)
}

func (suite *KeeperTestSuite) TestGetStorageCost_WithATLPrice() {
	params := types.DefaultParams()
	suite.Require().NoError(suite.keeper.Params.Set(suite.ctx, params))

	// cost_uatl = pricePerGbDay * gbs * days / atlPrice
	cases := []struct {
		name     string
		price    string
		gbs      int64
		days     int64
		expected int64
	}{
		{name: "atl at 0.0004 usd", price: "0.0004", gbs: 1, days: 1, expected: 1_000_000},
		{name: "atl at 0.0004 usd bulk", price: "0.0004", gbs: 5, days: 10, expected: 50_000_000},
		{name: "atl at 2 usd", price: "2", gbs: 1, days: 30, expected: 6_000},
		{name: "atl at 0.5 usd", price: "0.5", gbs: 3, days: 7, expected: 16_800},
		{name: "canine style price:marketcap", price: "0.0004:123456789", gbs: 1, days: 1, expected: 1_000_000},
		{name: "invalid price falls back to 1", price: "garbage", gbs: 1, days: 1, expected: 400},
		{name: "zero price falls back to 1", price: "0.000000", gbs: 1, days: 1, expected: 400},
	}

	for _, tc := range cases {
		suite.Run(tc.name, func() {
			suite.Require().NoError(suite.keeper.Price.Set(suite.ctx, tc.price))

			cost := suite.keeper.GetStorageCost(suite.ctx, tc.gbs, tc.days)
			suite.Require().Equal(math.NewInt(tc.expected), cost)
		})
	}
}
