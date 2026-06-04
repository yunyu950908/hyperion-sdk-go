package hyperion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAptosViewExecutorPostsRESTViewPayload(t *testing.T) {
	t.Parallel()

	payload := EntryFunctionPayload{
		Function:          "0x1::timestamp::now_seconds",
		TypeArguments:     []string{"0x1::aptos_coin::AptosCoin"},
		FunctionArguments: []any{"0x1"},
	}
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/view" {
			t.Fatalf("path = %s, want /v1/view", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer aptos-key" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content type = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_ = json.NewEncoder(w).Encode([]any{"123", "456"})
	}))
	defer server.Close()

	executor, err := NewAptosViewExecutor(server.URL, "aptos-key", server.Client())
	if err != nil {
		t.Fatalf("NewAptosViewExecutor returned error: %v", err)
	}
	values, err := executor.View(context.Background(), payload)
	if err != nil {
		t.Fatalf("View returned error: %v", err)
	}

	expectedBody := map[string]any{
		"function":       payload.Function,
		"type_arguments": []any{"0x1::aptos_coin::AptosCoin"},
		"arguments":      []any{"0x1"},
	}
	if !reflect.DeepEqual(captured, expectedBody) {
		t.Fatalf("request body = %#v, want %#v", captured, expectedBody)
	}
	assertArguments(t, values, []any{"123", "456"})
}

func TestAptosViewExecutorAcceptsVersionedFullNodeURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/view" {
			t.Fatalf("path = %s, want /v1/view", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]any{"ok"})
	}))
	defer server.Close()

	executor, err := NewAptosViewExecutor(server.URL+"/v1/", "", server.Client())
	if err != nil {
		t.Fatalf("NewAptosViewExecutor returned error: %v", err)
	}
	values, err := executor.View(context.Background(), EntryFunctionPayload{
		Function: "0x1::timestamp::now_seconds",
	})
	if err != nil {
		t.Fatalf("View returned error: %v", err)
	}
	assertArguments(t, values, []any{"ok"})
}

func TestAptosViewExecutorReportsHTTPStatusErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"bad view"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	executor, err := NewAptosViewExecutor(server.URL, "", server.Client())
	if err != nil {
		t.Fatalf("NewAptosViewExecutor returned error: %v", err)
	}
	_, err = executor.View(context.Background(), EntryFunctionPayload{
		Function: "0x1::timestamp::now_seconds",
	})
	if err == nil {
		t.Fatal("View returned nil error for HTTP 400")
	}

	var statusErr *ViewStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("error type = %T, want *ViewStatusError", err)
	}
	if statusErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status code = %d", statusErr.StatusCode)
	}
	if statusErr.Body == "" {
		t.Fatal("status error body is empty")
	}
}

func TestClientViewRequiresConfiguredExecutor(t *testing.T) {
	t.Parallel()

	sdk := newMainnetClientForPayloads(t)
	_, err := sdk.View(context.Background(), EntryFunctionPayload{
		Function: "0x1::timestamp::now_seconds",
	})
	if !errors.Is(err, ErrViewExecutorNotConfigured) {
		t.Fatalf("View error = %v, want ErrViewExecutorNotConfigured", err)
	}
}

func TestNewBuildsViewExecutorFromAptosFullNodeURL(t *testing.T) {
	t.Parallel()

	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode([]any{"ok"})
	}))
	defer server.Close()

	sdk, err := New(Options{
		Network:                    NetworkMainnet,
		ContractAddress:            MainnetContractAddress,
		HyperionFullNodeIndexerURL: server.URL,
		HyperionAPIHost:            server.URL,
		AptosFullNodeURL:           server.URL,
		AptosAPIKey:                "aptos-key",
		HTTPClient:                 server.Client(),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	values, err := sdk.View(context.Background(), EntryFunctionPayload{
		Function: "0x1::timestamp::now_seconds",
	})
	if err != nil {
		t.Fatalf("View returned error: %v", err)
	}
	if capturedAuth != "Bearer aptos-key" {
		t.Fatalf("authorization header = %q", capturedAuth)
	}
	assertArguments(t, values, []any{"ok"})
}
