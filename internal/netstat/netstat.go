package netstat

import (
	"fmt"
	"sync"
	"time"
)

// TrafficStats represents a snapshot of live network bandwidth and throughput.
type TrafficStats struct {
	RXBytesPerSec   float64
	TXBytesPerSec   float64
	RXMbps          float64
	TXMbps          float64
	TotalRXBytes    uint64
	TotalTXBytes    uint64
	DeltaRXBytes    uint64
	DeltaTXBytes    uint64
	DeltaWiFiBytes  uint64
	DeltaOtherBytes uint64
	PeakRXMbps      float64
	PeakTXMbps      float64
	HistoryRX       []float64
	HistoryTX       []float64
}

// FormatSpeed formats bytes/sec into a human-readable throughput string.
func FormatSpeed(bytesPerSec float64) string {
	if bytesPerSec >= 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
	}
	if bytesPerSec >= 1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	}
	if bytesPerSec >= 1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

// FormatMbps formats Mbps with 1 decimal place below 1000.
func FormatMbps(mbps float64) string {
	if mbps >= 1000 {
		return fmt.Sprintf("%.0f Mbps", mbps)
	}
	return fmt.Sprintf("%.1f Mbps", mbps)
}

// Monitor continuously polls system network interface traffic.
type Monitor struct {
	mu           sync.RWMutex
	stats        TrafficStats
	stopChan     chan struct{}
	historyLen   int
	lastInBytes  uint64
	lastOutBytes uint64
	lastWiFiIn   uint64
	lastWiFiOut  uint64
	lastOtherIn  uint64
	lastOtherOut uint64
	lastCheck    time.Time
	hasSample    bool
}

// NewMonitor initializes a new real-time network monitor.
func NewMonitor(historyLen int) *Monitor {
	if historyLen <= 0 {
		historyLen = 30
	}
	return &Monitor{
		historyLen: historyLen,
		stopChan:   make(chan struct{}),
		stats: TrafficStats{
			HistoryRX: make([]float64, historyLen),
			HistoryTX: make([]float64, historyLen),
		},
	}
}

// Start begins periodic polling of network interfaces.
func (m *Monitor) Start(interval time.Duration, onUpdate func(TrafficStats)) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopChan:
				return
			case t := <-ticker.C:
				inBytes, outBytes, wifiIn, wifiOut, otherIn, otherOut, err := readInterfaceOctets()
				if err != nil {
					continue
				}

				m.mu.Lock()
				if !m.hasSample {
					m.lastInBytes = inBytes
					m.lastOutBytes = outBytes
					m.lastWiFiIn = wifiIn
					m.lastWiFiOut = wifiOut
					m.lastOtherIn = otherIn
					m.lastOtherOut = otherOut
					m.lastCheck = t
					m.hasSample = true
					m.mu.Unlock()
					continue
				}

				elapsed := t.Sub(m.lastCheck).Seconds()
				if elapsed <= 0 {
					m.mu.Unlock()
					continue
				}

				deltaIn := uint64(0)
				if inBytes >= m.lastInBytes {
					deltaIn = inBytes - m.lastInBytes
				}
				deltaOut := uint64(0)
				if outBytes >= m.lastOutBytes {
					deltaOut = outBytes - m.lastOutBytes
				}

				deltaWiFi := uint64(0)
				if (wifiIn + wifiOut) >= (m.lastWiFiIn + m.lastWiFiOut) {
					deltaWiFi = (wifiIn + wifiOut) - (m.lastWiFiIn + m.lastWiFiOut)
				}
				deltaOther := uint64(0)
				if (otherIn + otherOut) >= (m.lastOtherIn + m.lastOtherOut) {
					deltaOther = (otherIn + otherOut) - (m.lastOtherIn + m.lastOtherOut)
				}

				m.lastInBytes = inBytes
				m.lastOutBytes = outBytes
				m.lastWiFiIn = wifiIn
				m.lastWiFiOut = wifiOut
				m.lastOtherIn = otherIn
				m.lastOtherOut = otherOut
				m.lastCheck = t

				rxBps := float64(deltaIn) / elapsed
				txBps := float64(deltaOut) / elapsed
				rxMbps := (rxBps * 8) / 1_000_000
				txMbps := (txBps * 8) / 1_000_000

				m.stats.RXBytesPerSec = rxBps
				m.stats.TXBytesPerSec = txBps
				m.stats.RXMbps = rxMbps
				m.stats.TXMbps = txMbps
				m.stats.TotalRXBytes = inBytes
				m.stats.TotalTXBytes = outBytes
				m.stats.DeltaRXBytes = deltaIn
				m.stats.DeltaTXBytes = deltaOut
				m.stats.DeltaWiFiBytes = deltaWiFi
				m.stats.DeltaOtherBytes = deltaOther

				if rxMbps > m.stats.PeakRXMbps {
					m.stats.PeakRXMbps = rxMbps
				}
				if txMbps > m.stats.PeakTXMbps {
					m.stats.PeakTXMbps = txMbps
				}

				m.stats.HistoryRX = append(m.stats.HistoryRX[1:], rxMbps)
				m.stats.HistoryTX = append(m.stats.HistoryTX[1:], txMbps)

				// The history slices are reused every tick, so the snapshot gets
				// its own copy; the view reads it while the next tick writes.
				snapshot := m.stats
				snapshot.HistoryRX = append([]float64(nil), m.stats.HistoryRX...)
				snapshot.HistoryTX = append([]float64(nil), m.stats.HistoryTX...)
				m.mu.Unlock()

				if onUpdate != nil {
					onUpdate(snapshot)
				}
			}
		}
	}()
}

// Stop terminates the polling goroutine.
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// Snapshot returns the latest traffic stats.
func (m *Monitor) Snapshot() TrafficStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}
