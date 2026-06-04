package hyperion

import "context"

// EstCurrencyAAmountFromB executes the pool estimate view for token A from a
// token B amount.
func (p *PoolService) EstCurrencyAAmountFromB(ctx context.Context, args EstCurrencyAAmountFromBArgs) ([]any, error) {
	payload, err := p.EstCurrencyAAmountFromBPayload(args)
	if err != nil {
		return nil, err
	}
	return p.client.View(ctx, payload)
}

// EstCurrencyBAmountFromA executes the pool estimate view for token B from a
// token A amount.
func (p *PoolService) EstCurrencyBAmountFromA(ctx context.Context, args EstCurrencyBAmountFromAArgs) ([]any, error) {
	payload, err := p.EstCurrencyBAmountFromAPayload(args)
	if err != nil {
		return nil, err
	}
	return p.client.View(ctx, payload)
}

// FetchTokensAmountByPositionID executes the position amount-by-liquidity view.
func (p *PositionService) FetchTokensAmountByPositionID(ctx context.Context, positionID string) ([]any, error) {
	return p.client.View(ctx, p.FetchTokensAmountByPositionIDPayload(positionID))
}

// FetchRewards executes the pending rewards view for a position.
func (r *RewardService) FetchRewards(ctx context.Context, positionID string) ([]any, error) {
	return r.client.View(ctx, r.FetchRewardsPayload(positionID))
}
