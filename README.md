# ipmax-go

[![Go Reference](https://pkg.go.dev/badge/github.com/IPMaxxing/go-sdk.svg)](https://pkg.go.dev/github.com/IPMaxxing/go-sdk)

Go client for the [IP-Max](https://ipm.ax) GeoIP and IP intelligence API. Standard library only, Go 1.24+.

## Install

```sh
go get github.com/IPMaxxing/go-sdk
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	ipmax "github.com/IPMaxxing/go-sdk"
)

func main() {
	client := ipmax.NewClient("")
	ctx := context.Background()

	geo, err := client.GeoIP(ctx, "8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(geo.Geo.City, geo.Geo.Timezone)

	intel, err := client.Intelligence(ctx, "8.8.8.8", ipmax.WithIdempotencyKey("order-1042-intel"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(intel.NetworkClass.Primary)
}
```

An empty API key falls back to the `IPMAX_API_KEY` environment variable. `Catalog` works without a key; `Account`, `GeoIP` and `Intelligence` need one.

## Configuration

| Option                 | Default              | Description                                                           |
| ---------------------- | -------------------- | --------------------------------------------------------------------- |
| `WithBaseURL(string)`  | `https://api.ipm.ax` | API origin.                                                           |
| `WithTimeout(d)`       | `10s`                | Per-attempt timeout.                                                  |
| `WithMaxRetries(n)`    | `2`                  | Retries on network errors, 408, 429 and 5xx, honouring `Retry-After`. |
| `WithCacheSize(n)`     | `1024`               | Successful lookups kept in memory; `0` disables the cache.            |
| `WithCacheTTL(d)`      | `5m`                 | Cache lifetime.                                                       |
| `WithHTTPClient(c)`    | `http.DefaultClient` | Custom client for proxies, transports or tests.                       |

Each lookup carries a fresh UUID v4 `Idempotency-Key` that is reused on every retry, so a retried request is never billed twice. Cached lookups make no request and cost nothing; `WithIdempotencyKey` skips the cache read. The client is safe for concurrent use.

## Error handling

Non-2xx responses are returned as `*ipmax.APIError`, which also matches sentinel errors with `errors.Is`:

```go
_, err := client.GeoIP(ctx, "192.0.2.1")
switch {
case errors.Is(err, ipmax.ErrNotFound):
	fmt.Println("not in the database")
case errors.Is(err, ipmax.ErrInsufficientBalance):
	fmt.Println("top up the geoip wallet")
case err != nil:
	var apiErr *ipmax.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message, apiErr.RequestID, apiErr.RetryAfter)
	}
	log.Fatal(err)
}
```

| Sentinel                 | Status        |
| ------------------------ | ------------- |
| `ErrInvalidRequest`      | 400, 413, 415 |
| `ErrUnauthorized`        | 401           |
| `ErrInsufficientBalance` | 402           |
| `ErrNotFound`            | 404           |
| `ErrConflict`            | 409           |
| `ErrRateLimited`         | 429           |
| `ErrServer`              | 5xx           |

Transport failures and timeouts are returned as wrapped `net/http` and `context` errors.

## License

[Apache-2.0](LICENSE)
