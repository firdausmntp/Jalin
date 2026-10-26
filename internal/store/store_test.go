package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_jalin.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	// 1. Test Speed Results
	res := &SpeedResult{
		CreatedAt:    time.Now(),
		DownloadMbps: 125.4,
		UploadMbps:   48.2,
		LatencyMs:    12.5,
		JitterMs:     1.8,
		Server:       "Cloudflare (Jakarta)",
		ViaProxy:     false,
		NetworkType:  "WiFi",
	}

	id, err := db.InsertSpeedResult(res)
	if err != nil {
		t.Fatalf("InsertSpeedResult failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Expected valid id, got %d", id)
	}

	list, err := db.ListSpeedResults(10)
	if err != nil {
		t.Fatalf("ListSpeedResults failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(list))
	}
	if list[0].DownloadMbps != 125.4 {
		t.Fatalf("Expected 125.4 mbps, got %f", list[0].DownloadMbps)
	}

	// 2. Test Proxy Profile
	profile := &ProxyProfile{
		Name:     "Test SOCKS5",
		Type:     "socks5",
		Host:     "127.0.0.1",
		Port:     1080,
		Username: "user",
		Password: "password",
	}

	pid, err := db.InsertProxyProfile(profile)
	if err != nil {
		t.Fatalf("InsertProxyProfile failed: %v", err)
	}

	if err := db.SetActiveProxy(pid); err != nil {
		t.Fatalf("SetActiveProxy failed: %v", err)
	}

	active, err := db.GetActiveProxy()
	if err != nil {
		t.Fatalf("GetActiveProxy failed: %v", err)
	}
	if active == nil || active.ID != pid {
		t.Fatalf("Expected active proxy with ID %d, got %+v", pid, active)
	}

	// 3. Test Settings
	if err := db.SetSetting("theme", "dark"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	val := db.GetSetting("theme", "light")
	if val != "dark" {
		t.Fatalf("Expected 'dark', got %q", val)
	}

	// 4. Test Daily Usage (Internet Speed Meter)
	today := time.Now().Format("2006-01-02")
	if err := db.AddDailyUsage(today, 1024*1024, 512*1024, 800*1024, 736*1024); err != nil {
		t.Fatalf("AddDailyUsage initial failed: %v", err)
	}
	// Accumulate more on the same day (test ON CONFLICT upsert)
	if err := db.AddDailyUsage(today, 2048*1024, 1024*1024, 1248*1024, 1824*1024); err != nil {
		t.Fatalf("AddDailyUsage accumulate failed: %v", err)
	}

	usage, err := db.GetDailyUsage(today)
	if err != nil {
		t.Fatalf("GetDailyUsage failed: %v", err)
	}
	if usage == nil || usage.RXBytes != 3072*1024 || usage.TXBytes != 1536*1024 || usage.WiFiBytes != 2048*1024 {
		t.Fatalf("Unexpected daily usage: %+v", usage)
	}

	dailyList, err := db.ListDailyUsage(10)
	if err != nil {
		t.Fatalf("ListDailyUsage failed: %v", err)
	}
	if len(dailyList) == 0 || dailyList[0].Date != today {
		t.Fatalf("Expected daily list to contain today, got %+v", dailyList)
	}
}
