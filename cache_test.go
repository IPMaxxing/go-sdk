package ipmax_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	ipmax "github.com/IPMaxxing/go-sdk"
)

func lookedUp(t *testing.T, srv *server) []string {
	t.Helper()
	var ips []string
	for _, request := range srv.recorded() {
		var body ipmax.LookupRequest
		if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
			t.Fatal(err)
		}
		ips = append(ips, request.Path+" "+body.IP)
	}
	return ips
}

func TestRepeatedLookupIsCached(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	ctx := context.Background()
	first, err := client.GeoIP(ctx, "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.GeoIP(ctx, " 8.8.8.8 ")
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(srv.recorded()) != 1 {
		t.Fatalf("requests = %d", len(srv.recorded()))
	}
}

func TestProductsAreCachedSeparately(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	ctx := context.Background()
	for _, call := range []func() error{
		func() error { _, err := client.GeoIP(ctx, "8.8.8.8"); return err },
		func() error { _, err := client.Intelligence(ctx, "8.8.8.8"); return err },
		func() error { _, err := client.Intelligence(ctx, "8.8.8.8"); return err },
	} {
		if err := call(); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"/api/v1/geoip 8.8.8.8", "/api/v1/intelligence 8.8.8.8"}
	if got := lookedUp(t, srv); !slices.Equal(got, want) {
		t.Fatalf("lookups = %v", got)
	}
}

func TestCacheTTL(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t), ipmax.WithCacheTTL(100*time.Millisecond))
	ctx := context.Background()
	for _, wait := range []time.Duration{0, 0, 150 * time.Millisecond} {
		time.Sleep(wait)
		if _, err := client.GeoIP(ctx, "8.8.8.8"); err != nil {
			t.Fatal(err)
		}
	}
	if len(srv.recorded()) != 2 {
		t.Fatalf("requests = %d, want 2", len(srv.recorded()))
	}
}

func TestCacheEvictsLeastRecentlyUsed(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t), ipmax.WithCacheSize(2))
	ctx := context.Background()
	for _, ip := range []string{"8.8.8.8", "10.1.2.3", "8.8.8.8", ipv6, "8.8.8.8", "10.1.2.3"} {
		if _, err := client.GeoIP(ctx, ip); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"/api/v1/geoip 8.8.8.8", "/api/v1/geoip 10.1.2.3", "/api/v1/geoip " + ipv6, "/api/v1/geoip 10.1.2.3"}
	if got := lookedUp(t, srv); !slices.Equal(got, want) {
		t.Fatalf("lookups = %v", got)
	}
}

func TestCacheSizeZeroDisablesCache(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t), ipmax.WithCacheSize(0))
	for range 2 {
		if _, err := client.GeoIP(context.Background(), "8.8.8.8"); err != nil {
			t.Fatal(err)
		}
	}
	if len(srv.recorded()) != 2 {
		t.Fatalf("requests = %d, want 2", len(srv.recorded()))
	}
}

func TestErrorsAreNotCached(t *testing.T) {
	client, srv := newClient(t, sequence(fixture(t, "error-not-found"), fixture(t, "geoip-8-8-8-8")))
	ctx := context.Background()
	if _, err := client.GeoIP(ctx, "8.8.8.8"); !errors.Is(err, ipmax.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := client.GeoIP(ctx, "8.8.8.8"); err != nil {
		t.Fatal(err)
	}
	if len(srv.recorded()) != 2 {
		t.Fatalf("requests = %d, want 2", len(srv.recorded()))
	}
}

func TestExplicitIdempotencyKeyBypassesCacheRead(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	ctx := context.Background()
	steps := []func() error{
		func() error { _, err := client.GeoIP(ctx, "8.8.8.8"); return err },
		func() error {
			_, err := client.GeoIP(ctx, "8.8.8.8", ipmax.WithIdempotencyKey("order-2026-0001-geoip"))
			return err
		},
		func() error {
			_, err := client.Intelligence(ctx, "8.8.8.8", ipmax.WithIdempotencyKey("order-2026-0001-intel"))
			return err
		},
		func() error { _, err := client.Intelligence(ctx, "8.8.8.8"); return err },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(srv.recorded()) != 3 {
		t.Fatalf("requests = %d, want 3", len(srv.recorded()))
	}
}

func TestCatalogAndAccountAreNotCached(t *testing.T) {
	client, srv := newClient(t, fixtureServer(t))
	ctx := context.Background()
	for range 2 {
		if _, err := client.Catalog(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Account(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if len(srv.recorded()) != 4 {
		t.Fatalf("requests = %d, want 4", len(srv.recorded()))
	}
}

func TestConcurrentLookups(t *testing.T) {
	client, _ := newClient(t, fixtureServer(t), ipmax.WithCacheSize(2))
	ips := []string{"8.8.8.8", "10.1.2.3", ipv6}
	var wg sync.WaitGroup
	for i := range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.GeoIP(context.Background(), ips[i%len(ips)]); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
