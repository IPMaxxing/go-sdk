package ipmax_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	ipmax "github.com/IPMaxxing/go-sdk"
)

const ipv6 = "240e:390:a1:3cd0:be24:11ff:fe46:aca3"

var replays = map[string]func(*ipmax.Client) (any, error){
	"catalog":               func(c *ipmax.Client) (any, error) { return c.Catalog(context.Background()) },
	"account":               func(c *ipmax.Client) (any, error) { return c.Account(context.Background()) },
	"error-unauthorized":    func(c *ipmax.Client) (any, error) { return c.Account(context.Background()) },
	"error-invalid-request": func(c *ipmax.Client) (any, error) { return c.GeoIP(context.Background(), "not-an-ip") },
	"error-not-found":       func(c *ipmax.Client) (any, error) { return c.Intelligence(context.Background(), "192.0.2.1") },
	"geoip-8-8-8-8":         func(c *ipmax.Client) (any, error) { return c.GeoIP(context.Background(), "8.8.8.8") },
	"geoip-10-1-2-3":        func(c *ipmax.Client) (any, error) { return c.GeoIP(context.Background(), "10.1.2.3") },
	"geoip-240e-390-a1-3cd0-be24-11ff-fe46-aca3": func(c *ipmax.Client) (any, error) {
		return c.GeoIP(context.Background(), ipv6)
	},
	"intelligence-8-8-8-8":  func(c *ipmax.Client) (any, error) { return c.Intelligence(context.Background(), "8.8.8.8") },
	"intelligence-10-1-2-3": func(c *ipmax.Client) (any, error) { return c.Intelligence(context.Background(), "10.1.2.3") },
	"intelligence-240e-390-a1-3cd0-be24-11ff-fe46-aca3": func(c *ipmax.Client) (any, error) {
		return c.Intelligence(context.Background(), ipv6)
	},
	"intelligence-rich": func(c *ipmax.Client) (any, error) { return c.Intelligence(context.Background(), "8.8.8.8") },
}

func TestEveryFixtureHasAReplay(t *testing.T) {
	names := fixtureNames(t)
	keys := make([]string, 0, len(replays))
	for name := range replays {
		keys = append(keys, name)
	}
	slices.Sort(keys)
	if !slices.Equal(names, keys) {
		t.Fatalf("fixtures %v, replays %v", names, keys)
	}
}

func TestFixturesRoundTrip(t *testing.T) {
	for _, name := range fixtureNames(t) {
		t.Run(name, func(t *testing.T) {
			r := fixture(t, name)
			client, srv := newClient(t, sequence(r))
			got, err := replays[name](client)
			if len(srv.recorded()) != 1 {
				t.Fatalf("requests = %d, want 1", len(srv.recorded()))
			}
			if r.Status != 200 {
				var apiErr *ipmax.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != r.Status {
					t.Fatalf("err = %v, want APIError %d", err, r.Status)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(r.Body, &envelope); err != nil {
				t.Fatal(err)
			}
			assertSameJSON(t, envelope.Data, got)
		})
	}
}

func assertSameJSON(t *testing.T, want json.RawMessage, got any) {
	t.Helper()
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var left, right any
	if err := json.Unmarshal(want, &left); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &right); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(withoutNulls(left), withoutNulls(right)) {
		t.Fatalf("decoded value does not round-trip\nwant %s\ngot  %s", want, encoded)
	}
}

func withoutNulls(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if item != nil {
				out[key] = withoutNulls(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = withoutNulls(item)
		}
		return out
	}
	return value
}

func TestCatalogKeepsFractionalPrices(t *testing.T) {
	client, _ := newClient(t, fixtureServer(t))
	catalog, err := client.Catalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Currency != "CNY" || !catalog.Available {
		t.Fatalf("catalog = %+v", catalog)
	}
	if catalog.Prices[0].Product != ipmax.ProductGeoIP || catalog.Prices[0].UnitMicros != 437.5 {
		t.Fatalf("geoip price = %+v", catalog.Prices[0])
	}
	if catalog.Prices[1].UnitMicros != 1920 {
		t.Fatalf("intelligence price = %+v", catalog.Prices[1])
	}
}

func TestAccountKeepsFractionalMicros(t *testing.T) {
	r := withData(t, fixture(t, "account"), func(data map[string]any) {
		data["wallets"] = []any{map[string]any{"product": "geoip", "balanceMicros": 999562.5}}
		data["entries"] = []any{map[string]any{
			"id": "entry-1", "product": "geoip", "kind": "charge", "amountMicros": 437.5,
			"balanceMicros": 999562.5, "reference": "receipt-1", "createdAt": "2026-09-29T08:30:00.000Z",
		}}
	})
	client, _ := newClient(t, sequence(r))
	account, err := client.Account(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if account.KeyPrefix != "ipmax_test" || account.Wallets[0].BalanceMicros != 999562.5 {
		t.Fatalf("account = %+v", account)
	}
	if entry := account.Entries[0]; entry.AmountMicros != 437.5 || entry.Kind != "charge" {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestGeoIPDeepFields(t *testing.T) {
	client, _ := newClient(t, fixtureServer(t))
	ctx := context.Background()

	geo, err := client.GeoIP(ctx, "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if geo.As.ASN != "AS15169" || geo.Geo.CountryCode != "US" || *geo.Geo.Radius != 25 || *geo.PrefixLength != 24 {
		t.Fatalf("geo = %+v", geo)
	}
	if geo.Currency.Code != "USD" || geo.TimeZone.Abbr != "PDT" || !geo.TimeZone.IsDST || *geo.CallingCode != "1" {
		t.Fatalf("enrichment = %+v %+v", geo.Currency, geo.TimeZone)
	}

	bogon, err := client.GeoIP(ctx, "10.1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if !bogon.IsBogon || bogon.Currency != nil || bogon.TimeZone != nil || bogon.CallingCode != nil || bogon.Geo.CountryZh != "" {
		t.Fatalf("bogon = %+v", bogon)
	}

	v6, err := client.GeoIP(ctx, ipv6)
	if err != nil {
		t.Fatal(err)
	}
	if v6.Geo.CountryZh != "中国" || v6.Currency.Code != "CNY" || v6.TimeZone.Abbr != "GMT+8" {
		t.Fatalf("v6 = %+v", v6)
	}
}

func TestIntelligenceDeepFields(t *testing.T) {
	client, _ := newClient(t, fixtureServer(t))
	ctx := context.Background()

	intel, err := client.Intelligence(ctx, "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	connectivity := intel.As.Connectivity
	if len(connectivity.Exchanges) != 30 || connectivity.EstimatedCapacity.LowerGbps != 3000 || connectivity.EstimatedCapacity.UpperGbps != nil {
		t.Fatalf("connectivity = %+v", connectivity)
	}
	if *connectivity.Exchanges[0].Ports[0].RSPeer != true {
		t.Fatalf("port = %+v", connectivity.Exchanges[0].Ports[0])
	}
	if intel.Rpki.Status != ipmax.RpkiStatusValid || intel.Rpki.Source != ipmax.RpkiSourceRTRLive || intel.Rpki.Rtr.Source != "Cloudflare" {
		t.Fatalf("rpki = %+v", intel.Rpki)
	}
	if *intel.Threat.Scores.TrustScore != 70 || intel.NetworkClass.Primary != ipmax.NetworkClassHosting {
		t.Fatalf("threat = %+v", intel.Threat)
	}
	wantRevoked := []ipmax.RevokedNetworkTag{ipmax.RevokedNetworkTagBusinessBroadband, ipmax.RevokedNetworkTagHomeISP, ipmax.RevokedNetworkTagMobileISP}
	if !slices.Equal(intel.Intelligence.RevokedTags, wantRevoked) {
		t.Fatalf("revoked = %v", intel.Intelligence.RevokedTags)
	}

	bogon, err := client.Intelligence(ctx, "10.1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if bogon.Threat.IsTor != nil || bogon.Threat.Scores.VPNScore != nil || bogon.Network != nil || bogon.Intelligence.UsageType != nil {
		t.Fatalf("bogon = %+v", bogon)
	}
}

func TestIntelligenceRich(t *testing.T) {
	client, _ := newClient(t, sequence(fixture(t, "intelligence-rich")))
	intel, err := client.Intelligence(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	ptr := intel.Intelligence.Ptr
	if len(ptr) != 3 {
		t.Fatalf("ptr = %+v", ptr)
	}
	located := ptr[0].Located
	if located == nil || located.Location.City != "Tokyo" || *located.Location.RegionCode != "13" ||
		located.Hint.Kind != ipmax.PtrHintCLLI || located.Confidence.Score != 0.92 ||
		located.Evidence.Sources[0].Dataset != ipmax.PtrDatasetGeonames {
		t.Fatalf("located = %+v", located)
	}
	if mismatch := ptr[1].AsnMismatch; mismatch == nil || mismatch.ObservedASN != "AS64501" || ptr[1].Hostname != "core1.example.net" {
		t.Fatalf("mismatch = %+v", ptr[1])
	}
	if ptr[2].Unmatched == nil || ptr[2].Hostname != "dns.google" || ptr[2].Located != nil {
		t.Fatalf("unmatched = %+v", ptr[2])
	}
	transfer := intel.Intelligence.Transfers[0]
	if transfer.SourceHolder != "Level 3" || transfer.RecipientHolder != "Google LLC" {
		t.Fatalf("transfer = %+v", transfer)
	}
	if !slices.Equal(intel.Intelligence.SegmentProbe.Ports, []int{53, 443}) {
		t.Fatalf("segment probe = %+v", intel.Intelligence.SegmentProbe)
	}
	if intel.Intelligence.Abuse.Categories[0].Count != 20 || *intel.Intelligence.Cpes[0].Version != "1.25.3" {
		t.Fatalf("abuse = %+v", intel.Intelligence.Abuse)
	}
}

func TestUnknownEnumsAndFields(t *testing.T) {
	r := withData(t, fixture(t, "intelligence-rich"), func(data map[string]any) {
		data["future_field"] = map[string]any{"nested": true}
		class := data["network_class"].(map[string]any)
		class["primary"] = "orbital"
		class["source"] = "telepathy"
		data["rpki"].(map[string]any)["status"] = "pending"
		live := data["intelligence"].(map[string]any)
		live["tags"] = []any{"quantum_relay"}
		live["ptr"] = []any{map[string]any{"hostname": "edge.example", "status": "triangulated", "extra": 1}}
	})
	client, _ := newClient(t, sequence(r))
	intel, err := client.Intelligence(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if intel.NetworkClass.Primary != "orbital" || intel.NetworkClass.Source != "telepathy" || intel.Rpki.Status != "pending" {
		t.Fatalf("enums = %+v %+v", intel.NetworkClass, intel.Rpki)
	}
	if !slices.Equal(intel.Intelligence.Tags, []ipmax.LiveIntelligenceTag{"quantum_relay"}) {
		t.Fatalf("tags = %v", intel.Intelligence.Tags)
	}
	ptr := intel.Intelligence.Ptr[0]
	if ptr.Status != "triangulated" || ptr.Hostname != "edge.example" || ptr.Located != nil || ptr.AsnMismatch != nil || ptr.Unmatched != nil {
		t.Fatalf("ptr = %+v", ptr)
	}
}
