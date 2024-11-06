package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const DEFAULT_TIMEOUT_DURATION = 60 * time.Second

func New(baseURL string) *Fetcher {
	httpClient := &http.Client{
		Timeout: DEFAULT_TIMEOUT_DURATION,
	}

	return &Fetcher{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (f *Fetcher) FetchRecentEvents(ctx context.Context, pageSize, offset int) ([]KillboardEvent, error) {
	url := fmt.Sprintf("%s/events?limit=%d&offset=%d&random=%s", f.baseURL, pageSize, offset, uuid.New())
	response, err := fetch[[]KillboardEvent](ctx, f, "GET", url)
	if err != nil {
		return nil, err
	}
	return *response, nil
}

func (f *Fetcher) FetchAllianceInfo(ctx context.Context, id string) (*Alliance, error) {
	url := fmt.Sprintf("%s/alliances/%s", f.baseURL, id)
	response, err := fetch[Alliance](ctx, f, "GET", url)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func fetch[T any](ctx context.Context, f *Fetcher, method, path string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	req.Header.Set("Pragma", "no-cache")

	httpResponse, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", httpResponse.Status)
	}

	response, err := parseResponse[T](httpResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, err
}

func parseResponse[T any](res *http.Response) (*T, error) {
	var response T

	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
