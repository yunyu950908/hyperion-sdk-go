package hyperion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTypedPoolRESTMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/data/pools/stats":
			if poolID := r.URL.Query().Get("poolId"); poolID != "" {
				_, _ = w.Write([]byte(`{"items":[{"poolId":"` + poolID + `","feeTier":"3","token1":"0x1","token2":"0x2"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"poolId":"pool-all","feeTier":2,"token1":"0x1","token2":"0x2","volume24h":"123.45"}]}`))
		case "/base/data/pools/by-token-pair":
			if r.URL.Query().Get("token1") != "0x1" || r.URL.Query().Get("token2") != "0x2" || r.URL.Query().Get("feeTier") != "2" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"item":{"poolId":"pair-pool","feeTier":"2","token1":"0x1","token2":"0x2"}}`))
		case "/base/data/pools/pool-1/liquidity-accumulation":
			_, _ = w.Write([]byte(`{"items":[{"tick":-120,"liquidity":"42","liquidityNet":"-1","liquidityGross":"10"}]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	sdk := newTestClient(t, server)
	pools, err := sdk.Pool.FetchAllPoolsTyped(context.Background())
	if err != nil {
		t.Fatalf("FetchAllPoolsTyped returned error: %v", err)
	}
	if pools[0].PoolID != "pool-all" || pools[0].FeeTier != "2" || pools[0].Token1 != "0x1" || pools[0].Token2 != "0x2" {
		t.Fatalf("typed pool = %#v", pools[0])
	}

	pools, err = sdk.Pool.FetchPoolByIDTyped(context.Background(), "pool-1")
	if err != nil {
		t.Fatalf("FetchPoolByIDTyped returned error: %v", err)
	}
	if pools[0].PoolID != "pool-1" || pools[0].FeeTier != "3" {
		t.Fatalf("typed pool by id = %#v", pools[0])
	}

	pool, err := sdk.Pool.GetPoolByTokenPairAndFeeTierTyped(context.Background(), "0x1", "0x2", FeeTier03Spacing60)
	if err != nil {
		t.Fatalf("GetPoolByTokenPairAndFeeTierTyped returned error: %v", err)
	}
	if pool.PoolID != "pair-pool" || pool.FeeTier != "2" {
		t.Fatalf("typed pair pool = %#v", pool)
	}

	ticks, err := sdk.Pool.FetchTicksTyped(context.Background(), "pool-1")
	if err != nil {
		t.Fatalf("FetchTicksTyped returned error: %v", err)
	}
	if ticks[0].Tick != "-120" || ticks[0].Liquidity != "42" || ticks[0].LiquidityNet != "-1" || ticks[0].LiquidityGross != "10" {
		t.Fatalf("typed tick = %#v", ticks[0])
	}
}

func TestTypedPositionAndRewardRESTMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/data/positions":
			if r.URL.Query().Get("address") != "0xabc" {
				t.Fatalf("address query = %q", r.URL.Query().Get("address"))
			}
			_, _ = w.Write([]byte(`[{"objectId":"position-list","ownerAddress":"0xabc","poolId":"pool-1","liquidity":"100","tickLower":-60,"tickUpper":"60"}]`))
		case "/base/data/liquidity/ownerships":
			if r.URL.Query().Get("objectId") != "pos-1" || r.URL.Query().Get("ownerAddress") != "0xabc" {
				t.Fatalf("unexpected ownership query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"items":[{"objectId":"pos-1","ownerAddress":"0xabc","poolId":"pool-1","feeTier":"2"}]}`))
		case "/base/data/rewards/claimed-fees":
			_, _ = w.Write([]byte(`{"items":[{"objectId":"pos-1","amount":"0"},{"objectId":"pos-1","token":"0xfee","amount":"12","timestamp":1710000000,"txVersion":"9"}]}`))
		case "/base/data/rewards/claimed-farms":
			_, _ = w.Write([]byte(`{"items":[{"objectId":"pos-1","amount":0},{"objectId":"pos-1","rewardToken":"0xreward","amount":"7","timestamp":"1710000001","txVersion":10}]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	sdk := newTestClient(t, server)

	positions, err := sdk.Position.FetchAllPositionsByAddressTyped(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("FetchAllPositionsByAddressTyped returned error: %v", err)
	}
	if positions[0].ObjectID != "position-list" || positions[0].OwnerAddress != "0xabc" || positions[0].TickLower != "-60" {
		t.Fatalf("typed position list = %#v", positions[0])
	}

	positions, err = sdk.Position.FetchPositionByIDTyped(context.Background(), "pos-1", "0xabc")
	if err != nil {
		t.Fatalf("FetchPositionByIDTyped returned error: %v", err)
	}
	if positions[0].ObjectID != "pos-1" || positions[0].FeeTier != "2" {
		t.Fatalf("typed position = %#v", positions[0])
	}

	fees, err := sdk.Position.FetchFeeHistoryTyped(context.Background(), "pos-1", "0xabc")
	if err != nil {
		t.Fatalf("FetchFeeHistoryTyped returned error: %v", err)
	}
	if len(fees) != 1 || fees[0].Amount != "12" || fees[0].Token != "0xfee" || fees[0].Timestamp != "1710000000" {
		t.Fatalf("typed fees = %#v", fees)
	}

	rewards, err := sdk.Reward.FetchRewardHistoryTyped(context.Background(), "pos-1", "0xabc")
	if err != nil {
		t.Fatalf("FetchRewardHistoryTyped returned error: %v", err)
	}
	if len(rewards) != 1 || rewards[0].Amount != "7" || rewards[0].RewardToken != "0xreward" || rewards[0].TxVersion != "10" {
		t.Fatalf("typed rewards = %#v", rewards)
	}
}

func TestTypedSwapQuoteMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/base/rate/getSwapInfo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("amount") != "1000" || r.URL.Query().Get("from") != "0x1" || r.URL.Query().Get("to") != "0x2" {
			t.Fatalf("unexpected quote query: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("flag") == "in" {
			_, _ = w.Write([]byte(`{"amount":"1000","from":"0x1","to":"0x2","flag":"in","safeMode":"false","amountIn":"1010","amountOut":"1000","path":["pool-1"]}`))
			return
		}
		_, _ = w.Write([]byte(`{"amount":"1000","from":"0x1","to":"0x2","flag":"out","safeMode":true,"amountIn":"1000","amountOut":"995","fee":"1","path":["pool-1"]}`))
	}))
	defer server.Close()

	sdk := newTestClient(t, server)

	out, err := sdk.Swap.EstFromAmountTyped(context.Background(), EstimateAmountArgs{
		Amount:   "1000",
		From:     "0x1",
		To:       "0x2",
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("EstFromAmountTyped returned error: %v", err)
	}
	if out.Flag != "out" || !out.SafeMode || out.ResolvedAmountOut() != "995" || out.ResolvedAmountIn() != "1000" || len(out.Path) != 1 {
		t.Fatalf("typed out quote = %#v", out)
	}

	in, err := sdk.Swap.EstToAmountTyped(context.Background(), EstimateAmountArgs{
		Amount: "1000",
		From:   "0x1",
		To:     "0x2",
	})
	if err != nil {
		t.Fatalf("EstToAmountTyped returned error: %v", err)
	}
	if in.Flag != "in" || in.SafeMode || in.ResolvedAmountIn() != "1010" || in.ResolvedAmountOut() != "1000" {
		t.Fatalf("typed in quote = %#v", in)
	}
}

func TestTypedRESTFixturesDecodeStableFields(t *testing.T) {
	t.Parallel()

	server := newParityFixtureServer(t)
	defer server.Close()

	sdk, err := New(Options{
		Network:                    NetworkMainnet,
		ContractAddress:            MainnetContractAddress,
		HyperionFullNodeIndexerURL: server.URL,
		HyperionAPIHost:            server.URL,
		HTTPClient:                 server.Client(),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	pools, err := sdk.Pool.FetchAllPoolsTyped(context.Background())
	if err != nil {
		t.Fatalf("FetchAllPoolsTyped returned error: %v", err)
	}
	if len(pools) != 1 || pools[0].PoolID != "pool-1" || pools[0].FeeTier != "2" {
		t.Fatalf("typed parity pools = %#v", pools)
	}

	quote, err := sdk.Swap.EstFromAmountTyped(context.Background(), EstimateAmountArgs{
		Amount:   "1000",
		From:     "0x1",
		To:       "0x2",
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("EstFromAmountTyped returned error: %v", err)
	}
	if quote.Flag != "out" || quote.OutputAmount != "995" || quote.ResolvedAmountOut() != "995" || quote.SafeMode != true {
		t.Fatalf("typed parity quote = %#v", quote)
	}
}

func TestTypedRESTMethodsRejectMalformedFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"poolId":{"inner":"pool-1"}}]}`))
	}))
	defer server.Close()

	sdk := newTestClient(t, server)
	_, err := sdk.Pool.FetchAllPoolsTyped(context.Background())
	if err == nil {
		t.Fatal("FetchAllPoolsTyped accepted malformed poolId")
	}
	if !strings.Contains(err.Error(), "poolId") {
		t.Fatalf("malformed field error = %v", err)
	}
}
