package hyperion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// PositionTokenAmounts is router_v3::get_amount_by_liquidity decoded output.
type PositionTokenAmounts struct {
	AmountA string `json:"amountA"`
	AmountB string `json:"amountB"`
}

// PendingFees is pool_v3::get_pending_fees decoded output.
type PendingFees struct {
	Amounts []string `json:"amounts"`
}

// PendingReward is one pool_v3::get_pending_rewards item.
type PendingReward struct {
	RewardFA   string `json:"rewardFA"`
	AmountOwed string `json:"amountOwed"`
}

// RewardRate is one pool_v3::get_position_emission_rate item.
type RewardRate struct {
	RewardFA string `json:"rewardFA"`
	Rate     string `json:"rate"`
}

// PositionPoolInfo is position_v3::get_pool_info decoded output.
type PositionPoolInfo struct {
	TokenA  string `json:"tokenA"`
	TokenB  string `json:"tokenB"`
	FeeTier uint8  `json:"feeTier"`
}

// SignedTick preserves the raw unsigned on-chain bits and signed tick value.
type SignedTick struct {
	Bits uint32 `json:"bits"`
	Tick int32  `json:"tick"`
}

// PositionTickRange is position_v3::get_tick decoded output.
type PositionTickRange struct {
	Lower SignedTick `json:"lower"`
	Upper SignedTick `json:"upper"`
}

// OptimalLiquidityAmountsArgs configures router_v3::optimal_liquidity_amounts.
type OptimalLiquidityAmountsArgs struct {
	TickLower          string
	TickUpper          string
	CurrencyA          string
	CurrencyB          string
	FeeTierIndex       string
	CurrencyAAmount    string
	CurrencyBAmount    string
	MinCurrencyAAmount string
	MinCurrencyBAmount string
}

// OptimalLiquidityAmounts is router_v3::optimal_liquidity_amounts decoded output.
type OptimalLiquidityAmounts struct {
	Liquidity       string `json:"liquidity"`
	CurrencyAAmount string `json:"currencyAAmount"`
	CurrencyBAmount string `json:"currencyBAmount"`
}

// OptimalLiquidityFromA is router_v3::optimal_liquidity_amounts_from_a decoded output.
type OptimalLiquidityFromA struct {
	Liquidity       string `json:"liquidity"`
	CurrencyBAmount string `json:"currencyBAmount"`
}

// OptimalLiquidityFromB is router_v3::optimal_liquidity_amounts_from_b decoded output.
type OptimalLiquidityFromB struct {
	Liquidity       string `json:"liquidity"`
	CurrencyAAmount string `json:"currencyAAmount"`
}

// PoolTokenPairArgs identifies a Hyperion pool by token metadata and fee tier.
type PoolTokenPairArgs struct {
	CurrencyA    string
	CurrencyB    string
	FeeTierIndex string
}

// LiquidityPoolAddress is pool_v3::liquidity_pool_address_safe decoded output.
type LiquidityPoolAddress struct {
	Exists  bool   `json:"exists"`
	Address string `json:"address"`
}

// PoolReserveAmount is pool_v3::pool_reserve_amount decoded output.
type PoolReserveAmount struct {
	AmountA string `json:"amountA"`
	AmountB string `json:"amountB"`
}

// CurrentTickAndPrice is pool_v3::current_tick_and_price decoded output.
type CurrentTickAndPrice struct {
	Tick  SignedTick `json:"tick"`
	Price string     `json:"price"`
}

// BatchAmountOutArgs configures router_v3::get_batch_amount_out.
type BatchAmountOutArgs struct {
	PoolRoute []string
	AmountIn  string
	TokenIn   string
	TokenOut  string
}

// BatchAmountInArgs configures router_v3::get_batch_amount_in.
type BatchAmountInArgs struct {
	PoolRoute []string
	AmountOut string
	TokenIn   string
	TokenOut  string
}

// PoolAmountOutArgs configures pool_v3::get_amount_out.
type PoolAmountOutArgs struct {
	PoolID   string
	TokenIn  string
	AmountIn string
}

// PoolAmountInArgs configures pool_v3::get_amount_in.
type PoolAmountInArgs struct {
	PoolID    string
	TokenOut  string
	AmountOut string
}

// PoolQuoteResult is pool_v3::get_amount_in/out decoded output.
type PoolQuoteResult struct {
	Amount    string `json:"amount"`
	FeeAmount string `json:"feeAmount"`
}

func (p *PositionService) PendingFeesPayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::get_pending_fees",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

func (p *PositionService) PendingRewardsPayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::get_pending_rewards",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

func (p *PositionService) PositionEmissionRatesPayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::get_position_emission_rate",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

func (p *PositionService) PositionLiquidityPayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::position_v3::get_liquidity",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

func (p *PositionService) PositionPoolInfoPayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::position_v3::get_pool_info",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

func (p *PositionService) PositionTickRangePayload(positionID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::position_v3::get_tick",
		TypeArguments:     []string{},
		FunctionArguments: []any{positionID},
	}
}

func (p *PoolService) OptimalLiquidityAmountsPayload(args OptimalLiquidityAmountsArgs) (EntryFunctionPayload, error) {
	if err := CheckCurrencyPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	feeTierIndex, err := parseUint8(args.FeeTierIndex)
	if err != nil {
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
	if err := requireRawIntegerStrings(map[string]string{
		"currencyAAmount":    args.CurrencyAAmount,
		"currencyBAmount":    args.CurrencyBAmount,
		"minCurrencyAAmount": args.MinCurrencyAAmount,
		"minCurrencyBAmount": args.MinCurrencyBAmount,
	}); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:      p.client.Options.ContractAddress + "::router_v3::optimal_liquidity_amounts",
		TypeArguments: []string{},
		FunctionArguments: []any{
			TickComplement(tickLower),
			TickComplement(tickUpper),
			args.CurrencyA,
			args.CurrencyB,
			feeTierIndex,
			args.CurrencyAAmount,
			args.CurrencyBAmount,
			args.MinCurrencyAAmount,
			args.MinCurrencyBAmount,
		},
	}, nil
}

func (p *PoolService) LiquidityPoolAddressSafePayload(args PoolTokenPairArgs) (EntryFunctionPayload, error) {
	return p.poolTokenPairPayload("liquidity_pool_address_safe", args)
}

func (p *PoolService) LiquidityPoolExistsPayload(args PoolTokenPairArgs) (EntryFunctionPayload, error) {
	return p.poolTokenPairPayload("liquidity_pool_exists", args)
}

func (p *PoolService) CurrentPricePayload(args PoolTokenPairArgs) (EntryFunctionPayload, error) {
	return p.poolTokenPairPayload("current_price", args)
}

func (p *PoolService) PoolReserveAmountPayload(poolID string) EntryFunctionPayload {
	return p.poolObjectPayload("pool_reserve_amount", poolID)
}

func (p *PoolService) CurrentTickAndPricePayload(poolID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::current_tick_and_price",
		TypeArguments:     []string{},
		FunctionArguments: []any{poolID},
	}
}

func (p *PoolService) PoolLiquidityPayload(poolID string) EntryFunctionPayload {
	return p.poolObjectPayload("get_pool_liquidity", poolID)
}

func (p *PoolService) FeeRatePayload(feeTierIndex string) (EntryFunctionPayload, error) {
	feeTier, err := parseUint8(feeTierIndex)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::get_fee_rate",
		TypeArguments:     []string{},
		FunctionArguments: []any{feeTier},
	}, nil
}

func (p *PoolService) SupportedInnerAssetsPayload(poolID string) EntryFunctionPayload {
	return p.poolObjectPayload("supported_inner_assets", poolID)
}

func (s *SwapService) BatchAmountOutPayload(args BatchAmountOutArgs) (EntryFunctionPayload, error) {
	if err := requireBatchQuoteArgs(args.PoolRoute, args.AmountIn, args.TokenIn, args.TokenOut, "amountIn"); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          s.client.Options.ContractAddress + "::router_v3::get_batch_amount_out",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.PoolRoute, args.AmountIn, args.TokenIn, args.TokenOut},
	}, nil
}

func (s *SwapService) BatchAmountInPayload(args BatchAmountInArgs) (EntryFunctionPayload, error) {
	if err := requireBatchQuoteArgs(args.PoolRoute, args.AmountOut, args.TokenIn, args.TokenOut, "amountOut"); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          s.client.Options.ContractAddress + "::router_v3::get_batch_amount_in",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.PoolRoute, args.AmountOut, args.TokenIn, args.TokenOut},
	}, nil
}

func (p *PoolService) PoolAmountOutPayload(args PoolAmountOutArgs) (EntryFunctionPayload, error) {
	if err := requirePoolQuoteArgs(args.PoolID, args.TokenIn, args.AmountIn, "tokenIn", "amountIn"); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::get_amount_out",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.PoolID, args.TokenIn, args.AmountIn},
	}, nil
}

func (p *PoolService) PoolAmountInPayload(args PoolAmountInArgs) (EntryFunctionPayload, error) {
	if err := requirePoolQuoteArgs(args.PoolID, args.TokenOut, args.AmountOut, "tokenOut", "amountOut"); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::get_amount_in",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.PoolID, args.TokenOut, args.AmountOut},
	}, nil
}

func (p *PositionService) FetchPositionTokenAmounts(ctx context.Context, positionID string) (PositionTokenAmounts, error) {
	values, err := p.client.View(ctx, p.FetchTokensAmountByPositionIDPayload(positionID))
	if err != nil {
		return PositionTokenAmounts{}, err
	}
	return decodePositionTokenAmounts(values)
}

func (p *PositionService) FetchPendingFees(ctx context.Context, positionID string) (PendingFees, error) {
	values, err := p.client.View(ctx, p.PendingFeesPayload(positionID))
	if err != nil {
		return PendingFees{}, err
	}
	amounts, err := decodeStringVectorResult(values, "pending fees")
	if err != nil {
		return PendingFees{}, err
	}
	return PendingFees{Amounts: amounts}, nil
}

func (p *PositionService) FetchPendingRewards(ctx context.Context, positionID string) ([]PendingReward, error) {
	values, err := p.client.View(ctx, p.PendingRewardsPayload(positionID))
	if err != nil {
		return nil, err
	}
	return decodePendingRewards(values)
}

func (p *PositionService) FetchPositionEmissionRates(ctx context.Context, positionID string) ([]RewardRate, error) {
	values, err := p.client.View(ctx, p.PositionEmissionRatesPayload(positionID))
	if err != nil {
		return nil, err
	}
	return decodeRewardRates(values)
}

func (p *PositionService) FetchPositionLiquidity(ctx context.Context, positionID string) (string, error) {
	values, err := p.client.View(ctx, p.PositionLiquidityPayload(positionID))
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "position liquidity")
}

func (p *PositionService) FetchPositionPoolInfo(ctx context.Context, positionID string) (PositionPoolInfo, error) {
	values, err := p.client.View(ctx, p.PositionPoolInfoPayload(positionID))
	if err != nil {
		return PositionPoolInfo{}, err
	}
	return decodePositionPoolInfo(values)
}

func (p *PositionService) FetchPositionTickRange(ctx context.Context, positionID string) (PositionTickRange, error) {
	values, err := p.client.View(ctx, p.PositionTickRangePayload(positionID))
	if err != nil {
		return PositionTickRange{}, err
	}
	return decodePositionTickRange(values)
}

func (p *PoolService) FetchOptimalLiquidityAmounts(ctx context.Context, args OptimalLiquidityAmountsArgs) (OptimalLiquidityAmounts, error) {
	payload, err := p.OptimalLiquidityAmountsPayload(args)
	if err != nil {
		return OptimalLiquidityAmounts{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return OptimalLiquidityAmounts{}, err
	}
	return decodeOptimalLiquidityAmounts(values)
}

func (p *PoolService) FetchOptimalLiquidityAmountsFromA(ctx context.Context, args EstCurrencyBAmountFromAArgs) (OptimalLiquidityFromA, error) {
	payload, err := p.EstCurrencyBAmountFromAPayload(args)
	if err != nil {
		return OptimalLiquidityFromA{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return OptimalLiquidityFromA{}, err
	}
	return decodeOptimalLiquidityFromA(values)
}

func (p *PoolService) FetchOptimalLiquidityAmountsFromB(ctx context.Context, args EstCurrencyAAmountFromBArgs) (OptimalLiquidityFromB, error) {
	payload, err := p.EstCurrencyAAmountFromBPayload(args)
	if err != nil {
		return OptimalLiquidityFromB{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return OptimalLiquidityFromB{}, err
	}
	return decodeOptimalLiquidityFromB(values)
}

func (p *PoolService) FetchLiquidityPoolAddressSafe(ctx context.Context, args PoolTokenPairArgs) (LiquidityPoolAddress, error) {
	payload, err := p.LiquidityPoolAddressSafePayload(args)
	if err != nil {
		return LiquidityPoolAddress{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return LiquidityPoolAddress{}, err
	}
	return decodeLiquidityPoolAddress(values)
}

func (p *PoolService) FetchLiquidityPoolExists(ctx context.Context, args PoolTokenPairArgs) (bool, error) {
	payload, err := p.LiquidityPoolExistsPayload(args)
	if err != nil {
		return false, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return false, err
	}
	return decodeSingleBoolResult(values, "liquidity pool exists")
}

func (p *PoolService) FetchPoolReserveAmount(ctx context.Context, poolID string) (PoolReserveAmount, error) {
	values, err := p.client.View(ctx, p.PoolReserveAmountPayload(poolID))
	if err != nil {
		return PoolReserveAmount{}, err
	}
	return decodePoolReserveAmount(values)
}

func (p *PoolService) FetchCurrentTickAndPrice(ctx context.Context, poolID string) (CurrentTickAndPrice, error) {
	values, err := p.client.View(ctx, p.CurrentTickAndPricePayload(poolID))
	if err != nil {
		return CurrentTickAndPrice{}, err
	}
	return decodeCurrentTickAndPrice(values)
}

func (p *PoolService) FetchCurrentPrice(ctx context.Context, args PoolTokenPairArgs) (string, error) {
	payload, err := p.CurrentPricePayload(args)
	if err != nil {
		return "", err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "current price")
}

func (p *PoolService) FetchPoolLiquidity(ctx context.Context, poolID string) (string, error) {
	values, err := p.client.View(ctx, p.PoolLiquidityPayload(poolID))
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "pool liquidity")
}

func (p *PoolService) FetchFeeRate(ctx context.Context, feeTierIndex string) (string, error) {
	payload, err := p.FeeRatePayload(feeTierIndex)
	if err != nil {
		return "", err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "fee rate")
}

func (p *PoolService) FetchSupportedInnerAssets(ctx context.Context, poolID string) ([]string, error) {
	values, err := p.client.View(ctx, p.SupportedInnerAssetsPayload(poolID))
	if err != nil {
		return nil, err
	}
	return decodeObjectAddressVectorResult(values, "supported inner assets")
}

func (s *SwapService) FetchBatchAmountOut(ctx context.Context, args BatchAmountOutArgs) (string, error) {
	payload, err := s.BatchAmountOutPayload(args)
	if err != nil {
		return "", err
	}
	values, err := s.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "batch amount out")
}

func (s *SwapService) FetchBatchAmountIn(ctx context.Context, args BatchAmountInArgs) (string, error) {
	payload, err := s.BatchAmountInPayload(args)
	if err != nil {
		return "", err
	}
	values, err := s.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "batch amount in")
}

func (p *PoolService) FetchPoolAmountOut(ctx context.Context, args PoolAmountOutArgs) (PoolQuoteResult, error) {
	payload, err := p.PoolAmountOutPayload(args)
	if err != nil {
		return PoolQuoteResult{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return PoolQuoteResult{}, err
	}
	return decodePoolQuoteResult(values)
}

func (p *PoolService) FetchPoolAmountIn(ctx context.Context, args PoolAmountInArgs) (PoolQuoteResult, error) {
	payload, err := p.PoolAmountInPayload(args)
	if err != nil {
		return PoolQuoteResult{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return PoolQuoteResult{}, err
	}
	return decodePoolQuoteResult(values)
}

func (p *PoolService) poolTokenPairPayload(function string, args PoolTokenPairArgs) (EntryFunctionPayload, error) {
	if err := requireMetadataPair(args.CurrencyA, args.CurrencyB); err != nil {
		return EntryFunctionPayload{}, err
	}
	feeTier, err := parseUint8(args.FeeTierIndex)
	if err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::" + function,
		TypeArguments:     []string{},
		FunctionArguments: []any{args.CurrencyA, args.CurrencyB, feeTier},
	}, nil
}

func (p *PoolService) poolObjectPayload(function, poolID string) EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::pool_v3::" + function,
		TypeArguments:     []string{},
		FunctionArguments: []any{poolID},
	}
}

func requireBatchQuoteArgs(route []string, amount, tokenIn, tokenOut, amountName string) error {
	if len(route) == 0 {
		return errors.New("poolRoute is required")
	}
	if err := requireMetadataPair(tokenIn, tokenOut); err != nil {
		return err
	}
	return requireRawIntegerString(amountName, amount)
}

func requirePoolQuoteArgs(poolID, token, amount, tokenName, amountName string) error {
	if err := requireNonEmpty("poolID", poolID); err != nil {
		return err
	}
	if err := requireNonEmpty(tokenName, token); err != nil {
		return err
	}
	if err := requireMetadataAddress(tokenName, token); err != nil {
		return err
	}
	return requireRawIntegerString(amountName, amount)
}

func decodePositionTokenAmounts(values []any) (PositionTokenAmounts, error) {
	if err := requireViewLen(values, 2, "position token amounts"); err != nil {
		return PositionTokenAmounts{}, err
	}
	amountA, err := viewString(values[0], "amountA")
	if err != nil {
		return PositionTokenAmounts{}, err
	}
	amountB, err := viewString(values[1], "amountB")
	if err != nil {
		return PositionTokenAmounts{}, err
	}
	return PositionTokenAmounts{AmountA: amountA, AmountB: amountB}, nil
}

func decodePositionPoolInfo(values []any) (PositionPoolInfo, error) {
	if err := requireViewLen(values, 3, "position pool info"); err != nil {
		return PositionPoolInfo{}, err
	}
	tokenA, err := viewObjectAddress(values[0], "tokenA")
	if err != nil {
		return PositionPoolInfo{}, err
	}
	tokenB, err := viewObjectAddress(values[1], "tokenB")
	if err != nil {
		return PositionPoolInfo{}, err
	}
	feeTier, err := viewUint8(values[2], "feeTier")
	if err != nil {
		return PositionPoolInfo{}, err
	}
	return PositionPoolInfo{TokenA: tokenA, TokenB: tokenB, FeeTier: feeTier}, nil
}

func decodePositionTickRange(values []any) (PositionTickRange, error) {
	if err := requireViewLen(values, 2, "position tick range"); err != nil {
		return PositionTickRange{}, err
	}
	lower, err := viewSignedTick(values[0], "tickLower")
	if err != nil {
		return PositionTickRange{}, err
	}
	upper, err := viewSignedTick(values[1], "tickUpper")
	if err != nil {
		return PositionTickRange{}, err
	}
	return PositionTickRange{Lower: lower, Upper: upper}, nil
}

func decodePendingRewards(values []any) ([]PendingReward, error) {
	items, err := viewVectorResult(values, "pending rewards")
	if err != nil {
		return nil, err
	}
	out := make([]PendingReward, 0, len(items))
	for i, item := range items {
		fields, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("pending rewards[%d] = %#v, want object", i, item)
		}
		rewardFA, err := viewObjectAddress(fields["reward_fa"], fmt.Sprintf("pending rewards[%d].reward_fa", i))
		if err != nil {
			return nil, err
		}
		amount, err := viewString(fields["amount_owed"], fmt.Sprintf("pending rewards[%d].amount_owed", i))
		if err != nil {
			return nil, err
		}
		out = append(out, PendingReward{RewardFA: rewardFA, AmountOwed: amount})
	}
	return out, nil
}

func decodeRewardRates(values []any) ([]RewardRate, error) {
	items, err := viewVectorResult(values, "position emission rates")
	if err != nil {
		return nil, err
	}
	out := make([]RewardRate, 0, len(items))
	for i, item := range items {
		fields, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("position emission rates[%d] = %#v, want object", i, item)
		}
		rewardFA, err := viewObjectAddress(fields["reward_fa"], fmt.Sprintf("position emission rates[%d].reward_fa", i))
		if err != nil {
			return nil, err
		}
		rate, err := viewString(fields["rate"], fmt.Sprintf("position emission rates[%d].rate", i))
		if err != nil {
			return nil, err
		}
		out = append(out, RewardRate{RewardFA: rewardFA, Rate: rate})
	}
	return out, nil
}

func decodeOptimalLiquidityAmounts(values []any) (OptimalLiquidityAmounts, error) {
	if err := requireViewLen(values, 3, "optimal liquidity amounts"); err != nil {
		return OptimalLiquidityAmounts{}, err
	}
	liquidity, err := viewString(values[0], "liquidity")
	if err != nil {
		return OptimalLiquidityAmounts{}, err
	}
	amountA, err := viewString(values[1], "currencyAAmount")
	if err != nil {
		return OptimalLiquidityAmounts{}, err
	}
	amountB, err := viewString(values[2], "currencyBAmount")
	if err != nil {
		return OptimalLiquidityAmounts{}, err
	}
	return OptimalLiquidityAmounts{Liquidity: liquidity, CurrencyAAmount: amountA, CurrencyBAmount: amountB}, nil
}

func decodeOptimalLiquidityFromA(values []any) (OptimalLiquidityFromA, error) {
	if err := requireViewLen(values, 2, "optimal liquidity from A"); err != nil {
		return OptimalLiquidityFromA{}, err
	}
	liquidity, err := viewString(values[0], "liquidity")
	if err != nil {
		return OptimalLiquidityFromA{}, err
	}
	amountB, err := viewString(values[1], "currencyBAmount")
	if err != nil {
		return OptimalLiquidityFromA{}, err
	}
	return OptimalLiquidityFromA{Liquidity: liquidity, CurrencyBAmount: amountB}, nil
}

func decodeOptimalLiquidityFromB(values []any) (OptimalLiquidityFromB, error) {
	if err := requireViewLen(values, 2, "optimal liquidity from B"); err != nil {
		return OptimalLiquidityFromB{}, err
	}
	liquidity, err := viewString(values[0], "liquidity")
	if err != nil {
		return OptimalLiquidityFromB{}, err
	}
	amountA, err := viewString(values[1], "currencyAAmount")
	if err != nil {
		return OptimalLiquidityFromB{}, err
	}
	return OptimalLiquidityFromB{Liquidity: liquidity, CurrencyAAmount: amountA}, nil
}

func decodeLiquidityPoolAddress(values []any) (LiquidityPoolAddress, error) {
	if err := requireViewLen(values, 2, "liquidity pool address"); err != nil {
		return LiquidityPoolAddress{}, err
	}
	exists, err := viewBool(values[0], "exists")
	if err != nil {
		return LiquidityPoolAddress{}, err
	}
	address, err := viewObjectAddress(values[1], "address")
	if err != nil {
		return LiquidityPoolAddress{}, err
	}
	return LiquidityPoolAddress{Exists: exists, Address: address}, nil
}

func decodePoolReserveAmount(values []any) (PoolReserveAmount, error) {
	if err := requireViewLen(values, 2, "pool reserve amount"); err != nil {
		return PoolReserveAmount{}, err
	}
	amountA, err := viewString(values[0], "amountA")
	if err != nil {
		return PoolReserveAmount{}, err
	}
	amountB, err := viewString(values[1], "amountB")
	if err != nil {
		return PoolReserveAmount{}, err
	}
	return PoolReserveAmount{AmountA: amountA, AmountB: amountB}, nil
}

func decodeCurrentTickAndPrice(values []any) (CurrentTickAndPrice, error) {
	if err := requireViewLen(values, 2, "current tick and price"); err != nil {
		return CurrentTickAndPrice{}, err
	}
	tick, err := viewSignedTick(values[0], "tick")
	if err != nil {
		return CurrentTickAndPrice{}, err
	}
	price, err := viewString(values[1], "price")
	if err != nil {
		return CurrentTickAndPrice{}, err
	}
	return CurrentTickAndPrice{Tick: tick, Price: price}, nil
}

func decodePoolQuoteResult(values []any) (PoolQuoteResult, error) {
	if err := requireViewLen(values, 2, "pool quote"); err != nil {
		return PoolQuoteResult{}, err
	}
	amount, err := viewString(values[0], "amount")
	if err != nil {
		return PoolQuoteResult{}, err
	}
	fee, err := viewString(values[1], "feeAmount")
	if err != nil {
		return PoolQuoteResult{}, err
	}
	return PoolQuoteResult{Amount: amount, FeeAmount: fee}, nil
}

func decodeSingleStringResult(values []any, name string) (string, error) {
	if err := requireViewLen(values, 1, name); err != nil {
		return "", err
	}
	return viewString(values[0], name)
}

func decodeSingleBoolResult(values []any, name string) (bool, error) {
	if err := requireViewLen(values, 1, name); err != nil {
		return false, err
	}
	return viewBool(values[0], name)
}

func decodeStringVectorResult(values []any, name string) ([]string, error) {
	items, err := viewVectorResult(values, name)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for i, item := range items {
		text, err := viewString(item, fmt.Sprintf("%s[%d]", name, i))
		if err != nil {
			return nil, err
		}
		out = append(out, text)
	}
	return out, nil
}

func decodeObjectAddressVectorResult(values []any, name string) ([]string, error) {
	items, err := viewVectorResult(values, name)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for i, item := range items {
		address, err := viewObjectAddress(item, fmt.Sprintf("%s[%d]", name, i))
		if err != nil {
			return nil, err
		}
		out = append(out, address)
	}
	return out, nil
}

func viewVectorResult(values []any, name string) ([]any, error) {
	if err := requireViewLen(values, 1, name); err != nil {
		return nil, err
	}
	items, ok := values[0].([]any)
	if !ok {
		return nil, fmt.Errorf("%s = %#v, want vector", name, values[0])
	}
	return items, nil
}

func requireViewLen(values []any, want int, name string) error {
	if len(values) != want {
		return fmt.Errorf("%s returned %d values, want %d", name, len(values), want)
	}
	return nil
}

func viewString(value any, name string) (string, error) {
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return "", fmt.Errorf("%s is empty", name)
		}
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case uint64:
		return strconv.FormatUint(typed, 10), nil
	case float64:
		if math.Trunc(typed) != typed {
			return "", fmt.Errorf("%s = %v, want integer", name, typed)
		}
		return strconv.FormatInt(int64(typed), 10), nil
	default:
		return "", fmt.Errorf("%s = %#v, want string", name, value)
	}
}

func viewBool(value any, name string) (bool, error) {
	valueBool, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s = %#v, want bool", name, value)
	}
	return valueBool, nil
}

func viewObjectAddress(value any, name string) (string, error) {
	if text, ok := value.(string); ok {
		if text == "" {
			return "", fmt.Errorf("%s is empty", name)
		}
		return text, nil
	}
	fields, ok := value.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%s = %#v, want object address", name, value)
	}
	inner, ok := fields["inner"]
	if !ok {
		return "", fmt.Errorf("%s missing inner address", name)
	}
	return viewString(inner, name+".inner")
}

func viewUint8(value any, name string) (uint8, error) {
	text, err := viewString(value, name)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(text, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return uint8(parsed), nil
}

func viewUint32(value any, name string) (uint32, error) {
	text, err := viewString(value, name)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return uint32(parsed), nil
}

func viewSignedTick(value any, name string) (SignedTick, error) {
	if fields, ok := value.(map[string]any); ok {
		bitsValue, ok := fields["bits"]
		if !ok {
			return SignedTick{}, fmt.Errorf("%s missing bits", name)
		}
		bits, err := viewUint32(bitsValue, name+".bits")
		if err != nil {
			return SignedTick{}, err
		}
		return SignedTick{Bits: bits, Tick: int32(bits)}, nil
	}
	bits, err := viewUint32(value, name)
	if err != nil {
		return SignedTick{}, err
	}
	return SignedTick{Bits: bits, Tick: int32(bits)}, nil
}
