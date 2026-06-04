package hyperion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEstAmountByAggregateSwapRequiresMainnet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("aggregate route endpoint should not be called on testnet")
	}))
	defer server.Close()

	sdk := newTestClient(t, server)
	_, err := sdk.Swap.EstAmountByAggregateSwap(context.Background(), AggregateSwapRouteArgs{
		Amount:   "1000",
		From:     "0x1",
		Input:    "0x1",
		Slippage: "0.5",
		To:       "0x2",
	})
	if err == nil {
		t.Fatal("EstAmountByAggregateSwap returned nil error on testnet")
	}
}

func TestEstAmountByAggregateSwapFetchesRoute(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/base/aggregator/getAggRoute" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("amount") != "1000" ||
			r.URL.Query().Get("from") != "0x1" ||
			r.URL.Query().Get("input") != "0x1" ||
			r.URL.Query().Get("slippage") != "0.5" ||
			r.URL.Query().Get("to") != "0x2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fromToken":{"address":"0x1"},
			"toToken":{"address":"0x2"},
			"exactIn":true,
			"feeAmount":"1",
			"fromTokenAmount":"1000",
			"minToTokenAmount":"990",
			"maxFromTokenAmount":"1000",
			"toTokenAmount":"995",
			"quotes":{"route":[{"amountIn":"1000","amountOut":"995","percentage":100,"feeAmount":"1","routeTaken":[{"fromToken":{"tokenType":"0x1"},"toToken":{"tokenType":"0x2"},"dexName":"Hyperion","poolId":"pool-1","a2b":true,"sqrtPriceLimit":"0","poolType":"","amountIn":"1000","amountOut":"995"}]}],"refundRoute":[]}
		}`))
	}))
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
	if !route.ExactIn || route.FromToken.Address != "0x1" || route.Quotes.Route[0].RouteTaken[0].DexName != "Hyperion" {
		t.Fatalf("route = %#v", route)
	}
}
