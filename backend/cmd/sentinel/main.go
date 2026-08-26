package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/analyzer"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/collector"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/db"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/handler"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/isp"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/pihole"
	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/router"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

func main() {
	port := flag.String("port", getEnv("PORT", "3001"), "HTTP service port (default: 3001)")
	dbPath := flag.String("db", getEnv("DB_PATH", "/var/lib/pi-sentinel/sentinel.db"), "SQLite database path")
	logPath := flag.String("log-path", getEnv("PIHOLE_LOG", "/var/log/pihole/pihole.log"), "Pi-hole log file path")
	safesearchConf := flag.String("safesearch-conf", getEnv("SAFESEARCH_CONF", "/etc/dnsmasq.d/05-safesearch.conf"), "SafeSearch config path")
	piholeBin := flag.String("pihole-bin", getEnv("PIHOLE_BIN", "/usr/local/bin/pihole"), "Pi-hole executable path")
	slaDown := flag.Float64("sla-down", getEnvFloat("SLA_DOWN_MBPS", 500.0), "Contracted SLA download Mbps")
	slaUp := flag.Float64("sla-up", getEnvFloat("SLA_UP_MBPS", 100.0), "Contracted SLA upload Mbps")
	ispName := flag.String("isp-name", getEnv("ISP_NAME", "Converge ICT FiberX"), "ISP provider name")

	// Huawei ONT Automation
	routerURL := flag.String("router-url", getEnv("ROUTER_URL", "http://192.168.100.1"), "Router web interface URL")
	routerUser := flag.String("router-user", getEnv("ROUTER_USER", "root"), "Router admin username")
	routerPass := flag.String("router-pass", getEnv("ROUTER_PASS", "admin 123"), "Router admin password")
	routerType := flag.String("router-type", getEnv("ROUTER_TYPE", "huawei_ont"), "Router driver: huawei_ont")
	routerAutoHeal := flag.Bool("router-auto-heal", getEnvBool("ROUTER_AUTO_HEAL", true), "Enable 120s WAN auto-reboot watchdog")

	flag.Parse()

	log.Printf("[AEGIS-SENTINEL] Starting Bare-Metal Sentinel Daemon on :%s for x86_64 Ubuntu Server...", *port)

	// 1. SQLite Store
	store, _ := db.NewStore(*dbPath)

	// 2. DGA Analyzer
	dgaAnalyzer := analyzer.NewDGAAnalyzer(0.75)

	// 3. Pi-hole CLI Client & SafeSearch
	piholeClient := pihole.NewClient(*piholeBin)
	safeSearch := pihole.NewSafeSearchManager(*safesearchConf)

	// 4. DNS Collector & 1-Click Unbreak Engine
	dnsEvents := make(chan collector.DNSQueryEvent, 100)
	tailer := collector.NewLogTailer(*logPath, dgaAnalyzer, dnsEvents)
	unbreakEngine := pihole.NewUnbreakEngine(piholeClient, store, tailer)

	// 5. Huawei ONT Session Manager & Watchdog
	routerCfg := router.RouterConfig{
		GatewayURL:     *routerURL,
		Username:       *routerUser,
		Password:       *routerPass,
		RouterType:     *routerType,
		AutoHeal:       *routerAutoHeal,
		HealTimeoutSec: 120,
	}
	routerClient := router.NewRouterClient(routerCfg)

	// 6. ISP Prober & Speedtest Watchdog
	pinger := isp.NewPinger([]string{"1.1.1.1:53", "8.8.8.8:53", "192.168.100.1:53"})
	speedtest := isp.NewSpeedtestRunner(store, nil, *slaDown, *slaUp, *ispName)

	// 7. Real-Time SSE Hub
	sseHub := handler.NewSSEHub(pinger, speedtest)

	// 8. Start Background Watchdogs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sseHub.Run()
	go pinger.Start(ctx)
	go speedtest.StartCron(ctx)
	_ = tailer.Start(ctx)

	// Threat Neutralization Loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-dnsEvents:
				sseHub.BroadcastQuery(ev)
				if ev.Threat != nil && ev.Threat.IsThreat {
					blockCtx, bCancel := context.WithTimeout(context.Background(), 5*time.Second)
					_ = piholeClient.BlacklistDomain(blockCtx, ev.Domain)
					_ = piholeClient.ReloadLists(blockCtx)
					bCancel()
				}
			}
		}
	}()

	// 9. Register Routes
	mux := http.NewServeMux()
	api := handler.NewAPIHandler(
		store,
		pinger,
		speedtest,
		unbreakEngine,
		piholeClient,
		routerClient,
		safeSearch,
		tailer,
		dgaAnalyzer,
		sseHub,
	)
	api.RegisterRoutes(mux)

	// Serve the Beautiful Friendly All-in-One Dashboard
	mux.Handle("/", handler.ServeStaticUI(nil, ""))

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		cancel()
		shutdownCtx, sCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer sCancel()
		_ = server.Shutdown(shutdownCtx)
		os.Exit(0)
	}()

	log.Printf("[AEGIS-SENTINEL] Sentinel Core live on http://0.0.0.0:%s", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[SERVER] Fatal HTTP server error: %v", err)
	}
}
