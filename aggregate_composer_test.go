package hyperion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateAggregateSwapTransactionScriptRecordsExactInHyperionRoute(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	recorder := NewAggregateSwapRecorder()
	route := aggregateComposerRoute(true)

	err := sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:    route,
		Composer: recorder,
	})
	if err != nil {
		t.Fatalf("GenerateAggregateSwapTransactionScript returned error: %v", err)
	}

	assertRecordedFunctions(t, recorder.Calls, []string{
		AggregateToolContractAddress + "::tool::get_signer_address",
		"0x1::object::address_to_object",
		"0x1::primary_fungible_store::withdraw",
		"0x1::fungible_asset::amount",
		"0x1::object::address_to_object",
		"0x1::fungible_asset::zero",
		"0x1::fungible_asset::extract",
		"0x1::object::address_to_object",
		"0x1::fungible_asset::zero",
		"0x1::fungible_asset::amount",
		MainnetContractAddress + "::partnership::swap",
		"0x1::primary_fungible_store::deposit",
		"0x1::fungible_asset::merge",
		"0x1::fungible_asset::merge",
		AggregateToolContractAddress + "::tool::fa_amount_check",
		"0x1::primary_fungible_store::deposit",
		"0x1::primary_fungible_store::deposit",
	})

	swap := recorder.Calls[10]
	assertCallArguments(t, swap.FunctionArguments, []AggregateSwapCallArgument{
		literalArg("pool-hyperion"),
		literalArg(true),
		literalArg(true),
		resultArg(9, 0),
		resultArg(6, 0),
		literalArg("0"),
		literalArg(AggregatorPartnerName),
	})

	depositReturnedFA := recorder.Calls[11]
	assertCallArguments(t, depositReturnedFA.FunctionArguments, []AggregateSwapCallArgument{
		resultArg(0, 0).Copy(),
		resultArg(10, 1),
	})
}

func TestGenerateAggregateSwapTransactionScriptRecordsExactOutRefundRoute(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	recorder := NewAggregateSwapRecorder()
	route := aggregateComposerRoute(false)
	route.Quotes.RefundRoute = []AggregateSwapRoute{
		{
			AmountIn:   "30",
			AmountOut:  "30",
			Percentage: 30,
			RouteTaken: []AggregateSwapRouteTaken{{
				DexName:   "Cellana",
				PoolID:    "cellana-pool",
				PoolType:  "stable",
				FromToken: TokenTypeInfo{TokenType: "0x2"},
				ToToken:   TokenTypeInfo{TokenType: "0x1"},
			}},
		},
		{
			AmountIn:   "70",
			AmountOut:  "70",
			Percentage: 70,
			RouteTaken: []AggregateSwapRouteTaken{{
				DexName:   "Cellana",
				PoolID:    "cellana-pool-2",
				PoolType:  "unstable",
				FromToken: TokenTypeInfo{TokenType: "0x2"},
				ToToken:   TokenTypeInfo{TokenType: "0x1"},
			}},
		},
	}

	err := sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:    route,
		Composer: recorder,
	})
	if err != nil {
		t.Fatalf("GenerateAggregateSwapTransactionScript returned error: %v", err)
	}

	depositExact := findRecordedCall(t, recorder.Calls, AggregateToolContractAddress+"::tool::deposit_fa_exact")
	assertCallArguments(t, depositExact.FunctionArguments[1:], []AggregateSwapCallArgument{
		resultArg(5, 0).BorrowMut(),
		literalArg("995"),
	})

	splitCalls := findRecordedCalls(recorder.Calls, AggregateToolContractAddress+"::tool::split_fa_proportionlly")
	if len(splitCalls) != 2 {
		t.Fatalf("split_fa_proportionlly calls = %d", len(splitCalls))
	}
	assertCallArguments(t, splitCalls[0].FunctionArguments[1:], []AggregateSwapCallArgument{
		resultArg(18, 0).Copy(),
		literalArg(3000),
	})
	assertCallArguments(t, splitCalls[1].FunctionArguments[1:], []AggregateSwapCallArgument{
		resultArg(18, 0).Copy(),
		literalArg(10000),
	})
}

func TestGenerateAggregateSwapTransactionScriptRecordsDEXAdapters(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	recorder := NewAggregateSwapRecorder()
	route := aggregateComposerRoute(true)
	route.Quotes.Route[0].RouteTaken = []AggregateSwapRouteTaken{
		{
			DexName:   "Cellana",
			PoolID:    "cellana-pool",
			PoolType:  "stable",
			FromToken: TokenTypeInfo{TokenType: "0x1"},
			ToToken:   TokenTypeInfo{TokenType: "0x2"},
		},
		{
			DexName:   "ThalaSwapV2",
			PoolID:    "thala-pool",
			PoolType:  "weighted",
			FromToken: TokenTypeInfo{TokenType: "0x2"},
			ToToken:   TokenTypeInfo{TokenType: "0x3"},
		},
		{
			DexName:       "EmojiCoin",
			PoolID:        "emoji-pool",
			FirstType:     "0x1::emoji::A",
			SecondType:    "0x1::emoji::B",
			IsSell:        true,
			Integrator:    "integrator-1",
			IntegratorFee: 7,
			FromToken:     TokenTypeInfo{TokenType: "0x3"},
			ToToken:       TokenTypeInfo{TokenType: "0x4"},
		},
	}

	err := sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:         route,
		Composer:      recorder,
		PartnershipID: "partner-override",
	})
	if err != nil {
		t.Fatalf("GenerateAggregateSwapTransactionScript returned error: %v", err)
	}

	cellana := findRecordedCall(t, recorder.Calls, "0x4bf51972879e3b95c4781a5cdcb9e1ee24ef483e7d22f2d903626f126df62bd1::router::swap")
	assertCallArguments(t, cellana.FunctionArguments[1:], []AggregateSwapCallArgument{
		literalArg(0),
		literalArg("0x2"),
		literalArg(true),
	})

	findRecordedCall(t, recorder.Calls, "0x7730cd28ee1cdc9e999336cbc430f99e7c44397c0aa77516f6f23a78559bb5::pool::swap_exact_in_weighted")

	emoji := findRecordedCall(t, recorder.Calls, AggregateToolContractAddress+"::tool::swap_in_emoji")
	if got, want := emoji.TypeArguments, []string{"0x1::emoji::A", "0x1::emoji::B"}; !stringSlicesEqual(got, want) {
		t.Fatalf("emoji type arguments = %#v, want %#v", got, want)
	}
	assertCallArguments(t, emoji.FunctionArguments[:5], []AggregateSwapCallArgument{
		signerArg(0),
		literalArg("emoji-pool"),
		literalArg(true),
		literalArg("integrator-1"),
		literalArg(7),
	})
}

func TestGenerateAggregateSwapTransactionScriptReportsComposerErrors(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	err := sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route: aggregateComposerRoute(true),
	})
	if err == nil || !strings.Contains(err.Error(), "composer is required") {
		t.Fatalf("nil composer error = %v", err)
	}

	recorder := NewAggregateSwapRecorder()
	route := aggregateComposerRoute(true)
	route.Quotes.Route[0].RouteTaken[0].DexName = "UnknownDEX"
	err = sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:    route,
		Composer: recorder,
	})
	if err == nil || !strings.Contains(err.Error(), "DEX not supported") {
		t.Fatalf("unsupported DEX error = %v", err)
	}

	recorder = NewAggregateSwapRecorder()
	route = aggregateComposerRoute(true)
	route.Quotes.Route[0].RouteTaken[0] = AggregateSwapRouteTaken{
		DexName:   "Cellana",
		PoolID:    "cellana-pool",
		PoolType:  "volatile",
		FromToken: TokenTypeInfo{TokenType: "0x1"},
		ToToken:   TokenTypeInfo{TokenType: "0x2"},
	}
	err = sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:    route,
		Composer: recorder,
	})
	if err == nil || !strings.Contains(err.Error(), "pool type mismatch") {
		t.Fatalf("pool type mismatch error = %v", err)
	}

	recorder = NewAggregateSwapRecorder()
	route = aggregateComposerRoute(false)
	route.Quotes.RefundRoute = []AggregateSwapRoute{
		{
			AmountIn:   "50",
			AmountOut:  "50",
			Percentage: 50,
			RouteTaken: []AggregateSwapRouteTaken{{
				DexName:   "Cellana",
				PoolID:    "cellana-pool",
				PoolType:  "stable",
				FromToken: TokenTypeInfo{TokenType: "0x2"},
				ToToken:   TokenTypeInfo{TokenType: "0x1"},
			}},
		},
	}
	err = sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:    route,
		Composer: recorder,
	})
	if err == nil || !strings.Contains(err.Error(), "not 100% refund") {
		t.Fatalf("refund percentage error = %v", err)
	}
}

func TestGenerateAggregateSwapTransactionScriptRequiresMainnet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()
	sdk := newTestClient(t, server)

	err := sdk.Swap.GenerateAggregateSwapTransactionScript(GenerateAggregateSwapTransactionScriptArgs{
		Route:    aggregateComposerRoute(true),
		Composer: NewAggregateSwapRecorder(),
	})
	if err == nil {
		t.Fatal("GenerateAggregateSwapTransactionScript returned nil error on testnet")
	}
}

func aggregateComposerRoute(exactIn bool) AggregateSwapInfoResult {
	return AggregateSwapInfoResult{
		FromToken:          TokenAddressInfo{Address: "0x1"},
		ToToken:            TokenAddressInfo{Address: "0x2"},
		ExactIn:            exactIn,
		FeeAmount:          "1",
		FromTokenAmount:    "1000",
		MinToTokenAmount:   "990",
		MaxFromTokenAmount: "1010",
		ToTokenAmount:      "995",
		Quotes: Quotes{
			Route: []AggregateSwapRoute{{
				AmountIn:   "1000",
				AmountOut:  "995",
				Percentage: 100,
				FeeAmount:  "1",
				RouteTaken: []AggregateSwapRouteTaken{{
					FromToken:      TokenTypeInfo{TokenType: "0x1"},
					ToToken:        TokenTypeInfo{TokenType: "0x2"},
					DexName:        "Hyperion",
					PoolID:         "pool-hyperion",
					A2B:            true,
					SqrtPriceLimit: "0",
					AmountIn:       "1000",
					AmountOut:      "995",
				}},
			}},
			RefundRoute: []AggregateSwapRoute{},
		},
	}
}

func assertRecordedFunctions(t *testing.T, calls []AggregateSwapComposerCall, expected []string) {
	t.Helper()
	if len(calls) != len(expected) {
		t.Fatalf("recorded calls = %d, want %d", len(calls), len(expected))
	}
	for i, call := range calls {
		if call.Function != expected[i] {
			t.Fatalf("call %d function = %q, want %q", i, call.Function, expected[i])
		}
	}
}

func assertCallArguments(t *testing.T, got, expected []AggregateSwapCallArgument) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("arguments = %#v, want %#v", got, expected)
	}
	for i := range expected {
		if !aggregateCallArgumentEqual(got[i], expected[i]) {
			t.Fatalf("argument %d = %#v, want %#v", i, got[i], expected[i])
		}
	}
}

func findRecordedCall(t *testing.T, calls []AggregateSwapComposerCall, function string) AggregateSwapComposerCall {
	t.Helper()
	for _, call := range calls {
		if call.Function == function {
			return call
		}
	}
	t.Fatalf("recorded call %q not found", function)
	return AggregateSwapComposerCall{}
}

func findRecordedCalls(calls []AggregateSwapComposerCall, function string) []AggregateSwapComposerCall {
	var out []AggregateSwapComposerCall
	for _, call := range calls {
		if call.Function == function {
			out = append(out, call)
		}
	}
	return out
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
