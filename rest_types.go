package hyperion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// PoolStats contains the stable, high-value fields returned by pool stats REST
// endpoints. Hyperion may return additional stats fields; use the JSONMap
// methods when a caller needs fields not listed here.
type PoolStats struct {
	PoolID  string `json:"poolId"`
	FeeTier string `json:"feeTier"`
	Token1  string `json:"token1"`
	Token2  string `json:"token2"`
}

// PoolLiquidityTick contains selected liquidity-accumulation fields for a pool
// tick. All numeric values stay as strings to avoid precision loss.
type PoolLiquidityTick struct {
	Tick           string `json:"tick"`
	Liquidity      string `json:"liquidity"`
	LiquidityNet   string `json:"liquidityNet"`
	LiquidityGross string `json:"liquidityGross"`
}

// PositionInfo contains selected REST fields for position and ownership reads.
// Dynamic UI/indexer fields are intentionally left to the JSONMap methods.
type PositionInfo struct {
	ObjectID     string `json:"objectId"`
	OwnerAddress string `json:"ownerAddress"`
	PoolID       string `json:"poolId"`
	Token1       string `json:"token1"`
	Token2       string `json:"token2"`
	FeeTier      string `json:"feeTier"`
	TickLower    string `json:"tickLower"`
	TickUpper    string `json:"tickUpper"`
	Liquidity    string `json:"liquidity"`
}

// ClaimedFeeHistoryItem contains selected claimed-fee history fields.
type ClaimedFeeHistoryItem struct {
	ObjectID     string `json:"objectId"`
	OwnerAddress string `json:"ownerAddress"`
	Token        string `json:"token"`
	Amount       string `json:"amount"`
	Timestamp    string `json:"timestamp"`
	TxVersion    string `json:"txVersion"`
}

// ClaimedRewardHistoryItem contains selected claimed-farm reward history fields.
type ClaimedRewardHistoryItem struct {
	ObjectID     string `json:"objectId"`
	OwnerAddress string `json:"ownerAddress"`
	Token        string `json:"token"`
	RewardToken  string `json:"rewardToken"`
	Amount       string `json:"amount"`
	Timestamp    string `json:"timestamp"`
	TxVersion    string `json:"txVersion"`
}

// SwapQuote contains selected fields from `/base/rate/getSwapInfo`. The route
// path is present only when the REST API returns a direct pool route.
type SwapQuote struct {
	Amount       string   `json:"amount"`
	From         string   `json:"from"`
	To           string   `json:"to"`
	Flag         string   `json:"flag"`
	SafeMode     bool     `json:"safeMode"`
	AmountIn     string   `json:"amountIn"`
	AmountOut    string   `json:"amountOut"`
	InputAmount  string   `json:"inputAmount"`
	OutputAmount string   `json:"outputAmount"`
	Fee          string   `json:"fee"`
	FeeAmount    string   `json:"feeAmount"`
	Path         []string `json:"path"`
}

// ResolvedAmountIn returns the most specific input amount field exposed by the
// quote response.
func (q SwapQuote) ResolvedAmountIn() string {
	if q.AmountIn != "" {
		return q.AmountIn
	}
	return q.InputAmount
}

// ResolvedAmountOut returns the most specific output amount field exposed by
// the quote response.
func (q SwapQuote) ResolvedAmountOut() string {
	if q.AmountOut != "" {
		return q.AmountOut
	}
	return q.OutputAmount
}

func (p *PoolStats) UnmarshalJSON(data []byte) error {
	fields, err := decodeRESTObjectFields(data, "PoolStats")
	if err != nil {
		return err
	}
	if p.PoolID, err = optionalRESTStringField(fields, "poolId"); err != nil {
		return err
	}
	if p.FeeTier, err = optionalRESTStringField(fields, "feeTier"); err != nil {
		return err
	}
	if p.Token1, err = optionalRESTStringField(fields, "token1"); err != nil {
		return err
	}
	if p.Token2, err = optionalRESTStringField(fields, "token2"); err != nil {
		return err
	}
	return nil
}

func (t *PoolLiquidityTick) UnmarshalJSON(data []byte) error {
	fields, err := decodeRESTObjectFields(data, "PoolLiquidityTick")
	if err != nil {
		return err
	}
	if t.Tick, err = optionalRESTStringField(fields, "tick"); err != nil {
		return err
	}
	if t.Liquidity, err = optionalRESTStringField(fields, "liquidity"); err != nil {
		return err
	}
	if t.LiquidityNet, err = optionalRESTStringField(fields, "liquidityNet"); err != nil {
		return err
	}
	if t.LiquidityGross, err = optionalRESTStringField(fields, "liquidityGross"); err != nil {
		return err
	}
	return nil
}

func (p *PositionInfo) UnmarshalJSON(data []byte) error {
	fields, err := decodeRESTObjectFields(data, "PositionInfo")
	if err != nil {
		return err
	}
	if p.ObjectID, err = optionalRESTStringField(fields, "objectId"); err != nil {
		return err
	}
	if p.OwnerAddress, err = optionalRESTStringField(fields, "ownerAddress"); err != nil {
		return err
	}
	if p.PoolID, err = optionalRESTStringField(fields, "poolId"); err != nil {
		return err
	}
	if p.Token1, err = optionalRESTStringField(fields, "token1"); err != nil {
		return err
	}
	if p.Token2, err = optionalRESTStringField(fields, "token2"); err != nil {
		return err
	}
	if p.FeeTier, err = optionalRESTStringField(fields, "feeTier"); err != nil {
		return err
	}
	if p.TickLower, err = optionalRESTStringField(fields, "tickLower"); err != nil {
		return err
	}
	if p.TickUpper, err = optionalRESTStringField(fields, "tickUpper"); err != nil {
		return err
	}
	if p.Liquidity, err = optionalRESTStringField(fields, "liquidity"); err != nil {
		return err
	}
	return nil
}

func (h *ClaimedFeeHistoryItem) UnmarshalJSON(data []byte) error {
	fields, err := decodeRESTObjectFields(data, "ClaimedFeeHistoryItem")
	if err != nil {
		return err
	}
	if h.ObjectID, err = optionalRESTStringField(fields, "objectId"); err != nil {
		return err
	}
	if h.OwnerAddress, err = optionalRESTStringField(fields, "ownerAddress"); err != nil {
		return err
	}
	if h.Token, err = optionalRESTStringField(fields, "token"); err != nil {
		return err
	}
	if h.Amount, err = optionalRESTStringField(fields, "amount"); err != nil {
		return err
	}
	if h.Timestamp, err = optionalRESTStringField(fields, "timestamp"); err != nil {
		return err
	}
	if h.TxVersion, err = optionalRESTStringField(fields, "txVersion"); err != nil {
		return err
	}
	return nil
}

func (h *ClaimedRewardHistoryItem) UnmarshalJSON(data []byte) error {
	fields, err := decodeRESTObjectFields(data, "ClaimedRewardHistoryItem")
	if err != nil {
		return err
	}
	if h.ObjectID, err = optionalRESTStringField(fields, "objectId"); err != nil {
		return err
	}
	if h.OwnerAddress, err = optionalRESTStringField(fields, "ownerAddress"); err != nil {
		return err
	}
	if h.Token, err = optionalRESTStringField(fields, "token"); err != nil {
		return err
	}
	if h.RewardToken, err = optionalRESTStringField(fields, "rewardToken"); err != nil {
		return err
	}
	if h.Amount, err = optionalRESTStringField(fields, "amount"); err != nil {
		return err
	}
	if h.Timestamp, err = optionalRESTStringField(fields, "timestamp"); err != nil {
		return err
	}
	if h.TxVersion, err = optionalRESTStringField(fields, "txVersion"); err != nil {
		return err
	}
	return nil
}

func (q *SwapQuote) UnmarshalJSON(data []byte) error {
	fields, err := decodeRESTObjectFields(data, "SwapQuote")
	if err != nil {
		return err
	}
	if q.Amount, err = optionalRESTStringField(fields, "amount"); err != nil {
		return err
	}
	if q.From, err = optionalRESTStringField(fields, "from"); err != nil {
		return err
	}
	if q.To, err = optionalRESTStringField(fields, "to"); err != nil {
		return err
	}
	if q.Flag, err = optionalRESTStringField(fields, "flag"); err != nil {
		return err
	}
	if q.SafeMode, err = optionalRESTBoolField(fields, "safeMode"); err != nil {
		return err
	}
	if q.AmountIn, err = optionalRESTStringField(fields, "amountIn"); err != nil {
		return err
	}
	if q.AmountOut, err = optionalRESTStringField(fields, "amountOut"); err != nil {
		return err
	}
	if q.InputAmount, err = optionalRESTStringField(fields, "inputAmount"); err != nil {
		return err
	}
	if q.OutputAmount, err = optionalRESTStringField(fields, "outputAmount"); err != nil {
		return err
	}
	if q.Fee, err = optionalRESTStringField(fields, "fee"); err != nil {
		return err
	}
	if q.FeeAmount, err = optionalRESTStringField(fields, "feeAmount"); err != nil {
		return err
	}
	if q.Path, err = optionalRESTStringSliceField(fields, "path"); err != nil {
		return err
	}
	return nil
}

func decodeItemsAs[T any](raw json.RawMessage) ([]T, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("decode REST items: empty response")
	}
	if trimmed[0] == '[' {
		var direct []T
		if err := json.Unmarshal(trimmed, &direct); err != nil {
			return nil, fmt.Errorf("decode REST item array: %w", err)
		}
		return direct, nil
	}
	if trimmed[0] != '{' {
		return nil, fmt.Errorf("decode REST items: want object or array, got %q", string(trimmed[:1]))
	}

	var wrapped struct {
		Items []T `json:"items"`
	}
	if err := json.Unmarshal(trimmed, &wrapped); err != nil {
		return nil, fmt.Errorf("decode REST items wrapper: %w", err)
	}
	if wrapped.Items == nil {
		return []T{}, nil
	}
	return wrapped.Items, nil
}

func decodeItemAs[T any](raw json.RawMessage) (T, error) {
	var out T
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return out, fmt.Errorf("decode REST item: empty response")
	}
	if trimmed[0] != '{' {
		return out, fmt.Errorf("decode REST item: want object, got %q", string(trimmed[:1]))
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return out, fmt.Errorf("decode REST item wrapper: %w", err)
	}
	if item, ok := fields["item"]; ok {
		if err := json.Unmarshal(item, &out); err != nil {
			return out, fmt.Errorf("decode REST item.item: %w", err)
		}
		return out, nil
	}

	if err := json.Unmarshal(trimmed, &out); err != nil {
		return out, fmt.Errorf("decode REST item object: %w", err)
	}
	return out, nil
}

func filterNonZeroClaimedFees(items []ClaimedFeeHistoryItem) []ClaimedFeeHistoryItem {
	filtered := make([]ClaimedFeeHistoryItem, 0, len(items))
	for _, item := range items {
		if !amountIsZero(item.Amount) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterNonZeroClaimedRewards(items []ClaimedRewardHistoryItem) []ClaimedRewardHistoryItem {
	filtered := make([]ClaimedRewardHistoryItem, 0, len(items))
	for _, item := range items {
		if !amountIsZero(item.Amount) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func decodeRESTObjectFields(data []byte, typeName string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("decode %s: %w", typeName, err)
	}
	if fields == nil {
		return nil, fmt.Errorf("decode %s: want object", typeName)
	}
	return fields, nil
}

func optionalRESTStringField(fields map[string]json.RawMessage, name string) (string, error) {
	raw, ok := fields[name]
	if !ok || isJSONNull(raw) {
		return "", nil
	}
	value, err := decodeRESTScalar(raw)
	if err != nil {
		return "", fmt.Errorf("decode REST field %q: %w", name, err)
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("decode REST field %q: got %s, want string, number, bool, or null", name, previewJSON(raw))
	}
}

func optionalRESTBoolField(fields map[string]json.RawMessage, name string) (bool, error) {
	raw, ok := fields[name]
	if !ok || isJSONNull(raw) {
		return false, nil
	}
	value, err := decodeRESTScalar(raw)
	if err != nil {
		return false, fmt.Errorf("decode REST field %q: %w", name, err)
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("decode REST field %q: got %q, want bool", name, v)
		}
		return parsed, nil
	case json.Number:
		switch v.String() {
		case "0":
			return false, nil
		case "1":
			return true, nil
		default:
			return false, fmt.Errorf("decode REST field %q: got %s, want bool", name, v.String())
		}
	default:
		return false, fmt.Errorf("decode REST field %q: got %s, want bool", name, previewJSON(raw))
	}
}

func optionalRESTStringSliceField(fields map[string]json.RawMessage, name string) ([]string, error) {
	raw, ok := fields[name]
	if !ok || isJSONNull(raw) {
		return nil, nil
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("decode REST field %q: %w", name, err)
	}
	out := make([]string, 0, len(values))
	for i, value := range values {
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return nil, fmt.Errorf("decode REST field %q[%d]: got %s, want string", name, i, previewJSON(value))
		}
		out = append(out, text)
	}
	return out, nil
}

func decodeRESTScalar(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func previewJSON(raw json.RawMessage) string {
	const max = 48
	trimmed := string(bytes.TrimSpace(raw))
	if len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max] + "..."
}
