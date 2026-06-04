package hyperion

import "testing"

func TestRewardPayloadBuilders(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	pending := sdk.Reward.FetchRewardsPayload("pos-1")
	if pending.Function != MainnetContractAddress+"::pool_v3::get_pending_rewards" {
		t.Fatalf("pending function = %q", pending.Function)
	}
	assertArguments(t, pending.FunctionArguments, []any{"pos-1"})

	claim := sdk.Reward.ClaimRewardPayload("pos-1", "0xabc")
	if claim.Function != MainnetContractAddress+"::router_v3::claim_rewards" {
		t.Fatalf("claim function = %q", claim.Function)
	}
	assertArguments(t, claim.FunctionArguments, []any{"pos-1", "0xabc"})
}

func TestPositionClaimPayloadBuilders(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	fee := sdk.Position.ClaimFeeTransactionPayload("pos-1", "0xabc")
	if fee.Function != MainnetContractAddress+"::router_v3::claim_fees" {
		t.Fatalf("fee function = %q", fee.Function)
	}
	assertArguments(t, fee.FunctionArguments, []any{[]string{"pos-1"}, "0xabc"})

	reward := sdk.Position.ClaimRewardTransactionPayload("pos-1", "0xabc")
	if reward.Function != MainnetContractAddress+"::router_v3::claim_rewards" {
		t.Fatalf("reward function = %q", reward.Function)
	}
	assertArguments(t, reward.FunctionArguments, []any{"pos-1", "0xabc"})

	all := sdk.Position.ClaimAllRewardsTransactionPayload("pos-1", "0xabc")
	if all.Function != MainnetContractAddress+"::router_v3::claim_fees_and_rewards" {
		t.Fatalf("all function = %q", all.Function)
	}
	assertArguments(t, all.FunctionArguments, []any{[]string{"pos-1"}, "0xabc"})
}

func TestSwapPayloadBuilders(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	payload, err := sdk.Swap.SwapTransactionPayload(SwapTransactionPayloadArgs{
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
	if payload.Function != MainnetContractAddress+"::router_v3::swap_batch" {
		t.Fatalf("swap function = %q", payload.Function)
	}
	assertArguments(t, payload.FunctionArguments, []any{[]string{"pool-1"}, "0x1", "0x2", "1000", "995", "0xabc"})

	partnership, err := sdk.Swap.SwapWithPartnershipTransactionPayload(SwapWithPartnershipTransactionPayloadArgs{
		SwapTransactionPayloadArgs: SwapTransactionPayloadArgs{
			CurrencyA:       "0x1",
			CurrencyB:       "0x2",
			CurrencyAAmount: "1000",
			CurrencyBAmount: "1000",
			Slippage:        "0.5",
			PoolRoute:       []string{"pool-1"},
			Recipient:       "0xabc",
		},
		Partnership: "partner-1",
	})
	if err != nil {
		t.Fatalf("SwapWithPartnershipTransactionPayload returned error: %v", err)
	}
	if partnership.Function != MainnetContractAddress+"::partnership::swap_batch_directly_deposit" {
		t.Fatalf("partnership function = %q", partnership.Function)
	}
	assertArguments(t, partnership.FunctionArguments, []any{[]string{"pool-1"}, "0x1", "0x2", "1000", "995", "partner-1"})
}

func TestSwapPayloadRejectsCoinTypesUntilFAConversionIsImplemented(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	_, err := sdk.Swap.SwapTransactionPayload(SwapTransactionPayloadArgs{
		CurrencyA:       "0x1::aptos_coin::AptosCoin",
		CurrencyB:       "0x2",
		CurrencyAAmount: "1000",
		CurrencyBAmount: "1000",
		Slippage:        "0.5",
		PoolRoute:       []string{"pool-1"},
		Recipient:       "0xabc",
	})
	if err == nil {
		t.Fatal("SwapTransactionPayload accepted coin type without FA conversion")
	}
}

func TestCreatePoolTransactionPayload(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	payload, err := sdk.Pool.CreatePoolTransactionPayload(CreatePoolTransactionPayloadArgs{
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
	if payload.Function != MainnetContractAddress+"::router_adapter::create_liquidity_entry" {
		t.Fatalf("function = %q", payload.Function)
	}
	assertArgumentPrefix(t, payload.FunctionArguments, []any{
		"0x1",
		"0x2",
		"2",
		PoolStableType,
		TickComplement(-60),
		TickComplement(60),
		TickComplement(0),
		"1000",
		"2000",
		"995",
		"1990",
	})
	assertDeadlineArgument(t, payload.FunctionArguments[len(payload.FunctionArguments)-1])
}

func TestAddAndRemoveLiquidityTransactionPayload(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	add, err := sdk.Position.AddLiquidityTransactionPayload(AddLiquidityTransactionPayloadArgs{
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
	if add.Function != MainnetContractAddress+"::router_adapter::add_liquidity_entry" {
		t.Fatalf("add function = %q", add.Function)
	}
	assertArgumentPrefix(t, add.FunctionArguments, []any{
		"pos-1",
		"0x1",
		"0x2",
		"2",
		PoolStableType,
		"1000",
		"2000",
		"995",
		"1990",
	})
	assertDeadlineArgument(t, add.FunctionArguments[len(add.FunctionArguments)-1])

	remove, err := sdk.Position.RemoveLiquidityTransactionPayload(RemoveLiquidityTransactionPayloadArgs{
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
	if remove.Function != MainnetContractAddress+"::router_adapter::remove_liquidity_entry_v2" {
		t.Fatalf("remove function = %q", remove.Function)
	}
	assertArgumentPrefix(t, remove.FunctionArguments, []any{
		"pos-1",
		"12345",
		"995",
		"1990",
		"0x0000000000000000000000000000000000000000000000000000000000000abc",
	})
	assertDeadlineArgument(t, remove.FunctionArguments[len(remove.FunctionArguments)-1])
}

func newMainnetClientForPayloads(t *testing.T) *Client {
	t.Helper()
	sdk, err := Init(InitOptions{Network: NetworkMainnet})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	return sdk
}

func assertArguments(t *testing.T, got []any, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("argument length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		switch wantValue := want[i].(type) {
		case []string:
			gotValue, ok := got[i].([]string)
			if !ok {
				t.Fatalf("arg %d type = %T, want []string", i, got[i])
			}
			if len(gotValue) != len(wantValue) {
				t.Fatalf("arg %d length = %d, want %d", i, len(gotValue), len(wantValue))
			}
			for j := range wantValue {
				if gotValue[j] != wantValue[j] {
					t.Fatalf("arg %d[%d] = %#v, want %#v", i, j, gotValue[j], wantValue[j])
				}
			}
		default:
			if got[i] != want[i] {
				t.Fatalf("arg %d = %#v, want %#v", i, got[i], want[i])
			}
		}
	}
}

func assertArgumentPrefix(t *testing.T, got []any, want []any) {
	t.Helper()
	if len(got) < len(want) {
		t.Fatalf("argument length = %d, want at least %d: %#v", len(got), len(want), got)
	}
	assertArguments(t, got[:len(want)], want)
}

func assertDeadlineArgument(t *testing.T, got any) {
	t.Helper()
	deadline, ok := got.(int64)
	if !ok {
		t.Fatalf("deadline type = %T, want int64", got)
	}
	if deadline <= 100*365*24*60*60 {
		t.Fatalf("deadline = %d, want unix timestamp plus 100 years", deadline)
	}
}
