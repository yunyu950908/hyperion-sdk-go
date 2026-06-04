package hyperion

import "testing"

func TestTickComplementMatchesUnsignedInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tick int64
		want uint32
	}{
		{name: "positive", tick: 443636, want: 443636},
		{name: "minus one", tick: -1, want: 4294967295},
		{name: "lowest tick", tick: -443636, want: 4294523660},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TickComplement(tt.tick); got != tt.want {
				t.Fatalf("TickComplement(%d) = %d, want %d", tt.tick, got, tt.want)
			}
		})
	}
}

func TestSlippageCalculatorRoundsToWholeUnits(t *testing.T) {
	t.Parallel()

	got, err := SlippageCalculator("100000", "0.5")
	if err != nil {
		t.Fatalf("SlippageCalculator returned error: %v", err)
	}
	if got != "99500" {
		t.Fatalf("slippage result = %q, want 99500", got)
	}
}

func TestSlippageCheckBounds(t *testing.T) {
	t.Parallel()

	if err := CheckSlippage("-0.1"); err == nil {
		t.Fatal("CheckSlippage accepted negative slippage")
	}
	if err := CheckSlippage("20.1"); err == nil {
		t.Fatal("CheckSlippage accepted slippage above 20")
	}
	if err := CheckSlippage("20"); err != nil {
		t.Fatalf("CheckSlippage rejected upper bound: %v", err)
	}
}

func TestCurrencyCheckRequiresAptosStyleAddresses(t *testing.T) {
	t.Parallel()

	if err := CheckCurrencyPair("0x1", "0x2"); err != nil {
		t.Fatalf("CheckCurrencyPair returned error: %v", err)
	}
	if err := CheckCurrencyPair("apt", "0x2"); err == nil {
		t.Fatal("CheckCurrencyPair accepted invalid currencyA")
	}
}

func TestRoundTickBySpacingMatchesTypeScriptSDK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tick    float64
		feeTier FeeTierIndex
		want    string
	}{
		{name: "rounds positive to spacing", tick: 123.4, feeTier: FeeTier03Spacing60, want: "120"},
		{name: "rounds positive half up", tick: 150, feeTier: FeeTier03Spacing60, want: "180"},
		{name: "rounds negative half away from zero", tick: -150, feeTier: FeeTier03Spacing60, want: "-180"},
		{name: "matches lowest tick spacing", tick: -443636, feeTier: FeeTier005Spacing5, want: "-443640"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := RoundTickBySpacing(tt.tick, tt.feeTier)
			if err != nil {
				t.Fatalf("RoundTickBySpacing returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("RoundTickBySpacing(%v, %d) = %q, want %q", tt.tick, tt.feeTier, got, tt.want)
			}
		})
	}
}

func TestPriceToTickMatchesTypeScriptSDK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		price         float64
		feeTier       FeeTierIndex
		decimalsRatio float64
		want          string
	}{
		{name: "unit price maps to zero tick", price: 1, feeTier: FeeTier03Spacing60, decimalsRatio: 1, want: "0"},
		{name: "price above one rounds to spacing", price: 1.01, feeTier: FeeTier03Spacing60, decimalsRatio: 1, want: "120"},
		{name: "price below one rounds negative", price: 0.99, feeTier: FeeTier03Spacing60, decimalsRatio: 1, want: "-120"},
		{name: "decimals ratio is applied before log", price: 100, feeTier: FeeTier03Spacing60, decimalsRatio: 100, want: "0"},
		{name: "upper clamp follows rounded highest tick", price: 1e100, feeTier: FeeTier03Spacing60, decimalsRatio: 1, want: "443640"},
		{name: "lower clamp follows rounded lowest tick", price: 1e-100, feeTier: FeeTier03Spacing60, decimalsRatio: 1, want: "-443640"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok, err := PriceToTick(PriceToTickArgs{
				Price:         tt.price,
				FeeTierIndex:  tt.feeTier,
				DecimalsRatio: tt.decimalsRatio,
			})
			if err != nil {
				t.Fatalf("PriceToTick returned error: %v", err)
			}
			if !ok {
				t.Fatal("PriceToTick returned ok=false")
			}
			if got != tt.want {
				t.Fatalf("PriceToTick(%v) = %q, want %q", tt.price, got, tt.want)
			}
		})
	}
}

func TestPriceToTickRejectsNaNLikeInputs(t *testing.T) {
	t.Parallel()

	if _, ok, err := PriceToTick(PriceToTickArgs{
		Price:         -1,
		FeeTierIndex:  FeeTier03Spacing60,
		DecimalsRatio: 1,
	}); err != nil || ok {
		t.Fatalf("PriceToTick negative price = ok %v err %v, want ok=false err=nil", ok, err)
	}
}

func TestTickToPriceMatchesTypeScriptSDK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tick          float64
		decimalsRatio float64
		want          string
	}{
		{name: "zero tick", tick: 0, decimalsRatio: 1, want: "1"},
		{name: "decimals ratio is multiplied", tick: 0, decimalsRatio: 100, want: "100"},
		{name: "positive tick", tick: 60, decimalsRatio: 1, want: "1.0060177342688164"},
		{name: "fractional tick is rounded first", tick: 60.4, decimalsRatio: 1, want: "1.0060177342688164"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TickToPrice(TickToPriceArgs{
				Tick:          tt.tick,
				DecimalsRatio: tt.decimalsRatio,
			})
			if got != tt.want {
				t.Fatalf("TickToPrice(%v) = %q, want %q", tt.tick, got, tt.want)
			}
		})
	}
}

func TestU64MaxMatchesTypeScriptSDKConstant(t *testing.T) {
	t.Parallel()

	if U64Max != "184467440737095516" {
		t.Fatalf("U64Max = %q", U64Max)
	}
}
