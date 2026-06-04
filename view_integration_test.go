package hyperion

import (
	"context"
	"testing"
	"time"
)

func TestIntegrationAptosViewExecutorTimestamp(t *testing.T) {
	cfg := loadIntegrationConfig(t)
	if cfg.AptosFullNodeURL == "" {
		t.Skip("set APTOS_FULLNODE_URL to run live Aptos view integration tests")
	}

	sdk := newIntegrationClient(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	values, err := sdk.View(ctx, EntryFunctionPayload{
		Function: "0x1::timestamp::now_seconds",
	})
	if err != nil {
		t.Fatalf("Client.View returned error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("timestamp view returned %d values, want 1", len(values))
	}
}
