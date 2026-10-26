package speedtest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSpeedtestEngineWithMockServer(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Stream 512KB data
			chunk := make([]byte, 32*1024)
			for i := 0; i < 16; i++ {
				_, _ = w.Write(chunk)
			}
		case http.MethodPost:
			_, _ = io.Copy(io.Discard, r.Body)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockServer.Close()

	serverConfig := ServerConfig{
		ID:          "mock-server",
		Name:        "Local Mock Server",
		Location:    "Localhost",
		PingURL:     mockServer.URL,
		DownloadURL: mockServer.URL,
		UploadURL:   mockServer.URL,
	}

	engine := NewEngine(2, 500*time.Millisecond, nil)
	engine.WarmupDuration = 100 * time.Millisecond

	var sampleCount int
	res, err := engine.Run(context.Background(), serverConfig, func(s Sample) {
		sampleCount++
	})

	if err != nil {
		t.Fatalf("Speedtest engine failed: %v", err)
	}
	if res == nil {
		t.Fatalf("Expected non-nil result")
	}
	if sampleCount == 0 {
		t.Fatalf("Expected sample callbacks to be called")
	}
}

func TestSpeedtestEngineCancellation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	serverConfig := ServerConfig{
		ID:          "mock-server",
		Name:        "Local Mock Server",
		PingURL:     mockServer.URL,
		DownloadURL: mockServer.URL,
		UploadURL:   mockServer.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	engine := NewEngine(2, 3*time.Second, nil)

	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	_, err := engine.Run(ctx, serverConfig, func(s Sample) {})
	if err == nil {
		t.Fatalf("Expected error on canceled context, got nil")
	}
}
