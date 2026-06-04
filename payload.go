package hyperion

import (
	"errors"
	"strconv"
	"strings"
)

// EntryFunctionPayload describes an Aptos entry function payload in the same
// shape returned by the upstream TypeScript SDK.
type EntryFunctionPayload struct {
	Function          string   `json:"function"`
	TypeArguments     []string `json:"typeArguments"`
	FunctionArguments []any    `json:"functionArguments"`
}

// SwapTransactionPayloadArgs configures a swap entry-function payload.
type SwapTransactionPayloadArgs struct {
	CurrencyA       string
	CurrencyB       string
	CurrencyAAmount string
	CurrencyBAmount string
	Slippage        string
	PoolRoute       []string
	Recipient       string
}

// SwapWithPartnershipTransactionPayloadArgs configures a partnership swap payload.
type SwapWithPartnershipTransactionPayloadArgs struct {
	SwapTransactionPayloadArgs
	Partnership string
}

// CreatePoolTransactionPayloadArgs configures a create-liquidity-pool payload.
type CreatePoolTransactionPayloadArgs struct {
	CurrencyA        string
	CurrencyB        string
	CurrencyAAmount  string
	CurrencyBAmount  string
	FeeTierIndex     string
	CurrentPriceTick string
	TickLower        string
	TickUpper        string
	Slippage         string
}

// AddLiquidityTransactionPayloadArgs configures an add-liquidity payload.
type AddLiquidityTransactionPayloadArgs struct {
	PositionID      string
	CurrencyA       string
	CurrencyB       string
	CurrencyAAmount string
	CurrencyBAmount string
	Slippage        string
	FeeTierIndex    string
}

// RemoveLiquidityTransactionPayloadArgs configures a remove-liquidity payload.
type RemoveLiquidityTransactionPayloadArgs struct {
	PositionID      string
	CurrencyA       string
	CurrencyB       string
	CurrencyAAmount string
	CurrencyBAmount string
	DeltaLiquidity  string
	Slippage        string
	Recipient       string
}

type poolEstAmountArgs struct {
	CurrencyA        string
	CurrencyB        string
	FeeTierIndex     string
	TickLower        string
	TickUpper        string
	CurrentPriceTick string
}

// EstCurrencyAAmountFromBArgs configures a currency-A-from-B view payload.
type EstCurrencyAAmountFromBArgs struct {
	CurrencyA        string
	CurrencyB        string
	FeeTierIndex     string
	TickLower        string
	TickUpper        string
	CurrentPriceTick string
	CurrencyBAmount  string
}

// EstCurrencyBAmountFromAArgs configures a currency-B-from-A view payload.
type EstCurrencyBAmountFromAArgs struct {
	CurrencyA        string
	CurrencyB        string
	FeeTierIndex     string
	TickLower        string
	TickUpper        string
	CurrentPriceTick string
	CurrencyAAmount  string
}

// EstCurrencyAAmountFromBPayload builds the view payload for estimating token A from token B.
func (p *PoolService) EstCurrencyAAmountFromBPayload(args EstCurrencyAAmountFromBArgs) (EntryFunctionPayload, error) {
	tickLower, tickUpper, currentPriceTick, err := parseEstimateTicks(poolEstAmountArgs{
		CurrencyA:        args.CurrencyA,
		CurrencyB:        args.CurrencyB,
		FeeTierIndex:     args.FeeTierIndex,
		TickLower:        args.TickLower,
		TickUpper:        args.TickUpper,
		CurrentPriceTick: args.CurrentPriceTick,
	})
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:      p.client.Options.ContractAddress + "::router_v3::optimal_liquidity_amounts_from_b",
		TypeArguments: []string{},
		FunctionArguments: []any{
			TickComplement(tickLower),
			TickComplement(tickUpper),
			TickComplement(currentPriceTick),
			args.CurrencyA,
			args.CurrencyB,
			args.FeeTierIndex,
			args.CurrencyBAmount,
			0,
			0,
		},
	}, nil
}

// EstCurrencyBAmountFromAPayload builds the view payload for estimating token B from token A.
func (p *PoolService) EstCurrencyBAmountFromAPayload(args EstCurrencyBAmountFromAArgs) (EntryFunctionPayload, error) {
	tickLower, tickUpper, currentPriceTick, err := parseEstimateTicks(poolEstAmountArgs{
		CurrencyA:        args.CurrencyA,
		CurrencyB:        args.CurrencyB,
		FeeTierIndex:     args.FeeTierIndex,
		TickLower:        args.TickLower,
		TickUpper:        args.TickUpper,
		CurrentPriceTick: args.CurrentPriceTick,
	})
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:      p.client.Options.ContractAddress + "::router_v3::optimal_liquidity_amounts_from_a",
		TypeArguments: []string{},
		FunctionArguments: []any{
			TickComplement(tickLower),
			TickComplement(tickUpper),
			TickComplement(currentPriceTick),
			args.CurrencyA,
			args.CurrencyB,
			args.FeeTierIndex,
			args.CurrencyAAmount,
			0,
			0,
		},
	}, nil
}

// CreatePoolTransactionPayload builds a create liquidity pool payload.
func (p *PoolService) CreatePoolTransactionPayload(args CreatePoolTransactionPayloadArgs) (EntryFunctionPayload, error) {
	if err := CheckCurrencyPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	slippage := normalizeSlippage(args.Slippage)
	if err := CheckSlippage(slippage); err != nil {
		return EntryFunctionPayload{}, err
	}

	tickLower, err := parseInt64(args.TickLower)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	tickUpper, err := parseInt64(args.TickUpper)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	currentPriceTick, err := parseInt64(args.CurrentPriceTick)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	slippageA, slippageB, err := slippageAmounts(args.CurrencyAAmount, args.CurrencyBAmount, slippage)
	if err != nil {
		return EntryFunctionPayload{}, err
	}

	params := []any{
		args.FeeTierIndex,
		PoolStableType,
		TickComplement(tickLower),
		TickComplement(tickUpper),
		TickComplement(currentPriceTick),
		args.CurrencyAAmount,
		args.CurrencyBAmount,
		slippageA,
		slippageB,
		PoolDeadline(),
	}
	paramsReverse := append([]any(nil), params...)
	paramsReverse[5], paramsReverse[6] = paramsReverse[6], paramsReverse[5]
	paramsReverse[7], paramsReverse[8] = paramsReverse[8], paramsReverse[7]
	paramsReverse[2], paramsReverse[3], paramsReverse[4] = TickComplement(-tickUpper), TickComplement(-tickLower), TickComplement(-currentPriceTick)

	contract := p.client.Options.ContractAddress
	return selectTokenPairPayload(args.CurrencyA, args.CurrencyB, []EntryFunctionPayload{
		{
			Function:          contract + "::router_adapter::create_liquidity_entry",
			TypeArguments:     []string{},
			FunctionArguments: append([]any{args.CurrencyA, args.CurrencyB}, params...),
		},
		{
			Function:          contract + "::router_adapter::create_liquidity_both_coin_entry",
			TypeArguments:     []string{args.CurrencyA, args.CurrencyB},
			FunctionArguments: params,
		},
		{
			Function:          contract + "::router_adapter::create_liquidity_coin_entry",
			TypeArguments:     []string{args.CurrencyA},
			FunctionArguments: append([]any{args.CurrencyB}, params...),
		},
		{
			Function:          contract + "::router_adapter::create_liquidity_coin_entry",
			TypeArguments:     []string{args.CurrencyB},
			FunctionArguments: append([]any{args.CurrencyA}, paramsReverse...),
		},
	}), nil
}

// AddLiquidityTransactionPayload builds an add-liquidity entry-function payload.
func (p *PositionService) AddLiquidityTransactionPayload(args AddLiquidityTransactionPayloadArgs) (EntryFunctionPayload, error) {
	if err := CheckCurrencyPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	slippage := normalizeSlippage(args.Slippage)
	if err := CheckSlippage(slippage); err != nil {
		return EntryFunctionPayload{}, err
	}
	slippageA, slippageB, err := slippageAmounts(args.CurrencyAAmount, args.CurrencyBAmount, slippage)
	if err != nil {
		return EntryFunctionPayload{}, err
	}

	params := []any{
		args.FeeTierIndex,
		PoolStableType,
		args.CurrencyAAmount,
		args.CurrencyBAmount,
		slippageA,
		slippageB,
		PoolDeadline(),
	}
	paramsReverse := append([]any(nil), params...)
	paramsReverse[2], paramsReverse[3] = paramsReverse[3], paramsReverse[2]
	paramsReverse[4], paramsReverse[5] = paramsReverse[5], paramsReverse[4]

	contract := p.client.Options.ContractAddress
	return selectTokenPairPayload(args.CurrencyA, args.CurrencyB, []EntryFunctionPayload{
		{
			Function:          contract + "::router_adapter::add_liquidity_entry",
			TypeArguments:     []string{},
			FunctionArguments: append([]any{args.PositionID, args.CurrencyA, args.CurrencyB}, params...),
		},
		{
			Function:          contract + "::router_adapter::add_liquidity_both_coin_entry",
			TypeArguments:     []string{args.CurrencyA, args.CurrencyB},
			FunctionArguments: append([]any{args.PositionID}, params...),
		},
		{
			Function:          contract + "::router_adapter::add_liquidity_coin_entry",
			TypeArguments:     []string{args.CurrencyA},
			FunctionArguments: append([]any{args.PositionID, args.CurrencyB}, params...),
		},
		{
			Function:          contract + "::router_adapter::add_liquidity_coin_entry",
			TypeArguments:     []string{args.CurrencyB},
			FunctionArguments: append([]any{args.PositionID, args.CurrencyA}, paramsReverse...),
		},
	}), nil
}

// RemoveLiquidityTransactionPayload builds a remove-liquidity entry-function payload.
func (p *PositionService) RemoveLiquidityTransactionPayload(args RemoveLiquidityTransactionPayloadArgs) (EntryFunctionPayload, error) {
	if err := CheckCurrencyPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	slippage := normalizeSlippage(args.Slippage)
	if err := CheckSlippage(slippage); err != nil {
		return EntryFunctionPayload{}, err
	}
	if !strings.HasPrefix(args.Recipient, "0x") {
		return EntryFunctionPayload{}, errors.New("invalid recipient address")
	}
	deltaLiquidity, err := roundDecimalString(args.DeltaLiquidity)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	slippageA, slippageB, err := slippageAmounts(args.CurrencyAAmount, args.CurrencyBAmount, slippage)
	if err != nil {
		return EntryFunctionPayload{}, err
	}

	functionArguments := []any{
		args.PositionID,
		deltaLiquidity,
		slippageA,
		slippageB,
		args.Recipient,
		PoolDeadline(),
	}
	contract := p.client.Options.ContractAddress
	return selectTokenPairPayload(args.CurrencyA, args.CurrencyB, []EntryFunctionPayload{
		{
			Function:          contract + "::router_adapter::remove_liquidity_entry_v2",
			TypeArguments:     []string{},
			FunctionArguments: functionArguments,
		},
		{
			Function:          contract + "::router_adapter::remove_liquidity_both_coins_entry_v2",
			TypeArguments:     []string{args.CurrencyA, args.CurrencyB},
			FunctionArguments: functionArguments,
		},
		{
			Function:          contract + "::router_adapter::remove_liquidity_coin_entry_v2",
			TypeArguments:     []string{args.CurrencyA},
			FunctionArguments: functionArguments,
		},
		{
			Function:          contract + "::router_adapter::remove_liquidity_coin_entry_v2",
			TypeArguments:     []string{args.CurrencyB},
			FunctionArguments: functionArguments,
		},
	}), nil
}

// FetchRewardsPayload builds a pending rewards Aptos view payload.
func (r *RewardService) FetchRewardsPayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::pool_v3::get_pending_rewards",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

// ClaimRewardPayload builds a reward claim entry-function payload.
func (r *RewardService) ClaimRewardPayload(positionID, recipient string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::router_v3::claim_rewards",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID, recipient},
	}
}

// ClaimFeeTransactionPayload builds a fee claim entry-function payload.
func (p *PositionService) ClaimFeeTransactionPayload(positionID, recipient string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::router_v3::claim_fees",
		TypeArguments:     []string{},
		FunctionArguments: []any{[]string{positionID}, recipient},
	}
}

// ClaimRewardTransactionPayload builds a reward claim entry-function payload.
func (p *PositionService) ClaimRewardTransactionPayload(positionID, recipient string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::router_v3::claim_rewards",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID, recipient},
	}
}

// ClaimAllRewardsTransactionPayload builds a combined fees and rewards claim payload.
func (p *PositionService) ClaimAllRewardsTransactionPayload(positionID, recipient string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::router_v3::claim_fees_and_rewards",
		TypeArguments:     []string{},
		FunctionArguments: []any{[]string{positionID}, recipient},
	}
}

// SwapTransactionPayload builds a basic swap entry-function payload.
func (s *SwapService) SwapTransactionPayload(args SwapTransactionPayloadArgs) (EntryFunctionPayload, error) {
	if err := CheckCurrencyPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	if isCoinType(args.CurrencyA) || isCoinType(args.CurrencyB) {
		return EntryFunctionPayload{}, errors.New("swap payload coin type to fungible asset conversion is not implemented")
	}
	slippage := normalizeSlippage(args.Slippage)
	if err := CheckSlippage(slippage); err != nil {
		return EntryFunctionPayload{}, err
	}
	amountB, err := SlippageCalculator(args.CurrencyBAmount, slippage)
	if err != nil {
		return EntryFunctionPayload{}, err
	}

	params := []any{args.PoolRoute, args.CurrencyA, args.CurrencyB, args.CurrencyAAmount, amountB, args.Recipient}
	contract := s.client.Options.ContractAddress
	return selectTokenPairPayload(args.CurrencyA, args.CurrencyB, []EntryFunctionPayload{
		{
			Function:          contract + "::router_v3::swap_batch",
			TypeArguments:     []string{},
			FunctionArguments: params,
		},
		{
			Function:          contract + "::router_v3::swap_batch_coin_entry",
			TypeArguments:     []string{args.CurrencyA},
			FunctionArguments: params,
		},
		{
			Function:          contract + "::router_v3::swap_batch_coin_entry",
			TypeArguments:     []string{args.CurrencyA},
			FunctionArguments: params,
		},
		{
			Function:          contract + "::router_v3::swap_batch",
			TypeArguments:     []string{},
			FunctionArguments: params,
		},
	}), nil
}

// SwapWithPartnershipTransactionPayload builds a partnership swap payload.
func (s *SwapService) SwapWithPartnershipTransactionPayload(args SwapWithPartnershipTransactionPayloadArgs) (EntryFunctionPayload, error) {
	if err := CheckCurrencyPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	if isCoinType(args.CurrencyA) || isCoinType(args.CurrencyB) {
		return EntryFunctionPayload{}, errors.New("swap payload coin type to fungible asset conversion is not implemented")
	}
	slippage := normalizeSlippage(args.Slippage)
	if err := CheckSlippage(slippage); err != nil {
		return EntryFunctionPayload{}, err
	}
	if args.Partnership == "" {
		return EntryFunctionPayload{}, errors.New("partnership is required")
	}
	amountB, err := SlippageCalculator(args.CurrencyBAmount, slippage)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:      s.client.Options.ContractAddress + "::partnership::swap_batch_directly_deposit",
		TypeArguments: []string{},
		FunctionArguments: []any{
			args.PoolRoute,
			args.CurrencyA,
			args.CurrencyB,
			args.CurrencyAAmount,
			amountB,
			args.Partnership,
		},
	}, nil
}

func selectTokenPairPayload(currencyA, currencyB string, variants []EntryFunctionPayload) EntryFunctionPayload {
	if len(variants) != 4 {
		return EntryFunctionPayload{}
	}

	aCoin := isCoinType(currencyA)
	bCoin := isCoinType(currencyB)
	switch {
	case !aCoin && !bCoin:
		return variants[0]
	case aCoin && bCoin:
		return variants[1]
	case aCoin:
		return variants[2]
	default:
		return variants[3]
	}
}

func isCoinType(currency string) bool {
	return strings.Contains(currency, "::")
}

func slippageAmounts(amountA, amountB, slippage string) (string, string, error) {
	slippageA, err := SlippageCalculator(amountA, slippage)
	if err != nil {
		return "", "", err
	}
	slippageB, err := SlippageCalculator(amountB, slippage)
	if err != nil {
		return "", "", err
	}
	return slippageA, slippageB, nil
}

func normalizeSlippage(slippage string) string {
	if slippage == "" {
		return "0.5"
	}
	if _, err := parseRat(slippage); err != nil {
		return "0.5"
	}
	return slippage
}

func roundDecimalString(value string) (string, error) {
	rat, err := parseRat(value)
	if err != nil {
		return "", err
	}
	return roundRatHalfUp(rat), nil
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseEstimateTicks(args poolEstAmountArgs) (int64, int64, int64, error) {
	tickLower, err := parseInt64(args.TickLower)
	if err != nil {
		return 0, 0, 0, err
	}
	tickUpper, err := parseInt64(args.TickUpper)
	if err != nil {
		return 0, 0, 0, err
	}
	currentPriceTick, err := parseInt64(args.CurrentPriceTick)
	if err != nil {
		return 0, 0, 0, err
	}
	return tickLower, tickUpper, currentPriceTick, nil
}
