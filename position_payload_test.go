package hyperion

import "testing"

func TestPositionAmountByLiquidityViewPayloadMatchesGoldenFixture(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	payload := sdk.Position.FetchTokensAmountByPositionIDPayload("pos-1")

	assertGoldenJSON(t, "testdata/parity/payloads/position_amount_by_liquidity_view.json", normalizePayloadSnapshot(payload))
}

func TestRemoveLiquidityRejectsInvalidRecipientAddress(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	_, err := sdk.Position.RemoveLiquidityTransactionPayload(RemoveLiquidityTransactionPayloadArgs{
		PositionID:      "pos-1",
		CurrencyA:       "0x1",
		CurrencyB:       "0x2",
		CurrencyAAmount: "1000",
		CurrencyBAmount: "2000",
		DeltaLiquidity:  "12345",
		Slippage:        "0.5",
		Recipient:       "0xabc",
	})
	if err == nil {
		t.Fatal("RemoveLiquidityTransactionPayload accepted a non-strict recipient address")
	}
}

func TestRemoveLiquidityAcceptsStrictRecipientAddress(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	payload, err := sdk.Position.RemoveLiquidityTransactionPayload(RemoveLiquidityTransactionPayloadArgs{
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
		t.Fatalf("RemoveLiquidityTransactionPayload rejected strict recipient address: %v", err)
	}
	if payload.Function != MainnetContractAddress+"::router_adapter::remove_liquidity_entry_v2" {
		t.Fatalf("function = %q", payload.Function)
	}
}
