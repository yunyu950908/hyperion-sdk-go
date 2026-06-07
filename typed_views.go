package hyperion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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

// PricePreviewArgs configures price_hub::get_price_preview.
type PricePreviewArgs struct {
	Asset  string
	Amount string
}

// PricePreview is price_hub::get_price_preview decoded output. The contract
// returns two source values without naming them, so the SDK preserves their
// positional meaning.
type PricePreview struct {
	First  string `json:"first"`
	Second string `json:"second"`
}

// AggPrice is price_hub::AggPrice decoded output.
type AggPrice struct {
	Price     string `json:"price"`
	Precision string `json:"precision"`
}

// PriceSourceComparison is price_hub::compare_two_source decoded output.
type PriceSourceComparison struct {
	First  AggPrice `json:"first"`
	Second AggPrice `json:"second"`
}

// RateLimitStatus is a remain/capacity/interval limiter tuple decoded from
// rate_limiter_check views. Values are raw on-chain integers.
type RateLimitStatus struct {
	Remain   string `json:"remain"`
	Capacity string `json:"capacity"`
	Interval string `json:"interval"`
}

// AssetRateLimitStatus is rate_limiter_check::LimiterNumber decoded output.
type AssetRateLimitStatus struct {
	Asset    string `json:"asset"`
	Remain   string `json:"remain"`
	Capacity string `json:"capacity"`
	Interval string `json:"interval"`
}

// UserAssetRateLimiterArgs configures rate_limiter_check::user_asset_rate_limiter.
type UserAssetRateLimiterArgs struct {
	User  string
	Asset string
}

// UserAssetRateLimiterBatchArgs configures rate_limiter_check::user_asset_rate_limiter_batch.
type UserAssetRateLimiterBatchArgs struct {
	User   string
	Assets []string
}

// PoolUPriceLimiterStatus is rate_limiter_check pool u-price limiter output.
type PoolUPriceLimiterStatus struct {
	Exists   bool   `json:"exists"`
	Remain   string `json:"remain"`
	Capacity string `json:"capacity"`
	Interval string `json:"interval"`
}

// GlobalUPriceLimiterStatus is rate_limiter_check::global_u_price_limiter decoded output.
type GlobalUPriceLimiterStatus struct {
	Exists   bool     `json:"exists"`
	Remain   string   `json:"remain"`
	Capacity string   `json:"capacity"`
	Interval string   `json:"interval"`
	Assets   []string `json:"assets"`
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

func (p *PriceHubService) PricePreviewPayload(args PricePreviewArgs) (EntryFunctionPayload, error) {
	if err := requireMetadataAddress("asset", args.Asset); err != nil {
		return EntryFunctionPayload{}, err
	}
	if err := requireRawIntegerString("amount", args.Amount); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::price_hub::get_price_preview",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.Asset, args.Amount},
	}, nil
}

func (p *PriceHubService) PriceSourceComparisonPayload(asset string) (EntryFunctionPayload, error) {
	if err := requireMetadataAddress("asset", asset); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::price_hub::compare_two_source",
		TypeArguments:     []string{},
		FunctionArguments: []any{asset},
	}, nil
}

func (p *PriceHubService) IsTokenInPriceHubPayload(asset string) (EntryFunctionPayload, error) {
	if err := requireMetadataAddress("asset", asset); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::price_hub::is_token_in_hub",
		TypeArguments:     []string{},
		FunctionArguments: []any{asset},
	}, nil
}

func (p *PriceHubService) TokenInHubListPayload() EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::price_hub::token_in_hub_list",
		TypeArguments:     []string{},
		FunctionArguments: []any{},
	}
}

func (p *PriceHubService) PriceHubFeedSourcePayload() EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          p.client.Options.ContractAddress + "::price_hub::feed_source",
		TypeArguments:     []string{},
		FunctionArguments: []any{},
	}
}

func (r *RateLimiterService) GlobalAssetRateLimiterPayload(asset string) (EntryFunctionPayload, error) {
	if err := requireMetadataAddress("asset", asset); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::global_asset_rate_limiter",
		TypeArguments:     []string{},
		FunctionArguments: []any{asset},
	}, nil
}

func (r *RateLimiterService) GlobalAssetRateLimiterBatchPayload(assets []string) (EntryFunctionPayload, error) {
	if err := requireMetadataAddressVector("assets", assets); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::global_asset_rate_limiter_batch",
		TypeArguments:     []string{},
		FunctionArguments: []any{assets},
	}, nil
}

func (r *RateLimiterService) UserAssetRateLimiterPayload(args UserAssetRateLimiterArgs) (EntryFunctionPayload, error) {
	if err := requireAddressLike("user", args.User); err != nil {
		return EntryFunctionPayload{}, err
	}
	if err := requireMetadataAddress("asset", args.Asset); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::user_asset_rate_limiter",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.User, args.Asset},
	}, nil
}

func (r *RateLimiterService) UserAssetRateLimiterBatchPayload(args UserAssetRateLimiterBatchArgs) (EntryFunctionPayload, error) {
	if err := requireAddressLike("user", args.User); err != nil {
		return EntryFunctionPayload{}, err
	}
	if err := requireMetadataAddressVector("assets", args.Assets); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::user_asset_rate_limiter_batch",
		TypeArguments:     []string{},
		FunctionArguments: []any{args.User, args.Assets},
	}, nil
}

func (r *RateLimiterService) PoolUPriceLimiterPayload(pool string) (EntryFunctionPayload, error) {
	if err := requireAddressLike("pool", pool); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::pool_u_price_limiter",
		TypeArguments:     []string{},
		FunctionArguments: []any{pool},
	}, nil
}

func (r *RateLimiterService) PoolUPriceLimiterBatchPayload(pools []string) (EntryFunctionPayload, error) {
	if err := requireAddressLikeVector("pools", pools); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::pool_u_price_limiter_batch",
		TypeArguments:     []string{},
		FunctionArguments: []any{pools},
	}, nil
}

func (r *RateLimiterService) GlobalUPriceLimiterPayload() EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          r.client.Options.ContractAddress + "::rate_limiter_check::global_u_price_limiter",
		TypeArguments:     []string{},
		FunctionArguments: []any{},
	}
}

func (c *CoinWrapperService) IsWrapperPayload(asset string) (EntryFunctionPayload, error) {
	return c.coinWrapperAssetPayload("is_wrapper", asset)
}

func (c *CoinWrapperService) OriginalAssetPayload(asset string) (EntryFunctionPayload, error) {
	return c.coinWrapperAssetPayload("get_original", asset)
}

func (c *CoinWrapperService) CoinTypePayload(asset string) (EntryFunctionPayload, error) {
	return c.coinWrapperAssetPayload("get_coin_type", asset)
}

func (c *CoinWrapperService) FormattedFungibleAssetPayload(asset string) (EntryFunctionPayload, error) {
	return c.coinWrapperAssetPayload("format_fungible_asset", asset)
}

func (c *CoinWrapperService) CoinWrapperSupportedPayload() EntryFunctionPayload {
	return EntryFunctionPayload{
		Function:          c.client.Options.ContractAddress + "::coin_wrapper::is_supported",
		TypeArguments:     []string{},
		FunctionArguments: []any{},
	}
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

func (p *PriceHubService) FetchPricePreview(ctx context.Context, args PricePreviewArgs) (PricePreview, error) {
	payload, err := p.PricePreviewPayload(args)
	if err != nil {
		return PricePreview{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return PricePreview{}, err
	}
	return decodePricePreview(values)
}

func (p *PriceHubService) FetchPriceSourceComparison(ctx context.Context, asset string) (PriceSourceComparison, error) {
	payload, err := p.PriceSourceComparisonPayload(asset)
	if err != nil {
		return PriceSourceComparison{}, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return PriceSourceComparison{}, err
	}
	return decodePriceSourceComparison(values)
}

func (p *PriceHubService) FetchIsTokenInPriceHub(ctx context.Context, asset string) (bool, error) {
	payload, err := p.IsTokenInPriceHubPayload(asset)
	if err != nil {
		return false, err
	}
	values, err := p.client.View(ctx, payload)
	if err != nil {
		return false, err
	}
	return decodeSingleBoolResult(values, "token in price hub")
}

func (p *PriceHubService) FetchTokenInHubList(ctx context.Context) ([]string, error) {
	values, err := p.client.View(ctx, p.TokenInHubListPayload())
	if err != nil {
		return nil, err
	}
	return decodeObjectAddressVectorResult(values, "token in hub list")
}

func (p *PriceHubService) FetchPriceHubFeedSource(ctx context.Context) (string, error) {
	values, err := p.client.View(ctx, p.PriceHubFeedSourcePayload())
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "price hub feed source")
}

func (r *RateLimiterService) FetchGlobalAssetRateLimiter(ctx context.Context, asset string) (RateLimitStatus, error) {
	payload, err := r.GlobalAssetRateLimiterPayload(asset)
	if err != nil {
		return RateLimitStatus{}, err
	}
	values, err := r.client.View(ctx, payload)
	if err != nil {
		return RateLimitStatus{}, err
	}
	return decodeRateLimitStatusResult(values, "global asset rate limiter")
}

func (r *RateLimiterService) FetchGlobalAssetRateLimiterBatch(ctx context.Context, assets []string) ([]AssetRateLimitStatus, error) {
	payload, err := r.GlobalAssetRateLimiterBatchPayload(assets)
	if err != nil {
		return nil, err
	}
	values, err := r.client.View(ctx, payload)
	if err != nil {
		return nil, err
	}
	return decodeAssetRateLimitStatusVector(values, "global asset rate limiter batch")
}

func (r *RateLimiterService) FetchUserAssetRateLimiter(ctx context.Context, args UserAssetRateLimiterArgs) (RateLimitStatus, error) {
	payload, err := r.UserAssetRateLimiterPayload(args)
	if err != nil {
		return RateLimitStatus{}, err
	}
	values, err := r.client.View(ctx, payload)
	if err != nil {
		return RateLimitStatus{}, err
	}
	return decodeRateLimitStatusResult(values, "user asset rate limiter")
}

func (r *RateLimiterService) FetchUserAssetRateLimiterBatch(ctx context.Context, args UserAssetRateLimiterBatchArgs) ([]AssetRateLimitStatus, error) {
	payload, err := r.UserAssetRateLimiterBatchPayload(args)
	if err != nil {
		return nil, err
	}
	values, err := r.client.View(ctx, payload)
	if err != nil {
		return nil, err
	}
	return decodeAssetRateLimitStatusVector(values, "user asset rate limiter batch")
}

func (r *RateLimiterService) FetchPoolUPriceLimiter(ctx context.Context, pool string) (PoolUPriceLimiterStatus, error) {
	payload, err := r.PoolUPriceLimiterPayload(pool)
	if err != nil {
		return PoolUPriceLimiterStatus{}, err
	}
	values, err := r.client.View(ctx, payload)
	if err != nil {
		return PoolUPriceLimiterStatus{}, err
	}
	return decodePoolUPriceLimiterStatusResult(values, "pool u price limiter")
}

func (r *RateLimiterService) FetchPoolUPriceLimiterBatch(ctx context.Context, pools []string) ([]PoolUPriceLimiterStatus, error) {
	payload, err := r.PoolUPriceLimiterBatchPayload(pools)
	if err != nil {
		return nil, err
	}
	values, err := r.client.View(ctx, payload)
	if err != nil {
		return nil, err
	}
	return decodePoolUPriceLimiterStatusVector(values, "pool u price limiter batch")
}

func (r *RateLimiterService) FetchGlobalUPriceLimiter(ctx context.Context) (GlobalUPriceLimiterStatus, error) {
	values, err := r.client.View(ctx, r.GlobalUPriceLimiterPayload())
	if err != nil {
		return GlobalUPriceLimiterStatus{}, err
	}
	return decodeGlobalUPriceLimiterStatus(values)
}

func (c *CoinWrapperService) FetchIsWrapper(ctx context.Context, asset string) (bool, error) {
	payload, err := c.IsWrapperPayload(asset)
	if err != nil {
		return false, err
	}
	values, err := c.client.View(ctx, payload)
	if err != nil {
		return false, err
	}
	return decodeSingleBoolResult(values, "is wrapper")
}

func (c *CoinWrapperService) FetchOriginalAsset(ctx context.Context, asset string) (string, error) {
	payload, err := c.OriginalAssetPayload(asset)
	if err != nil {
		return "", err
	}
	values, err := c.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "original asset")
}

func (c *CoinWrapperService) FetchCoinType(ctx context.Context, asset string) (string, error) {
	payload, err := c.CoinTypePayload(asset)
	if err != nil {
		return "", err
	}
	values, err := c.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "coin type")
}

func (c *CoinWrapperService) FetchFormattedFungibleAsset(ctx context.Context, asset string) (string, error) {
	payload, err := c.FormattedFungibleAssetPayload(asset)
	if err != nil {
		return "", err
	}
	values, err := c.client.View(ctx, payload)
	if err != nil {
		return "", err
	}
	return decodeSingleStringResult(values, "formatted fungible asset")
}

func (c *CoinWrapperService) FetchCoinWrapperSupported(ctx context.Context) (bool, error) {
	values, err := c.client.View(ctx, c.CoinWrapperSupportedPayload())
	if err != nil {
		return false, err
	}
	return decodeSingleBoolResult(values, "coin wrapper supported")
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

func (c *CoinWrapperService) coinWrapperAssetPayload(function, asset string) (EntryFunctionPayload, error) {
	if err := requireMetadataAddress("asset", asset); err != nil {
		return EntryFunctionPayload{}, err
	}
	return EntryFunctionPayload{
		Function:          c.client.Options.ContractAddress + "::coin_wrapper::" + function,
		TypeArguments:     []string{},
		FunctionArguments: []any{asset},
	}, nil
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

func requireMetadataAddressVector(name string, values []string) error {
	if len(values) == 0 {
		return errors.New(name + " is required")
	}
	for i, value := range values {
		if err := requireMetadataAddress(fmt.Sprintf("%s[%d]", name, i), value); err != nil {
			return err
		}
	}
	return nil
}

func requireAddressLikeVector(name string, values []string) error {
	if len(values) == 0 {
		return errors.New(name + " is required")
	}
	for i, value := range values {
		if err := requireAddressLike(fmt.Sprintf("%s[%d]", name, i), value); err != nil {
			return err
		}
	}
	return nil
}

func requireAddressLike(name, value string) error {
	if err := requireNonEmpty(name, value); err != nil {
		return err
	}
	if !strings.HasPrefix(value, "0x") || isCoinType(value) {
		return errors.New(name + " must be an Aptos address")
	}
	return nil
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

func decodePricePreview(values []any) (PricePreview, error) {
	if err := requireViewLen(values, 2, "price preview"); err != nil {
		return PricePreview{}, err
	}
	first, err := viewString(values[0], "price preview first")
	if err != nil {
		return PricePreview{}, err
	}
	second, err := viewString(values[1], "price preview second")
	if err != nil {
		return PricePreview{}, err
	}
	return PricePreview{First: first, Second: second}, nil
}

func decodePriceSourceComparison(values []any) (PriceSourceComparison, error) {
	if err := requireViewLen(values, 2, "price source comparison"); err != nil {
		return PriceSourceComparison{}, err
	}
	first, err := viewAggPrice(values[0], "price source comparison first")
	if err != nil {
		return PriceSourceComparison{}, err
	}
	second, err := viewAggPrice(values[1], "price source comparison second")
	if err != nil {
		return PriceSourceComparison{}, err
	}
	return PriceSourceComparison{First: first, Second: second}, nil
}

func decodeRateLimitStatusResult(values []any, name string) (RateLimitStatus, error) {
	if err := requireViewLen(values, 3, name); err != nil {
		return RateLimitStatus{}, err
	}
	return viewRateLimitStatus(values[0], values[1], values[2], name)
}

func decodeAssetRateLimitStatusVector(values []any, name string) ([]AssetRateLimitStatus, error) {
	items, err := viewVectorResult(values, name)
	if err != nil {
		return nil, err
	}
	out := make([]AssetRateLimitStatus, 0, len(items))
	for i, item := range items {
		decoded, err := viewAssetRateLimitStatus(item, fmt.Sprintf("%s[%d]", name, i))
		if err != nil {
			return nil, err
		}
		out = append(out, decoded)
	}
	return out, nil
}

func decodePoolUPriceLimiterStatusResult(values []any, name string) (PoolUPriceLimiterStatus, error) {
	if err := requireViewLen(values, 4, name); err != nil {
		return PoolUPriceLimiterStatus{}, err
	}
	return viewPoolUPriceLimiterStatus(values[0], values[1], values[2], values[3], name)
}

func decodePoolUPriceLimiterStatusVector(values []any, name string) ([]PoolUPriceLimiterStatus, error) {
	items, err := viewVectorResult(values, name)
	if err != nil {
		return nil, err
	}
	out := make([]PoolUPriceLimiterStatus, 0, len(items))
	for i, item := range items {
		decoded, err := viewPoolUPriceLimiterStatusObject(item, fmt.Sprintf("%s[%d]", name, i))
		if err != nil {
			return nil, err
		}
		out = append(out, decoded)
	}
	return out, nil
}

func decodeGlobalUPriceLimiterStatus(values []any) (GlobalUPriceLimiterStatus, error) {
	if err := requireViewLen(values, 5, "global u price limiter"); err != nil {
		return GlobalUPriceLimiterStatus{}, err
	}
	status, err := viewPoolUPriceLimiterStatus(values[0], values[1], values[2], values[3], "global u price limiter")
	if err != nil {
		return GlobalUPriceLimiterStatus{}, err
	}
	assets, err := viewObjectAddressVector(values[4], "global u price limiter assets")
	if err != nil {
		return GlobalUPriceLimiterStatus{}, err
	}
	return GlobalUPriceLimiterStatus{
		Exists:   status.Exists,
		Remain:   status.Remain,
		Capacity: status.Capacity,
		Interval: status.Interval,
		Assets:   assets,
	}, nil
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
	if err := requireViewLen(values, 1, name); err != nil {
		return nil, err
	}
	return viewObjectAddressVector(values[0], name)
}

func viewObjectAddressVector(value any, name string) ([]string, error) {
	items, err := viewVector(value, name)
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
	return viewVector(values[0], name)
}

func viewVector(value any, name string) ([]any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s = %#v, want vector", name, value)
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

func viewAggPrice(value any, name string) (AggPrice, error) {
	fields, ok := value.(map[string]any)
	if !ok {
		return AggPrice{}, fmt.Errorf("%s = %#v, want object", name, value)
	}
	price, err := viewString(fields["price"], name+".price")
	if err != nil {
		return AggPrice{}, err
	}
	precision, err := viewString(fields["precision"], name+".precision")
	if err != nil {
		return AggPrice{}, err
	}
	return AggPrice{Price: price, Precision: precision}, nil
}

func viewRateLimitStatus(remainValue, capacityValue, intervalValue any, name string) (RateLimitStatus, error) {
	remain, err := viewString(remainValue, name+".remain")
	if err != nil {
		return RateLimitStatus{}, err
	}
	capacity, err := viewString(capacityValue, name+".capacity")
	if err != nil {
		return RateLimitStatus{}, err
	}
	interval, err := viewString(intervalValue, name+".interval")
	if err != nil {
		return RateLimitStatus{}, err
	}
	return RateLimitStatus{Remain: remain, Capacity: capacity, Interval: interval}, nil
}

func viewAssetRateLimitStatus(value any, name string) (AssetRateLimitStatus, error) {
	fields, ok := value.(map[string]any)
	if !ok {
		return AssetRateLimitStatus{}, fmt.Errorf("%s = %#v, want object", name, value)
	}
	asset, err := viewObjectAddress(fields["asset"], name+".asset")
	if err != nil {
		return AssetRateLimitStatus{}, err
	}
	status, err := viewRateLimitStatus(fields["remain"], fields["capacity"], fields["interval"], name)
	if err != nil {
		return AssetRateLimitStatus{}, err
	}
	return AssetRateLimitStatus{
		Asset:    asset,
		Remain:   status.Remain,
		Capacity: status.Capacity,
		Interval: status.Interval,
	}, nil
}

func viewPoolUPriceLimiterStatus(existsValue, remainValue, capacityValue, intervalValue any, name string) (PoolUPriceLimiterStatus, error) {
	exists, err := viewBool(existsValue, name+".exists")
	if err != nil {
		return PoolUPriceLimiterStatus{}, err
	}
	status, err := viewRateLimitStatus(remainValue, capacityValue, intervalValue, name)
	if err != nil {
		return PoolUPriceLimiterStatus{}, err
	}
	return PoolUPriceLimiterStatus{
		Exists:   exists,
		Remain:   status.Remain,
		Capacity: status.Capacity,
		Interval: status.Interval,
	}, nil
}

func viewPoolUPriceLimiterStatusObject(value any, name string) (PoolUPriceLimiterStatus, error) {
	fields, ok := value.(map[string]any)
	if !ok {
		return PoolUPriceLimiterStatus{}, fmt.Errorf("%s = %#v, want object", name, value)
	}
	return viewPoolUPriceLimiterStatus(fields["exist"], fields["remain"], fields["capacity"], fields["interval"], name)
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
