package hyperion

import (
	"context"
	"errors"
	"testing"
)

type queuedViewExecutor struct {
	payloads  []EntryFunctionPayload
	responses [][]any
	err       error
}

func (e *queuedViewExecutor) View(ctx context.Context, payload EntryFunctionPayload) ([]any, error) {
	e.payloads = append(e.payloads, payload)
	if e.err != nil {
		return nil, e.err
	}
	if len(e.responses) == 0 {
		return nil, errors.New("no queued response")
	}
	out := e.responses[0]
	e.responses = e.responses[1:]
	return out, nil
}

func TestTypedPositionViewHelpersDecode(t *testing.T) {
	t.Parallel()

	executor := &queuedViewExecutor{responses: [][]any{
		{"100", "200"},
		{[]any{"10", "20"}},
		{[]any{map[string]any{"reward_fa": map[string]any{"inner": "0xreward"}, "amount_owed": "30"}}},
		{[]any{map[string]any{"reward_fa": "0xreward", "rate": "40"}}},
		{"12345"},
		{map[string]any{"inner": "0xa"}, map[string]any{"inner": "0xb"}, float64(2)},
		{map[string]any{"bits": float64(4294967293)}, map[string]any{"bits": "4294967295"}},
	}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	ctx := context.Background()

	amounts, err := sdk.Position.FetchPositionTokenAmounts(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPositionTokenAmounts returned error: %v", err)
	}
	if amounts.AmountA != "100" || amounts.AmountB != "200" {
		t.Fatalf("amounts = %#v", amounts)
	}

	fees, err := sdk.Position.FetchPendingFees(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPendingFees returned error: %v", err)
	}
	assertArguments(t, stringAnySlice(fees.Amounts), []any{"10", "20"})

	rewards, err := sdk.Position.FetchPendingRewards(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPendingRewards returned error: %v", err)
	}
	if len(rewards) != 1 || rewards[0].RewardFA != "0xreward" || rewards[0].AmountOwed != "30" {
		t.Fatalf("rewards = %#v", rewards)
	}

	rates, err := sdk.Position.FetchPositionEmissionRates(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPositionEmissionRates returned error: %v", err)
	}
	if len(rates) != 1 || rates[0].RewardFA != "0xreward" || rates[0].Rate != "40" {
		t.Fatalf("rates = %#v", rates)
	}

	liquidity, err := sdk.Position.FetchPositionLiquidity(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPositionLiquidity returned error: %v", err)
	}
	if liquidity != "12345" {
		t.Fatalf("liquidity = %q", liquidity)
	}

	poolInfo, err := sdk.Position.FetchPositionPoolInfo(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPositionPoolInfo returned error: %v", err)
	}
	if poolInfo.TokenA != "0xa" || poolInfo.TokenB != "0xb" || poolInfo.FeeTier != 2 {
		t.Fatalf("poolInfo = %#v", poolInfo)
	}

	ticks, err := sdk.Position.FetchPositionTickRange(ctx, "pos-1")
	if err != nil {
		t.Fatalf("FetchPositionTickRange returned error: %v", err)
	}
	if ticks.Lower.Bits != 4294967293 || ticks.Lower.Tick != -3 || ticks.Upper.Tick != -1 {
		t.Fatalf("ticks = %#v", ticks)
	}

	if executor.payloads[0].Function != MainnetContractAddress+"::router_v3::get_amount_by_liquidity" {
		t.Fatalf("amount function = %q", executor.payloads[0].Function)
	}
	if executor.payloads[5].Function != MainnetContractAddress+"::position_v3::get_pool_info" {
		t.Fatalf("pool info function = %q", executor.payloads[5].Function)
	}
}

func TestTypedOptimalLiquidityViewHelpersDecode(t *testing.T) {
	t.Parallel()

	executor := &queuedViewExecutor{responses: [][]any{
		{"123", "100", "200"},
		{"456", "220"},
		{"789", "330"},
	}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	ctx := context.Background()

	optimal, err := sdk.Pool.FetchOptimalLiquidityAmounts(ctx, OptimalLiquidityAmountsArgs{
		TickLower:          "-60",
		TickUpper:          "60",
		CurrencyA:          "0x1",
		CurrencyB:          "0x2",
		FeeTierIndex:       "2",
		CurrencyAAmount:    "100",
		CurrencyBAmount:    "200",
		MinCurrencyAAmount: "0",
		MinCurrencyBAmount: "0",
	})
	if err != nil {
		t.Fatalf("FetchOptimalLiquidityAmounts returned error: %v", err)
	}
	if optimal.Liquidity != "123" || optimal.CurrencyAAmount != "100" || optimal.CurrencyBAmount != "200" {
		t.Fatalf("optimal = %#v", optimal)
	}

	fromA, err := sdk.Pool.FetchOptimalLiquidityAmountsFromA(ctx, EstCurrencyBAmountFromAArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyAAmount:  "1000",
	})
	if err != nil {
		t.Fatalf("FetchOptimalLiquidityAmountsFromA returned error: %v", err)
	}
	if fromA.Liquidity != "456" || fromA.CurrencyBAmount != "220" {
		t.Fatalf("fromA = %#v", fromA)
	}

	fromB, err := sdk.Pool.FetchOptimalLiquidityAmountsFromB(ctx, EstCurrencyAAmountFromBArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyBAmount:  "2000",
	})
	if err != nil {
		t.Fatalf("FetchOptimalLiquidityAmountsFromB returned error: %v", err)
	}
	if fromB.Liquidity != "789" || fromB.CurrencyAAmount != "330" {
		t.Fatalf("fromB = %#v", fromB)
	}
}

func TestTypedPoolStateAndQuoteViewHelpersDecode(t *testing.T) {
	t.Parallel()

	executor := &queuedViewExecutor{responses: [][]any{
		{true, "0xpool"},
		{true},
		{"100", "200"},
		{float64(4294967293), "123456789"},
		{"222"},
		{"333"},
		{"30"},
		{[]any{map[string]any{"inner": "0xa"}, "0xb"}},
		{"900"},
		{"901"},
		{"902", "3"},
		{"903", "4"},
	}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	ctx := context.Background()

	pair := PoolTokenPairArgs{CurrencyA: "0x1", CurrencyB: "0x2", FeeTierIndex: "2"}
	address, err := sdk.Pool.FetchLiquidityPoolAddressSafe(ctx, pair)
	if err != nil {
		t.Fatalf("FetchLiquidityPoolAddressSafe returned error: %v", err)
	}
	if !address.Exists || address.Address != "0xpool" {
		t.Fatalf("address = %#v", address)
	}

	exists, err := sdk.Pool.FetchLiquidityPoolExists(ctx, pair)
	if err != nil {
		t.Fatalf("FetchLiquidityPoolExists returned error: %v", err)
	}
	if !exists {
		t.Fatal("exists = false")
	}

	reserve, err := sdk.Pool.FetchPoolReserveAmount(ctx, "0xpool")
	if err != nil {
		t.Fatalf("FetchPoolReserveAmount returned error: %v", err)
	}
	if reserve.AmountA != "100" || reserve.AmountB != "200" {
		t.Fatalf("reserve = %#v", reserve)
	}

	tickPrice, err := sdk.Pool.FetchCurrentTickAndPrice(ctx, "0xpool")
	if err != nil {
		t.Fatalf("FetchCurrentTickAndPrice returned error: %v", err)
	}
	if tickPrice.Tick.Tick != -3 || tickPrice.Price != "123456789" {
		t.Fatalf("tickPrice = %#v", tickPrice)
	}

	price, err := sdk.Pool.FetchCurrentPrice(ctx, pair)
	if err != nil {
		t.Fatalf("FetchCurrentPrice returned error: %v", err)
	}
	if price != "222" {
		t.Fatalf("price = %q", price)
	}

	liquidity, err := sdk.Pool.FetchPoolLiquidity(ctx, "0xpool")
	if err != nil {
		t.Fatalf("FetchPoolLiquidity returned error: %v", err)
	}
	if liquidity != "333" {
		t.Fatalf("liquidity = %q", liquidity)
	}

	feeRate, err := sdk.Pool.FetchFeeRate(ctx, "2")
	if err != nil {
		t.Fatalf("FetchFeeRate returned error: %v", err)
	}
	if feeRate != "30" {
		t.Fatalf("feeRate = %q", feeRate)
	}

	assets, err := sdk.Pool.FetchSupportedInnerAssets(ctx, "0xpool")
	if err != nil {
		t.Fatalf("FetchSupportedInnerAssets returned error: %v", err)
	}
	assertArguments(t, stringAnySlice(assets), []any{"0xa", "0xb"})

	batchOut, err := sdk.Swap.FetchBatchAmountOut(ctx, BatchAmountOutArgs{
		PoolRoute: []string{"0xpool"},
		AmountIn:  "1000",
		TokenIn:   "0x1",
		TokenOut:  "0x2",
	})
	if err != nil {
		t.Fatalf("FetchBatchAmountOut returned error: %v", err)
	}
	if batchOut != "900" {
		t.Fatalf("batchOut = %q", batchOut)
	}

	batchIn, err := sdk.Swap.FetchBatchAmountIn(ctx, BatchAmountInArgs{
		PoolRoute: []string{"0xpool"},
		AmountOut: "1000",
		TokenIn:   "0x1",
		TokenOut:  "0x2",
	})
	if err != nil {
		t.Fatalf("FetchBatchAmountIn returned error: %v", err)
	}
	if batchIn != "901" {
		t.Fatalf("batchIn = %q", batchIn)
	}

	poolOut, err := sdk.Pool.FetchPoolAmountOut(ctx, PoolAmountOutArgs{PoolID: "0xpool", TokenIn: "0x1", AmountIn: "1000"})
	if err != nil {
		t.Fatalf("FetchPoolAmountOut returned error: %v", err)
	}
	if poolOut.Amount != "902" || poolOut.FeeAmount != "3" {
		t.Fatalf("poolOut = %#v", poolOut)
	}

	poolIn, err := sdk.Pool.FetchPoolAmountIn(ctx, PoolAmountInArgs{PoolID: "0xpool", TokenOut: "0x2", AmountOut: "1000"})
	if err != nil {
		t.Fatalf("FetchPoolAmountIn returned error: %v", err)
	}
	if poolIn.Amount != "903" || poolIn.FeeAmount != "4" {
		t.Fatalf("poolIn = %#v", poolIn)
	}
}

func TestTypedViewHelpersReturnDecodeErrors(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientWithViewExecutor(t, &queuedViewExecutor{responses: [][]any{{"only-one"}}})
	if _, err := sdk.Position.FetchPositionTokenAmounts(context.Background(), "pos-1"); err == nil {
		t.Fatal("FetchPositionTokenAmounts accepted malformed result")
	}
}

func stringAnySlice(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}
