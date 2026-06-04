package hyperion

import (
	"context"
	"errors"
)

// AggregateToolContractAddress is the upstream aggregate helper contract address.
const AggregateToolContractAddress = "0xc3290d98d622b9ab354277e0b1c5e66552f9784d2d4fb473ff321b9338485117"

// AggregatorPartnerName is the default partnership identifier used by upstream aggregate swaps.
const AggregatorPartnerName = "hyperion-aggregator"

// AggregateSwapRouteArgs configures an aggregate swap route request.
type AggregateSwapRouteArgs struct {
	Amount   string
	From     string
	Input    string
	Slippage string
	To       string
}

// AggregateSwapInfoResult is the Hyperion aggregate route response.
type AggregateSwapInfoResult struct {
	FromToken          TokenAddressInfo `json:"fromToken"`
	ToToken            TokenAddressInfo `json:"toToken"`
	ExactIn            bool             `json:"exactIn"`
	FeeAmount          string           `json:"feeAmount"`
	FromTokenAmount    string           `json:"fromTokenAmount"`
	MinToTokenAmount   string           `json:"minToTokenAmount"`
	MaxFromTokenAmount string           `json:"maxFromTokenAmount"`
	ToTokenAmount      string           `json:"toTokenAmount"`
	Quotes             Quotes           `json:"quotes"`
}

// Quotes contains aggregate swap route and refund route splits.
type Quotes struct {
	Route       []AggregateSwapRoute `json:"route"`
	RefundRoute []AggregateSwapRoute `json:"refundRoute"`
}

// AggregateSwapRoute is one aggregate route split.
type AggregateSwapRoute struct {
	AmountIn   string                    `json:"amountIn"`
	AmountOut  string                    `json:"amountOut"`
	Percentage int                       `json:"percentage"`
	FeeAmount  string                    `json:"feeAmount"`
	RouteTaken []AggregateSwapRouteTaken `json:"routeTaken"`
}

// AggregateSwapRouteTaken is one DEX hop inside a route split.
type AggregateSwapRouteTaken struct {
	FromToken      TokenTypeInfo `json:"fromToken"`
	ToToken        TokenTypeInfo `json:"toToken"`
	DexName        string        `json:"dexName"`
	PoolID         string        `json:"poolId"`
	A2B            bool          `json:"a2b"`
	SqrtPriceLimit string        `json:"sqrtPriceLimit"`
	PoolType       string        `json:"poolType"`
	AmountIn       string        `json:"amountIn"`
	AmountOut      string        `json:"amountOut"`
	FirstType      string        `json:"firstType,omitempty"`
	SecondType     string        `json:"secondType,omitempty"`
	IsSell         bool          `json:"isSell,omitempty"`
	Integrator     string        `json:"integrator,omitempty"`
	IntegratorFee  int           `json:"integratorFee,omitempty"`
}

// TokenTypeInfo describes a token type returned inside route hops.
type TokenTypeInfo struct {
	TokenType string `json:"tokenType"`
}

// TokenAddressInfo describes a token address returned by aggregate route APIs.
type TokenAddressInfo struct {
	Address string `json:"address"`
}

// EstAmountByAggregateSwap fetches the aggregate swap route. Upstream supports
// this helper only on mainnet, so testnet clients return an error before making
// a network request.
func (s *SwapService) EstAmountByAggregateSwap(ctx context.Context, args AggregateSwapRouteArgs) (*AggregateSwapInfoResult, error) {
	if s.client.Options.Network != NetworkMainnet {
		return nil, errors.New("aggregate swap is only supported on MAINNET")
	}

	var out AggregateSwapInfoResult
	err := s.client.APIRequest.Get(ctx, "/base/aggregator/getAggRoute", QueryParams{
		"slippage": args.Slippage,
		"amount":   args.Amount,
		"from":     args.From,
		"input":    args.Input,
		"to":       args.To,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
