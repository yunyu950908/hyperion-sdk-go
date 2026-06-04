package hyperion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// QueryParams contains optional REST query values. Nil values are omitted.
type QueryParams map[string]any

// RequestClient performs Hyperion REST API requests.
type RequestClient struct {
	apiHost    string
	httpClient *http.Client
}

// HTTPStatusError reports a non-2xx Hyperion REST response.
type HTTPStatusError struct {
	StatusCode int
	Status     string
	Path       string
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("hyperion API request failed: %s", e.Status)
}

// NewRequestClient creates a REST client using the provided API host.
func NewRequestClient(apiHost string, httpClient *http.Client) *RequestClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &RequestClient{
		apiHost:    normalizeAPIHost(apiHost),
		httpClient: httpClient,
	}
}

// APIHost returns the normalized Hyperion REST API host.
func (r *RequestClient) APIHost() string {
	return r.apiHost
}

// Get sends a GET request and decodes the JSON response into out.
func (r *RequestClient) Get(ctx context.Context, path string, params QueryParams, out any) error {
	requestURL, err := r.buildURL(path, params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Path:       path,
			Body:       string(body),
		}
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (r *RequestClient) buildURL(path string, params QueryParams) (string, error) {
	base, err := url.Parse(r.apiHost + "/")
	if err != nil {
		return "", err
	}
	rel, err := url.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return "", err
	}
	u := base.ResolveReference(rel)
	query := u.Query()
	for key, value := range params {
		if value == nil {
			continue
		}
		query.Set(key, fmt.Sprint(value))
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func normalizeAPIHost(rawURL string) string {
	rawURL = strings.TrimSuffix(rawURL, "/")
	rawURL = strings.TrimSuffix(rawURL, "/v1/graphql")
	return strings.TrimSuffix(rawURL, "/")
}
