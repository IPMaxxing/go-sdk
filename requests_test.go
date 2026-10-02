package ipmax_test

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	ipmax "github.com/IPMaxxing/go-sdk"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestLookupRequestShape(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	ctx := context.Background()
	if _, err := client.GeoIP(ctx, "  8.8.8.8 "); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Intelligence(ctx, "8.8.8.8"); err != nil {
		t.Fatal(err)
	}
	requests := srv.recorded()
	if requests[0].Path != "/api/v1/geoip" || requests[1].Path != "/api/v1/intelligence" {
		t.Fatalf("paths = %s, %s", requests[0].Path, requests[1].Path)
	}
	for _, request := range requests {
		if request.Method != http.MethodPost || request.Body != `{"ip":"8.8.8.8"}` {
			t.Fatalf("request = %s %q", request.Method, request.Body)
		}
		want := map[string]string{
			"Authorization": "Bearer ipmax_test_key",
			"Content-Type":  "application/json",
			"Accept":        "application/json",
			"User-Agent":    "ipmax-go/0.1.0",
		}
		for name, value := range want {
			if got := request.Header.Get(name); got != value {
				t.Fatalf("%s = %q, want %q", name, got, value)
			}
		}
		if key := request.Header.Get("Idempotency-Key"); !uuidV4.MatchString(key) {
			t.Fatalf("Idempotency-Key = %q is not a UUID v4", key)
		}
	}
	if requests[0].Header.Get("Idempotency-Key") == requests[1].Header.Get("Idempotency-Key") {
		t.Fatal("each lookup needs a fresh idempotency key")
	}
}

func TestExplicitIdempotencyKey(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	if _, err := client.GeoIP(context.Background(), "8.8.8.8", ipmax.WithIdempotencyKey("order-2026-0001-geoip")); err != nil {
		t.Fatal(err)
	}
	if got := srv.recorded()[0].Header.Get("Idempotency-Key"); got != "order-2026-0001-geoip" {
		t.Fatalf("Idempotency-Key = %q", got)
	}
}

func TestCatalogWithoutKey(t *testing.T) {
	t.Setenv("IPMAX_API_KEY", "")
	srv := serve(t, fixtureServer(t))
	client := ipmax.NewClient("", ipmax.WithBaseURL(srv.URL+"/"))
	if _, err := client.Catalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	request := srv.recorded()[0]
	if request.Method != http.MethodGet || request.Path != "/api/v1/catalog" || request.Body != "" {
		t.Fatalf("request = %+v", request)
	}
	for _, name := range []string{"Authorization", "Content-Type", "Idempotency-Key"} {
		if value := request.Header.Get(name); value != "" {
			t.Fatalf("%s = %q, want none", name, value)
		}
	}
}

func TestAccountSendsKey(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	if _, err := client.Account(context.Background()); err != nil {
		t.Fatal(err)
	}
	request := srv.recorded()[0]
	if request.Method != http.MethodGet || request.Path != "/api/v1/account" || request.Header.Get("Authorization") != "Bearer ipmax_test_key" {
		t.Fatalf("request = %+v", request)
	}
}

func TestAPIKeyFromEnvironment(t *testing.T) {
	t.Setenv("IPMAX_API_KEY", "ipmax_from_env")
	srv := serve(t, fixtureServer(t))
	client := ipmax.NewClient("", ipmax.WithBaseURL(srv.URL))
	if _, err := client.Account(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := srv.recorded()[0].Header.Get("Authorization"); got != "Bearer ipmax_from_env" {
		t.Fatalf("Authorization = %q", got)
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestInjectedHTTPClient(t *testing.T) {
	srv := serve(t, fixtureServer(t))
	var seen []string
	httpClient := &http.Client{Transport: roundTripper(func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.URL.String())
		r.URL.Host = srv.URL[len("http://"):]
		r.URL.Scheme = "http"
		return http.DefaultTransport.RoundTrip(r)
	})}
	client := ipmax.NewClient("ipmax_test_key", ipmax.WithHTTPClient(httpClient))
	geo, err := client.GeoIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if geo.IP != "8.8.8.8" || len(seen) != 1 || seen[0] != "https://api.ipm.ax/api/v1/geoip" {
		t.Fatalf("seen = %v", seen)
	}
}

func TestContextCancellationStopsRetries(t *testing.T) {
	client, srv := newClient(t, func(recorded) reply {
		time.Sleep(200 * time.Millisecond)
		return fixture(t, "catalog")
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := client.Catalog(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if n := len(srv.recorded()); n != 1 {
		t.Fatalf("requests = %d, want 1", n)
	}
}

func TestAttemptTimeout(t *testing.T) {
	client, srv := newClient(t, func(recorded) reply {
		time.Sleep(time.Second)
		return fixture(t, "catalog")
	}, ipmax.WithTimeout(100*time.Millisecond), ipmax.WithMaxRetries(1))
	_, err := client.Catalog(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want a deadline error", err)
	}
	if n := len(srv.recorded()); n != 2 {
		t.Fatalf("requests = %d, want 2", n)
	}
}

func TestSlowAttemptIsRetried(t *testing.T) {
	var calls atomic.Int32
	client, _ := newClient(t, func(recorded) reply {
		if calls.Add(1) == 1 {
			time.Sleep(time.Second)
		}
		return fixture(t, "catalog")
	}, ipmax.WithTimeout(150*time.Millisecond), ipmax.WithMaxRetries(1))
	catalog, err := client.Catalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Currency != "CNY" {
		t.Fatalf("catalog = %+v", catalog)
	}
}
