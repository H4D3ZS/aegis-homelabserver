package pihole

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/collector"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/db"
)

var knownMediaCDNs = []string{
	"kwik.cx",
	"doodstream.com",
	"mp4upload.com",
	"vidsrc.me",
	"streamtape.com",
	"animeflv.net",
	"crunchyroll.com",
	"netflix.com",
}

// UnbreakRule represents an active temporary whitelist bypass.
type UnbreakRule struct {
	Domain    string    `json:"domain"`
	ClientIP  string    `json:"client_ip"`
	ExpiresAt time.Time `json:"expires_at"`
	Timer     *time.Timer `json:"-"`
}

// UnbreakEngine manages 1-Click Smart Unbreak with 15-minute auto-eviction.
type UnbreakEngine struct {
	Client       *Client
	Store        *db.Store
	Tailer       *collector.LogTailer
	activeRules  map[string]*UnbreakRule
	mu           sync.RWMutex
	RuleTTL      time.Duration
}

// NewUnbreakEngine initializes the unbreaker.
func NewUnbreakEngine(client *Client, store *db.Store, tailer *collector.LogTailer) *UnbreakEngine {
	return &UnbreakEngine{
		Client:      client,
		Store:       store,
		Tailer:      tailer,
		activeRules: make(map[string]*UnbreakRule),
		RuleTTL:     15 * time.Minute,
	}
}

// UnbreakLastBlocked scans recent blocked CDN queries for the requesting client and temporarily whitelists them.
func (u *UnbreakEngine) UnbreakLastBlocked(ctx context.Context, clientIP string) (map[string]interface{}, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	var unblockedDomains []string

	if u.Tailer != nil {
		events := u.Tailer.GetBlockedEventsForIP(clientIP, 120*time.Second)
		for _, ev := range events {
			domain := ev.Domain
			if u.isMediaCDN(domain) && u.activeRules[domain] == nil {
				_ = u.Client.WhitelistDomain(ctx, domain)
				unblockedDomains = append(unblockedDomains, domain)
				u.scheduleEviction(domain, clientIP)
			}
		}
	}

	if len(unblockedDomains) == 0 {
		defaultCDNs := []string{"kwik.cx", "doodstream.com", "mp4upload.com"}
		for _, d := range defaultCDNs {
			if u.activeRules[d] == nil {
				_ = u.Client.WhitelistDomain(ctx, d)
				unblockedDomains = append(unblockedDomains, d)
				u.scheduleEviction(d, clientIP)
			}
		}
	}

	_ = u.Client.ReloadLists(ctx)

	return map[string]interface{}{
		"success":           true,
		"unblocked_domains": unblockedDomains,
		"ttl_minutes":       15,
		"message":           fmt.Sprintf("Temporarily whitelisted %d media streams for 15 minutes", len(unblockedDomains)),
	}, nil
}

func (u *UnbreakEngine) isMediaCDN(domain string) bool {
	dLower := strings.ToLower(domain)
	for _, cdn := range knownMediaCDNs {
		if strings.Contains(dLower, cdn) {
			return true
		}
	}
	return false
}

func (u *UnbreakEngine) scheduleEviction(domain, clientIP string) {
	expiresAt := time.Now().Add(u.RuleTTL)

	timer := time.AfterFunc(u.RuleTTL, func() {
		u.mu.Lock()
		delete(u.activeRules, domain)
		u.mu.Unlock()

		log.Printf("[UNBREAK] 15m TTL expired for domain: %s. Auto-evicting whitelist...", domain)
		_ = u.Client.RemoveWhitelist(context.Background(), domain)
		_ = u.Client.ReloadLists(context.Background())
	})

	u.activeRules[domain] = &UnbreakRule{
		Domain:    domain,
		ClientIP:  clientIP,
		ExpiresAt: expiresAt,
		Timer:     timer,
	}
}
