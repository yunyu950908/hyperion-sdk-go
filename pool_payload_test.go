package hyperion

import "testing"

func TestPoolEstimateViewPayloadsMatchGoldenFixtures(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)

	fromB, err := sdk.Pool.EstCurrencyAAmountFromBPayload(EstCurrencyAAmountFromBArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyBAmount:  "2000",
	})
	if err != nil {
		t.Fatalf("EstCurrencyAAmountFromBPayload returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/payloads/est_currency_a_from_b_view.json", normalizePayloadSnapshot(fromB))

	fromA, err := sdk.Pool.EstCurrencyBAmountFromAPayload(EstCurrencyBAmountFromAArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyAAmount:  "1000",
	})
	if err != nil {
		t.Fatalf("EstCurrencyBAmountFromAPayload returned error: %v", err)
	}
	assertGoldenJSON(t, "testdata/parity/payloads/est_currency_b_from_a_view.json", normalizePayloadSnapshot(fromA))
}

func TestPoolEstimateViewPayloadUsesNumericFeeTier(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	payload, err := sdk.Pool.EstCurrencyAAmountFromBPayload(EstCurrencyAAmountFromBArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyBAmount:  "2000",
	})
	if err != nil {
		t.Fatalf("EstCurrencyAAmountFromBPayload returned error: %v", err)
	}

	feeTier, ok := payload.FunctionArguments[5].(uint8)
	if !ok {
		t.Fatalf("fee tier argument type = %T, want uint8", payload.FunctionArguments[5])
	}
	if feeTier != 2 {
		t.Fatalf("fee tier argument = %d, want 2", feeTier)
	}
}

func TestPoolEstimateViewPayloadUsesStringU64Placeholders(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	payload, err := sdk.Pool.EstCurrencyAAmountFromBPayload(EstCurrencyAAmountFromBArgs{
		CurrencyA:        "0x1",
		CurrencyB:        "0x2",
		FeeTierIndex:     "2",
		TickLower:        "-60",
		TickUpper:        "60",
		CurrentPriceTick: "0",
		CurrencyBAmount:  "2000",
	})
	if err != nil {
		t.Fatalf("EstCurrencyAAmountFromBPayload returned error: %v", err)
	}

	if payload.FunctionArguments[7] != "0" {
		t.Fatalf("argument 7 = %#v, want string zero", payload.FunctionArguments[7])
	}
	if payload.FunctionArguments[8] != "0" {
		t.Fatalf("argument 8 = %#v, want string zero", payload.FunctionArguments[8])
	}
}
