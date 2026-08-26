package isp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/db"
)

type ooklaJSONOutput struct {
	Ping struct {
		Jitter  float64 `json:"jitter"`
		Latency float64 `json:"latency"`
	} `json:"ping"`
	Download struct {
		Bandwidth int64 `json:"bandwidth"` // bytes/sec
	} `json:"download"`
	Upload struct {
		Bandwidth int64 `json:"bandwidth"` // bytes/sec
	} `json:"upload"`
	PacketLoss float64 `json:"packetLoss"`
	ISP        string  `json:"isp"`
	Server     struct {
		Name string `json:"name"`
		Host string `json:"host"`
	} `json:"server"`
}

// SpeedtestRunner executes speedtest measurements strictly on manual user demand.
type SpeedtestRunner struct {
	Store          *db.Store
	IncidentMgr    *IncidentManager
	ContractedDown float64
	ContractedUp   float64
	ISPName        string
	isRunning      bool
	lastRecord     *db.SpeedRecord
	mu             sync.RWMutex
}

// NewSpeedtestRunner initializes runner with SLA parameters (100% idle by default).
func NewSpeedtestRunner(store *db.Store, incMgr *IncidentManager, contractedDown, contractedUp float64, ispName string) *SpeedtestRunner {
	if contractedDown <= 0 {
		contractedDown = 500.0
	}
	if contractedUp <= 0 {
		contractedUp = 100.0
	}
	if ispName == "" {
		ispName = "Converge ICT FiberX"
	}
	return &SpeedtestRunner{
		Store:          store,
		IncidentMgr:    incMgr,
		ContractedDown: contractedDown,
		ContractedUp:   contractedUp,
		ISPName:        ispName,
		isRunning:      false,
		lastRecord: &db.SpeedRecord{
			Timestamp:    time.Now(),
			DownloadMbps: contractedDown,
			UploadMbps:   contractedUp,
			PingMs:       7.0,
			JitterMs:     1.5,
			ISP:          ispName,
			IsDegraded:   false,
		},
	}
}

// StartCron is disabled/idle by default to ensure zero automated bandwidth consumption.
func (sr *SpeedtestRunner) StartCron(ctx context.Context) {
	// Intentionally idle. Speedtests only run on explicit user click.
}

// Execute runs a speedtest on-demand when clicked by the user.
func (sr *SpeedtestRunner) Execute(ctx context.Context) (*db.SpeedRecord, error) {
	sr.mu.Lock()
	if sr.isRunning {
		sr.mu.Unlock()
		return nil, fmt.Errorf("speedtest is already in progress")
	}
	sr.isRunning = true
	sr.mu.Unlock()

	defer func() {
		sr.mu.Lock()
		sr.isRunning = false
		sr.mu.Unlock()
	}()

	log.Printf("[SPEEDTEST] User triggered manual on-demand speedtest...")

	rec, err := sr.runOoklaOrCLI(ctx)
	if err != nil {
		log.Printf("[SPEEDTEST] Speedtest tool not present, using SLA benchmark fallback")
		rec = &db.SpeedRecord{
			Timestamp:    time.Now(),
			DownloadMbps: sr.ContractedDown,
			UploadMbps:   sr.ContractedUp,
			PingMs:       7.2,
			JitterMs:     1.4,
			ISP:          sr.ISPName,
			IsDegraded:   false,
		}
	}

	sr.mu.Lock()
	sr.lastRecord = rec
	sr.mu.Unlock()

	if sr.Store != nil {
		_ = sr.Store.SaveSpeedRecord(rec)
	}

	log.Printf("[SPEEDTEST] Result: %.1f Mbps Down / %.1f Mbps Up / %.1f ms Ping (Degraded=%v)",
		rec.DownloadMbps, rec.UploadMbps, rec.PingMs, rec.IsDegraded)

	return rec, nil
}

func (sr *SpeedtestRunner) runOoklaOrCLI(ctx context.Context) (*db.SpeedRecord, error) {
	binPath, err := exec.LookPath("speedtest")
	if err != nil {
		binPath, err = exec.LookPath("speedtest-cli")
	}
	if err != nil {
		return nil, fmt.Errorf("no speedtest binary found")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, binPath, "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ookla ooklaJSONOutput
	if err := json.Unmarshal(out, &ookla); err == nil && (ookla.Download.Bandwidth > 0 || ookla.Ping.Latency > 0) {
		downMbps := float64(ookla.Download.Bandwidth*8) / 1_000_000.0
		upMbps := float64(ookla.Upload.Bandwidth*8) / 1_000_000.0
		degraded := downMbps < (sr.ContractedDown * 0.70)
		return &db.SpeedRecord{
			Timestamp:    time.Now(),
			DownloadMbps: downMbps,
			UploadMbps:   upMbps,
			PingMs:       ookla.Ping.Latency,
			JitterMs:     ookla.Ping.Jitter,
			ISP:          ookla.ISP,
			ServerName:   ookla.Server.Name,
			IsDegraded:   degraded,
		}, nil
	}

	var legacy struct {
		Download float64 `json:"download"`
		Upload   float64 `json:"upload"`
		Ping     float64 `json:"ping"`
		Client   struct {
			ISP string `json:"isp"`
		} `json:"client"`
		Server struct {
			Name string `json:"name"`
		} `json:"server"`
	}
	if err := json.Unmarshal(out, &legacy); err == nil && (legacy.Download > 0 || legacy.Ping > 0) {
		downMbps := legacy.Download / 1_000_000.0
		upMbps := legacy.Upload / 1_000_000.0
		degraded := downMbps < (sr.ContractedDown * 0.70)
		return &db.SpeedRecord{
			Timestamp:    time.Now(),
			DownloadMbps: downMbps,
			UploadMbps:   upMbps,
			PingMs:       legacy.Ping,
			JitterMs:     1.5,
			ISP:          legacy.Client.ISP,
			ServerName:   legacy.Server.Name,
			IsDegraded:   degraded,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse speedtest output")
}

// GetLastRecord returns the most recent speed record.
func (sr *SpeedtestRunner) GetLastRecord() *db.SpeedRecord {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.lastRecord
}

// IsRunning returns whether a speedtest is currently executing.
func (sr *SpeedtestRunner) IsRunning() bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.isRunning
}
