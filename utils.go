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
	// U64Max matches the upstream TypeScript SDK's u64Max constant.
	U64Max = "184467440737095516"
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

// PriceToTickArgs configures PriceToTick.
type PriceToTickArgs struct {
	Price         float64
	FeeTierIndex  FeeTierIndex
	DecimalsRatio float64
}

// PriceToTick converts a price to the nearest usable tick for a fee tier.
//
// The boolean return is false when the upstream TypeScript SDK would return
// null because the logarithm result is NaN.
func PriceToTick(args PriceToTickArgs) (string, bool, error) {
	ret := LogBase(args.Price / args.DecimalsRatio)
	if math.IsNaN(ret) {
		return "", false, nil
	}
	lowest, highest, err := roundedTickBounds(args.FeeTierIndex)
	if err != nil {
		return "", false, err
	}
	if math.IsInf(ret, -1) {
		return lowest, true, nil
	}
	if math.IsInf(ret, 1) {
		return highest, true, nil
	}

	rounded, err := RoundTickBySpacing(ret, args.FeeTierIndex)
	if err != nil {
		return "", false, err
	}

	roundedFloat, err := strconv.ParseFloat(rounded, 64)
	if err != nil {
		return "", false, err
	}
	lowestFloat, err := strconv.ParseFloat(lowest, 64)
	if err != nil {
		return "", false, err
	}
	highestFloat, err := strconv.ParseFloat(highest, 64)
	if err != nil {
		return "", false, err
	}
	if ret < 0 {
		return strconv.FormatFloat(math.Max(roundedFloat, lowestFloat), 'f', 0, 64), true, nil
	}
	return strconv.FormatFloat(math.Min(roundedFloat, highestFloat), 'f', 0, 64), true, nil
}

// TickToPriceArgs configures TickToPrice.
type TickToPriceArgs struct {
	Tick          float64
	DecimalsRatio float64
}

// TickToPrice converts a tick to a price adjusted by decimals ratio.
func TickToPrice(args TickToPriceArgs) string {
	tick := math.Round(args.Tick)
	price := math.Pow(Base, tick) * args.DecimalsRatio
	return strconv.FormatFloat(price, 'g', -1, 64)
}

// RoundTickBySpacing rounds a tick to the fee tier's configured spacing.
func RoundTickBySpacing(tick float64, feeTierIndex FeeTierIndex) (string, error) {
	index := int(feeTierIndex)
	if index < 0 || index >= len(FeeTierStep) {
		return "", errors.New("fee tier index out of range")
	}
	if math.IsNaN(tick) || math.IsInf(tick, 0) {
		return "", errors.New("tick must be finite")
	}
	step := float64(FeeTierStep[index])
	return strconv.FormatInt(int64(math.Round(tick/step)*step), 10), nil
}

func roundedTickBounds(feeTierIndex FeeTierIndex) (string, string, error) {
	lowestTick, err := strconv.ParseFloat(LowestTick, 64)
	if err != nil {
		return "", "", err
	}
	highestTick, err := strconv.ParseFloat(HighestTick, 64)
	if err != nil {
		return "", "", err
	}
	lowest, err := RoundTickBySpacing(lowestTick, feeTierIndex)
	if err != nil {
		return "", "", err
	}
	highest, err := RoundTickBySpacing(highestTick, feeTierIndex)
	if err != nil {
		return "", "", err
	}
	return lowest, highest, nil
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
