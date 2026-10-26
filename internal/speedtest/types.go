package speedtest

import "time"

// Phase indicates the current stage of a speed test run.
type Phase string

const (
	PhaseIdle     Phase = "idle"
	PhasePing     Phase = "ping"
	PhaseDownload Phase = "download"
	PhaseUpload   Phase = "upload"
	PhaseDone     Phase = "done"
	PhaseError    Phase = "error"
)

// ServerConfig defines the endpoints used for testing speed and latency.
type ServerConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	PingURL     string `json:"ping_url"`
	DownloadURL string `json:"download_url"`
	UploadURL   string `json:"upload_url"`
}

// Sample represents a live streaming telemetry update emitted during a test.
type Sample struct {
	Phase        Phase   `json:"phase"`
	Mbps         float64 `json:"mbps"`
	LatencyMs    float64 `json:"latency_ms"`
	JitterMs     float64 `json:"jitter_ms"`
	Progress     float64 `json:"progress"` // 0.0 to 1.0
	ElapsedMs    int64   `json:"elapsed_ms"`
	BytesTotal   int64   `json:"bytes_total"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// Result is the consolidated summary produced upon test completion.
type Result struct {
	ServerName   string    `json:"server_name"`
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	LatencyMs    float64   `json:"latency_ms"`
	JitterMs     float64   `json:"jitter_ms"`
	ViaProxy     bool      `json:"via_proxy"`
	CompletedAt  time.Time `json:"completed_at"`
}
