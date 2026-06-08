package hyperion

import (
	"context"
	"encoding/json"
	"fmt"
)

// JSONMap is a decoded Hyperion API object. The upstream TypeScript SDK returns
// untyped REST objects for these endpoints, so the Go SDK keeps the boundary
// flexible until Hyperion publishes stable response schemas.
type JSONMap map[string]any

// EstimateAmountArgs configures Hyperion swap quote requests.
type EstimateAmountArgs struct {
	Amount   string
	From     string
	To       string
	SafeMode bool
}

// FetchAllPools returns all pool stats from `/base/data/pools/stats`.
func (p *PoolService) FetchAllPools(ctx context.Context) ([]JSONMap, error) {
	raw, err := p.fetchPoolStatsRaw(ctx, nil)
	if err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchAllPoolsTyped returns typed pool stats for stable, high-value fields.
func (p *PoolService) FetchAllPoolsTyped(ctx context.Context) ([]PoolStats, error) {
	raw, err := p.fetchPoolStatsRaw(ctx, nil)
	if err != nil {
		return nil, err
	}
	return decodeItemsAs[PoolStats](raw)
}

// FetchPoolByID returns pool stats filtered by pool ID.
func (p *PoolService) FetchPoolByID(ctx context.Context, poolID string) ([]JSONMap, error) {
	raw, err := p.fetchPoolStatsRaw(ctx, QueryParams{"poolId": poolID})
	if err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchPoolByIDTyped returns typed pool stats filtered by pool ID.
func (p *PoolService) FetchPoolByIDTyped(ctx context.Context, poolID string) ([]PoolStats, error) {
	raw, err := p.fetchPoolStatsRaw(ctx, QueryParams{"poolId": poolID})
	if err != nil {
		return nil, err
	}
	return decodeItemsAs[PoolStats](raw)
}

// GetPoolByTokenPairAndFeeTier returns the pool matching a token pair and fee tier.
func (p *PoolService) GetPoolByTokenPairAndFeeTier(ctx context.Context, token1, token2 string, feeTier FeeTierIndex) (JSONMap, error) {
	raw, err := p.getPoolByTokenPairAndFeeTierRaw(ctx, token1, token2, feeTier)
	if err != nil {
		return nil, err
	}
	return decodeItem(raw)
}

// GetPoolByTokenPairAndFeeTierTyped returns the typed pool matching a token pair and fee tier.
func (p *PoolService) GetPoolByTokenPairAndFeeTierTyped(ctx context.Context, token1, token2 string, feeTier FeeTierIndex) (PoolStats, error) {
	raw, err := p.getPoolByTokenPairAndFeeTierRaw(ctx, token1, token2, feeTier)
	if err != nil {
		return PoolStats{}, err
	}
	return decodeItemAs[PoolStats](raw)
}

// FetchTicks returns pool liquidity accumulation ticks.
func (p *PoolService) FetchTicks(ctx context.Context, poolID string) ([]JSONMap, error) {
	raw, err := p.fetchTicksRaw(ctx, poolID)
	if err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchTicksTyped returns typed pool liquidity accumulation ticks.
func (p *PoolService) FetchTicksTyped(ctx context.Context, poolID string) ([]PoolLiquidityTick, error) {
	raw, err := p.fetchTicksRaw(ctx, poolID)
	if err != nil {
		return nil, err
	}
	return decodeItemsAs[PoolLiquidityTick](raw)
}

// FetchAllPositionsByAddress returns all positions owned by an address.
func (p *PositionService) FetchAllPositionsByAddress(ctx context.Context, address string) ([]JSONMap, error) {
	raw, err := p.fetchAllPositionsByAddressRaw(ctx, address)
	if err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchAllPositionsByAddressTyped returns typed positions owned by an address.
func (p *PositionService) FetchAllPositionsByAddressTyped(ctx context.Context, address string) ([]PositionInfo, error) {
	raw, err := p.fetchAllPositionsByAddressRaw(ctx, address)
	if err != nil {
		return nil, err
	}
	return decodeItemsAs[PositionInfo](raw)
}

// FetchPositionByID returns ownership information for a position and owner.
func (p *PositionService) FetchPositionByID(ctx context.Context, positionID, address string) ([]JSONMap, error) {
	raw, err := p.fetchPositionByIDRaw(ctx, positionID, address)
	if err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchPositionByIDTyped returns typed ownership information for a position and owner.
func (p *PositionService) FetchPositionByIDTyped(ctx context.Context, positionID, address string) ([]PositionInfo, error) {
	raw, err := p.fetchPositionByIDRaw(ctx, positionID, address)
	if err != nil {
		return nil, err
	}
	return decodeItemsAs[PositionInfo](raw)
}

// FetchFeeHistory returns non-zero claimed fee history for a position.
func (p *PositionService) FetchFeeHistory(ctx context.Context, positionID, address string) ([]JSONMap, error) {
	raw, err := p.fetchFeeHistoryRaw(ctx, positionID, address)
	if err != nil {
		return nil, err
	}
	items, err := decodeItems(raw)
	if err != nil {
		return nil, err
	}
	return filterNonZeroAmount(items), nil
}

// FetchFeeHistoryTyped returns typed non-zero claimed fee history for a position.
func (p *PositionService) FetchFeeHistoryTyped(ctx context.Context, positionID, address string) ([]ClaimedFeeHistoryItem, error) {
	raw, err := p.fetchFeeHistoryRaw(ctx, positionID, address)
	if err != nil {
		return nil, err
	}
	items, err := decodeItemsAs[ClaimedFeeHistoryItem](raw)
	if err != nil {
		return nil, err
	}
	return filterNonZeroClaimedFees(items), nil
}

// FetchRewardHistory returns non-zero claimed farm reward history for a position.
func (r *RewardService) FetchRewardHistory(ctx context.Context, positionID, address string) ([]JSONMap, error) {
	raw, err := r.fetchRewardHistoryRaw(ctx, positionID, address)
	if err != nil {
		return nil, err
	}
	items, err := decodeItems(raw)
	if err != nil {
		return nil, err
	}
	return filterNonZeroAmount(items), nil
}

// FetchRewardHistoryTyped returns typed non-zero claimed farm reward history for a position.
func (r *RewardService) FetchRewardHistoryTyped(ctx context.Context, positionID, address string) ([]ClaimedRewardHistoryItem, error) {
	raw, err := r.fetchRewardHistoryRaw(ctx, positionID, address)
	if err != nil {
		return nil, err
	}
	items, err := decodeItemsAs[ClaimedRewardHistoryItem](raw)
	if err != nil {
		return nil, err
	}
	return filterNonZeroClaimedRewards(items), nil
}

// EstFromAmount estimates output amount for an exact input swap.
func (s *SwapService) EstFromAmount(ctx context.Context, args EstimateAmountArgs) (JSONMap, error) {
	return s.estimateAmount(ctx, args, "out")
}

// EstFromAmountTyped estimates output amount for an exact input swap and decodes stable quote fields.
func (s *SwapService) EstFromAmountTyped(ctx context.Context, args EstimateAmountArgs) (SwapQuote, error) {
	return s.estimateAmountTyped(ctx, args, "out")
}

// EstToAmount estimates input amount for an exact output swap.
func (s *SwapService) EstToAmount(ctx context.Context, args EstimateAmountArgs) (JSONMap, error) {
	return s.estimateAmount(ctx, args, "in")
}

// EstToAmountTyped estimates input amount for an exact output swap and decodes stable quote fields.
func (s *SwapService) EstToAmountTyped(ctx context.Context, args EstimateAmountArgs) (SwapQuote, error) {
	return s.estimateAmountTyped(ctx, args, "in")
}

func (s *SwapService) estimateAmount(ctx context.Context, args EstimateAmountArgs, flag string) (JSONMap, error) {
	raw, err := s.estimateAmountRaw(ctx, args, flag)
	if err != nil {
		return nil, err
	}
	var out JSONMap
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SwapService) estimateAmountTyped(ctx context.Context, args EstimateAmountArgs, flag string) (SwapQuote, error) {
	raw, err := s.estimateAmountRaw(ctx, args, flag)
	if err != nil {
		return SwapQuote{}, err
	}
	var out SwapQuote
	if err := json.Unmarshal(raw, &out); err != nil {
		return SwapQuote{}, err
	}
	return out, nil
}

func (p *PoolService) fetchPoolStatsRaw(ctx context.Context, params QueryParams) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := p.client.Request.Get(ctx, "/base/data/pools/stats", params, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *PoolService) getPoolByTokenPairAndFeeTierRaw(ctx context.Context, token1, token2 string, feeTier FeeTierIndex) (json.RawMessage, error) {
	var raw json.RawMessage
	err := p.client.Request.Get(ctx, "/base/data/pools/by-token-pair", QueryParams{
		"token1":  token1,
		"token2":  token2,
		"feeTier": int(feeTier),
	}, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *PoolService) fetchTicksRaw(ctx context.Context, poolID string) (json.RawMessage, error) {
	var raw json.RawMessage
	path := fmt.Sprintf("/base/data/pools/%s/liquidity-accumulation", poolID)
	if err := p.client.Request.Get(ctx, path, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *PositionService) fetchAllPositionsByAddressRaw(ctx context.Context, address string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := p.client.Request.Get(ctx, "/base/data/positions", QueryParams{"address": address}, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *PositionService) fetchPositionByIDRaw(ctx context.Context, positionID, address string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := p.client.Request.Get(ctx, "/base/data/liquidity/ownerships", QueryParams{
		"objectId":     positionID,
		"ownerAddress": address,
	}, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *PositionService) fetchFeeHistoryRaw(ctx context.Context, positionID, address string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := p.client.Request.Get(ctx, "/base/data/rewards/claimed-fees", QueryParams{
		"objectId":     positionID,
		"ownerAddress": address,
	}, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (r *RewardService) fetchRewardHistoryRaw(ctx context.Context, positionID, address string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := r.client.Request.Get(ctx, "/base/data/rewards/claimed-farms", QueryParams{
		"objectId":     positionID,
		"ownerAddress": address,
	}, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *SwapService) estimateAmountRaw(ctx context.Context, args EstimateAmountArgs, flag string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := s.client.APIRequest.Get(ctx, "/base/rate/getSwapInfo", QueryParams{
		"amount":   args.Amount,
		"from":     args.From,
		"to":       args.To,
		"safeMode": args.SafeMode,
		"flag":     flag,
	}, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func decodeItems(raw json.RawMessage) ([]JSONMap, error) {
	var direct []JSONMap
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, nil
	}

	var wrapped struct {
		Items []JSONMap `json:"items"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Items == nil {
		return []JSONMap{}, nil
	}
	return wrapped.Items, nil
}

func decodeItem(raw json.RawMessage) (JSONMap, error) {
	var wrapped struct {
		Item JSONMap `json:"item"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Item != nil {
		return wrapped.Item, nil
	}

	var direct JSONMap
	if err := json.Unmarshal(raw, &direct); err != nil {
		return nil, err
	}
	return direct, nil
}

func filterNonZeroAmount(items []JSONMap) []JSONMap {
	filtered := make([]JSONMap, 0, len(items))
	for _, item := range items {
		if !amountIsZero(item["amount"]) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func amountIsZero(value any) bool {
	if value == nil {
		return false
	}
	rat, err := parseRat(fmt.Sprint(value))
	if err != nil {
		return false
	}
	return rat.Sign() == 0
}
