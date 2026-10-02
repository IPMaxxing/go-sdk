package ipmax

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	mathrand "math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	version        = "0.1.0"
	userAgent      = "ipmax-go/" + version
	defaultBaseURL = "https://api.ipm.ax"
	retryBase      = 500 * time.Millisecond
	retryCap       = 8 * time.Second
)

type Client struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
	cache      *lruCache
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.timeout = timeout }
}

func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) { c.maxRetries = maxRetries }
}

func WithCacheSize(size int) Option {
	return func(c *Client) { c.cache.capacity = size }
}

func WithCacheTTL(ttl time.Duration) Option {
	return func(c *Client) { c.cache.ttl = ttl }
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

type CallOption func(*callOptions)

type callOptions struct {
	idempotencyKey string
}

func WithIdempotencyKey(key string) CallOption {
	return func(o *callOptions) { o.idempotencyKey = key }
}

func NewClient(apiKey string, opts ...Option) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("IPMAX_API_KEY")
	}
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		timeout:    10 * time.Second,
		maxRetries: 2,
		httpClient: http.DefaultClient,
		cache:      newLRUCache(1024, 5*time.Minute),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Catalog(ctx context.Context) (*Catalog, error) {
	var response CatalogResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/catalog", nil, "", &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) Account(ctx context.Context) (*Account, error) {
	var response AccountResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/account", nil, "", &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) GeoIP(ctx context.Context, ip string, opts ...CallOption) (*GeoIPData, error) {
	return lookup(ctx, c, ProductGeoIP, ip, opts, func(r *GeoIPResponse) *GeoIPData { return &r.Data })
}

func (c *Client) Intelligence(ctx context.Context, ip string, opts ...CallOption) (*IntelligenceData, error) {
	return lookup(ctx, c, ProductIntelligence, ip, opts, func(r *IntelligenceResponse) *IntelligenceData { return &r.Data })
}

func lookup[R, T any](ctx context.Context, c *Client, product Product, ip string, opts []CallOption, data func(*R) *T) (*T, error) {
	var call callOptions
	for _, opt := range opts {
		opt(&call)
	}
	ip = strings.TrimSpace(ip)
	key := string(product) + " " + ip
	if call.idempotencyKey == "" {
		if cached, ok := c.cache.get(key); ok {
			return cached.(*T), nil
		}
		call.idempotencyKey = newIdempotencyKey()
	}
	body, err := json.Marshal(LookupRequest{IP: ip})
	if err != nil {
		return nil, err
	}
	var response R
	if err := c.do(ctx, http.MethodPost, "/api/v1/"+string(product), body, call.idempotencyKey, &response); err != nil {
		return nil, err
	}
	result := data(&response)
	c.cache.set(key, result)
	return result, nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte, idempotencyKey string, out any) error {
	for attempt := 0; ; attempt++ {
		status, header, payload, err := c.send(ctx, method, path, body, idempotencyKey)
		if err == nil && status >= 200 && status < 300 {
			if err := json.Unmarshal(payload, out); err != nil {
				return fmt.Errorf("ipmax: decode %s response: %w", path, err)
			}
			return nil
		}
		wait := backoff(attempt)
		if err == nil {
			err = newAPIError(status, header, payload)
			if delay, ok := retryAfter(header); ok {
				wait = delay
			}
		} else {
			err = fmt.Errorf("ipmax: %s %s: %w", method, path, err)
		}
		if ctx.Err() != nil || attempt >= c.maxRetries || !retryable(err) {
			return err
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (c *Client) send(ctx context.Context, method, path string, body []byte, idempotencyKey string) (int, http.Header, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", userAgent)
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, nil, nil, err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return response.StatusCode, response.Header, payload, nil
}

func retryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return retryableStatus(apiErr.StatusCode)
	}
	return true
}

func backoff(attempt int) time.Duration {
	limit := retryCap
	if attempt < 4 {
		limit = retryBase << attempt
	}
	return mathrand.N(limit)
}

func retryAfter(header http.Header) (time.Duration, bool) {
	value := header.Get("Retry-After")
	if value == "" {
		return 0, false
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 {
		return 0, false
	}
	return time.Duration(seconds * float64(time.Second)), true
}

func newIdempotencyKey() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
