package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
	"jalin/internal/store"
)

// DefaultHealthCheckURL is used to test reachability through the proxy.
const DefaultHealthCheckURL = "https://1.1.1.1/cdn-cgi/trace"

// TransportForProfile returns an http.RoundTripper configured with the given proxy profile.
// If profile is nil or not active, it returns a clean default transport.
func TransportForProfile(p *store.ProxyProfile) (http.RoundTripper, error) {
	if p == nil {
		return &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}, nil
	}

	proto := strings.ToLower(strings.TrimSpace(p.Type))
	hostPort := net.JoinHostPort(p.Host, strconv.Itoa(p.Port))

	switch proto {
	case "socks5":
		var auth *proxy.Auth
		if p.Username != "" || p.Password != "" {
			auth = &proxy.Auth{
				User:     p.Username,
				Password: p.Password,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", hostPort, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("create socks5 dialer: %w", err)
		}

		contextDialer, ok := dialer.(proxy.ContextDialer)
		var dialContext func(ctx context.Context, network, addr string) (net.Conn, error)
		if ok {
			dialContext = contextDialer.DialContext
		} else {
			dialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		}

		return &http.Transport{
			DialContext:           dialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}, nil

	case "http", "https":
		var userinfo *url.Userinfo
		if p.Username != "" || p.Password != "" {
			userinfo = url.UserPassword(p.Username, p.Password)
		}
		proxyURL := &url.URL{
			Scheme: "http",
			Host:   hostPort,
			User:   userinfo,
		}
		return &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy protocol: %s", p.Type)
	}
}

// ClientForProfile returns an http.Client configured with timeout and the proxy profile.
func ClientForProfile(p *store.ProxyProfile, timeout time.Duration) (*http.Client, error) {
	tr, err := TransportForProfile(p)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}, nil
}

// TestConnection tests whether a proxy profile is reachable and can transmit HTTP requests.
// It returns the round-trip latency and any error encountered.
func TestConnection(ctx context.Context, p *store.ProxyProfile) (time.Duration, error) {
	client, err := ClientForProfile(p, 10*time.Second)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DefaultHealthCheckURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Jalin/1.0 (Network Utility)")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusUnauthorized {
		return elapsed, fmt.Errorf("upstream responded with HTTP %d", resp.StatusCode)
	}

	return elapsed, nil
}
