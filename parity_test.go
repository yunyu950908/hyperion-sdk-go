package hyperion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParityGoldenPayloads(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	createPool, err := sdk.Pool.CreatePoolTransactionPayload(CreatePoolTransactionPayloadArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		CurrencyAAmount:  "1000",
		CurrencyBAmount:  "2000",
		FeeTierIndex:     "2",
		CurrentPriceTick: "0",
		TickLower:        "-60",
		TickUpper:        "60",
		Slippage:         "0.5",
	})
	if err != nil {
		t.Fatalf("CreatePoolTransactionPayload returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/payloads/create_pool_fa_to_fa.json", normalizePayloadSnapshot(createPool))

	addLiquidity, err := sdk.Position.AddLiquidityTransactionPayload(AddLiquidityTransactionPayloadArgs{
		PositionID:      "pos-1",
		CurrencyA:       "0x1",
		CurrencyB:       "0x2",
		CurrencyAAmount: "1000",
		CurrencyBAmount: "2000",
		Slippage:        "0.5",
		FeeTierIndex:    "2",
	})
	if err != nil {
		t.Fatalf("AddLiquidityTransactionPayload returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/payloads/add_liquidity_fa_to_fa.json", normalizePayloadSnapshot(addLiquidity))

	swap, err := sdk.Swap.SwapTransactionPayload(SwapTransactionPayloadArgs{
		CurrencyA:       "0x1",
		CurrencyB:       "0x2",
		CurrencyAAmount: "1000",
		CurrencyBAmount: "1000",
		Slippage:        "0.5",
		PoolRoute:       []string{"pool-1"},
		Recipient:       "0xabc",
	})
	if err != nil {
		t.Fatalf("SwapTransactionPayload returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/payloads/swap_fa_to_fa.json", normalizePayloadSnapshot(swap))

	swapCoin, err := sdk.Swap.SwapTransactionPayload(SwapTransactionPayloadArgs{
		CurrencyA:       aptosCoinType,
		CurrencyB:       exampleCoinType,
		CurrencyAAmount: "1000",
		CurrencyBAmount: "1000",
		Slippage:        "0.5",
		PoolRoute:       []string{"pool-1"},
		Recipient:       "0xabc",
	})
	if err != nil {
		t.Fatalf("SwapTransactionPayload coin route returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/payloads/swap_coin_to_coin.json", normalizePayloadSnapshot(swapCoin))
}

func TestParityGoldenAdditionalPayloads(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	removeLiquidity, err := sdk.Position.RemoveLiquidityTransactionPayload(RemoveLiquidityTransactionPayloadArgs{
		PositionID:      "pos-1",
		CurrencyA:       "0x1",
		CurrencyB:       "0x2",
		CurrencyAAmount: "1000",
		CurrencyBAmount: "2000",
		DeltaLiquidity:  "12345",
		Slippage:        "0.5",
		Recipient:       "0x0000000000000000000000000000000000000000000000000000000000000abc",
	})
	if err != nil {
		t.Fatalf("RemoveLiquidityTransactionPayload returned error: %v", err)
	}

	snapshots := map[string]any{
		"removeLiquidity": normalizePayloadSnapshot(removeLiquidity),
		"claimFee":        normalizePayloadSnapshot(sdk.Position.ClaimFeeTransactionPayload("pos-1", "0xabc")),
		"claimReward":     normalizePayloadSnapshot(sdk.Position.ClaimRewardTransactionPayload("pos-1", "0xabc")),
		"claimAllRewards": normalizePayloadSnapshot(sdk.Position.ClaimAllRewardsTransactionPayload("pos-1", "0xabc")),
		"pendingRewards":  normalizePayloadSnapshot(sdk.Reward.FetchRewardsPayload("pos-1")),
	}
	assertGoldenJSON(t, "testdata/parity/payloads/position_reward_payloads.json", snapshots)
}

func TestParityGoldenRESTFixtures(t *testing.T) {
	t.Parallel()

	server := newParityFixtureServer(t)
	defer server.Close()

	sdk, err := New(Options{
		Network:                    NetworkMainnet,
		ContractAddress:            MainnetContractAddress,
		HyperionFullNodeIndexerURL: server.URL,
		HyperionAPIHost:            server.URL,
		HTTPClient:                 server.Client(),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	pools, err := sdk.Pool.FetchAllPools(context.Background())
	if err != nil {
		t.Fatalf("FetchAllPools returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/rest/pools_stats_items.json", pools)

	quote, err := sdk.Swap.EstFromAmount(context.Background(), EstimateAmountArgs{
		Amount:   "1000",
		From:     "0x1",
		To:       "0x2",
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("EstFromAmount returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/rest/swap_quote_out.json", quote)

	route, err := sdk.Swap.EstAmountByAggregateSwap(context.Background(), AggregateSwapRouteArgs{
		Amount:   "1000",
		From:     "0x1",
		Input:    "0x1",
		Slippage: "0.5",
		To:       "0x2",
	})
	if err != nil {
		t.Fatalf("EstAmountByAggregateSwap returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/rest/aggregate_route.json", route)
}

func newParityFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/data/pools/stats":
			if err := writeFixture(w, "testdata/parity/rest/pools_stats_response.json"); err != nil {
				t.Errorf("write pool fixture: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "/base/rate/getSwapInfo":
			if r.URL.Query().Get("flag") != "out" || r.URL.Query().Get("safeMode") != "true" {
				t.Errorf("unexpected swap quote query: %s", r.URL.RawQuery)
				http.Error(w, "unexpected query", http.StatusBadRequest)
				return
			}
			if err := writeFixture(w, "testdata/parity/rest/swap_quote_out.json"); err != nil {
				t.Errorf("write swap quote fixture: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "/base/aggregator/getAggRoute":
			if err := writeFixture(w, "testdata/parity/rest/aggregate_route.json"); err != nil {
				t.Errorf("write aggregate route fixture: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			t.Errorf("unexpected fixture path: %s", r.URL.Path)
			http.Error(w, "unexpected path", http.StatusNotFound)
		}
	}))
}
