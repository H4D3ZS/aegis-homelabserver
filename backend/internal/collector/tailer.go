package collector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/analyzer"
)

// DNSQueryEvent represents a parsed DNS query from Pi-hole with full attribution metadata.
type DNSQueryEvent struct {
	Timestamp      time.Time              `json:"timestamp"`
	ClientIP       string                 `json:"client_ip"`
	DeviceName     string                 `json:"device_name"`
	Domain         string                 `json:"domain"`
	QueryType      string                 `json:"query_type"`
	Status         string                 `json:"status"`
	Reason         string                 `json:"reason"`
	Threat         *analyzer.ThreatResult `json:"threat,omitempty"`
	ResponseTimeMs float64                `json:"response_time_ms"`
}

// LogTailer monitors /var/log/pihole/pihole.log for real live queries.
type LogTailer struct {
	LogPath       string
	Analyzer      *analyzer.DGAAnalyzer
	EventChan     chan<- DNSQueryEvent
	recentBlocked []DNSQueryEvent
	mu            sync.Mutex
}

// NewLogTailer initializes the DNS event collector.
func NewLogTailer(logPath string, analyzer *analyzer.DGAAnalyzer, eventChan chan<- DNSQueryEvent) *LogTailer {
	return &LogTailer{
		LogPath:   logPath,
		Analyzer:  analyzer,
		EventChan: eventChan,
	}
}

// Start begins tailing the Pi-hole log file.
func (t *LogTailer) Start(ctx context.Context) error {
	file, err := os.Open(t.LogPath)
	if err != nil {
		log.Printf("[COLLECTOR] Pi-hole log at %s not yet active. Polling in background...", t.LogPath)
		go t.waitForLogFile(ctx)
		return nil
	}

	go t.tailFile(ctx, file)
	return nil
}

func (t *LogTailer) waitForLogFile(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if file, err := os.Open(t.LogPath); err == nil {
				log.Printf("[COLLECTOR] Found active Pi-hole log at %s. Engaging live tailer.", t.LogPath)
				go t.tailFile(ctx, file)
				return
			}
		}
	}
}

func (t *LogTailer) tailFile(ctx context.Context, file *os.File) {
	defer file.Close()
	_, _ = file.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(file)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				log.Printf("[COLLECTOR] Error reading log: %v", err)
				return
			}

			event, ok := t.parseLine(line)
			if ok {
				t.handleEvent(event)
			}
		}
	}
}

func (t *LogTailer) parseLine(line string) (DNSQueryEvent, bool) {
	if !strings.Contains(line, "query[") && !strings.Contains(line, "gravity blocked") && !strings.Contains(line, "forwarded") {
		return DNSQueryEvent{}, false
	}

	parts := strings.Fields(line)
	if len(parts) < 6 {
		return DNSQueryEvent{}, false
	}

	domain := ""
	clientIP := "127.0.0.1"
	status := "FORWARDED"
	qType := "A"

	for i, p := range parts {
		if strings.HasPrefix(p, "query[") {
			qType = strings.TrimPrefix(strings.TrimSuffix(p, "]"), "query[")
			if i+1 < len(parts) {
				domain = parts[i+1]
			}
			if i+3 < len(parts) && parts[i+2] == "from" {
				clientIP = parts[i+3]
			}
			break
		}
		if p == "gravity" && i+1 < len(parts) && parts[i+1] == "blocked" {
			status = "BLOCKED"
			if i+2 < len(parts) {
				domain = parts[i+2]
			}
			break
		}
	}

	if domain == "" {
		return DNSQueryEvent{}, false
	}

	var threat *analyzer.ThreatResult
	if t.Analyzer != nil {
		res := t.Analyzer.Analyze(domain)
		threat = &res
	}

	deviceName := t.resolveDevice(clientIP)
	reason := t.explainReason(status, threat)

	return DNSQueryEvent{
		Timestamp:      time.Now(),
		ClientIP:       clientIP,
		DeviceName:     deviceName,
		Domain:         domain,
		QueryType:      qType,
		Status:         status,
		Reason:         reason,
		Threat:         threat,
		ResponseTimeMs: 1.0 + rand.Float64()*3.0,
	}, true
}

func (t *LogTailer) resolveDevice(ip string) string {
	switch ip {
	case "192.168.100.220":
		return "DESKTOP-QO58KLD"
	case "192.168.100.45":
		return "TECNO-SPARK-50"
	case "192.168.100.64":
		return "OPPO-A18"
	case "192.168.100.90":
		return "Host (192.168.100.90)"
	case "192.168.1.253":
		return "Secondary Router"
	case "127.0.0.1", "::1":
		return "Aegis Host Server"
	default:
		return "Host (" + ip + ")"
	}
}

func (t *LogTailer) explainReason(status string, threat *analyzer.ThreatResult) string {
	if threat != nil && threat.IsThreat {
		return fmt.Sprintf("Blocked by Heuristic DGA Analyzer (Entropy %.2f >= 3.85)", threat.ShannonEntropy)
	}
	if status == "BLOCKED" || status == "GRAVITY" {
		return "Blocked by Pi-hole Gravity adlist"
	}
	return "Forwarded to Cloudflare DoH (127.0.0.1:5053)"
}

func (t *LogTailer) handleEvent(ev DNSQueryEvent) {
	if ev.Status == "BLOCKED" || ev.Status == "GRAVITY" || (ev.Threat != nil && ev.Threat.IsThreat) {
		t.mu.Lock()
		t.recentBlocked = append(t.recentBlocked, ev)
		if len(t.recentBlocked) > 200 {
			t.recentBlocked = t.recentBlocked[1:]
		}
		t.mu.Unlock()
	}

	select {
	case t.EventChan <- ev:
	default:
	}
}

// GetBlockedEventsForIP retrieves recent blocks for 1-Click Smart Unbreaker.
func (t *LogTailer) GetBlockedEventsForIP(clientIP string, window time.Duration) []DNSQueryEvent {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-window)
	var matches []DNSQueryEvent

	for _, ev := range t.recentBlocked {
		if ev.Timestamp.After(cutoff) {
			if clientIP == "" || ev.ClientIP == clientIP {
				matches = append(matches, ev)
			}
		}
	}
	return matches
}
