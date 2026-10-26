package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"jalin/internal/netstat"
	"jalin/internal/proxy"
	"jalin/internal/speedtest"
	"jalin/internal/store"
)

// App holds the complete reactive UI and application state.
type App struct {
	mu             sync.Mutex
	win            *mygo.Window
	miniWin        *mygo.Window
	db             *store.DB
	tray           *mygo.Tray
	traySpeedItem  *mygo.MenuItem
	trayUsageItem  *mygo.MenuItem
	lastTrayUpdate time.Time

	// Navigation
	SelectedNav string // "test", "monitor", "history", "proxy", "settings"

	// Live Network Traffic Monitor
	NetMonitor *netstat.Monitor
	LiveStats  netstat.TrafficStats

	// Speed Test State
	IsTesting        bool
	CancelFunc       context.CancelFunc
	CurrentPhase     speedtest.Phase
	CurrentMbps      float64
	LatencyMs        float64
	JitterMs         float64
	DownloadMbps     float64
	UploadMbps       float64
	Progress         float64
	StatusText       string
	SelectedServerID string
	ChartHistory     []float64
	LastResult       *store.SpeedResult

	// Proxy State
	Profiles         []store.ProxyProfile
	ActiveProxy      *store.ProxyProfile
	IsTestingProxy   bool
	ShowAddModal     bool
	NewProxyName     string
	NewProxyType     string // "http" or "socks5"
	NewProxyHost     string
	NewProxyPort     string
	NewProxyUser     string
	NewProxyPass     string
	ProxyTestLatency string
	ProxyTestError   string

	// History State
	History []store.SpeedResult

	// Daily Usage State (Internet Speed Meter)
	DailyUsages []store.DailyUsage
	TodayUsage  store.DailyUsage
	accumRX     uint64
	accumTX     uint64
	accumWiFi   uint64
	accumOther  uint64
	lastFlush   time.Time

	// Settings State
	StreamsInput    string
	DurationInput   string
	CustomServerURL string
	SettingsMessage string

	// Floating Mini Statusbar Position
	MiniX int
	MiniY int
}

// NewApp initializes application state with the database and live traffic monitor.
func NewApp(db *store.DB) *App {
	mon := netstat.NewMonitor(30)
	a := &App{
		db:               db,
		SelectedNav:      "test",
		NetMonitor:       mon,
		CurrentPhase:     speedtest.PhaseIdle,
		SelectedServerID: "cloudflare-global",
		StatusText:       "Ready to test",
		ChartHistory:     make([]float64, 30),
		NewProxyType:     "http",
		NewProxyPort:     "8080",
		StreamsInput:     db.GetSetting("streams", "4"),
		DurationInput:    db.GetSetting("duration", "8"),
		CustomServerURL:  db.GetSetting("custom_server_url", ""),
	}
	a.ReloadData()
	return a
}

// SetWindow binds the native window to the state coordinator and starts live monitoring.
func (a *App) SetWindow(win *mygo.Window) {
	a.win = win
	a.lastFlush = time.Now()

	// Start live background traffic sampler
	a.NetMonitor.Start(500*time.Millisecond, func(stats netstat.TrafficStats) {
		a.mu.Lock()
		a.LiveStats = stats

		// Accumulate live deltas for Internet Speed Meter daily tracking
		a.accumRX += stats.DeltaRXBytes
		a.accumTX += stats.DeltaTXBytes
		a.accumWiFi += stats.DeltaWiFiBytes
		a.accumOther += stats.DeltaOtherBytes

		// Update in-memory TodayUsage immediately for real-time display
		a.TodayUsage.RXBytes += stats.DeltaRXBytes
		a.TodayUsage.TXBytes += stats.DeltaTXBytes
		a.TodayUsage.WiFiBytes += stats.DeltaWiFiBytes
		a.TodayUsage.OtherBytes += stats.DeltaOtherBytes
		a.TodayUsage.TotalBytes = a.TodayUsage.RXBytes + a.TodayUsage.TXBytes

		now := time.Now()
		shouldFlush := (now.Sub(a.lastFlush) >= 3*time.Second || a.accumRX >= 256*1024 || a.accumTX >= 256*1024) &&
			(a.accumRX > 0 || a.accumTX > 0) && a.db != nil

		var flushRX, flushTX, flushWiFi, flushOther uint64
		todayDate := now.Format("2006-01-02")
		if shouldFlush {
			flushRX = a.accumRX
			flushTX = a.accumTX
			flushWiFi = a.accumWiFi
			flushOther = a.accumOther
			a.accumRX = 0
			a.accumTX = 0
			a.accumWiFi = 0
			a.accumOther = 0
			a.lastFlush = now
		}
		a.mu.Unlock()

		if shouldFlush {
			_ = a.db.AddDailyUsage(todayDate, flushRX, flushTX, flushWiFi, flushOther)
		}

		// Update Windows System Tray ToolTip & Live Menu Items throttled to every 2 seconds
		now = time.Now()
		if now.Sub(a.lastTrayUpdate) >= 2*time.Second {
			a.lastTrayUpdate = now
			if a.tray != nil {
				tip := fmt.Sprintf("Jalin Speed Meter\n↓ RX: %s\n↑ TX: %s\nToday: %s",
					netstat.FormatSpeed(stats.RXBytesPerSec),
					netstat.FormatSpeed(stats.TXBytesPerSec),
					netstat.FormatSpeed(float64(a.TodayUsage.TotalBytes)))
				a.tray.SetToolTip(tip)
			}
			if a.traySpeedItem != nil {
				a.traySpeedItem.SetLabel(fmt.Sprintf("Live: ↓ %s   ↑ %s",
					netstat.FormatSpeed(stats.RXBytesPerSec),
					netstat.FormatSpeed(stats.TXBytesPerSec)))
			}
			if a.trayUsageItem != nil {
				a.trayUsageItem.SetLabel(fmt.Sprintf("Today: %s",
					netstat.FormatSpeed(float64(a.TodayUsage.TotalBytes))))
			}
		}

		if a.win != nil && a.win.IsVisible() {
			a.win.Update(func() {})
		}
		if a.miniWin != nil && a.miniWin.IsVisible() {
			a.miniWin.Update(func() {})
		}
	})
}

// SetTray binds the native system tray and context menu live telemetry items.
func (a *App) SetTray(tray *mygo.Tray, speedItem, usageItem *mygo.MenuItem) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tray = tray
	a.traySpeedItem = speedItem
	a.trayUsageItem = usageItem
}

// SetMiniWindow binds the floating taskbar speed meter window.
func (a *App) SetMiniWindow(w *mygo.Window) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.miniWin = w
}

// RestoreFromTray hides the mini floating widget and restores the main application window.
func (a *App) RestoreFromTray() {
	if a.miniWin != nil {
		a.miniWin.Hide()
	}
	if a.win != nil {
		a.win.Show()
		a.win.Restore()
		a.win.Focus()
	}
}

// UpdateMiniPosition records and persists the floating speed meter's coordinates.
func (a *App) UpdateMiniPosition(x, y int) {
	a.mu.Lock()
	a.MiniX = x
	a.MiniY = y
	a.mu.Unlock()
	if a.db != nil {
		_ = a.db.SetSetting("mini_x", strconv.Itoa(x))
		_ = a.db.SetSetting("mini_y", strconv.Itoa(y))
	}
}

// HideToTray hides the main window and activates the docked floating speed meter widget.
func (a *App) HideToTray() {
	if a.win != nil {
		a.win.Hide()
	}
	if a.miniWin != nil {
		a.mu.Lock()
		mx, my := a.MiniX, a.MiniY
		a.mu.Unlock()
		if mx > 0 && my > 0 {
			a.miniWin.SetPosition(mx, my)
		}
		a.miniWin.ShowInactive()
		a.miniWin.SetAlwaysOnTop(true)
	}
}

// Exit cleanly closes windows and terminates the application.
func (a *App) Exit() {
	if a.miniWin != nil {
		a.miniWin.Close()
	}
	if a.win != nil {
		a.win.Close()
	}
	mygo.App.Quit()
}

// ReloadData fetches the latest records from SQLite.
func (a *App) ReloadData() {
	if a.db == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	hist, _ := a.db.ListSpeedResults(50)
	a.History = hist

	profs, _ := a.db.ListProxyProfiles()
	a.Profiles = profs

	active, _ := a.db.GetActiveProxy()
	a.ActiveProxy = active

	daily, _ := a.db.ListDailyUsage(30)
	a.DailyUsages = daily

	todayStr := time.Now().Format("2006-01-02")
	today, _ := a.db.GetDailyUsage(todayStr)
	if today != nil {
		a.TodayUsage = *today
	}
}

// StartSpeedTest initiates the speed test in a background goroutine.
func (a *App) StartSpeedTest() {
	a.mu.Lock()
	if a.IsTesting {
		a.mu.Unlock()
		return
	}

	a.IsTesting = true
	a.CurrentPhase = speedtest.PhasePing
	a.CurrentMbps = 0
	a.LatencyMs = 0
	a.JitterMs = 0
	a.DownloadMbps = 0
	a.UploadMbps = 0
	a.Progress = 0.05
	a.StatusText = "Measuring latency & jitter..."
	a.ChartHistory = make([]float64, 30)

	ctx, cancel := context.WithCancel(context.Background())
	a.CancelFunc = cancel

	streams, _ := strconv.Atoi(a.StreamsInput)
	if streams <= 0 {
		streams = 4
	}
	durSec, _ := strconv.Atoi(a.DurationInput)
	if durSec <= 0 {
		durSec = 8
	}

	server := speedtest.GetServerOrDefault(a.SelectedServerID)
	activeProxy := a.ActiveProxy
	a.mu.Unlock()

	engine := speedtest.NewEngine(streams, time.Duration(durSec)*time.Second, activeProxy)

	go func() {
		result, err := engine.Run(ctx, server, func(s speedtest.Sample) {
			if a.win != nil {
				a.win.Update(func() {
					a.mu.Lock()
					defer a.mu.Unlock()
					a.CurrentPhase = s.Phase
					a.Progress = s.Progress
					if s.LatencyMs > 0 {
						a.LatencyMs = s.LatencyMs
					}
					if s.JitterMs > 0 {
						a.JitterMs = s.JitterMs
					}

					switch s.Phase {
					case speedtest.PhasePing:
						a.StatusText = fmt.Sprintf("Testing latency: %.1f ms", s.LatencyMs)
					case speedtest.PhaseDownload:
						a.CurrentMbps = s.Mbps
						a.DownloadMbps = s.Mbps
						a.StatusText = fmt.Sprintf("Testing download: %.1f Mbps", s.Mbps)
						a.ChartHistory = append(a.ChartHistory[1:], s.Mbps)
					case speedtest.PhaseUpload:
						a.CurrentMbps = s.Mbps
						a.UploadMbps = s.Mbps
						a.StatusText = fmt.Sprintf("Testing upload: %.1f Mbps", s.Mbps)
						a.ChartHistory = append(a.ChartHistory[1:], s.Mbps)
					case speedtest.PhaseDone:
						a.StatusText = "Test complete"
					}
				})
			}
		})

		if a.win != nil {
			a.win.Update(func() {
				a.mu.Lock()
				defer a.mu.Unlock()
				a.IsTesting = false
				a.CancelFunc = nil

				if err != nil {
					if ctx.Err() == context.Canceled {
						a.StatusText = "Test cancelled"
					} else {
						a.StatusText = fmt.Sprintf("Test failed: %v", err)
					}
					return
				}

				if result != nil {
					a.DownloadMbps = result.DownloadMbps
					a.UploadMbps = result.UploadMbps
					a.LatencyMs = result.LatencyMs
					a.JitterMs = result.JitterMs
					a.StatusText = fmt.Sprintf("Complete: ↓ %.1f Mbps  ↑ %.1f Mbps", result.DownloadMbps, result.UploadMbps)

					// Save into SQLite
					dbResult := &store.SpeedResult{
						CreatedAt:    time.Now(),
						DownloadMbps: result.DownloadMbps,
						UploadMbps:   result.UploadMbps,
						LatencyMs:    result.LatencyMs,
						JitterMs:     result.JitterMs,
						Server:       result.ServerName,
						ViaProxy:     result.ViaProxy,
						NetworkType:  "Ethernet/WiFi",
					}
					if a.db != nil {
						_, _ = a.db.InsertSpeedResult(dbResult)
						a.LastResult = dbResult
						hist, _ := a.db.ListSpeedResults(50)
						a.History = hist
					}
				}
			})
		}
	}()
}

// CancelSpeedTest cancels the currently running test.
func (a *App) CancelSpeedTest() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.CancelFunc != nil {
		a.CancelFunc()
		a.CancelFunc = nil
	}
	a.IsTesting = false
	a.StatusText = "Cancelled"
}

// ToggleProxy switches a proxy profile's active status.
func (a *App) ToggleProxy(profile *store.ProxyProfile) {
	if a.db == nil || profile == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	var newActiveID int64
	if !profile.Active {
		newActiveID = profile.ID
	}
	_ = a.db.SetActiveProxy(newActiveID)

	profs, _ := a.db.ListProxyProfiles()
	a.Profiles = profs
	active, _ := a.db.GetActiveProxy()
	a.ActiveProxy = active
}

// DeleteProxy removes a saved proxy profile.
func (a *App) DeleteProxy(id int64) {
	if a.db == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	_ = a.db.DeleteProxyProfile(id)
	profs, _ := a.db.ListProxyProfiles()
	a.Profiles = profs
	active, _ := a.db.GetActiveProxy()
	a.ActiveProxy = active
}

// SaveNewProxy adds a proxy profile from modal inputs.
func (a *App) SaveNewProxy() {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := strings.TrimSpace(a.NewProxyName)
	if name == "" {
		name = "Custom Proxy"
	}
	host := strings.TrimSpace(a.NewProxyHost)
	if host == "" {
		return
	}
	port, _ := strconv.Atoi(strings.TrimSpace(a.NewProxyPort))
	if port <= 0 {
		port = 8080
	}

	p := &store.ProxyProfile{
		Name:      name,
		Type:      a.NewProxyType,
		Host:      host,
		Port:      port,
		Username:  strings.TrimSpace(a.NewProxyUser),
		Password:  strings.TrimSpace(a.NewProxyPass),
		CreatedAt: time.Now(),
	}

	if a.db != nil {
		_, _ = a.db.InsertProxyProfile(p)
		profs, _ := a.db.ListProxyProfiles()
		a.Profiles = profs
	}

	a.ShowAddModal = false
	a.NewProxyName = ""
	a.NewProxyHost = ""
	a.NewProxyPort = "8080"
	a.NewProxyUser = ""
	a.NewProxyPass = ""
}

// TestProxyConnection runs a live connectivity check on a profile.
func (a *App) TestProxyConnection(p *store.ProxyProfile) {
	a.mu.Lock()
	if a.IsTestingProxy {
		a.mu.Unlock()
		return
	}
	a.IsTestingProxy = true
	a.ProxyTestLatency = "Testing..."
	a.ProxyTestError = ""
	a.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		defer cancel()

		lat, err := proxy.TestConnection(ctx, p)

		if a.win != nil {
			a.win.Update(func() {
				a.mu.Lock()
				defer a.mu.Unlock()
				a.IsTestingProxy = false
				if err != nil {
					a.ProxyTestLatency = ""
					a.ProxyTestError = err.Error()
				} else {
					a.ProxyTestLatency = fmt.Sprintf("%.0f ms", float64(lat.Milliseconds()))
					a.ProxyTestError = ""
				}
			})
		}
	}()
}

// ClearAllHistory deletes all saved history results.
func (a *App) ClearAllHistory() {
	if a.db == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = a.db.ClearSpeedResults()
	a.History = nil
}

// DeleteHistoryItem deletes one result from history.
func (a *App) DeleteHistoryItem(id int64) {
	if a.db == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = a.db.DeleteSpeedResult(id)
	hist, _ := a.db.ListSpeedResults(50)
	a.History = hist
}

// SaveSettings persists configuration options to SQLite.
func (a *App) SaveSettings() {
	if a.db == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	_ = a.db.SetSetting("streams", strings.TrimSpace(a.StreamsInput))
	_ = a.db.SetSetting("duration", strings.TrimSpace(a.DurationInput))
	_ = a.db.SetSetting("custom_server_url", strings.TrimSpace(a.CustomServerURL))

	a.SettingsMessage = "Settings saved successfully!"
}
