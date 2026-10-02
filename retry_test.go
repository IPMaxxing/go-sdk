package ipmax_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	ipmax "github.com/IPMaxxing/go-sdk"
)

func unavailable(retryAfter string) reply {
	return reply{
		Status:  503,
		Headers: map[string]string{"Retry-After": retryAfter},
		Body:    []byte(`{"status":false,"data":null,"error":{"code":1400,"message":"Service unavailable","request_id":"req-503","retryable":true}}`),
	}
}

func idempotencyKeys(srv *server) map[string]int {
	keys := map[string]int{}
	for _, request := range srv.recorded() {
		keys[request.Header.Get("Idempotency-Key")]++
	}
	return keys
}

func TestRetryReusesIdempotencyKey(t *testing.T) {
	client, srv := newClient(t, sequence(unavailable("0"), fixture(t, "geoip-8-8-8-8")))
	geo, err := client.GeoIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if geo.IP != "8.8.8.8" {
		t.Fatalf("geo = %+v", geo)
	}
	keys := idempotencyKeys(srv)
	if len(srv.recorded()) != 2 || len(keys) != 1 {
		t.Fatalf("requests = %d, keys = %v", len(srv.recorded()), keys)
	}
}

func TestRetryAfterIsHonoured(t *testing.T) {
	client, _ := newClient(t, sequence(unavailable("0.15"), fixture(t, "catalog")))
	started := time.Now()
	if _, err := client.Catalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 150*time.Millisecond {
		t.Fatalf("elapsed = %v, want at least 150ms", elapsed)
	}
}

func TestRetriesStopAfterMax(t *testing.T) {
	client, srv := newClient(t, sequence(unavailable("0")), ipmax.WithMaxRetries(2))
	_, err := client.GeoIP(context.Background(), "8.8.8.8")
	if !errors.Is(err, ipmax.ErrServer) {
		t.Fatalf("err = %v", err)
	}
	if len(srv.recorded()) != 3 || len(idempotencyKeys(srv)) != 1 {
		t.Fatalf("requests = %d, keys = %v", len(srv.recorded()), idempotencyKeys(srv))
	}
}

func TestZeroRetries(t *testing.T) {
	client, srv := newClient(t, sequence(unavailable("0")), ipmax.WithMaxRetries(0))
	if _, err := client.Catalog(context.Background()); err == nil {
		t.Fatal("want an error")
	}
	if len(srv.recorded()) != 1 {
		t.Fatalf("requests = %d", len(srv.recorded()))
	}
}

func TestRetryableStatuses(t *testing.T) {
	for _, status := range []int{408, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			failure := unavailable("0")
			failure.Status = status
			client, srv := newClient(t, sequence(failure, fixture(t, "catalog")))
			if _, err := client.Catalog(context.Background()); err != nil {
				t.Fatal(err)
			}
			if len(srv.recorded()) != 2 {
				t.Fatalf("requests = %d", len(srv.recorded()))
			}
		})
	}
}

func TestClientErrorsAreNotRetried(t *testing.T) {
	client, srv := newClient(t, sequence(fixture(t, "error-not-found")))
	if _, err := client.GeoIP(context.Background(), "192.0.2.1"); !errors.Is(err, ipmax.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if len(srv.recorded()) != 1 {
		t.Fatalf("requests = %d", len(srv.recorded()))
	}
}

func TestDroppedConnectionIsRetried(t *testing.T) {
	client, srv := newClient(t, sequence(reply{Drop: true}, fixture(t, "geoip-8-8-8-8")), ipmax.WithMaxRetries(1))
	geo, err := client.GeoIP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if geo.Currency.Code != "USD" || len(idempotencyKeys(srv)) != 1 {
		t.Fatalf("geo = %+v, keys = %v", geo, idempotencyKeys(srv))
	}
}
