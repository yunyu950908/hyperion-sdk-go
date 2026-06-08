package hyperion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestBuildAggregateSwapSubmitPlanReturnsComposerPlan(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	route := aggregateComposerRouteWithAllDEXAdapters()

	plan, err := sdk.Swap.BuildAggregateSwapSubmitPlan(BuildAggregateSwapSubmitPlanArgs{
		Route:         route,
		PartnershipID: "partner-override",
	})
	if err != nil {
		t.Fatalf("BuildAggregateSwapSubmitPlan returned error: %v", err)
	}

	if plan.PartnershipID != "partner-override" || plan.RouteSplits != 1 || plan.RefundRouteSplits != 0 {
		t.Fatalf("submit plan metadata = %#v", plan)
	}
	if len(plan.Calls) != 29 {
		t.Fatalf("submit plan calls = %d", len(plan.Calls))
	}

	findRecordedCall(t, plan.Calls, MainnetContractAddress+"::partnership::swap")
	findRecordedCall(t, plan.Calls, cellanaContract+"::router::swap")
	findRecordedCall(t, plan.Calls, thalaSwapV2Contract+"::pool::swap_exact_in_weighted")
	findRecordedCall(t, plan.Calls, AggregateToolContractAddress+"::tool::swap_in_emoji")

	plan.Route.Quotes.Route[0].RouteTaken[0].DexName = "Mutated"
	if route.Quotes.Route[0].RouteTaken[0].DexName != "Hyperion" {
		t.Fatal("submit plan mutation changed caller route")
	}
}

func TestBuildAggregateSwapSubmitTransactionPassesComposerPlanToAdapter(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	adapter := &recordingSubmitAdapter{}
	route := aggregateComposerRouteWithAllDEXAdapters()

	result, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(context.Background(), BuildAggregateSwapSubmitTransactionArgs{
		Route:         route,
		PartnershipID: "partner-override",
		Adapter:       adapter,
	})
	if err != nil {
		t.Fatalf("BuildAggregateSwapSubmitTransaction returned error: %v", err)
	}

	if result.PayloadType != "test/aggregate-submit" || string(result.PayloadBytes) != "calls:29" {
		t.Fatalf("submit result = %#v", result)
	}
	if adapter.plan.PartnershipID != "partner-override" || adapter.plan.RouteSplits != 1 || adapter.plan.RefundRouteSplits != 0 {
		t.Fatalf("submit plan metadata = %#v", adapter.plan)
	}
	if len(adapter.plan.Calls) != 29 {
		t.Fatalf("submit plan calls = %d", len(adapter.plan.Calls))
	}

	findRecordedCall(t, adapter.plan.Calls, MainnetContractAddress+"::partnership::swap")
	findRecordedCall(t, adapter.plan.Calls, cellanaContract+"::router::swap")
	findRecordedCall(t, adapter.plan.Calls, thalaSwapV2Contract+"::pool::swap_exact_in_weighted")
	findRecordedCall(t, adapter.plan.Calls, AggregateToolContractAddress+"::tool::swap_in_emoji")

	adapter.plan.Route.Quotes.Route[0].RouteTaken[0].DexName = "Mutated"
	if route.Quotes.Route[0].RouteTaken[0].DexName != "Hyperion" {
		t.Fatal("adapter plan mutation changed caller route")
	}
}

func TestBuildAggregateSwapSubmitTransactionReportsUnsupportedAdapterGap(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	_, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(context.Background(), BuildAggregateSwapSubmitTransactionArgs{
		Route:   aggregateComposerRoute(true),
		Adapter: NewUnsupportedAggregateSwapSubmitAdapter(),
	})
	if err == nil {
		t.Fatal("BuildAggregateSwapSubmitTransaction returned nil error")
	}
	if !errors.Is(err, ErrAggregateSubmitAdapterUnsupported) {
		t.Fatalf("unsupported adapter error = %v", err)
	}
	if !strings.Contains(err.Error(), "Dynamic Script Composer") {
		t.Fatalf("unsupported adapter error = %v", err)
	}
}

func TestBuildAggregateSwapSubmitTransactionValidatesInputs(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	_, err := sdk.Swap.BuildAggregateSwapSubmitTransaction(context.Background(), BuildAggregateSwapSubmitTransactionArgs{
		Route: aggregateComposerRoute(true),
	})
	if err == nil || !strings.Contains(err.Error(), "submit adapter is required") {
		t.Fatalf("nil adapter error = %v", err)
	}

	testnet := newTestClient(t, newEmptyTestServer(t))
	_, err = testnet.Swap.BuildAggregateSwapSubmitTransaction(context.Background(), BuildAggregateSwapSubmitTransactionArgs{
		Route:   aggregateComposerRoute(true),
		Adapter: &recordingSubmitAdapter{},
	})
	if err == nil || !strings.Contains(err.Error(), "only supported on MAINNET") {
		t.Fatalf("testnet error = %v", err)
	}
}

type recordingSubmitAdapter struct {
	plan AggregateSwapSubmitPlan
}

func (a *recordingSubmitAdapter) BuildAggregateSwapSubmitTransaction(_ context.Context, plan AggregateSwapSubmitPlan) (*AggregateSwapSubmitTransaction, error) {
	a.plan = plan
	return &AggregateSwapSubmitTransaction{
		PayloadType:  "test/aggregate-submit",
		PayloadBytes: []byte("calls:" + strconv.Itoa(len(plan.Calls))),
	}, nil
}

func aggregateComposerRouteWithAllDEXAdapters() AggregateSwapInfoResult {
	route := aggregateComposerRoute(true)
	route.ToToken = TokenAddressInfo{Address: "0x5"}
	route.Quotes.Route[0].RouteTaken = []AggregateSwapRouteTaken{
		{
			FromToken:      TokenTypeInfo{TokenType: "0x1"},
			ToToken:        TokenTypeInfo{TokenType: "0x2"},
			DexName:        "Hyperion",
			PoolID:         "pool-hyperion",
			A2B:            true,
			SqrtPriceLimit: "0",
			AmountIn:       "1000",
			AmountOut:      "995",
		},
		{
			DexName:   "Cellana",
			PoolID:    "cellana-pool",
			PoolType:  "stable",
			FromToken: TokenTypeInfo{TokenType: "0x2"},
			ToToken:   TokenTypeInfo{TokenType: "0x3"},
		},
		{
			DexName:   "ThalaSwapV2",
			PoolID:    "thala-pool",
			PoolType:  "weighted",
			FromToken: TokenTypeInfo{TokenType: "0x3"},
			ToToken:   TokenTypeInfo{TokenType: "0x4"},
		},
		{
			DexName:       "EmojiCoin",
			PoolID:        "emoji-pool",
			FirstType:     "0x1::emoji::A",
			SecondType:    "0x1::emoji::B",
			IsSell:        true,
			Integrator:    "integrator-1",
			IntegratorFee: 7,
			FromToken:     TokenTypeInfo{TokenType: "0x4"},
			ToToken:       TokenTypeInfo{TokenType: "0x5"},
		},
	}
	return route
}

func newEmptyTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.NewServeMux())
	t.Cleanup(server.Close)
	return server
}
