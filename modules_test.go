package hyperion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPoolServiceFetchMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/data/pools/stats":
			if poolID := r.URL.Query().Get("poolId"); poolID != "" {
				_, _ = w.Write([]byte(`{"items":[{"poolId":"` + poolID + `"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"poolId":"pool-all"}]}`))
		case "/base/data/pools/by-token-pair":
			if r.URL.Query().Get("token1") != "0x1" || r.URL.Query().Get("token2") != "0x2" || r.URL.Query().Get("feeTier") != "2" {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"item":{"poolId":"pair-pool"}}`))
		case "/base/data/pools/pool-1/liquidity-accumulation":
			_, _ = w.Write([]byte(`{"items":[{"tick":"1"}]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	sdk := newTestClient(t, server)

	pools, err := sdk.Pool.FetchAllPools(context.Background())
	if err != nil {
		t.Fatalf("FetchAllPools returned error: %v", err)
	}
	if pools[0]["poolId"] != "pool-all" {
		t.Fatalf("pool id = %#v", pools[0]["poolId"])
	}

	pools, err = sdk.Pool.FetchPoolByID(context.Background(), "pool-1")
	if err != nil {
		t.Fatalf("FetchPoolByID returned error: %v", err)
	}
	if pools[0]["poolId"] != "pool-1" {
		t.Fatalf("pool id = %#v", pools[0]["poolId"])
	}

	pool, err := sdk.Pool.GetPoolByTokenPairAndFeeTier(context.Background(), "0x1", "0x2", FeeTier03Spacing60)
	if err != nil {
		t.Fatalf("GetPoolByTokenPairAndFeeTier returned error: %v", err)
	}
	if pool["poolId"] != "pair-pool" {
		t.Fatalf("pool id = %#v", pool["poolId"])
	}

	ticks, err := sdk.Pool.FetchTicks(context.Background(), "pool-1")
	if err != nil {
		t.Fatalf("FetchTicks returned error: %v", err)
	}
	if ticks[0]["tick"] != "1" {
		t.Fatalf("tick = %#v", ticks[0]["tick"])
	}
}

func TestPositionAndRewardHistoryFilterZeroAmounts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/base/data/positions":
			if r.URL.Query().Get("address") != "0xabc" {
				t.Fatalf("address query = %q", r.URL.Query().Get("address"))
			}
			_, _ = w.Write([]byte(`[{"objectId":"position-list"}]`))
		case "/base/data/liquidity/ownerships":
			if r.URL.Query().Get("objectId") != "pos-1" || r.URL.Query().Get("ownerAddress") != "0xabc" {
				t.Fatalf("unexpected ownership query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"items":[{"objectId":"pos-1"}]}`))
		case "/base/data/rewards/claimed-fees":
			_, _ = w.Write([]byte(`{"items":[{"amount":"0"},{"amount":"12"}]}`))
		case "/base/data/rewards/claimed-farms":
			_, _ = w.Write([]byte(`{"items":[{"amount":0},{"amount":"7"}]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	sdk := newTestClient(t, server)

	positions, err := sdk.Position.FetchAllPositionsByAddress(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("FetchAllPositionsByAddress returned error: %v", err)
	}
	if positions[0]["objectId"] != "position-list" {
		t.Fatalf("position id = %#v", positions[0]["objectId"])
	}

	positions, err = sdk.Position.FetchPositionByID(context.Background(), "pos-1", "0xabc")
	if err != nil {
		t.Fatalf("FetchPositionByID returned error: %v", err)
	}
	if positions[0]["objectId"] != "pos-1" {
		t.Fatalf("position id = %#v", positions[0]["objectId"])
	}

	fees, err := sdk.Position.FetchFeeHistory(context.Background(), "pos-1", "0xabc")
	if err != nil {
		t.Fatalf("FetchFeeHistory returned error: %v", err)
	}
	if len(fees) != 1 || fees[0]["amount"] != "12" {
		t.Fatalf("fee history = %#v", fees)
	}

	rewards, err := sdk.Reward.FetchRewardHistory(context.Background(), "pos-1", "0xabc")
	if err != nil {
		t.Fatalf("FetchRewardHistory returned error: %v", err)
	}
	if len(rewards) != 1 || rewards[0]["amount"] != "7" {
		t.Fatalf("reward history = %#v", rewards)
	}
}

func TestSwapServiceQuoteMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/base/rate/getSwapInfo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("amount") != "1000" || r.URL.Query().Get("from") != "0x1" || r.URL.Query().Get("to") != "0x2" {
			t.Fatalf("unexpected quote query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"flag":"` + r.URL.Query().Get("flag") + `","safeMode":"` + r.URL.Query().Get("safeMode") + `"}`))
	}))
	defer server.Close()

	sdk := newTestClient(t, server)

	out, err := sdk.Swap.EstFromAmount(context.Background(), EstimateAmountArgs{
		Amount:   "1000",
		From:     "0x1",
		To:       "0x2",
		SafeMode: true,
	})
	if err != nil {
		t.Fatalf("EstFromAmount returned error: %v", err)
	}
	if out["flag"] != "out" || out["safeMode"] != "true" {
		t.Fatalf("EstFromAmount response = %#v", out)
	}

	in, err := sdk.Swap.EstToAmount(context.Background(), EstimateAmountArgs{
		Amount: "1000",
		From:   "0x1",
		To:     "0x2",
	})
	if err != nil {
		t.Fatalf("EstToAmount returned error: %v", err)
	}
	if in["flag"] != "in" || in["safeMode"] != "false" {
		t.Fatalf("EstToAmount response = %#v", in)
	}
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	sdk, err := New(Options{
		Network:                    NetworkTestnet,
		ContractAddress:            TestnetContractAddress,
		HyperionFullNodeIndexerURL: server.URL,
		HyperionAPIHost:            server.URL,
		HTTPClient:                 server.Client(),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return sdk
}
