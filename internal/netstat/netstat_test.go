package netstat

import (
	"testing"
	"time"
)

func TestNetstatMonitor(t *testing.T) {
	mon := NewMonitor(20)
	done := make(chan bool, 1)

	mon.Start(50*time.Millisecond, func(s TrafficStats) {
		if len(s.HistoryRX) == 20 {
			select {
			case done <- true:
			default:
			}
		}
	})
	defer mon.Stop()

	select {
	case <-done:
		snap := mon.Snapshot()
		if len(snap.HistoryRX) != 20 {
			t.Fatalf("Expected 20 history items, got %d", len(snap.HistoryRX))
		}
	case <-time.After(2 * time.Second):
		// On non-windows or quiet env, ensure at least Snapshot works
		snap := mon.Snapshot()
		if len(snap.HistoryRX) != 20 {
			t.Fatalf("Expected 20 history items, got %d", len(snap.HistoryRX))
		}
	}
}

func TestFormatHelpers(t *testing.T) {
	if s := FormatSpeed(500); s != "500 B/s" {
		t.Fatalf("Expected 500 B/s, got %s", s)
	}
	if s := FormatSpeed(2048); s != "2.0 KB/s" {
		t.Fatalf("Expected 2.0 KB/s, got %s", s)
	}
	if s := FormatSpeed(5 * 1024 * 1024); s != "5.00 MB/s" {
		t.Fatalf("Expected 5.00 MB/s, got %s", s)
	}
	if s := FormatMbps(124.5); s != "124.5 Mbps" {
		t.Fatalf("Expected 124.5 Mbps, got %s", s)
	}
}
