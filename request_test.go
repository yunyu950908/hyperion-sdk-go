package hyperion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestClientGetBuildsQueryAndDecodesJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/base/data/pools/stats" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("poolId"); got != "pool-1" {
			t.Fatalf("poolId query = %q", got)
		}
		if got := r.URL.Query().Get("skip"); got != "" {
			t.Fatalf("skip query = %q, want omitted", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"pool-1"}]}`))
	}))
	defer server.Close()

	req := NewRequestClient(server.URL+"/v1/graphql/", server.Client())
	var out struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}

	err := req.Get(context.Background(), "/base/data/pools/stats", QueryParams{
		"poolId": "pool-1",
		"skip":   nil,
	}, &out)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].ID != "pool-1" {
		t.Fatalf("decoded items = %#v", out.Items)
	}
}

func TestRequestClientGetReturnsStatusError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request body", http.StatusBadRequest)
	}))
	defer server.Close()

	req := NewRequestClient(server.URL, server.Client())
	var out map[string]any
	err := req.Get(context.Background(), "/broken", nil, &out)
	if err == nil {
		t.Fatal("Get returned nil error")
	}

	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("error type = %T, want *HTTPStatusError", err)
	}
	if statusErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", statusErr.StatusCode)
	}
	if statusErr.Path != "/broken" {
		t.Fatalf("path = %q", statusErr.Path)
	}
}
