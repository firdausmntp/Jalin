package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"jalin/internal/store"
)

func TestDirectClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := ClientForProfile(nil, 5*time.Second)
	if err != nil {
		t.Fatalf("ClientForProfile failed: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("client.Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestHTTPProxyProfile(t *testing.T) {
	// Dummy proxy server
	hitProxy := false
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitProxy = true
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	u, err := url.Parse(proxyServer.URL)
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())

	profile := &store.ProxyProfile{
		Name: "Local HTTP Proxy",
		Type: "http",
		Host: u.Hostname(),
		Port: port,
	}

	client, err := ClientForProfile(profile, 5*time.Second)
	if err != nil {
		t.Fatalf("ClientForProfile failed: %v", err)
	}

	// Request through proxy
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/test", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do failed: %v", err)
	}
	defer resp.Body.Close()

	if !hitProxy {
		t.Fatalf("Expected proxy to be contacted")
	}
}
