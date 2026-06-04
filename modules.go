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
	var raw json.RawMessage
	if err := p.client.Request.Get(ctx, "/base/data/pools/stats", nil, &raw); err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchPoolByID returns pool stats filtered by pool ID.
func (p *PoolService) FetchPoolByID(ctx context.Context, poolID string) ([]JSONMap, error) {
	var raw json.RawMessage
	if err := p.client.Request.Get(ctx, "/base/data/pools/stats", QueryParams{"poolId": poolID}, &raw); err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// GetPoolByTokenPairAndFeeTier returns the pool matching a token pair and fee tier.
func (p *PoolService) GetPoolByTokenPairAndFeeTier(ctx context.Context, token1, token2 string, feeTier FeeTierIndex) (JSONMap, error) {
	var raw json.RawMessage
	err := p.client.Request.Get(ctx, "/base/data/pools/by-token-pair", QueryParams{
		"token1":  token1,
		"token2":  token2,
		"feeTier": int(feeTier),
	}, &raw)
	if err != nil {
		return nil, err
	}
	return decodeItem(raw)
}

// FetchTicks returns pool liquidity accumulation ticks.
func (p *PoolService) FetchTicks(ctx context.Context, poolID string) ([]JSONMap, error) {
	var raw json.RawMessage
	path := fmt.Sprintf("/base/data/pools/%s/liquidity-accumulation", poolID)
	if err := p.client.Request.Get(ctx, path, nil, &raw); err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchAllPositionsByAddress returns all positions owned by an address.
func (p *PositionService) FetchAllPositionsByAddress(ctx context.Context, address string) ([]JSONMap, error) {
	var raw json.RawMessage
	if err := p.client.Request.Get(ctx, "/base/data/positions", QueryParams{"address": address}, &raw); err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchPositionByID returns ownership information for a position and owner.
func (p *PositionService) FetchPositionByID(ctx context.Context, positionID, address string) ([]JSONMap, error) {
	var raw json.RawMessage
	err := p.client.Request.Get(ctx, "/base/data/liquidity/ownerships", QueryParams{
		"objectId":     positionID,
		"ownerAddress": address,
	}, &raw)
	if err != nil {
		return nil, err
	}
	return decodeItems(raw)
}

// FetchFeeHistory returns non-zero claimed fee history for a position.
func (p *PositionService) FetchFeeHistory(ctx context.Context, positionID, address string) ([]JSONMap, error) {
	var raw json.RawMessage
	err := p.client.Request.Get(ctx, "/base/data/rewards/claimed-fees", QueryParams{
		"objectId":     positionID,
		"ownerAddress": address,
	}, &raw)
	if err != nil {
		return nil, err
	}
	items, err := decodeItems(raw)
	if err != nil {
		return nil, err
	}
	return filterNonZeroAmount(items), nil
}

// FetchRewardHistory returns non-zero claimed farm reward history for a position.
func (r *RewardService) FetchRewardHistory(ctx context.Context, positionID, address string) ([]JSONMap, error) {
	var raw json.RawMessage
	err := r.client.Request.Get(ctx, "/base/data/rewards/claimed-farms", QueryParams{
		"objectId":     positionID,
		"ownerAddress": address,
	}, &raw)
	if err != nil {
		return nil, err
	}
	items, err := decodeItems(raw)
	if err != nil {
		return nil, err
	}
	return filterNonZeroAmount(items), nil
}

// EstFromAmount estimates output amount for an exact input swap.
func (s *SwapService) EstFromAmount(ctx context.Context, args EstimateAmountArgs) (JSONMap, error) {
	return s.estimateAmount(ctx, args, "out")
}

// EstToAmount estimates input amount for an exact output swap.
func (s *SwapService) EstToAmount(ctx context.Context, args EstimateAmountArgs) (JSONMap, error) {
	return s.estimateAmount(ctx, args, "in")
}

func (s *SwapService) estimateAmount(ctx context.Context, args EstimateAmountArgs, flag string) (JSONMap, error) {
	var out JSONMap
	err := s.client.APIRequest.Get(ctx, "/base/rate/getSwapInfo", QueryParams{
		"amount":   args.Amount,
		"from":     args.From,
		"to":       args.To,
		"safeMode": args.SafeMode,
		"flag":     flag,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
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
