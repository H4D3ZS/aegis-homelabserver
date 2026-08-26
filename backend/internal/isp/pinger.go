package isp

import (
	"context"
	"math"
	"net"
	"sync"
	"time"
)

// PingTargetResult records single probe result for a target.
type PingTargetResult struct {
	Target     string    `json:"target"`
	PingMs     float64   `json:"ping_ms"`
	PacketLoss float64   `json:"packet_loss"`
	Success    bool      `json:"success"`
	Timestamp  time.Time `json:"timestamp"`
}

// NetworkVitals aggregates rolling latency, jitter, and packet loss metrics.
type NetworkVitals struct {
	Timestamp      time.Time          `json:"timestamp"`
	PingMs         float64            `json:"ping_ms"`
	JitterMs       float64            `json:"jitter_ms"`
	PacketLoss     float64            `json:"packet_loss"`
	PrimaryTarget  string             `json:"primary_target"`
	Targets        []PingTargetResult `json:"targets"`
	IsDegraded     bool               `json:"is_degraded"`
	DegradedReason string             `json:"degraded_reason,omitempty"`
}

// Pinger performs continuous UDP/TCP dial probes to calculate rolling jitter (RFC 3550).
type Pinger struct {
	Targets      []string
	vitals       NetworkVitals
	mu           sync.RWMutex
	lastDiff     float64
	currentJitter float64
	recentPings  []float64
	lossWindow   []bool
}

// NewPinger initializes the prober with target list.
func NewPinger(targets []string) *Pinger {
	if len(targets) == 0 {
		targets = []string{"1.1.1.1:53", "8.8.8.8:53", "192.168.100.1:53"}
	}
	return &Pinger{
		Targets:       targets,
		recentPings:   make([]float64, 0, 30),
		lossWindow:    make([]bool, 0, 30),
		currentJitter: 1.5,
	}
}

// Start launches the continuous probe loop.
func (p *Pinger) Start(ctx context.Context) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.probeAll(ctx)
		}
	}
}

func (p *Pinger) probeAll(ctx context.Context) {
	var targetResults []PingTargetResult
	var successCount int
	var totalLatency float64
	var primaryPing float64

	for i, target := range p.Targets {
		start := time.Now()
		d := net.Dialer{Timeout: 800 * time.Millisecond}
		conn, err := d.DialContext(ctx, "udp", target)

		var success bool
		var latency float64

		if err == nil {
			_ = conn.Close()
			latency = float64(time.Since(start).Microseconds()) / 1000.0
			success = true
			successCount++
			totalLatency += latency
			if i == 0 {
				primaryPing = latency
			}
		} else {
			latency = 0.0
			success = false
		}

		targetResults = append(targetResults, PingTargetResult{
			Target:     target,
			PingMs:     latency,
			PacketLoss: func() float64 { if success { return 0 }; return 100 }(),
			Success:    success,
			Timestamp:  time.Now(),
		})
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Rolling loss window
	anySuccess := successCount > 0
	p.lossWindow = append(p.lossWindow, anySuccess)
	if len(p.lossWindow) > 30 {
		p.lossWindow = p.lossWindow[1:]
	}

	lostPings := 0
	for _, ok := range p.lossWindow {
		if !ok {
			lostPings++
		}
	}
	lossRate := (float64(lostPings) / float64(len(p.lossWindow))) * 100.0

	// Rolling RFC 3550 Jitter calculation: J_i = J_{i-1} + (|D(i-1, i)| - J_{i-1}) / 16
	avgPing := 0.0
	if successCount > 0 {
		avgPing = totalLatency / float64(successCount)
		if primaryPing > 0 {
			avgPing = primaryPing
		}

		if len(p.recentPings) > 0 {
			prev := p.recentPings[len(p.recentPings)-1]
			diff := math.Abs(avgPing - prev)
			p.currentJitter = p.currentJitter + (diff-p.currentJitter)/16.0
		}
		p.recentPings = append(p.recentPings, avgPing)
		if len(p.recentPings) > 30 {
			p.recentPings = p.recentPings[1:]
		}
	}

	isDegraded := false
	degradedReason := ""

	if lossRate > 5.0 {
		isDegraded = true
		degradedReason = "High Packet Loss (>5%)"
	} else if p.currentJitter > 15.0 {
		isDegraded = true
		degradedReason = "Severe RFC Jitter Spikes (>15ms)"
	} else if avgPing > 80.0 {
		isDegraded = true
		degradedReason = "Elevated WAN Latency (>80ms)"
	}

	p.vitals = NetworkVitals{
		Timestamp:      time.Now(),
		PingMs:         avgPing,
		JitterMs:       p.currentJitter,
		PacketLoss:     lossRate,
		PrimaryTarget:  p.Targets[0],
		Targets:        targetResults,
		IsDegraded:     isDegraded,
		DegradedReason: degradedReason,
	}
}

// GetVitals returns a point-in-time snapshot of rolling network vitals.
func (p *Pinger) GetVitals() NetworkVitals {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.vitals
}
