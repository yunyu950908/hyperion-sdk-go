package hyperion

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegrationAptosViewExecutorTimestamp(t *testing.T) {
	if os.Getenv("HYPERION_INTEGRATION") != "1" {
		t.Skip("set HYPERION_INTEGRATION=1 to run live Aptos view integration tests")
	}
	fullNodeURL := os.Getenv("APTOS_FULLNODE_URL")
	if fullNodeURL == "" {
		t.Skip("set APTOS_FULLNODE_URL to run live Aptos view integration tests")
	}

	executor, err := NewAptosViewExecutor(fullNodeURL, os.Getenv("APTOS_API_KEY"), nil)
	if err != nil {
		t.Fatalf("NewAptosViewExecutor returned error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	values, err := executor.View(ctx, EntryFunctionPayload{
		Function: "0x1::timestamp::now_seconds",
	})
	if err != nil {
		t.Fatalf("View returned error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("timestamp view returned %d values, want 1", len(values))
	}
}
