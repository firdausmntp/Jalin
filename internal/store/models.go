package store

import "time"

// SpeedResult represents a completed network speed test record.
type SpeedResult struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	LatencyMs    float64   `json:"latency_ms"`
	JitterMs     float64   `json:"jitter_ms"`
	Server       string    `json:"server"`
	ViaProxy     bool      `json:"via_proxy"`
	NetworkType  string    `json:"network_type"`
}

// ProxyProfile represents a saved proxy server configuration.
type ProxyProfile struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "http" or "socks5"
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// DailyUsage represents the accumulated data transfer volume for a specific day.
type DailyUsage struct {
	Date       string `json:"date"` // "YYYY-MM-DD"
	WiFiBytes  uint64 `json:"wifi_bytes"`
	OtherBytes uint64 `json:"other_bytes"`
	RXBytes    uint64 `json:"rx_bytes"`
	TXBytes    uint64 `json:"tx_bytes"`
	TotalBytes uint64 `json:"total_bytes"`
}
