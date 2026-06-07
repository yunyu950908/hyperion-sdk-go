package hyperion

import (
	"context"
	"testing"
)

func TestTypedPriceHubViewHelpersDecode(t *testing.T) {
	t.Parallel()

	asset := "0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b"
	executor := &queuedViewExecutor{responses: [][]any{
		{"111", "222"},
		{map[string]any{"price": "333", "precision": "1000000"}, map[string]any{"price": "444", "precision": float64(1000000)}},
		{true},
		{[]any{"0xa", map[string]any{"inner": "0xb"}}},
		{"chainlink"},
	}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	ctx := context.Background()

	preview, err := sdk.PriceHub.FetchPricePreview(ctx, PricePreviewArgs{
		Asset:  asset,
		Amount: "1000000000000000000",
	})
	if err != nil {
		t.Fatalf("FetchPricePreview returned error: %v", err)
	}
	if preview.First != "111" || preview.Second != "222" {
		t.Fatalf("preview = %#v", preview)
	}

	comparison, err := sdk.PriceHub.FetchPriceSourceComparison(ctx, asset)
	if err != nil {
		t.Fatalf("FetchPriceSourceComparison returned error: %v", err)
	}
	if comparison.First.Price != "333" || comparison.First.Precision != "1000000" {
		t.Fatalf("comparison.First = %#v", comparison.First)
	}
	if comparison.Second.Price != "444" || comparison.Second.Precision != "1000000" {
		t.Fatalf("comparison.Second = %#v", comparison.Second)
	}

	inHub, err := sdk.PriceHub.FetchIsTokenInPriceHub(ctx, asset)
	if err != nil {
		t.Fatalf("FetchIsTokenInPriceHub returned error: %v", err)
	}
	if !inHub {
		t.Fatal("inHub = false")
	}

	assets, err := sdk.PriceHub.FetchTokenInHubList(ctx)
	if err != nil {
		t.Fatalf("FetchTokenInHubList returned error: %v", err)
	}
	assertArguments(t, stringAnySlice(assets), []any{"0xa", "0xb"})

	source, err := sdk.PriceHub.FetchPriceHubFeedSource(ctx)
	if err != nil {
		t.Fatalf("FetchPriceHubFeedSource returned error: %v", err)
	}
	if source != "chainlink" {
		t.Fatalf("source = %q", source)
	}

	if executor.payloads[0].Function != MainnetContractAddress+"::price_hub::get_price_preview" {
		t.Fatalf("price preview function = %q", executor.payloads[0].Function)
	}
	assertArguments(t, executor.payloads[0].FunctionArguments, []any{asset, "1000000000000000000"})
	if executor.payloads[1].Function != MainnetContractAddress+"::price_hub::compare_two_source" {
		t.Fatalf("price comparison function = %q", executor.payloads[1].Function)
	}
	if executor.payloads[4].Function != MainnetContractAddress+"::price_hub::feed_source" {
		t.Fatalf("feed source function = %q", executor.payloads[4].Function)
	}
}

func TestTypedRateLimiterViewHelpersDecode(t *testing.T) {
	t.Parallel()

	user := "0x1230000000000000000000000000000000000000000000000000000000000000"
	assetA := "0xa"
	assetB := "0xb"
	poolA := "0x1110000000000000000000000000000000000000000000000000000000000000"
	poolB := "0x2220000000000000000000000000000000000000000000000000000000000000"
	executor := &queuedViewExecutor{responses: [][]any{
		{"10", "100", "60"},
		{[]any{
			map[string]any{"asset": map[string]any{"inner": assetA}, "remain": "9", "capacity": "90", "interval": "30"},
			map[string]any{"asset": assetB, "remain": float64(8), "capacity": "80", "interval": "20"},
		}},
		{"7", "70", "15"},
		{[]any{map[string]any{"asset": assetA, "remain": "6", "capacity": "60", "interval": "10"}}},
		{true, "5", "50", "5"},
		{[]any{
			map[string]any{"exist": true, "remain": "4", "capacity": "40", "interval": "4"},
			map[string]any{"exist": false, "remain": "0", "capacity": "0", "interval": "0"},
		}},
		{true, "3", "30", "3", []any{assetA, map[string]any{"inner": assetB}}},
	}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	ctx := context.Background()

	globalAsset, err := sdk.RateLimiter.FetchGlobalAssetRateLimiter(ctx, assetA)
	if err != nil {
		t.Fatalf("FetchGlobalAssetRateLimiter returned error: %v", err)
	}
	if globalAsset.Remain != "10" || globalAsset.Capacity != "100" || globalAsset.Interval != "60" {
		t.Fatalf("globalAsset = %#v", globalAsset)
	}

	globalBatch, err := sdk.RateLimiter.FetchGlobalAssetRateLimiterBatch(ctx, []string{assetA, assetB})
	if err != nil {
		t.Fatalf("FetchGlobalAssetRateLimiterBatch returned error: %v", err)
	}
	if len(globalBatch) != 2 || globalBatch[1].Asset != assetB || globalBatch[1].Remain != "8" {
		t.Fatalf("globalBatch = %#v", globalBatch)
	}

	userAsset, err := sdk.RateLimiter.FetchUserAssetRateLimiter(ctx, UserAssetRateLimiterArgs{
		User:  user,
		Asset: assetA,
	})
	if err != nil {
		t.Fatalf("FetchUserAssetRateLimiter returned error: %v", err)
	}
	if userAsset.Remain != "7" || userAsset.Capacity != "70" || userAsset.Interval != "15" {
		t.Fatalf("userAsset = %#v", userAsset)
	}

	userBatch, err := sdk.RateLimiter.FetchUserAssetRateLimiterBatch(ctx, UserAssetRateLimiterBatchArgs{
		User:   user,
		Assets: []string{assetA},
	})
	if err != nil {
		t.Fatalf("FetchUserAssetRateLimiterBatch returned error: %v", err)
	}
	if len(userBatch) != 1 || userBatch[0].Asset != assetA || userBatch[0].Remain != "6" {
		t.Fatalf("userBatch = %#v", userBatch)
	}

	pool, err := sdk.RateLimiter.FetchPoolUPriceLimiter(ctx, poolA)
	if err != nil {
		t.Fatalf("FetchPoolUPriceLimiter returned error: %v", err)
	}
	if !pool.Exists || pool.Remain != "5" || pool.Capacity != "50" || pool.Interval != "5" {
		t.Fatalf("pool = %#v", pool)
	}

	poolBatch, err := sdk.RateLimiter.FetchPoolUPriceLimiterBatch(ctx, []string{poolA, poolB})
	if err != nil {
		t.Fatalf("FetchPoolUPriceLimiterBatch returned error: %v", err)
	}
	if len(poolBatch) != 2 || poolBatch[1].Exists || poolBatch[1].Remain != "0" {
		t.Fatalf("poolBatch = %#v", poolBatch)
	}

	globalUPrice, err := sdk.RateLimiter.FetchGlobalUPriceLimiter(ctx)
	if err != nil {
		t.Fatalf("FetchGlobalUPriceLimiter returned error: %v", err)
	}
	if !globalUPrice.Exists || globalUPrice.Remain != "3" || globalUPrice.Capacity != "30" || globalUPrice.Interval != "3" {
		t.Fatalf("globalUPrice = %#v", globalUPrice)
	}
	assertArguments(t, stringAnySlice(globalUPrice.Assets), []any{assetA, assetB})

	if executor.payloads[0].Function != MainnetContractAddress+"::rate_limiter_check::global_asset_rate_limiter" {
		t.Fatalf("global asset function = %q", executor.payloads[0].Function)
	}
	assertArguments(t, executor.payloads[1].FunctionArguments, []any{[]string{assetA, assetB}})
	assertArguments(t, executor.payloads[2].FunctionArguments, []any{user, assetA})
	if executor.payloads[6].Function != MainnetContractAddress+"::rate_limiter_check::global_u_price_limiter" {
		t.Fatalf("global u price function = %q", executor.payloads[6].Function)
	}
}

func TestTypedCoinWrapperViewHelpersDecode(t *testing.T) {
	t.Parallel()

	asset := "0x3330000000000000000000000000000000000000000000000000000000000000"
	executor := &queuedViewExecutor{responses: [][]any{
		{true},
		{"0x4440000000000000000000000000000000000000000000000000000000000000"},
		{"0x1::aptos_coin::AptosCoin"},
		{"0x3330000000000000000000000000000000000000000000000000000000000000"},
		{true},
	}}
	sdk := newMainnetClientWithViewExecutor(t, executor)
	ctx := context.Background()

	isWrapper, err := sdk.CoinWrapper.FetchIsWrapper(ctx, asset)
	if err != nil {
		t.Fatalf("FetchIsWrapper returned error: %v", err)
	}
	if !isWrapper {
		t.Fatal("isWrapper = false")
	}

	original, err := sdk.CoinWrapper.FetchOriginalAsset(ctx, asset)
	if err != nil {
		t.Fatalf("FetchOriginalAsset returned error: %v", err)
	}
	if original != "0x4440000000000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("original = %q", original)
	}

	coinType, err := sdk.CoinWrapper.FetchCoinType(ctx, asset)
	if err != nil {
		t.Fatalf("FetchCoinType returned error: %v", err)
	}
	if coinType != "0x1::aptos_coin::AptosCoin" {
		t.Fatalf("coinType = %q", coinType)
	}

	formatted, err := sdk.CoinWrapper.FetchFormattedFungibleAsset(ctx, asset)
	if err != nil {
		t.Fatalf("FetchFormattedFungibleAsset returned error: %v", err)
	}
	if formatted != "0x3330000000000000000000000000000000000000000000000000000000000000" {
		t.Fatalf("formatted = %q", formatted)
	}

	supported, err := sdk.CoinWrapper.FetchCoinWrapperSupported(ctx)
	if err != nil {
		t.Fatalf("FetchCoinWrapperSupported returned error: %v", err)
	}
	if !supported {
		t.Fatal("supported = false")
	}

	if executor.payloads[0].Function != MainnetContractAddress+"::coin_wrapper::is_wrapper" {
		t.Fatalf("is wrapper function = %q", executor.payloads[0].Function)
	}
	assertArguments(t, executor.payloads[0].FunctionArguments, []any{asset})
	if executor.payloads[4].Function != MainnetContractAddress+"::coin_wrapper::is_supported" {
		t.Fatalf("is supported function = %q", executor.payloads[4].Function)
	}
}

func TestLaterScopeTypedViewHelpersReturnDecodeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*Client) error
	}{
		{
			name: "price comparison malformed agg price",
			call: func(sdk *Client) error {
				_, err := sdk.PriceHub.FetchPriceSourceComparison(context.Background(), "0xa")
				return err
			},
		},
		{
			name: "rate limiter batch missing interval",
			call: func(sdk *Client) error {
				_, err := sdk.RateLimiter.FetchGlobalAssetRateLimiterBatch(context.Background(), []string{"0xa"})
				return err
			},
		},
		{
			name: "coin wrapper bool malformed",
			call: func(sdk *Client) error {
				_, err := sdk.CoinWrapper.FetchIsWrapper(context.Background(), "0xa")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sdk := newMainnetClientWithViewExecutor(t, &queuedViewExecutor{responses: [][]any{
				{map[string]any{"price": "1"}},
			}})
			if err := tt.call(sdk); err == nil {
				t.Fatal("accepted malformed result")
			}
		})
	}
}

func TestLaterScopeTypedViewHelpersValidatePayloadArgs(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientWithViewExecutor(t, &queuedViewExecutor{})

	if _, err := sdk.PriceHub.PricePreviewPayload(PricePreviewArgs{
		Asset:  "0x1::aptos_coin::AptosCoin",
		Amount: "1000",
	}); err == nil {
		t.Fatal("PricePreviewPayload accepted coin type asset")
	}

	if _, err := sdk.PriceHub.PricePreviewPayload(PricePreviewArgs{
		Asset:  "0xa",
		Amount: "1.5",
	}); err == nil {
		t.Fatal("PricePreviewPayload accepted decimal amount")
	}

	if _, err := sdk.RateLimiter.GlobalAssetRateLimiterBatchPayload(nil); err == nil {
		t.Fatal("GlobalAssetRateLimiterBatchPayload accepted empty assets")
	}

	if _, err := sdk.RateLimiter.PoolUPriceLimiterBatchPayload([]string{"pool-without-prefix"}); err == nil {
		t.Fatal("PoolUPriceLimiterBatchPayload accepted non-address pool")
	}

	if _, err := sdk.CoinWrapper.CoinTypePayload("0x1::aptos_coin::AptosCoin"); err == nil {
		t.Fatal("CoinTypePayload accepted coin type asset")
	}
}
