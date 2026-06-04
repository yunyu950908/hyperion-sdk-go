package hyperion

import (
	"context"
	"testing"
	"time"
)

func TestIntegrationHyperionFetchAllPools(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	sdk := newIntegrationClient(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pools, err := sdk.Pool.FetchAllPools(ctx)
	if err != nil {
		t.Fatalf("FetchAllPools returned error: %v", err)
	}
	if pools == nil {
		t.Fatal("FetchAllPools returned a nil slice")
	}
}

func TestIntegrationHyperionSwapQuote(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	if cfg.SwapFrom == "" || cfg.SwapTo == "" {
		t.Skip("set HYPERION_SWAP_FROM and HYPERION_SWAP_TO to run live swap quote integration tests")
	}

	sdk := newIntegrationClient(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	quote, err := sdk.Swap.EstFromAmount(ctx, EstimateAmountArgs{
		Amount:   cfg.SwapAmount,
		From:     cfg.SwapFrom,
		To:       cfg.SwapTo,
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("EstFromAmount returned error: %v", err)
	}
	if len(quote) == 0 {
		t.Fatal("EstFromAmount returned an empty quote object")
	}
}

func TestIntegrationHyperionAggregateRoute(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	if cfg.Network != NetworkMainnet {
		t.Skip("aggregate swap route integration test only runs on mainnet")
	}
	if cfg.AggregateFrom == "" || cfg.AggregateInput == "" || cfg.AggregateTo == "" {
		t.Skip("set HYPERION_AGG_FROM, HYPERION_AGG_INPUT, and HYPERION_AGG_TO to run aggregate route integration tests")
	}

	sdk := newIntegrationClient(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	route, err := sdk.Swap.EstAmountByAggregateSwap(ctx, AggregateSwapRouteArgs{
		Amount:   cfg.AggregateAmount,
		From:     cfg.AggregateFrom,
		Input:    cfg.AggregateInput,
		Slippage: "0.5",
		To:       cfg.AggregateTo,
	})
	if err != nil {
		t.Fatalf("EstAmountByAggregateSwap returned error: %v", err)
	}
	if route.FromToken.Address == "" || route.ToToken.Address == "" {
		t.Fatalf("aggregate route token addresses are empty: %#v", route)
	}
	if route.Quotes.Route == nil {
		t.Fatalf("aggregate route quotes are nil: %#v", route)
	}
}
