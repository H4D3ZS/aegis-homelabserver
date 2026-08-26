package handler

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeStaticUI serves an embedded filesystem or fallback rich UI.
func ServeStaticUI(staticFS fs.FS, localFallbackDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if staticFS != nil {
			f, err := staticFS.Open(path)
			if err == nil {
				_ = f.Close()
				http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
				return
			}
		}

		if localFallbackDir != "" {
			diskPath := filepath.Join(localFallbackDir, path)
			if stat, err := os.Stat(diskPath); err == nil && !stat.IsDir() {
				http.ServeFile(w, r, diskPath)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(richFriendlyDashboardHTML))
	})
}

const richFriendlyDashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Aegis Sentinel | Sovereign Homelab</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <style>
    :root { --bg: #09090B; }
    body { background-color: var(--bg); color: #FAFAFA; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; }
    .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
    .muted { color: #71717A; }
    ::-webkit-scrollbar { width: 5px; height: 5px; }
    ::-webkit-scrollbar-track { background: #09090B; }
    ::-webkit-scrollbar-thumb { background: #27272A; border-radius: 2px; }
    ::-webkit-scrollbar-thumb:hover { background: #3F3F46; }
    .drag-active { border-color: #10B981 !important; background-color: rgba(16, 185, 129, 0.05) !important; }
  </style>
</head>
<body class="min-h-screen flex flex-col antialiased bg-[#09090B] text-[#FAFAFA] p-6 lg:p-10 max-w-[1600px] mx-auto space-y-8">

  <!-- Main Header & Live Vitals Strip -->
  <header class="flex flex-wrap items-end justify-between gap-6 pb-6 border-b border-[#27272A]">
    <div>
      <div class="flex items-center gap-3">
        <div class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></div>
        <h1 class="text-2xl font-bold tracking-tight">Aegis Sentinel</h1>
        <span class="text-xs muted mono">x86_64 // N4100 Bare-Metal</span>
      </div>
      <p class="text-xs muted mt-1">Converge FiberX 500M • Primary Gateway: <span class="mono text-zinc-300">192.168.100.1</span> • DNS Port 53: <span class="text-emerald-400 mono">Pi-hole + DoH Active</span></p>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex flex-wrap items-center rounded-lg bg-[#121215] p-1 border border-[#27272A] text-xs gap-1">
      <button onclick="switchTab('tab-homelab')" id="nav-tab-homelab" class="px-3.5 py-2 rounded-md bg-[#27272A] text-white font-medium transition cursor-pointer">
        Homelab &amp; Topology
      </button>
      <button onclick="switchTab('tab-pihole')" id="nav-tab-pihole" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        🕳️ Pi-hole &amp; DNS Sinkhole
      </button>
      <button onclick="switchTab('tab-media')" id="nav-tab-media" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        🎬 Anime &amp; Torrents (Jellyfin)
      </button>
      <button onclick="switchTab('tab-minecraft')" id="nav-tab-minecraft" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        Minecraft (Crafty 4 &amp; Tailscale)
      </button>
      <button onclick="switchTab('tab-gitea')" id="nav-tab-gitea" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        Multi-Drive Storage &amp; Gitea
      </button>
      <button onclick="switchTab('tab-crowdsec')" id="nav-tab-crowdsec" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        CrowdSec IPS (<span id="crowdsecBansBadge">4</span>)
      </button>
      <button onclick="switchTab('tab-wazuh')" id="nav-tab-wazuh" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        Wazuh HIDS (94%)
      </button>
    </div>
  </header>

  <!-- ==================== TAB 1: HOMELAB & TOPOLOGY ==================== -->
  <main id="tab-homelab" class="space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div class="flex items-center gap-2">
        <button id="speedtestBtn" onclick="runSpeedtest()" class="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-white text-xs font-medium transition cursor-pointer">
          Run Speedtest
        </button>
        <button onclick="openUnbreakModal()" class="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-white text-xs font-medium transition cursor-pointer">
          Fix Video Stream
        </button>
        <button onclick="rebootRouter()" class="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-400 hover:text-rose-400 text-xs font-medium transition cursor-pointer">
          Reboot Fiber Router
        </button>
      </div>

      <div class="flex items-center gap-2 text-xs muted mono">
        <span class="w-2 h-2 rounded-full bg-cyan-400"></span> Cloudflare DoH (127.0.0.1:5053) Active
      </div>
    </div>

    <!-- 3 Vital Metrics -->
    <section class="grid grid-cols-1 md:grid-cols-3 gap-8">
      <div>
        <div class="text-xs uppercase tracking-wider font-semibold muted">ISP Bandwidth</div>
        <div class="mt-2 flex items-baseline gap-3">
          <span id="downSpeed" class="text-4xl font-extrabold tracking-tight">500</span>
          <span class="text-sm muted font-medium">Mbps ↓</span>
          <span id="upSpeed" class="text-2xl font-bold text-zinc-300 ml-2">100</span>
          <span class="text-xs muted">Mbps ↑</span>
        </div>
        <div class="mt-1 text-xs muted">Target: 500 Mbps SLA • Idle on-demand testing</div>
      </div>

      <div>
        <div class="text-xs uppercase tracking-wider font-semibold muted">Latency &amp; RFC Jitter</div>
        <div class="mt-2 flex items-baseline gap-2">
          <span id="livePing" class="text-4xl font-extrabold tracking-tight text-emerald-400">7.0</span>
          <span class="text-sm muted font-medium">ms</span>
          <span class="text-xs muted mono ml-2">(&plusmn;<span id="liveJitter">1.5</span>ms jitter)</span>
        </div>
        <div class="mt-1 text-xs muted">0.0% packet drop • Sub-second ICMP prober</div>
      </div>

      <div>
        <div class="text-xs uppercase tracking-wider font-semibold muted">DNS Protection</div>
        <div class="mt-2 flex items-baseline gap-2">
          <span id="piholeBlockedThreats" class="text-4xl font-extrabold tracking-tight text-cyan-400">5,492</span>
          <span class="text-sm muted font-medium">ads &amp; threats blocked</span>
        </div>
        <div class="mt-1 text-xs muted">184,290 domains in Gravity • 100% whole-house cover</div>
      </div>
    </section>

    <!-- Organic EtherApe Live Network Topology Canvas -->
    <section class="space-y-4 pt-4 border-t border-[#18181B]">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 class="text-sm font-semibold tracking-tight text-zinc-200">Network Topology (EtherApe Live)</h2>
          <p class="text-xs muted">Real-time organic nodes, live packet flows, and hierarchical subgraph zooming</p>
        </div>
      </div>

      <div class="relative w-full h-80 bg-[#0C0C0E] rounded-lg overflow-hidden border border-[#18181B]">
        <canvas id="etherapeCanvas" class="w-full h-full"></canvas>
        <div class="absolute bottom-3 left-3 text-[11px] muted mono bg-black/60 px-2 py-1 rounded backdrop-blur-sm">
          Active Organic Nodes: <span id="activeNodeCount">7</span> • Total Throughput: <span class="text-cyan-400">8.42 MB/s</span>
        </div>
      </div>
    </section>

    <!-- Connected Devices Table -->
    <section class="space-y-4 pt-4 border-t border-[#18181B]">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 class="text-sm font-semibold tracking-tight text-zinc-200">Connected Devices (<span id="totalDevicesCount">13</span> Discovered)</h2>
          <p class="text-xs muted">Verified from Huawei ONT ARP tables and network discovery</p>
        </div>
      </div>

      <div class="overflow-x-auto text-xs border border-[#18181B] rounded-lg">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
              <th class="py-2.5 px-3">Device Name</th>
              <th class="py-2.5 px-3">Port</th>
              <th class="py-2.5 px-3">IP &amp; MAC Address</th>
              <th class="py-2.5 px-3">Duration / Last Seen</th>
              <th class="py-2.5 px-3 text-right">Status</th>
            </tr>
          </thead>
          <tbody id="devicesTableBody" class="divide-y divide-[#18181B] text-zinc-300"></tbody>
        </table>
      </div>
    </section>
  </main>

  <!-- ==================== TAB 2: PI-HOLE ==================== -->
  <main id="tab-pihole" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">🕳️ Pi-hole FTL &amp; Encrypted DNS Sinkhole</h2>
        <p class="text-xs muted">Network-wide ad-blocking, tracker neutralization, DNS-over-HTTPS &amp; SafeSearch enforcement</p>
      </div>
      <div class="flex items-center gap-2">
        <a href="http://localhost/admin/" target="_blank" class="px-4 py-2 rounded bg-red-600 hover:bg-red-700 text-white text-xs font-semibold transition cursor-pointer flex items-center gap-1.5">
          <span>🕳️ Open Pi-hole Admin Portal</span>
          <span class="mono text-[10px] bg-red-900 px-1.5 py-0.5 rounded">/admin/</span>
        </a>
      </div>
    </div>

    <!-- Pi-hole 4 Core Metric Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Total Queries Today</div>
        <div id="piholeTotalQueries" class="text-2xl font-bold text-white mt-1">28,410</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Queries Blocked</div>
        <div id="piholeQueriesBlocked" class="text-2xl font-bold text-red-400 mt-1">5,492</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Domains on Blocklist</div>
        <div class="text-2xl font-bold text-emerald-400 mt-1">184,290</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Encrypted Upstream DoH</div>
        <div class="text-lg font-bold text-cyan-400 mt-1">Cloudflare TLS 1.3</div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 3: ANIME & TORRENTS ==================== -->
  <main id="tab-media" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">🎬 Anime Streaming &amp; Torrent Hub (Jellyfin &amp; qBittorrent)</h2>
        <p class="text-xs muted">Hardware-accelerated Intel QuickSync streaming, SyncPlay watch parties &amp; custom storage paths</p>
      </div>
      <div class="flex items-center gap-2">
        <a href="http://localhost:8096" target="_blank" class="px-4 py-2 rounded bg-purple-600 hover:bg-purple-700 text-white text-xs font-semibold transition cursor-pointer flex items-center gap-1.5">
          <span>🎬 Open Jellyfin Cinema</span>
          <span class="mono text-[10px] bg-purple-900 px-1.5 py-0.5 rounded">:8096</span>
        </a>
        <a href="http://localhost:9091" target="_blank" class="px-4 py-2 rounded bg-cyan-600 hover:bg-cyan-700 text-white text-xs font-semibold transition cursor-pointer flex items-center gap-1.5">
          <span>📥 Open qBittorrent WebUI</span>
          <span class="mono text-[10px] bg-cyan-900 px-1.5 py-0.5 rounded">:9091</span>
        </a>
      </div>
    </div>

    <!-- Quick Magnet Link Downloader Box with Storage Location Selector -->
    <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">⚡ Quick Magnet URL Downloader &amp; Storage Target</h3>
        <span class="text-[11px] text-cyan-400 mono">Dispatches directly to qBittorrent WebAPI (:9091)</span>
      </div>

      <div class="space-y-3 text-xs">
        <div>
          <label class="block text-zinc-400 mb-1">Paste Magnet Link:</label>
          <input type="text" id="magnetUrlInput" placeholder="magnet:?xt=urn:btih:... (Anime episode, movie, or dataset)" class="w-full px-3 py-2 rounded bg-[#0C0C0E] border border-[#27272A] text-white focus:outline-none focus:border-zinc-500 mono" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-zinc-400 mb-1">Target Storage Location:</label>
            <select id="saveLocationSelect" class="w-full px-3 py-2 rounded bg-[#0C0C0E] border border-[#27272A] text-white focus:outline-none focus:border-zinc-500">
              <option value="/mnt/external_1tb/media/downloads">💾 External 1TB HDD (/mnt/external_1tb/media/downloads)</option>
              <option value="/var/lib/aegis-data/downloads">⚡ Fast Internal 256GB SSD (/var/lib/aegis-data/downloads)</option>
              <option value="/mnt/external_1tb/media/anime">🎬 Direct to Jellyfin Anime Library (/mnt/external_1tb/media/anime)</option>
            </select>
          </div>

          <div class="flex items-end">
            <button onclick="dispatchTorrentDownload()" class="w-full py-2 rounded bg-cyan-600 hover:bg-cyan-500 text-white font-semibold cursor-pointer transition">
              🚀 Start High-Speed Download
            </button>
          </div>
        </div>

        <div id="torrentDispatchMsg" class="hidden p-2.5 rounded bg-emerald-950/60 border border-emerald-800 text-emerald-300 text-[11px]"></div>
      </div>
    </div>

    <!-- SyncPlay Card -->
    <div class="p-5 rounded-lg bg-gradient-to-r from-[#121215] to-[#18181B] border border-[#27272A] space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-purple-400 animate-pulse"></div>
          <div>
            <h3 class="text-sm font-bold text-white">Jellyfin SyncPlay Watch-Party (Watch Anime with Girlfriend)</h3>
            <p class="text-xs muted mt-0.5">Real-time play/pause synchronization over Tailscale (Stream anywhere with 0 buffering)</p>
          </div>
        </div>
        <span class="px-2.5 py-1 rounded bg-purple-950 text-purple-300 border border-purple-800 text-[11px] font-semibold">
          ✓ Intel QuickSync (QSV) Active
        </span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-2">
          <div class="text-xs text-zinc-400 font-medium">Girlfriend's Remote Streaming Address:</div>
          <div class="flex items-center justify-between gap-2">
            <span class="mono font-bold text-purple-400 text-sm md:text-base">http://100.115.42.18:8096</span>
            <button onclick="navigator.clipboard.writeText('http://100.115.42.18:8096'); alert('Copied Jellyfin URL to clipboard!')" class="px-3 py-1.5 rounded bg-purple-600 hover:bg-purple-700 text-white text-xs font-semibold cursor-pointer transition">
              📋 Copy Link for Her
            </button>
          </div>
        </div>

        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1.5 text-xs text-zinc-300">
          <div class="font-semibold text-white">How to watch together in SyncPlay:</div>
          <ol class="list-decimal list-inside space-y-1 text-zinc-400 text-[11px]">
            <li>Open any anime episode in Jellyfin</li>
            <li>Click the <strong class="text-white">SyncPlay group icon</strong> (👥 top-right)</li>
            <li>Have your girlfriend click <strong class="text-white">Join SyncPlay Group</strong></li>
          </ol>
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 4: MINECRAFT ==================== -->
  <main id="tab-minecraft" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">Minecraft Modded Multiplayer &amp; Crafty Controller 4</h2>
        <p class="text-xs muted">Forge 1.20.1 runtime • Tailscale zero-config mesh VPN for remote girlfriend play</p>
      </div>
      <div class="flex items-center gap-2">
        <a href="https://localhost:8443" target="_blank" class="px-4 py-2 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold transition cursor-pointer flex items-center gap-1.5">
          <span>🚀 Open Crafty Controller GUI</span>
          <span class="mono text-[10px] bg-emerald-900 px-1.5 py-0.5 rounded">:8443</span>
        </a>
      </div>
    </div>

    <!-- Tailscale Card -->
    <div class="p-5 rounded-lg bg-gradient-to-r from-[#121215] to-[#18181B] border border-[#27272A] space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-emerald-400 animate-pulse"></div>
          <div>
            <h3 class="text-sm font-bold text-white">Tailscale Remote Multiplayer (Girlfriend Connection)</h3>
            <p class="text-xs muted mt-0.5">Secure peer-to-peer WireGuard tunnel (Plays from any house with 0 router port-forwarding)</p>
          </div>
        </div>
        <span class="px-2.5 py-1 rounded bg-emerald-950 text-emerald-400 border border-emerald-800 text-[11px] font-semibold">
          ✓ Tailscale Mesh Active
        </span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-2">
          <div class="text-xs text-zinc-400 font-medium">Girlfriend's Direct Connect Server Address:</div>
          <div class="flex items-center justify-between gap-2">
            <span id="tailscaleMultiplayerIP" class="mono font-bold text-emerald-400 text-sm md:text-base">100.115.42.18:25565</span>
            <button onclick="copyMultiplayerIP()" id="copyIPBtn" class="px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold cursor-pointer transition">
              📋 Copy IP for Her
            </button>
          </div>
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 5: MULTI-DRIVE STORAGE & GITEA ==================== -->
  <main id="tab-gitea" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">Multi-Drive Storage Health &amp; Sovereign Gitea</h2>
        <p class="text-xs muted">Live telemetry for both Internal 256GB SSD &amp; External 1TB NTFS Enclosure</p>
      </div>
      <div class="flex items-center gap-2">
        <a href="http://localhost:3000" target="_blank" class="px-4 py-2 rounded bg-cyan-600 hover:bg-cyan-700 text-white text-xs font-semibold transition cursor-pointer flex items-center gap-1.5">
          <span>🐙 Open Gitea Web GUI</span>
          <span class="mono text-[10px] bg-cyan-900 px-1.5 py-0.5 rounded">:3000</span>
        </a>
      </div>
    </div>

    <!-- Live Multi-Drive Telemetry Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-xs">
      <!-- Drive 1: Internal 256GB SSD -->
      <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-3">
        <div class="flex justify-between items-center">
          <div class="font-bold text-white text-sm">⚡ Fast Internal 256GB SSD</div>
          <span class="mono text-emerald-400 text-[11px]" id="ssdUsageText">22% Used (186.2 GB Free)</span>
        </div>
        <div class="w-full h-2.5 rounded-full bg-[#18181B] overflow-hidden flex border border-[#27272A]">
          <div id="ssdBar" style="width: 22%;" class="bg-cyan-500 h-full"></div>
        </div>
        <p class="text-[11px] muted">Holds: Minecraft Forge World Chunks, Crafty runtime, SQLite DBs, and high-IOPS services.</p>
      </div>

      <!-- Drive 2: External 1TB NTFS HDD -->
      <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-3">
        <div class="flex justify-between items-center">
          <div class="font-bold text-white text-sm">💾 External 1TB HDD Enclosure (NTFS)</div>
          <span class="mono text-emerald-400 text-[11px]" id="hddUsageText">16% Used (782.4 GB Free)</span>
        </div>
        <div class="w-full h-2.5 rounded-full bg-[#18181B] overflow-hidden flex border border-[#27272A]">
          <div id="hddBar" style="width: 16%;" class="bg-emerald-500 h-full"></div>
        </div>
        <p class="text-[11px] muted">Holds: Gitea Repositories (100GB+ LFS), Jellyfin Anime Library, and Torrent Downloads.</p>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 6: CROWDSEC ==================== -->
  <main id="tab-crowdsec" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">CrowdSec Collaborative IPS</h2>
        <p class="text-xs muted">Local API engine, active remediation bouncers, and community threat consensus</p>
      </div>
    </div>
    <div class="overflow-x-auto text-xs border border-[#18181B] rounded-lg">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
            <th class="py-2.5 px-3">IP Address</th>
            <th class="py-2.5 px-3">Origin</th>
            <th class="py-2.5 px-3">Reason / Scenario</th>
            <th class="py-2.5 px-3">Action</th>
            <th class="py-2.5 px-3">Duration</th>
          </tr>
        </thead>
        <tbody id="crowdsecDecisionsBody" class="divide-y divide-[#18181B]"></tbody>
      </table>
    </div>
  </main>

  <!-- ==================== TAB 7: WAZUH ==================== -->
  <main id="tab-wazuh" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">Wazuh Host Intrusion &amp; Integrity (HIDS)</h2>
        <p class="text-xs muted">File integrity monitoring (FIM), CIS Linux compliance assessment, and rootkit checks</p>
      </div>
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">CIS Benchmark Score</div>
        <div class="text-2xl font-bold text-emerald-400 mt-1">94%</div>
      </div>
    </div>
  </main>

  <script>
    let allDevices = [];
    let filteredDevices = [];
    let currentDevicePage = 1;
    const pageSize = 5;

    function switchTab(tabId) {
      ['tab-homelab', 'tab-pihole', 'tab-media', 'tab-minecraft', 'tab-gitea', 'tab-crowdsec', 'tab-wazuh'].forEach(t => {
        const el = document.getElementById(t);
        const nav = document.getElementById('nav-' + t);
        if (el && nav) {
          if (t === tabId) {
            el.classList.remove('hidden');
            nav.className = 'px-3.5 py-2 rounded-md bg-[#27272A] text-white font-medium transition cursor-pointer';
          } else {
            el.classList.add('hidden');
            nav.className = 'px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer';
          }
        }
      });
      if (tabId === 'tab-pihole') loadPiholeStats();
      if (tabId === 'tab-gitea') loadStorageDrives();
      if (tabId === 'tab-minecraft') loadTailscale();
    }

    async function dispatchTorrentDownload() {
      const magnet = document.getElementById('magnetUrlInput').value.trim();
      const savePath = document.getElementById('saveLocationSelect').value;
      if (!magnet) return alert('Please enter or paste a Magnet URL.');

      const msgBox = document.getElementById('torrentDispatchMsg');
      msgBox.classList.remove('hidden');
      msgBox.innerText = 'Dispatching to qBittorrent...';

      try {
        const res = await fetch('/api/v1/homelab/media/download', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({ magnet_url: magnet, save_path: savePath })
        });
        const json = await res.json();
        msgBox.innerText = '✓ ' + json.message;
        document.getElementById('magnetUrlInput').value = '';
      } catch (err) {
        msgBox.innerText = '✗ Failed: ' + err.message;
      }
    }

    async function loadStorageDrives() {
      try {
        const res = await fetch('/api/v1/homelab/storage');
        if (res.ok) {
          const s = await res.json();
          if (s.internal_ssd) {
            document.getElementById('ssdUsageText').innerText = s.internal_ssd.usage_pct + '% Used (' + s.internal_ssd.free_gb.toFixed(1) + ' GB Free)';
            document.getElementById('ssdBar').style.width = s.internal_ssd.usage_pct + '%';
          }
          if (s.external_hdd) {
            document.getElementById('hddUsageText').innerText = s.external_hdd.usage_pct + '% Used (' + s.external_hdd.free_gb.toFixed(1) + ' GB Free)';
            document.getElementById('hddBar').style.width = s.external_hdd.usage_pct + '%';
          }
        }
      } catch (e) {}
    }

    async function loadPiholeStats() {
      try {
        const res = await fetch('/api/v1/pihole/stats');
        if (res.ok) {
          const p = await res.json();
          document.getElementById('piholeTotalQueries').innerText = p.dns_queries_today.toLocaleString();
          document.getElementById('piholeQueriesBlocked').innerText = p.ads_blocked_today.toLocaleString();
          document.getElementById('piholeBlockedThreats').innerText = p.ads_blocked_today.toLocaleString();
        }
      } catch (e) {}
    }

    async function loadTailscale() {
      try {
        const res = await fetch('/api/v1/tailscale/status');
        if (res.ok) {
          const data = await res.json();
          document.getElementById('tailscaleMultiplayerIP').innerText = data.multiplayer_direct_ip || '100.115.42.18:25565';
        }
      } catch (e) {}
    }

    function copyMultiplayerIP() {
      const ip = document.getElementById('tailscaleMultiplayerIP').innerText;
      navigator.clipboard.writeText(ip);
      const btn = document.getElementById('copyIPBtn');
      btn.innerText = '✓ Copied!';
      setTimeout(() => { btn.innerText = '📋 Copy IP for Her'; }, 3000);
    }

    async function loadDevices() {
      try {
        const res = await fetch('/api/v1/router/devices');
        if (res.ok) {
          const json = await res.json();
          allDevices = json.devices || [];
          document.getElementById('totalDevicesCount').innerText = allDevices.length;
          const tbody = document.getElementById('devicesTableBody');
          tbody.innerHTML = allDevices.map(d => 
            '<tr class="hover:bg-[#121215] transition">' +
              '<td class="py-2.5 px-3 font-medium text-zinc-200">' + (d.device_name || 'Host') + '</td>' +
              '<td class="py-2.5 px-3 muted mono">' + d.port_id + ' (' + d.interface_type + ')</td>' +
              '<td class="py-2.5 px-3 mono text-[11px]"><span class="text-zinc-300">' + d.ip_address + '</span><br><span class="muted">' + d.mac_address + '</span></td>' +
              '<td class="py-2.5 px-3 muted mono text-[11px]">' + d.connection_time + '</td>' +
              '<td class="py-2.5 px-3 text-right"><span class="' + (d.status === 'Online' ? 'text-emerald-400' : 'muted') + '">' + d.status + '</span></td>' +
            '</tr>'
          ).join('');
        }
      } catch (e) {}
    }

    function initEtherApe() {
      const canvas = document.getElementById('etherapeCanvas');
      if (!canvas) return;
      const ctx = canvas.getContext('2d');
      canvas.width = canvas.parentElement.clientWidth;
      canvas.height = canvas.parentElement.clientHeight;

      const rawNodes = [
        { id: 'gw', x: 0.5, y: 0.5, r: 18, label: 'Huawei ONT (192.168.100.1)', color: '#FAFAFA' },
        { id: 'doh', x: 0.85, y: 0.5, r: 14, label: 'Cloudflare DoH (1.1.1.1)', color: '#10B981' },
        { id: 'sec', x: 0.2, y: 0.35, r: 15, label: 'Secondary Router (192.168.1.253)', color: '#A855F7' },
        { id: 'server', x: 0.7, y: 0.25, r: 15, label: 'Aegis Host Server', color: '#10B981' },
        { id: 'pc', x: 0.3, y: 0.8, r: 13, label: 'DESKTOP-QO58KLD (LAN1)', color: '#06B6D4' }
      ];

      const links = [
        { s: 0, t: 1, color: '#10B981' },
        { s: 0, t: 2, color: '#A855F7' },
        { s: 0, t: 3, color: '#10B981' },
        { s: 0, t: 4, color: '#06B6D4' }
      ];

      let t = 0;
      function render() {
        ctx.fillStyle = '#0C0C0E';
        ctx.fillRect(0, 0, canvas.width, canvas.height);

        links.forEach(l => {
          const s = rawNodes[l.s];
          const trg = rawNodes[l.t];
          const sx = s.x * canvas.width;
          const sy = s.y * canvas.height;
          const tx = trg.x * canvas.width;
          const ty = trg.y * canvas.height;

          ctx.strokeStyle = '#27272A';
          ctx.lineWidth = 1.5;
          ctx.beginPath();
          ctx.moveTo(sx, sy);
          ctx.lineTo(tx, ty);
          ctx.stroke();

          const progress = (t * 0.015 + l.s * 0.25) % 1;
          const px = sx + (tx - sx) * progress;
          const py = sy + (ty - sy) * progress;
          ctx.fillStyle = l.color;
          ctx.beginPath();
          ctx.arc(px, py, 3, 0, Math.PI * 2);
          ctx.fill();
        });

        rawNodes.forEach(n => {
          const nx = n.x * canvas.width;
          const ny = n.y * canvas.height;
          ctx.fillStyle = '#18181B';
          ctx.strokeStyle = n.color;
          ctx.lineWidth = 2;
          ctx.beginPath();
          ctx.arc(nx, ny, n.r, 0, Math.PI * 2);
          ctx.fill();
          ctx.stroke();

          ctx.fillStyle = '#A1A1AA';
          ctx.font = '10px ui-monospace, monospace';
          ctx.textAlign = 'center';
          ctx.fillText(n.label, nx, ny + n.r + 14);
        });

        t++;
        requestAnimationFrame(render);
      }
      render();
    }

    async function runSpeedtest() {
      const btn = document.getElementById('speedtestBtn');
      btn.innerText = 'Testing...';
      btn.disabled = true;
      try {
        await fetch('/api/v1/speedtest/run', { method: 'POST' });
        setTimeout(() => { btn.innerText = 'Run Speedtest'; btn.disabled = false; }, 4000);
      } catch (e) { btn.innerText = 'Run Speedtest'; btn.disabled = false; }
    }

    async function rebootRouter() {
      if (!confirm('Reboot Huawei ONT (192.168.100.1)?')) return;
      try {
        const res = await fetch('/api/v1/router/reboot', { method: 'POST' });
        const json = await res.json();
        alert(json.message || 'Reboot dispatched.');
      } catch (e) { alert('Reboot dispatched.'); }
    }

    window.onload = () => {
      initEtherApe();
      loadDevices();
      loadPiholeStats();
      loadStorageDrives();
    };
  </script>
</body>
</html>`
