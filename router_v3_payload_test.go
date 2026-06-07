package hyperion

import "testing"

const (
	calibratedCurrencyA = "0x357b0b74bc833e95a115ad22604854d6b0fca151cecd94111770e5d6ffc9dc2b"
	calibratedCurrencyB = "0xbae207659db88bea0cbead6da0ed00aac12edcdda169e591cd41c94180b46f3b"
)

func TestRouterV3LiquidityPayloadsMatchGoldenFixture(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	createSingle, err := sdk.Pool.CreateLiquiditySinglePayload(CreateLiquiditySinglePayloadArgs{
		CurrencyA:            calibratedCurrencyA,
		CurrencyB:            calibratedCurrencyB,
		FeeTierIndex:         "0",
		TickLower:            "-3",
		TickUpper:            "-1",
		Amount:               "25086537",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	})
	if err != nil {
		t.Fatalf("CreateLiquiditySinglePayload returned error: %v", err)
	}

	addSingle, err := sdk.Position.AddLiquiditySinglePayload(AddLiquiditySinglePayloadArgs{
		PositionID:           "0xa836eccaeb80072e69f46a01b6899cfb8797d31fdff4dc3ac104296d5370d62a",
		FromCurrency:         calibratedCurrencyA,
		ToCurrency:           calibratedCurrencyB,
		Amount:               "25086537",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	})
	if err != nil {
		t.Fatalf("AddLiquiditySinglePayload returned error: %v", err)
	}

	addSingleCoins, err := sdk.Position.AddLiquiditySingleCoinsPayload(AddLiquiditySingleCoinsPayloadArgs{
		PositionID:           "0xa836eccaeb80072e69f46a01b6899cfb8797d31fdff4dc3ac104296d5370d62a",
		CoinType:             aptosCoinType,
		PairedCurrency:       calibratedCurrencyB,
		Amount:               "25086537",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	})
	if err != nil {
		t.Fatalf("AddLiquiditySingleCoinsPayload returned error: %v", err)
	}

	remove, err := sdk.Position.RemoveLiquidityMultiAgentDirectlyDepositPayload(RemoveLiquidityMultiAgentDirectlyDepositPayloadArgs{
		PositionID:     "0x2ada87ed27ac997b88dfdeb430d0b8c1abbd40b28be11aacb745da5d3fca26f9",
		DeltaLiquidity: "1259530800761551",
		MinAmountA:     "0",
		MinAmountB:     "62307822006",
		Deadline:       "4936478833",
		SecondarySignerAddresses: []string{
			"0xce702f163f8f9a95e842584ad9377855030442837416045d8990c2eabb0e0ace",
			"0xfb644c0c2cadc6ea2041ebb19492e6e071709665a55a2cf302290460c144fcdf",
		},
	})
	if err != nil {
		t.Fatalf("RemoveLiquidityMultiAgentDirectlyDepositPayload returned error: %v", err)
	}

	assertGoldenJSON(t, "testdata/parity/payloads/router_v3_liquidity_payloads.json", map[string]any{
		"createSingle":   normalizePayloadSnapshot(createSingle),
		"addSingle":      normalizePayloadSnapshot(addSingle),
		"addSingleCoins": normalizePayloadSnapshot(addSingleCoins),
		"removeMultiAgent": map[string]any{
			"payload":                  normalizePayloadSnapshot(remove.Payload),
			"secondarySignerAddresses": remove.SecondarySignerAddresses,
		},
	})
}

func TestRouterV3LiquidityPayloadValidation(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	if _, err := sdk.Pool.CreateLiquiditySinglePayload(CreateLiquiditySinglePayloadArgs{
		CurrencyA:            calibratedCurrencyA,
		CurrencyB:            calibratedCurrencyB,
		FeeTierIndex:         "not-a-tier",
		TickLower:            "-3",
		TickUpper:            "-1",
		Amount:               "25086537",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	}); err == nil {
		t.Fatal("CreateLiquiditySinglePayload accepted invalid fee tier")
	}
	if _, err := sdk.Pool.CreateLiquiditySinglePayload(CreateLiquiditySinglePayloadArgs{
		CurrencyA:            calibratedCurrencyA,
		CurrencyB:            calibratedCurrencyB,
		FeeTierIndex:         "0",
		TickLower:            "invalid",
		TickUpper:            "-1",
		Amount:               "25086537",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	}); err == nil {
		t.Fatal("CreateLiquiditySinglePayload accepted invalid tick")
	}
	if _, err := sdk.Pool.CreateLiquiditySinglePayload(CreateLiquiditySinglePayloadArgs{
		CurrencyA:            aptosCoinType,
		CurrencyB:            calibratedCurrencyB,
		FeeTierIndex:         "0",
		TickLower:            "-3",
		TickUpper:            "-1",
		Amount:               "25086537",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	}); err == nil {
		t.Fatal("CreateLiquiditySinglePayload accepted coin type metadata")
	}
	if _, err := sdk.Position.AddLiquiditySinglePayload(AddLiquiditySinglePayloadArgs{
		PositionID:           "pos-1",
		FromCurrency:         calibratedCurrencyA,
		ToCurrency:           calibratedCurrencyB,
		Amount:               "",
		SlippageNumerator:    "99",
		SlippageDenominator:  "100",
		ThresholdNumerator:   "1",
		ThresholdDenominator: "1",
	}); err == nil {
		t.Fatal("AddLiquiditySinglePayload accepted empty amount")
	}
	if _, err := sdk.Position.RemoveLiquidityMultiAgentDirectlyDepositPayload(RemoveLiquidityMultiAgentDirectlyDepositPayloadArgs{
		PositionID:               "pos-1",
		DeltaLiquidity:           "1",
		MinAmountA:               "0",
		MinAmountB:               "1",
		Deadline:                 "4936478833",
		SecondarySignerAddresses: []string{"0x1"},
	}); err == nil {
		t.Fatal("RemoveLiquidityMultiAgentDirectlyDepositPayload accepted invalid secondary signers")
	}
}
