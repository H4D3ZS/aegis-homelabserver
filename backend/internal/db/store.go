package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps the local SQLite database for speedtest history and incidents.
type Store struct {
	db *sql.DB
}

// SpeedRecord represents an executed speed measurement.
type SpeedRecord struct {
	ID           int64     `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	PingMs       float64   `json:"ping_ms"`
	JitterMs     float64   `json:"jitter_ms"`
	PacketLoss   float64   `json:"packet_loss"`
	ISP          string    `json:"isp"`
	ServerName   string    `json:"server_name"`
	ServerHost   string    `json:"server_host"`
	IsDegraded   bool      `json:"is_degraded"`
}

// IncidentRecord represents a recorded ISP outage or SLA breach.
type IncidentRecord struct {
	ID          int64     `json:"id"`
	StartedAt   time.Time `json:"started_at"`
	ResolvedAt  *time.Time`json:"resolved_at,omitempty"`
	Type        string    `json:"type"` // "OUTAGE", "DEGRADATION", "HIGH_JITTER"
	Description string    `json:"description"`
	MetricsJSON string    `json:"metrics_json"`
}

// NewStore initializes SQLite schema.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// High concurrency PRAGMAs
	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA synchronous=NORMAL;")

	schema := `
	CREATE TABLE IF NOT EXISTS speed_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		download_mbps REAL NOT NULL,
		upload_mbps REAL NOT NULL,
		ping_ms REAL NOT NULL,
		jitter_ms REAL NOT NULL,
		packet_loss REAL NOT NULL,
		isp TEXT NOT NULL,
		server_name TEXT,
		server_host TEXT,
		is_degraded BOOLEAN NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS incidents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		started_at DATETIME NOT NULL,
		resolved_at DATETIME,
		type TEXT NOT NULL,
		description TEXT NOT NULL,
		metrics_json TEXT
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	return &Store{db: db}, nil
}

// SaveSpeedRecord writes a speedtest result to the database.
func (s *Store) SaveSpeedRecord(rec *SpeedRecord) error {
	query := `
	INSERT INTO speed_records (timestamp, download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss, isp, server_name, server_host, is_degraded)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, rec.Timestamp, rec.DownloadMbps, rec.UploadMbps, rec.PingMs, rec.JitterMs, rec.PacketLoss, rec.ISP, rec.ServerName, rec.ServerHost, rec.IsDegraded)
	return err
}

// GetSpeedHistory returns speed records within the given hours window.
func (s *Store) GetSpeedHistory(hours int) ([]SpeedRecord, error) {
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	rows, err := s.db.Query(`SELECT id, timestamp, download_mbps, upload_mbps, ping_ms, jitter_ms, packet_loss, isp, server_name, server_host, is_degraded FROM speed_records WHERE timestamp >= ? ORDER BY timestamp ASC`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SpeedRecord
	for rows.Next() {
		var r SpeedRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.DownloadMbps, &r.UploadMbps, &r.PingMs, &r.JitterMs, &r.PacketLoss, &r.ISP, &r.ServerName, &r.ServerHost, &r.IsDegraded); err == nil {
			records = append(records, r)
		}
	}
	return records, nil
}

// LogIncident creates a new incident record.
func (s *Store) LogIncident(inc *IncidentRecord) (int64, error) {
	query := `INSERT INTO incidents (started_at, type, description, metrics_json) VALUES (?, ?, ?, ?)`
	res, err := s.db.Exec(query, inc.StartedAt, inc.Type, inc.Description, inc.MetricsJSON)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ResolveIncident marks an incident as resolved.
func (s *Store) ResolveIncident(id int64, resolvedAt time.Time) error {
	_, err := s.db.Exec(`UPDATE incidents SET resolved_at = ? WHERE id = ?`, resolvedAt, id)
	return err
}

// GetRecentIncidents returns the latest incidents.
func (s *Store) GetRecentIncidents(limit int) ([]IncidentRecord, error) {
	rows, err := s.db.Query(`SELECT id, started_at, resolved_at, type, description, metrics_json FROM incidents ORDER BY started_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []IncidentRecord
	for rows.Next() {
		var inc IncidentRecord
		if err := rows.Scan(&inc.ID, &inc.StartedAt, &inc.ResolvedAt, &inc.Type, &inc.Description, &inc.MetricsJSON); err == nil {
			incidents = append(incidents, inc)
		}
	}
	return incidents, nil
}
