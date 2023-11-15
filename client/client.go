package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
)

type cacheStore interface {
	Get(_ context.Context, key any) (any, error)
	GetWithTTL(ctx context.Context, key any) (any, time.Duration, error)
	Set(ctx context.Context, key any, value any, options ...store.Option) error
	Delete(_ context.Context, key any) error
	Invalidate(ctx context.Context, options ...store.InvalidateOption) error
	Clear(_ context.Context) error
	GetType() string
}

type TestResponse struct {
	Items []string `json:"items"`
}

type getTestCache struct {
	*cache.Cache[TestResponse]
	expiration time.Duration
}

type Client struct {
	logger       *slog.Logger
	httpCli      *http.Client
	getTestCache getTestCache
	baseURL      string
}

func New(logger *slog.Logger, httpCli *http.Client, cacheStore cacheStore, baseURL string) *Client {
	return &Client{
		logger:  logger,
		httpCli: httpCli,
		baseURL: baseURL,
		getTestCache: getTestCache{
			cache.New[TestResponse](cacheStore),
			time.Hour,
		},
	}
}

func (c *Client) GetTest(ctx context.Context) (*TestResponse, error) {
	cached, err := c.getTestCache.Get(ctx, "test")
	if err == nil {
		c.logger.Info("cache hit")
		return &cached, nil
	}

	u, err := url.JoinPath(c.baseURL, "/test")
	if err != nil {
		return nil, fmt.Errorf("could not join url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var testResp TestResponse
	if err := json.NewDecoder(resp.Body).Decode(&testResp); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}

	if err := c.getTestCache.Set(ctx, "test", testResp, store.WithExpiration(time.Hour)); err != nil {
		return nil, fmt.Errorf("could not set cache: %w", err)
	}

	return &testResp, nil
}
