package ipmax_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	ipmax "github.com/IPMaxxing/go-sdk"
)

type reply struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
	Raw     string            `json:"-"`
	Drop    bool              `json:"-"`
}

type recorded struct {
	Method string
	Path   string
	Header http.Header
	Body   string
}

type handler func(recorded) reply

type server struct {
	URL      string
	mu       sync.Mutex
	requests []recorded
}

func (s *server) recorded() []recorded {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]recorded(nil), s.requests...)
}

func fixtureNames(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, strings.TrimSuffix(filepath.Base(path), ".json"))
	}
	return names
}

func fixture(t *testing.T, name string) reply {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var r reply
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func fixtureServer(t *testing.T) handler {
	return func(request recorded) reply {
		product := request.Path[strings.LastIndex(request.Path, "/")+1:]
		if request.Method == http.MethodGet {
			return fixture(t, product)
		}
		var body ipmax.LookupRequest
		if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
			t.Error(err)
		}
		return fixture(t, product+"-"+strings.NewReplacer(".", "-", ":", "-").Replace(body.IP))
	}
}

func sequence(replies ...reply) handler {
	var mu sync.Mutex
	index := 0
	return func(recorded) reply {
		mu.Lock()
		defer mu.Unlock()
		r := replies[min(index, len(replies)-1)]
		index++
		return r
	}
}

func withData(t *testing.T, r reply, change func(data map[string]any)) reply {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	change(body["data"].(map[string]any))
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	r.Body = encoded
	return r
}

func serve(t *testing.T, h handler) *server {
	t.Helper()
	s := &server{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		request := recorded{Method: r.Method, Path: r.URL.Path, Header: r.Header.Clone(), Body: string(body)}
		s.mu.Lock()
		s.requests = append(s.requests, request)
		s.mu.Unlock()
		out := h(request)
		if out.Drop {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				conn.Close()
			}
			return
		}
		for name, value := range out.Headers {
			w.Header().Set(name, value)
		}
		payload := []byte(out.Raw)
		if out.Raw == "" {
			w.Header().Set("Content-Type", "application/json")
			payload = out.Body
		}
		w.WriteHeader(out.Status)
		_, _ = w.Write(payload)
	}))
	t.Cleanup(ts.Close)
	s.URL = ts.URL
	return s
}

func newClient(t *testing.T, h handler, opts ...ipmax.Option) (*ipmax.Client, *server) {
	t.Helper()
	s := serve(t, h)
	return ipmax.NewClient("ipmax_test_key", append([]ipmax.Option{ipmax.WithBaseURL(s.URL)}, opts...)...), s
}
