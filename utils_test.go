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
