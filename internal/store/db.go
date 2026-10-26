package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection with helper methods.
type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

// DefaultDBPath returns the platform-specific default path for jalin.db.
func DefaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return "jalin.db"
	}
	appDir := filepath.Join(configDir, "jalin")
	_ = os.MkdirAll(appDir, 0o755)
	return filepath.Join(appDir, "jalin.db")
}

// OpenDB opens a SQLite database at the specified path and runs migrations.
func OpenDB(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", filepath.ToSlash(path))
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", path, err)
	}

	conn.SetMaxOpenConns(1) // SQLite single-writer safe concurrency

	d := &DB{db: conn}
	if err := d.migrate(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate sqlite db: %w", err)
	}

	return d, nil
}

// Close closes the underlying database handle.
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.db.Close()
}

func (d *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS speed_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL,
			download_mbps REAL NOT NULL,
			upload_mbps REAL NOT NULL,
			latency_ms REAL NOT NULL,
			jitter_ms REAL NOT NULL,
			server TEXT NOT NULL DEFAULT '',
			via_proxy INTEGER NOT NULL DEFAULT 0,
			network_type TEXT NOT NULL DEFAULT 'Ethernet/WiFi'
		);`,
		`CREATE TABLE IF NOT EXISTS proxy_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'http',
			host TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 8080,
			username TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL DEFAULT '',
			active INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS daily_usage (
			date TEXT PRIMARY KEY,
			wifi_bytes INTEGER NOT NULL DEFAULT 0,
			other_bytes INTEGER NOT NULL DEFAULT 0,
			rx_bytes INTEGER NOT NULL DEFAULT 0,
			tx_bytes INTEGER NOT NULL DEFAULT 0
		);`,
	}

	for _, q := range queries {
		if _, err := d.db.Exec(q); err != nil {
			return err
		}
	}

	// Idempotent migrations for existing installations
	_, _ = d.db.Exec("ALTER TABLE daily_usage ADD COLUMN wifi_bytes INTEGER NOT NULL DEFAULT 0;")
	_, _ = d.db.Exec("ALTER TABLE daily_usage ADD COLUMN other_bytes INTEGER NOT NULL DEFAULT 0;")
	return nil
}

// InsertSpeedResult inserts a new speed test result.
func (d *DB) InsertSpeedResult(r *SpeedResult) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}

	viaProxyInt := 0
	if r.ViaProxy {
		viaProxyInt = 1
	}

	query := `INSERT INTO speed_results (created_at, download_mbps, upload_mbps, latency_ms, jitter_ms, server, via_proxy, network_type)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	res, err := d.db.Exec(query, r.CreatedAt.Format(time.RFC3339), r.DownloadMbps, r.UploadMbps, r.LatencyMs, r.JitterMs, r.Server, viaProxyInt, r.NetworkType)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	r.ID = id
	return id, nil
}

// ListSpeedResults returns the latest N speed results ordered by date descending.
func (d *DB) ListSpeedResults(limit int) ([]SpeedResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `SELECT id, created_at, download_mbps, upload_mbps, latency_ms, jitter_ms, server, via_proxy, network_type 
	          FROM speed_results ORDER BY id DESC LIMIT ?;`
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SpeedResult
	for rows.Next() {
		var r SpeedResult
		var createdAtStr string
		var viaProxyInt int
		if err := rows.Scan(&r.ID, &createdAtStr, &r.DownloadMbps, &r.UploadMbps, &r.LatencyMs, &r.JitterMs, &r.Server, &viaProxyInt, &r.NetworkType); err != nil {
			return nil, err
		}
		r.CreatedAt = parseDBTime(createdAtStr)
		r.ViaProxy = viaProxyInt != 0
		results = append(results, r)
	}
	return results, rows.Err()
}

// DeleteSpeedResult deletes a single result record by ID.
func (d *DB) DeleteSpeedResult(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(`DELETE FROM speed_results WHERE id = ?;`, id)
	return err
}

// ClearSpeedResults removes all saved speed test results.
func (d *DB) ClearSpeedResults() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(`DELETE FROM speed_results;`)
	return err
}

// InsertProxyProfile stores a new proxy configuration profile.
func (d *DB) InsertProxyProfile(p *ProxyProfile) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	activeInt := 0
	if p.Active {
		activeInt = 1
	}

	query := `INSERT INTO proxy_profiles (name, type, host, port, username, password, active, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	res, err := d.db.Exec(query, p.Name, p.Type, p.Host, p.Port, p.Username, p.Password, activeInt, p.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	p.ID = id
	return id, nil
}

// UpdateProxyProfile updates an existing proxy profile.
func (d *DB) UpdateProxyProfile(p *ProxyProfile) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	activeInt := 0
	if p.Active {
		activeInt = 1
	}

	query := `UPDATE proxy_profiles SET name = ?, type = ?, host = ?, port = ?, username = ?, password = ?, active = ?
	          WHERE id = ?;`
	_, err := d.db.Exec(query, p.Name, p.Type, p.Host, p.Port, p.Username, p.Password, activeInt, p.ID)
	return err
}

// DeleteProxyProfile removes a proxy profile.
func (d *DB) DeleteProxyProfile(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(`DELETE FROM proxy_profiles WHERE id = ?;`, id)
	return err
}

// ListProxyProfiles retrieves all saved proxy profiles.
func (d *DB) ListProxyProfiles() ([]ProxyProfile, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, name, type, host, port, username, password, active, created_at 
	          FROM proxy_profiles ORDER BY id ASC;`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []ProxyProfile
	for rows.Next() {
		var p ProxyProfile
		var createdAtStr string
		var activeInt int
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.Host, &p.Port, &p.Username, &p.Password, &activeInt, &createdAtStr); err != nil {
			return nil, err
		}
		p.CreatedAt = parseDBTime(createdAtStr)
		p.Active = activeInt != 0
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// SetActiveProxy marks one proxy profile as active, or deactivates all if id <= 0.
func (d *DB) SetActiveProxy(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`UPDATE proxy_profiles SET active = 0;`); err != nil {
		return err
	}
	if id > 0 {
		if _, err := tx.Exec(`UPDATE proxy_profiles SET active = 1 WHERE id = ?;`, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetActiveProxy returns the currently active proxy profile, or nil if none.
func (d *DB) GetActiveProxy() (*ProxyProfile, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, name, type, host, port, username, password, active, created_at 
	          FROM proxy_profiles WHERE active = 1 LIMIT 1;`
	var p ProxyProfile
	var createdAtStr string
	var activeInt int
	err := d.db.QueryRow(query).Scan(&p.ID, &p.Name, &p.Type, &p.Host, &p.Port, &p.Username, &p.Password, &activeInt, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = parseDBTime(createdAtStr)
	p.Active = activeInt != 0
	return &p, nil
}

// GetSetting gets a persistent setting by key with a fallback default.
func (d *DB) GetSetting(key, defaultVal string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var val string
	err := d.db.QueryRow(`SELECT value FROM settings WHERE key = ?;`, key).Scan(&val)
	if err != nil {
		return defaultVal
	}
	return val
}

// SetSetting writes a persistent setting by key.
func (d *DB) SetSetting(key, val string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value;`, key, val)
	return err
}

// AddDailyUsage accumulates transferred delta bytes into a specific date's record.
func (d *DB) AddDailyUsage(date string, deltaRX, deltaTX, deltaWiFi, deltaOther uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `INSERT INTO daily_usage (date, rx_bytes, tx_bytes, wifi_bytes, other_bytes) 
	          VALUES (?, ?, ?, ?, ?) 
	          ON CONFLICT(date) DO UPDATE SET 
	              rx_bytes = rx_bytes + excluded.rx_bytes, 
	              tx_bytes = tx_bytes + excluded.tx_bytes,
	              wifi_bytes = wifi_bytes + excluded.wifi_bytes,
	              other_bytes = other_bytes + excluded.other_bytes;`
	_, err := d.db.Exec(query, date, deltaRX, deltaTX, deltaWiFi, deltaOther)
	return err
}

// GetDailyUsage returns the data usage recorded for a specific date (YYYY-MM-DD).
func (d *DB) GetDailyUsage(date string) (*DailyUsage, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var u DailyUsage
	err := d.db.QueryRow(`SELECT date, rx_bytes, tx_bytes, wifi_bytes, other_bytes FROM daily_usage WHERE date = ?;`, date).
		Scan(&u.Date, &u.RXBytes, &u.TXBytes, &u.WiFiBytes, &u.OtherBytes)
	if err == sql.ErrNoRows {
		return &DailyUsage{Date: date}, nil
	}
	if err != nil {
		return nil, err
	}
	u.TotalBytes = u.RXBytes + u.TXBytes
	return &u, nil
}

// ListDailyUsage retrieves the latest N daily usage records ordered by date descending.
func (d *DB) ListDailyUsage(limit int) ([]DailyUsage, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 {
		limit = 30
	}
	query := `SELECT date, rx_bytes, tx_bytes, wifi_bytes, other_bytes FROM daily_usage ORDER BY date DESC LIMIT ?;`
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []DailyUsage
	for rows.Next() {
		var u DailyUsage
		if err := rows.Scan(&u.Date, &u.RXBytes, &u.TXBytes, &u.WiFiBytes, &u.OtherBytes); err != nil {
			return nil, err
		}
		u.TotalBytes = u.RXBytes + u.TXBytes
		usages = append(usages, u)
	}
	return usages, rows.Err()
}

// GetTotalDailyUsage returns the aggregate RX and TX volume across all recorded days.
func (d *DB) GetTotalDailyUsage() (totalRX uint64, totalTX uint64, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT COALESCE(SUM(rx_bytes), 0), COALESCE(SUM(tx_bytes), 0) FROM daily_usage;`
	err = d.db.QueryRow(query).Scan(&totalRX, &totalTX)
	return totalRX, totalTX, err
}

func parseDBTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil && !t.IsZero() {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil && !t.IsZero() {
		return t
	}
	return time.Time{}
}
