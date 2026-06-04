package hyperion

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	// Base is the tick math base used by Hyperion pools.
	Base = 1.0001
	// LowestTick is the minimum supported pool tick.
	LowestTick = "-443636"
	// HighestTick is the maximum supported pool tick.
	HighestTick = "443636"
	// PoolStableType matches the upstream SDK's stable-pool flag for payload construction.
	PoolStableType = false
)

// FeeTierIndex identifies a Hyperion fee tier.
type FeeTierIndex int

const (
	FeeTier001Spacing1  FeeTierIndex = 0
	FeeTier005Spacing5  FeeTierIndex = 1
	FeeTier03Spacing60  FeeTierIndex = 2
	FeeTier1Spacing200  FeeTierIndex = 3
	FeeTier01Spacing20  FeeTierIndex = 4
	FeeTier025Spacing50 FeeTierIndex = 5
)

// FeeTierItems contains upstream fee tier labels in basis-point-like units.
var FeeTierItems = []string{"1", "5", "30", "100", "10", "25"}

// FeeTierStep contains the tick spacing for each fee tier.
var FeeTierStep = []int{1, 10, 60, 200, 20, 50}

// LowestTickByStep contains the lower tick bound for each fee tier.
var LowestTickByStep = []int{-443636, -443630, -443580, -443600, -443620, -443600}

// HighestTickByStep contains the upper tick bound for each fee tier.
var HighestTickByStep = []int{443636, 443630, 443580, 443600, 443620, 443600}

// TickComplement mirrors Long.fromInt(tick).toUnsigned().toInt() in the TypeScript SDK.
func TickComplement(tick int64) uint32 {
	return uint32(int32(tick))
}

// PoolDeadline returns the upstream SDK's pool deadline: roughly 100 years from now.
func PoolDeadline() int64 {
	return int64(100*365*24*60*60) + time.Now().Unix()
}

// CheckCurrencyPair validates that both currencies look like Aptos addresses or asset types.
func CheckCurrencyPair(currencyA, currencyB string) error {
	if currencyA == "" || currencyB == "" {
		return errors.New("currencyA and currencyB are required and can not be empty")
	}
	if !strings.HasPrefix(currencyA, "0x") || !strings.HasPrefix(currencyB, "0x") {
		return errors.New("currencyA and currencyB must be valid aptos account/token address")
	}
	return nil
}

// CheckSlippage validates slippage bounds. Empty slippage is accepted as the upstream default.
func CheckSlippage(slippage string) error {
	if slippage == "" {
		return nil
	}
	value, err := parseRat(slippage)
	if err != nil {
		return nil
	}
	if value.Sign() < 0 {
		return errors.New("slippage must be greater than 0")
	}
	if value.Cmp(big.NewRat(20, 1)) > 0 {
		return errors.New("slippage must be less than 20")
	}
	return nil
}

// SlippageCalculator subtracts slippage from amount and rounds to a whole unit.
func SlippageCalculator(amount, slippage string) (string, error) {
	amountRat, err := parseRat(amount)
	if err != nil {
		return "", fmt.Errorf("amount: %w", err)
	}
	slippageRat, err := parseRat(slippage)
	if err != nil {
		return "", fmt.Errorf("slippage: %w", err)
	}

	multiplier := new(big.Rat).Sub(big.NewRat(100, 1), slippageRat)
	result := new(big.Rat).Mul(amountRat, multiplier)
	result.Quo(result, big.NewRat(100, 1))
	return roundRatHalfUp(result), nil
}

// LogBase returns log(number) using the Hyperion tick base.
func LogBase(number float64) float64 {
	return math.Log(number) / math.Log(Base)
}

// RoundTickBySpacing rounds a tick to the fee tier's configured spacing.
func RoundTickBySpacing(tick float64, feeTierIndex FeeTierIndex) (string, error) {
	index := int(feeTierIndex)
	if index < 0 || index >= len(FeeTierStep) {
		return "", errors.New("fee tier index out of range")
	}
	step := float64(FeeTierStep[index])
	return strconv.FormatInt(int64(math.Round(tick/step)*step), 10), nil
}

func parseRat(value string) (*big.Rat, error) {
	rat, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid decimal %q", value)
	}
	return rat, nil
}

func roundRatHalfUp(value *big.Rat) string {
	num := new(big.Int).Set(value.Num())
	den := new(big.Int).Set(value.Denom())
	quotient, remainder := new(big.Int).QuoRem(num, den, new(big.Int))
	if value.Sign() >= 0 {
		doubled := new(big.Int).Mul(remainder, big.NewInt(2))
		if doubled.Cmp(den) >= 0 {
			quotient.Add(quotient, big.NewInt(1))
		}
	}
	return quotient.String()
}
