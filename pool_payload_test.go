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
