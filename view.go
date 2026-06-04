package hyperion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrViewExecutorNotConfigured is returned when Client.View is called without
// an injected ViewExecutor or Aptos fullnode URL.
var ErrViewExecutorNotConfigured = errors.New("aptos view executor is not configured")

// ViewExecutor executes Aptos view-function payloads.
type ViewExecutor interface {
	View(ctx context.Context, payload EntryFunctionPayload) ([]any, error)
}

// ViewStatusError reports a non-2xx Aptos fullnode view response.
type ViewStatusError struct {
	StatusCode int
	Status     string
	URL        string
	Body       string
}

func (e *ViewStatusError) Error() string {
	return fmt.Sprintf("aptos view request failed: %s", e.Status)
}

// AptosViewExecutor executes EntryFunctionPayload values against an Aptos
// fullnode REST API.
type AptosViewExecutor struct {
	fullNodeURL string
	aptosAPIKey string
	httpClient  *http.Client
}

// NewAptosViewExecutor creates a REST-backed Aptos view executor. The
// fullNodeURL may be either a host root or a versioned `/v1` REST base.
func NewAptosViewExecutor(fullNodeURL, aptosAPIKey string, httpClient *http.Client) (*AptosViewExecutor, error) {
	fullNodeURL = strings.TrimSpace(fullNodeURL)
	if fullNodeURL == "" {
		return nil, errors.New("aptos fullnode URL is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &AptosViewExecutor{
		fullNodeURL: normalizeAPIHost(fullNodeURL),
		aptosAPIKey: aptosAPIKey,
		httpClient:  httpClient,
	}, nil
}

type aptosViewRequest struct {
	Function      string   `json:"function"`
	TypeArguments []string `json:"type_arguments"`
	Arguments     []any    `json:"arguments"`
}

// View posts the payload to the configured Aptos fullnode `/v1/view` endpoint.
func (e *AptosViewExecutor) View(ctx context.Context, payload EntryFunctionPayload) ([]any, error) {
	requestBody := aptosViewRequest{
		Function:      payload.Function,
		TypeArguments: payload.TypeArguments,
		Arguments:     payload.FunctionArguments,
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		return nil, err
	}

	viewURL := e.viewURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, viewURL, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if e.aptosAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.aptosAPIKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, &ViewStatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			URL:        viewURL,
			Body:       string(body),
		}
	}

	var out []any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out == nil {
		return []any{}, nil
	}
	return out, nil
}

func (e *AptosViewExecutor) viewURL() string {
	base := strings.TrimRight(e.fullNodeURL, "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/view"
	}
	return base + "/v1/view"
}

// View executes an Aptos view-function payload through the configured executor.
func (c *Client) View(ctx context.Context, payload EntryFunctionPayload) ([]any, error) {
	if c.ViewClient == nil {
		return nil, ErrViewExecutorNotConfigured
	}
	return c.ViewClient.View(ctx, payload)
}
