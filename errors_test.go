package ipmax_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	ipmax "github.com/IPMaxxing/go-sdk"
)

func asAPIError(t *testing.T, err error) *ipmax.APIError {
	t.Helper()
	var apiErr *ipmax.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *ipmax.APIError", err)
	}
	return apiErr
}

func TestErrorFixtures(t *testing.T) {
	cases := []struct {
		fixture  string
		call     func(*ipmax.Client) error
		sentinel error
		status   int
		code     ipmax.ErrorCode
		message  string
		request  string
	}{
		{
			fixture:  "error-invalid-request",
			call:     func(c *ipmax.Client) error { _, err := c.GeoIP(context.Background(), "not-an-ip"); return err },
			sentinel: ipmax.ErrInvalidRequest,
			status:   400,
			code:     ipmax.ErrorCodeInvalidRequest,
			message:  "A valid IP and 16-120 character Idempotency-Key are required",
			request:  "ENT-9b178079-212d-4052-ae72-8edf564be6e1",
		},
		{
			fixture:  "error-not-found",
			call:     func(c *ipmax.Client) error { _, err := c.Intelligence(context.Background(), "192.0.2.1"); return err },
			sentinel: ipmax.ErrNotFound,
			status:   404,
			code:     ipmax.ErrorCodeIPNotFound,
			message:  "This IP address is not in the database",
			request:  "ENT-f1b619f1-4da9-48c6-bb41-8ac819ea863a",
		},
		{
			fixture:  "error-unauthorized",
			call:     func(c *ipmax.Client) error { _, err := c.Account(context.Background()); return err },
			sentinel: ipmax.ErrUnauthorized,
			status:   401,
			code:     ipmax.ErrorCodeInvalidAPIKey,
			message:  "Invalid API key",
			request:  "ENT-706ff971-cb85-4933-8b6a-c4e42e0ecce1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			client, _ := newClient(t, sequence(fixture(t, tc.fixture)))
			err := tc.call(client)
			apiErr := asAPIError(t, err)
			if apiErr.StatusCode != tc.status || apiErr.Code != tc.code || apiErr.Message != tc.message ||
				apiErr.RequestID != tc.request || apiErr.Retryable || apiErr.RetryAfter != 0 {
				t.Fatalf("apiErr = %+v", apiErr)
			}
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("errors.Is(%v, %v) = false", err, tc.sentinel)
			}
			if errors.Is(err, ipmax.ErrServer) {
				t.Fatal("client error matched ErrServer")
			}
		})
	}
}

func TestStatusSentinels(t *testing.T) {
	cases := map[int]error{
		400: ipmax.ErrInvalidRequest,
		401: ipmax.ErrUnauthorized,
		402: ipmax.ErrInsufficientBalance,
		403: nil,
		404: ipmax.ErrNotFound,
		409: ipmax.ErrConflict,
		413: ipmax.ErrInvalidRequest,
		415: ipmax.ErrInvalidRequest,
		429: ipmax.ErrRateLimited,
		500: ipmax.ErrServer,
		503: ipmax.ErrServer,
	}
	sentinels := []error{
		ipmax.ErrInvalidRequest, ipmax.ErrUnauthorized, ipmax.ErrInsufficientBalance, ipmax.ErrNotFound,
		ipmax.ErrConflict, ipmax.ErrRateLimited, ipmax.ErrServer,
	}
	for status, want := range cases {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			client, _ := newClient(t, sequence(reply{
				Status:  status,
				Headers: map[string]string{"Retry-After": "7"},
				Body:    []byte(`{"status":false,"data":null,"error":{"code":1999,"message":"boom","request_id":"req-1","retryable":true}}`),
			}), ipmax.WithMaxRetries(0))
			_, err := client.GeoIP(context.Background(), "8.8.8.8")
			apiErr := asAPIError(t, err)
			if apiErr.StatusCode != status || apiErr.Code != 1999 || !apiErr.Retryable || apiErr.RetryAfter != 7*time.Second {
				t.Fatalf("apiErr = %+v", apiErr)
			}
			for _, sentinel := range sentinels {
				if got := errors.Is(err, sentinel); got != (sentinel == want) {
					t.Fatalf("errors.Is(%v) = %v", sentinel, got)
				}
			}
		})
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	client, _ := newClient(t, sequence(reply{
		Status:  http.StatusBadGateway,
		Headers: map[string]string{"X-Request-Id": "edge-42"},
		Raw:     "<html>bad gateway</html>",
	}), ipmax.WithMaxRetries(0))
	_, err := client.Catalog(context.Background())
	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != 502 || apiErr.Code != 0 || apiErr.RequestID != "edge-42" || !apiErr.Retryable || apiErr.Message != "Bad Gateway" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
	if !errors.Is(err, ipmax.ErrServer) {
		t.Fatal("502 should match ErrServer")
	}
	if want := "ipmax: 502 Bad Gateway [request edge-42]"; err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestUnreachableServer(t *testing.T) {
	client := ipmax.NewClient("ipmax_test_key", ipmax.WithBaseURL("http://127.0.0.1:9"), ipmax.WithMaxRetries(0))
	_, err := client.Catalog(context.Background())
	var apiErr *ipmax.APIError
	if err == nil || errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want a transport error", err)
	}
}
