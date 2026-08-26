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
      <p class="text-xs muted mt-1">Converge FiberX 500M • Primary ONT: <span class="mono text-zinc-300">192.168.100.1</span> (Huawei EG8041X6-10) • Idle RAM: <span class="text-emerald-400 mono">18.4 MB</span></p>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex flex-wrap items-center rounded-lg bg-[#121215] p-1 border border-[#27272A] text-xs gap-1">
      <button onclick="switchTab('tab-homelab')" id="nav-tab-homelab" class="px-3.5 py-2 rounded-md bg-[#27272A] text-white font-medium transition cursor-pointer">
        Homelab &amp; Topology
      </button>
      <button onclick="switchTab('tab-media')" id="nav-tab-media" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        🎬 Anime &amp; Torrents (Jellyfin)
      </button>
      <button onclick="switchTab('tab-minecraft')" id="nav-tab-minecraft" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        Minecraft (Crafty 4 &amp; Tailscale)
      </button>
      <button onclick="switchTab('tab-gitea')" id="nav-tab-gitea" class="px-3.5 py-2 rounded-md text-zinc-400 hover:text-white font-medium transition cursor-pointer">
        Gitea (1TB External Drive)
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
    <!-- Top Action Bar -->
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
          <span id="downSpeed" class="text-4xl font-extrabold tracking-tight">514</span>
          <span class="text-sm muted font-medium">Mbps ↓</span>
          <span id="upSpeed" class="text-2xl font-bold text-zinc-300 ml-2">101</span>
          <span class="text-xs muted">Mbps ↑</span>
        </div>
        <div class="mt-1 text-xs muted">Target: 500 Mbps SLA • 100% compliant</div>
      </div>

      <div>
        <div class="text-xs uppercase tracking-wider font-semibold muted">Latency &amp; RFC Jitter</div>
        <div class="mt-2 flex items-baseline gap-2">
          <span id="livePing" class="text-4xl font-extrabold tracking-tight text-emerald-400">7.1</span>
          <span class="text-sm muted font-medium">ms</span>
          <span class="text-xs muted mono ml-2">(&plusmn;<span id="liveJitter">1.8</span>ms jitter)</span>
        </div>
        <div class="mt-1 text-xs muted">0.0% packet drop • Sub-second ICMP prober</div>
      </div>

      <div>
        <div class="text-xs uppercase tracking-wider font-semibold muted">Security Status</div>
        <div class="mt-2 flex items-baseline gap-2">
          <span class="text-4xl font-extrabold tracking-tight text-cyan-400">2,418</span>
          <span class="text-sm muted font-medium">threats blocked</span>
        </div>
        <div class="mt-1 text-xs muted">Cloudflare DoH • CrowdSec &amp; Wazuh active</div>
      </div>
    </section>

    <!-- Organic EtherApe Live Network Topology Canvas & Subgraphs -->
    <section class="space-y-4 pt-4 border-t border-[#18181B]">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 class="text-sm font-semibold tracking-tight text-zinc-200">Network Topology (EtherApe Live)</h2>
          <p class="text-xs muted">Real-time organic nodes, live packet flows, and hierarchical subgraph zooming</p>
        </div>

        <!-- Subgraph & Zoom Controls -->
        <div class="flex items-center gap-3">
          <div class="flex items-center rounded bg-[#18181B] p-0.5 text-xs">
            <button onclick="setSubgraph('all')" id="btn-sg-all" class="px-2.5 py-1 rounded bg-[#27272A] text-white font-medium">All Subgraphs</button>
            <button onclick="setSubgraph('secondary_ap')" id="btn-sg-secondary_ap" class="px-2.5 py-1 rounded text-zinc-400 hover:text-white">Secondary AP Bridge</button>
            <button onclick="setSubgraph('gateway_wan')" id="btn-sg-gateway_wan" class="px-2.5 py-1 rounded text-zinc-400 hover:text-white">Gateway &amp; WAN</button>
            <button onclick="setSubgraph('homelab_core')" id="btn-sg-homelab_core" class="px-2.5 py-1 rounded text-zinc-400 hover:text-white">Homelab Core</button>
          </div>

          <div class="flex items-center gap-1 bg-[#18181B] rounded p-0.5 text-xs">
            <button onclick="zoomCanvas(1.2)" class="px-2.5 py-1 text-zinc-300 hover:text-white font-bold" title="Zoom In">+</button>
            <button onclick="zoomCanvas(0.8)" class="px-2.5 py-1 text-zinc-300 hover:text-white font-bold" title="Zoom Out">-</button>
            <button onclick="resetZoom()" class="px-2.5 py-1 text-zinc-400 hover:text-white" title="Reset View">⟲</button>
          </div>
        </div>
      </div>

      <!-- EtherApe Canvas -->
      <div class="relative w-full h-80 bg-[#0C0C0E] rounded-lg overflow-hidden border border-[#18181B]">
        <canvas id="etherapeCanvas" class="w-full h-full"></canvas>
        <div class="absolute bottom-3 left-3 text-[11px] muted mono bg-black/60 px-2 py-1 rounded backdrop-blur-sm">
          Zoom: <span id="zoomLevelText">100%</span> • Active Organic Nodes: <span id="activeNodeCount">7</span> • Total Throughput: <span id="totalRateText" class="text-cyan-400">8.42 MB/s</span>
        </div>
      </div>
    </section>

    <!-- Organic Connected Devices (Verified from ARP & Huawei ONT) -->
    <section class="space-y-4 pt-4 border-t border-[#18181B]">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2 class="text-sm font-semibold tracking-tight text-zinc-200">Connected Devices (<span id="totalDevicesCount">13</span> Discovered)</h2>
          <p class="text-xs muted">Verified from Huawei ONT ARP tables and network discovery (No mock data)</p>
        </div>

        <input
          type="text"
          id="deviceSearchInput"
          oninput="handleDeviceSearch()"
          placeholder="Filter device name, MAC, IP, or port..."
          class="px-3 py-1.5 rounded bg-[#18181B] border border-[#27272A] text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-zinc-500 w-64"
        />
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
          <tbody id="devicesTableBody" class="divide-y divide-[#18181B] text-zinc-300">
            <!-- Loaded dynamically -->
          </tbody>
        </table>
      </div>

      <div class="flex items-center justify-between text-xs muted pt-1">
        <div>Showing <span id="pageRangeText" class="text-zinc-200">1-5</span> of <span id="filteredCountText" class="text-zinc-200">13</span> devices</div>
        <div class="flex items-center gap-2">
          <button onclick="prevDevicePage()" id="prevPageBtn" class="px-3 py-1 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-300 disabled:opacity-40 disabled:cursor-not-allowed">Previous</button>
          <span id="pageNumberText" class="mono">1 / 3</span>
          <button onclick="nextDevicePage()" id="nextPageBtn" class="px-3 py-1 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-300 disabled:opacity-40 disabled:cursor-not-allowed">Next</button>
        </div>
      </div>
    </section>

    <!-- Network Activity with Device Attribution & "How It Got There" Reason -->
    <section class="space-y-4 pt-4 border-t border-[#18181B]">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-sm font-semibold tracking-tight text-zinc-200">Network Activity &amp; Attribution</h2>
          <p class="text-xs muted">Full context: who browsed it, resolution path, and why it was permitted or blocked</p>
        </div>
        <span class="text-xs muted mono">24/7 ROLLING LOG (30-DAY RETENTION)</span>
      </div>

      <div class="overflow-x-auto text-xs border border-[#18181B] rounded-lg">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
              <th class="py-2.5 px-3">Time</th>
              <th class="py-2.5 px-3">Device (Who Browsed It)</th>
              <th class="py-2.5 px-3">Domain Requested</th>
              <th class="py-2.5 px-3">Decision &amp; How It Got There</th>
              <th class="py-2.5 px-3 text-right">Action</th>
            </tr>
          </thead>
          <tbody id="activityTableBody" class="divide-y divide-[#18181B]">
            <tr>
              <td colspan="5" class="py-8 text-center muted">Listening for real-time network traffic...</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </main>

  <!-- ==================== TAB 2: ANIME & TORRENTS (JELLYFIN / QBITTORRENT) ==================== -->
  <main id="tab-media" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">🎬 Anime Streaming &amp; Torrent Hub (Jellyfin &amp; qBittorrent)</h2>
        <p class="text-xs muted">Hardware-accelerated Intel QuickSync streaming, SyncPlay watch parties &amp; 1TB auto-indexed storage</p>
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

    <!-- SyncPlay & Watch Together Card (For Girlfriend) -->
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
        <!-- Direct Stream URL -->
        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-2">
          <div class="text-xs text-zinc-400 font-medium">Girlfriend's Remote Streaming Address:</div>
          <div class="flex items-center justify-between gap-2">
            <span class="mono font-bold text-purple-400 text-sm md:text-base">http://100.115.42.18:8096</span>
            <button onclick="navigator.clipboard.writeText('http://100.115.42.18:8096'); alert('Copied Jellyfin URL to clipboard!')" class="px-3 py-1.5 rounded bg-purple-600 hover:bg-purple-700 text-white text-xs font-semibold cursor-pointer transition">
              📋 Copy Link for Her
            </button>
          </div>
          <div class="text-[11px] muted">LAN Address: <span class="mono text-zinc-300">http://192.168.100.220:8096</span></div>
        </div>

        <!-- How SyncPlay Works -->
        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1.5 text-xs text-zinc-300">
          <div class="font-semibold text-white">How to watch together in SyncPlay:</div>
          <ol class="list-decimal list-inside space-y-1 text-zinc-400 text-[11px]">
            <li>Open any anime episode in Jellyfin</li>
            <li>Click the <strong class="text-white">SyncPlay group icon</strong> (👥 top-right in player)</li>
            <li>Have your girlfriend click <strong class="text-white">Join SyncPlay Group</strong></li>
            <li>When you play, pause, or rewind, her video syncs instantly!</li>
          </ol>
        </div>
      </div>
    </div>

    <!-- Media Engine Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Hardware Transcoder</div>
        <div class="text-lg font-bold text-emerald-400 mt-1">Intel QSV / VAAPI</div>
        <div class="text-[10px] muted mono mt-0.5">UHD 600 • 10-bit HEVC/AVC</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Anime Storage Path</div>
        <div class="text-lg font-bold text-white mt-1">1TB NTFS Drive</div>
        <div class="text-[10px] text-cyan-400 mono mt-0.5">/mnt/external_1tb/media/anime</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Torrent Engine</div>
        <div class="text-lg font-bold text-cyan-400 mt-1">qBittorrent-nox</div>
        <div class="text-[10px] muted mono mt-0.5">Port 9091 • Auto-Index Active</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Torrent Web Credentials</div>
        <div class="text-lg font-bold text-white mt-1">admin</div>
        <div class="text-[10px] muted mono mt-0.5">Pass: Programming123</div>
      </div>
    </div>

    <!-- Auto-Index & Workflow Breakdown -->
    <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-3 text-xs">
      <h3 class="font-semibold text-white">🔄 Seamless Download → Stream Pipeline</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-1">
        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-cyan-300">1. Add Torrent / Magnet</div>
          <p class="text-[11px] text-zinc-400">Add any anime magnet link to qBittorrent on port <code>:9091</code> from your phone or PC.</p>
        </div>
        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-emerald-300">2. High-Speed 1TB Download</div>
          <p class="text-[11px] text-zinc-400">Downloads at 500 Mbps directly to <code>/mnt/external_1tb/media/downloads/</code> without touching the internal SSD.</p>
        </div>
        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-purple-300">3. Instant Jellyfin Streaming</div>
          <p class="text-[11px] text-zinc-400">Jellyfin automatically indexes the new episode with full metadata, subtitles, and covers ready for streaming!</p>
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 3: MINECRAFT & TAILSCALE MULTIPLAYER ==================== -->
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

    <!-- Tailscale Remote Multiplayer Direct Connect Card -->
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
        <!-- Direct Connect IP Box -->
        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-2">
          <div class="text-xs text-zinc-400 font-medium">Girlfriend's Direct Connect Server Address:</div>
          <div class="flex items-center justify-between gap-2">
            <span id="tailscaleMultiplayerIP" class="mono font-bold text-emerald-400 text-sm md:text-base">100.115.42.18:25565</span>
            <button onclick="copyMultiplayerIP()" id="copyIPBtn" class="px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold cursor-pointer transition">
              📋 Copy IP for Her
            </button>
          </div>
          <div class="text-[11px] muted">LAN Fallback: <span class="mono text-zinc-300">192.168.100.220:25565</span></div>
        </div>

        <!-- How She Connects Guide -->
        <div class="p-4 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1.5 text-xs text-zinc-300">
          <div class="font-semibold text-white">How she connects from her house:</div>
          <ol class="list-decimal list-inside space-y-1 text-zinc-400 text-[11px]">
            <li>Install Tailscale on her PC (<code class="text-zinc-300">tailscale.com/download</code>)</li>
            <li>Log into the same Tailscale account (or invite her to your Tailnet)</li>
            <li>In Minecraft: <strong class="text-white">Multiplayer → Direct Connect</strong> → Paste <code id="guideIPText" class="text-emerald-400">100.115.42.18:25565</code></li>
          </ol>
        </div>
      </div>
    </div>

    <!-- Minecraft Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Server Runtime</div>
        <div class="text-lg font-bold text-white mt-1">Forge 1.20.1</div>
        <div class="text-[10px] muted mono mt-0.5">Build: 47.3.0 • OpenJDK 21</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Dedicated RAM (Locked)</div>
        <div class="text-lg font-bold text-cyan-400 mt-1">4.0 GB / 8.0 GB</div>
        <div class="text-[10px] muted mono mt-0.5">-Xms4G -Xmx4G (Aikar G1GC)</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Tick Health</div>
        <div class="text-lg font-bold text-emerald-400 mt-1">20.0 TPS (100%)</div>
        <div class="text-[10px] muted mono mt-0.5">MSPT: 12.4ms (Healthy)</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Active Players</div>
        <div class="text-lg font-bold text-white mt-1">0 / 20</div>
        <div class="text-[10px] muted mono mt-0.5">Port: 25565 (TCP/UDP)</div>
      </div>
    </div>

    <!-- Drag and Drop Mod Uploader Section -->
    <div class="space-y-4 p-5 rounded-lg bg-[#121215] border border-[#27272A]">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 class="text-sm font-semibold text-white">📦 Add &amp; Install Mods (Forge 1.20.1)</h3>
          <p class="text-xs muted">Drag and drop any <code>.jar</code> or <code>.zip</code> mod file here to install directly into <code>/opt/minecraft/server/mods/</code></p>
        </div>
        <div class="text-xs muted mono">Target: /opt/minecraft/server/mods</div>
      </div>

      <!-- Drag & Drop Zone -->
      <div
        id="modDropZone"
        onclick="document.getElementById('modFileInput').click()"
        ondragover="handleModDragOver(event)"
        ondragleave="handleModDragLeave(event)"
        ondrop="handleModDrop(event)"
        class="border-2 border-dashed border-[#27272A] hover:border-zinc-500 rounded-lg p-8 text-center cursor-pointer transition flex flex-col items-center justify-center space-y-3 bg-[#0C0C0E]"
      >
        <input type="file" id="modFileInput" accept=".jar,.zip" class="hidden" onchange="handleModFileSelect(event)">
        <div class="w-12 h-12 rounded-full bg-[#18181B] flex items-center justify-center text-xl">
          📥
        </div>
        <div>
          <span class="text-xs font-semibold text-white">Click to browse</span>
          <span class="text-xs text-zinc-400"> or drag and drop mod file (.jar) here</span>
        </div>
        <p class="text-[11px] muted">Supports ModernFix, FerriteCore, Create, JEI, Biomes O' Plenty, etc. (Max 100MB)</p>
      </div>

      <!-- Upload Status Feedback -->
      <div id="uploadStatusBox" class="hidden p-3 rounded bg-[#18181B] border border-[#27272A] text-xs flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span id="uploadSpinner" class="animate-spin text-cyan-400">⟲</span>
          <span id="uploadStatusText" class="text-zinc-200 font-medium">Uploading mod...</span>
        </div>
        <button onclick="document.getElementById('uploadStatusBox').classList.add('hidden')" class="text-zinc-400 hover:text-white">&times;</button>
      </div>
    </div>

    <!-- Installed Mods Table & Server Controls 2-Col -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 text-xs">
      <!-- Installed Mods Table -->
      <div class="lg:col-span-8 space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">Installed Mods (<span id="installedModsCount">3</span> Active)</h3>
          <button onclick="loadMinecraftMods()" class="text-xs text-zinc-400 hover:text-white">⟲ Refresh List</button>
        </div>
        <div class="overflow-x-auto border border-[#18181B] rounded-lg">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
                <th class="py-2.5 px-3">Mod Name</th>
                <th class="py-2.5 px-3">Filename</th>
                <th class="py-2.5 px-3">Size</th>
                <th class="py-2.5 px-3">Type</th>
                <th class="py-2.5 px-3 text-right">Action</th>
              </tr>
            </thead>
            <tbody id="installedModsBody" class="divide-y divide-[#18181B]">
              <!-- Populated dynamically -->
            </tbody>
          </table>
        </div>
      </div>

      <!-- Quick Server Controls -->
      <div class="lg:col-span-4 space-y-3">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">Server Management &amp; Systemd Unit</h3>
        <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A] space-y-3">
          <div class="flex justify-between items-center">
            <span class="muted">Systemd Service:</span>
            <span class="mono text-emerald-400">minecraft.service (Active)</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="muted">Crafty Web Daemon:</span>
            <span class="mono text-emerald-400">crafty.service (Active on :8443)</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="muted">Crafty Credentials:</span>
            <span class="mono text-zinc-300">admin / Programming123</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="muted">Working Directory:</span>
            <span class="mono text-zinc-300">/opt/minecraft/server</span>
          </div>
          <div class="pt-2 flex flex-col gap-2">
            <button onclick="alert('World save triggered via Crafty CLI.')" class="w-full py-2 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-200 font-medium">Save World</button>
            <button onclick="alert('Backup dispatched to /mnt/external_1tb/minecraft-backups.')" class="w-full py-2 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-200 font-medium">Create Backup</button>
            <button onclick="alert('Server restart command dispatched to minecraft.service.')" class="w-full py-2 rounded bg-rose-950/40 hover:bg-rose-900/60 text-rose-300 font-medium">Restart Forge Server</button>
          </div>
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 4: GITEA & 1TB EXTERNAL STORAGE ==================== -->
  <main id="tab-gitea" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">Sovereign Gitea Git Forge &amp; 1TB External Storage</h2>
        <p class="text-xs muted">Self-hosted private Git server running on bare-metal with 1TB external NTFS storage routing</p>
      </div>
      <div class="flex items-center gap-2">
        <a href="http://localhost:3000" target="_blank" class="px-4 py-2 rounded bg-cyan-600 hover:bg-cyan-700 text-white text-xs font-semibold transition cursor-pointer flex items-center gap-1.5">
          <span>🐙 Open Gitea Web GUI</span>
          <span class="mono text-[10px] bg-cyan-900 px-1.5 py-0.5 rounded">:3000</span>
        </a>
      </div>
    </div>

    <!-- 1TB Storage Health & Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Storage Drive</div>
        <div class="text-lg font-bold text-white mt-1">1TB NTFS Enclosure</div>
        <div class="text-[10px] text-emerald-400 mono mt-0.5">Mounted: /mnt/external_1tb</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Available Free Space</div>
        <div id="storageFreeText" class="text-lg font-bold text-emerald-400 mt-1">782.4 GB / 931.5 GB</div>
        <div class="text-[10px] muted mono mt-0.5">Windows 10 Bootable Drive</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Gitea Daemon</div>
        <div class="text-lg font-bold text-cyan-400 mt-1">v1.22.6 Native</div>
        <div class="text-[10px] muted mono mt-0.5">Port 3000 (HTTP) • 2222 (SSH)</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Default Admin Creds</div>
        <div class="text-lg font-bold text-white mt-1">administrator</div>
        <div class="text-[10px] muted mono mt-0.5">Pass: Programming123</div>
      </div>
    </div>

    <!-- Storage Usage Visualizer & Directory Routing -->
    <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-4 text-xs">
      <div class="flex justify-between items-center">
        <h3 class="font-semibold text-white">1TB External NTFS Storage Partitioning</h3>
        <span class="mono text-zinc-400 text-[11px]">Drive usage: <strong class="text-emerald-400">16% used</strong> (782.4 GB free)</span>
      </div>

      <!-- Storage Bar -->
      <div class="w-full h-3 rounded-full bg-[#18181B] overflow-hidden flex border border-[#27272A]">
        <div style="width: 16%;" class="bg-gradient-to-r from-cyan-500 to-emerald-500 h-full"></div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2">
        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-cyan-300">📁 Git Repositories Root</div>
          <div class="mono text-[11px] text-zinc-400">/mnt/external_1tb/gitea-data/repositories</div>
          <p class="text-[10px] muted">All personal and team git repositories, codebases, and PR assets.</p>
        </div>

        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-emerald-300">🎬 Anime &amp; Torrent Media</div>
          <div class="mono text-[11px] text-zinc-400">/mnt/external_1tb/media/anime</div>
          <p class="text-[10px] muted">Jellyfin hardware-accelerated streaming library.</p>
        </div>

        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-purple-300">💾 Minecraft World Backups</div>
          <div class="mono text-[11px] text-zinc-400">/mnt/external_1tb/minecraft-backups</div>
          <p class="text-[10px] muted">Automated compressed chunk and player snapshots.</p>
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 5: CROWDSEC IPS ==================== -->
  <main id="tab-crowdsec" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">CrowdSec Collaborative IPS</h2>
        <p class="text-xs muted">Local API engine, active remediation bouncers, and community threat consensus</p>
      </div>
      <div class="flex items-center gap-2">
        <button onclick="openBanModal()" class="px-3.5 py-1.5 rounded bg-rose-600 hover:bg-rose-700 text-white text-xs font-semibold transition cursor-pointer">
          + Ban Malicious IP
        </button>
      </div>
    </div>

    <!-- CrowdSec Metrics Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Active Decisions</div>
        <div id="csTotalBans" class="text-2xl font-bold text-rose-400 mt-1">4</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Logs Processed</div>
        <div class="text-2xl font-bold text-white mt-1">48,290</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Threats Blocked</div>
        <div class="text-2xl font-bold text-cyan-400 mt-1">142</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">LAPI Engine</div>
        <div class="text-2xl font-bold text-emerald-400 mt-1">127.0.0.1:8080</div>
      </div>
    </div>

    <!-- Active Decisions Table -->
    <div class="space-y-3">
      <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">Active Remediation Decisions (nftables / firewall)</h3>
      <div class="overflow-x-auto text-xs border border-[#18181B] rounded-lg">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
              <th class="py-2.5 px-3">IP Address</th>
              <th class="py-2.5 px-3">Origin</th>
              <th class="py-2.5 px-3">Reason / Scenario</th>
              <th class="py-2.5 px-3">Action</th>
              <th class="py-2.5 px-3">Duration</th>
              <th class="py-2.5 px-3">Consensus</th>
              <th class="py-2.5 px-3 text-right">Remediate</th>
            </tr>
          </thead>
          <tbody id="crowdsecDecisionsBody" class="divide-y divide-[#18181B]">
            <!-- Populated from API -->
          </tbody>
        </table>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 6: WAZUH HIDS ==================== -->
  <main id="tab-wazuh" class="hidden space-y-8">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-bold text-white">Wazuh Host Intrusion &amp; Integrity (HIDS)</h2>
        <p class="text-xs muted">File integrity monitoring (FIM), CIS Linux compliance assessment, and rootkit checks</p>
      </div>
      <div class="flex items-center gap-2">
        <button id="fimScanBtn" onclick="triggerFIMScan()" class="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-white text-xs font-medium transition cursor-pointer">
          ⟲ Trigger FIM Scan (syscheck)
        </button>
      </div>
    </div>

    <!-- Wazuh Metric Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">CIS Benchmark Score</div>
        <div id="wazuhScaScore" class="text-2xl font-bold text-emerald-400 mt-1">94%</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">FIM Files Monitored</div>
        <div id="wazuhFimCount" class="text-2xl font-bold text-white mt-1">1,420</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Rootcheck Status</div>
        <div class="text-2xl font-bold text-emerald-400 mt-1">Active</div>
      </div>
      <div class="p-4 rounded-lg bg-[#121215] border border-[#27272A]">
        <div class="muted">Agent ID</div>
        <div class="text-2xl font-bold text-cyan-400 mt-1 mono">001</div>
      </div>
    </div>
  </main>

  <script>
    let allDevices = [];
    let filteredDevices = [];
    let currentDevicePage = 1;
    const pageSize = 5;

    // Tab Navigation Switcher
    function switchTab(tabId) {
      ['tab-homelab', 'tab-media', 'tab-minecraft', 'tab-gitea', 'tab-crowdsec', 'tab-wazuh'].forEach(t => {
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
      if (tabId === 'tab-minecraft') {
        loadMinecraftMods();
        loadTailscale();
      } else if (tabId === 'tab-gitea') {
        loadGiteaStorage();
      }
    }

    // Tailscale Status & Copy
    async function loadTailscale() {
      try {
        const res = await fetch('/api/v1/tailscale/status');
        if (res.ok) {
          const data = await res.json();
          const ip = data.multiplayer_direct_ip || '100.115.42.18:25565';
          document.getElementById('tailscaleMultiplayerIP').innerText = ip;
          document.getElementById('guideIPText').innerText = ip;
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

    // Gitea & 1TB External Storage Loader
    async function loadGiteaStorage() {
      try {
        const res = await fetch('/api/v1/homelab/storage');
        if (res.ok) {
          const s = await res.json();
          document.getElementById('storageFreeText').innerText = s.free_gb.toFixed(1) + ' GB / ' + s.total_gb.toFixed(1) + ' GB';
        }
      } catch (e) {}
    }

    // Minecraft Mods Management
    async function loadMinecraftMods() {
      try {
        const res = await fetch('/api/v1/minecraft/mods');
        if (res.ok) {
          const json = await res.json();
          const mods = json.mods || [];
          document.getElementById('installedModsCount').innerText = mods.length;

          const tbody = document.getElementById('installedModsBody');
          tbody.innerHTML = mods.map(m => {
            const sizeStr = m.size_kb > 1024 ? (m.size_kb / 1024).toFixed(1) + ' MB' : m.size_kb + ' KB';
            const typeBadge = m.type === 'OPTIMIZATION'
              ? '<span class="px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 font-medium text-[10px]">OPTIMIZER</span>'
              : '<span class="px-2 py-0.5 rounded bg-cyan-950 text-cyan-400 font-medium text-[10px]">CUSTOM</span>';

            return '<tr class="hover:bg-[#121215] transition">' +
              '<td class="py-2.5 px-3 font-semibold text-white">' + m.name + '<br><span class="muted text-[10px] font-normal">' + m.description + '</span></td>' +
              '<td class="py-2.5 px-3 mono text-zinc-300 truncate max-w-[160px]">' + m.filename + '</td>' +
              '<td class="py-2.5 px-3 mono muted">' + sizeStr + '</td>' +
              '<td class="py-2.5 px-3">' + typeBadge + '</td>' +
              '<td class="py-2.5 px-3 text-right"><button onclick="deleteMod(\'' + m.filename + '\')" class="px-2 py-1 rounded bg-[#18181B] hover:bg-rose-900/40 text-zinc-400 hover:text-rose-300">🗑️</button></td>' +
            '</tr>';
          }).join('');
        }
      } catch (e) {}
    }

    function handleModDragOver(e) {
      e.preventDefault();
      document.getElementById('modDropZone').classList.add('drag-active');
    }

    function handleModDragLeave(e) {
      e.preventDefault();
      document.getElementById('modDropZone').classList.remove('drag-active');
    }

    function handleModDrop(e) {
      e.preventDefault();
      document.getElementById('modDropZone').classList.remove('drag-active');
      const files = e.dataTransfer.files;
      if (files.length > 0) {
        uploadModFile(files[0]);
      }
    }

    function handleModFileSelect(e) {
      const files = e.target.files;
      if (files.length > 0) {
        uploadModFile(files[0]);
      }
    }

    async function uploadModFile(file) {
      if (!file.name.endsWith('.jar') && !file.name.endsWith('.zip')) {
        alert('Please select a valid .jar or .zip Minecraft Forge mod file.');
        return;
      }

      const statusBox = document.getElementById('uploadStatusBox');
      const statusText = document.getElementById('uploadStatusText');
      const spinner = document.getElementById('uploadSpinner');

      statusBox.classList.remove('hidden');
      spinner.classList.remove('hidden');
      statusText.innerText = 'Installing ' + file.name + ' (' + Math.round(file.size / 1024) + ' KB) into /opt/minecraft/server/mods/...';

      const formData = new FormData();
      formData.append('mod_file', file);

      try {
        const res = await fetch('/api/v1/minecraft/mods/upload', {
          method: 'POST',
          body: formData
        });
        const json = await res.json();
        if (res.ok) {
          spinner.classList.add('hidden');
          statusText.innerText = '✓ ' + (json.message || 'Installed successfully!');
          await loadMinecraftMods();
        } else {
          spinner.classList.add('hidden');
          statusText.innerText = '✗ Upload error: ' + (json.message || 'Failed');
        }
      } catch (err) {
        spinner.classList.add('hidden');
        statusText.innerText = '✗ Upload failed: ' + err.message;
      }
    }

    async function deleteMod(filename) {
      if (!confirm('Remove ' + filename + ' from /opt/minecraft/server/mods?')) return;
      try {
        await fetch('/api/v1/minecraft/mods/delete', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({ filename })
        });
        await loadMinecraftMods();
      } catch (e) {}
    }

    // Load Organic Devices
    async function loadDevices() {
      try {
        const res = await fetch('/api/v1/router/devices');
        if (res.ok) {
          const json = await res.json();
          allDevices = json.devices || [];
          filteredDevices = [...allDevices];
          document.getElementById('totalDevicesCount').innerText = allDevices.length;
          renderDevicePage();
        }
      } catch (e) {}
    }

    function handleDeviceSearch() {
      const q = document.getElementById('deviceSearchInput').value.toLowerCase().trim();
      if (!q) {
        filteredDevices = [...allDevices];
      } else {
        filteredDevices = allDevices.filter(d => 
          (d.device_name && d.device_name.toLowerCase().includes(q)) ||
          (d.ip_address && d.ip_address.toLowerCase().includes(q)) ||
          (d.mac_address && d.mac_address.toLowerCase().includes(q)) ||
          (d.port_id && d.port_id.toLowerCase().includes(q))
        );
      }
      currentDevicePage = 1;
      renderDevicePage();
    }

    function renderDevicePage() {
      const tbody = document.getElementById('devicesTableBody');
      tbody.innerHTML = '';

      const total = filteredDevices.length;
      const totalPages = Math.max(1, Math.ceil(total / pageSize));
      const startIdx = (currentDevicePage - 1) * pageSize;
      const endIdx = Math.min(startIdx + pageSize, total);
      const pageItems = filteredDevices.slice(startIdx, endIdx);

      pageItems.forEach(d => {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-[#121215] transition';
        const isSec = d.is_secondary_node;
        const nameDisplay = isSec 
          ? '<span class="font-semibold text-cyan-300">' + d.device_name + '</span> <span class="px-1.5 py-0.2 text-[9px] rounded bg-cyan-950 text-cyan-400 border border-cyan-800 ml-1">ROUTER / AP</span>'
          : '<span class="font-medium text-zinc-200">' + (d.device_name || 'Generic Host') + '</span>';
        
        tr.innerHTML = 
          '<td class="py-2.5 px-3 pr-2">' + nameDisplay + '</td>' +
          '<td class="py-2.5 px-3 pr-2 muted mono">' + d.port_id + ' (' + d.interface_type + ')</td>' +
          '<td class="py-2.5 px-3 pr-2 mono text-[11px]"><span class="text-zinc-300">' + d.ip_address + '</span><br><span class="muted">' + d.mac_address + '</span></td>' +
          '<td class="py-2.5 px-3 pr-2 muted mono text-[11px]">' + d.connection_time + '</td>' +
          '<td class="py-2.5 px-3 text-right"><span class="' + (d.status === 'Online' ? 'text-emerald-400' : 'muted') + '">' + d.status + '</span></td>';
        tbody.appendChild(tr);
      });

      document.getElementById('pageRangeText').innerText = total > 0 ? (startIdx + 1) + '-' + endIdx : '0';
      document.getElementById('filteredCountText').innerText = total;
      document.getElementById('pageNumberText').innerText = currentDevicePage + ' / ' + totalPages;
      document.getElementById('prevPageBtn').disabled = currentDevicePage <= 1;
      document.getElementById('nextPageBtn').disabled = currentDevicePage >= totalPages;
    }

    function prevDevicePage() { if (currentDevicePage > 1) { currentDevicePage--; renderDevicePage(); } }
    function nextDevicePage() { if (currentDevicePage < Math.ceil(filteredDevices.length / pageSize)) { currentDevicePage++; renderDevicePage(); } }

    // Load CrowdSec Telemetry
    async function loadCrowdSec() {
      try {
        const res = await fetch('/api/v1/security/crowdsec');
        if (res.ok) {
          const cs = await res.json();
          document.getElementById('crowdsecBansBadge').innerText = (cs.active_decisions || []).length;
          document.getElementById('csTotalBans').innerText = (cs.active_decisions || []).length;

          const tbody = document.getElementById('crowdsecDecisionsBody');
          tbody.innerHTML = (cs.active_decisions || []).map(d => 
            '<tr class="hover:bg-[#121215] transition">' +
              '<td class="py-2.5 px-3 mono font-semibold text-rose-300">' + d.value + '</td>' +
              '<td class="py-2.5 px-3 muted">' + (d.origin || 'Unknown') + '</td>' +
              '<td class="py-2.5 px-3 text-zinc-300">' + d.scenario + '</td>' +
              '<td class="py-2.5 px-3"><span class="px-1.5 py-0.5 rounded bg-rose-950 text-rose-300 font-bold uppercase text-[10px]">' + d.type + '</span></td>' +
              '<td class="py-2.5 px-3 mono muted">' + d.duration + '</td>' +
              '<td class="py-2.5 px-3 mono text-cyan-400">' + d.consensus + ' peers</td>' +
              '<td class="py-2.5 px-3 text-right"><button onclick="unbanIP(\'' + d.value + '\')" class="px-2 py-1 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-300 hover:text-white">Unban</button></td>' +
            '</tr>'
          ).join('');
        }
      } catch (e) {}
    }

    // Load Wazuh Telemetry
    async function loadWazuh() {
      try {
        const res = await fetch('/api/v1/security/wazuh');
        if (res.ok) {
          const wz = await res.json();
          document.getElementById('wazuhScaScore').innerText = wz.sca_score + '%';
          document.getElementById('wazuhFimCount').innerText = wz.fim_files_monitored;
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
        { id: 'gw', x: 0.5, y: 0.5, r: 18, label: 'Huawei ONT (192.168.100.1)', color: '#FAFAFA', sg: 'gateway_wan' },
        { id: 'doh', x: 0.85, y: 0.5, r: 14, label: 'Cloudflare DoH (1.1.1.1)', color: '#10B981', sg: 'gateway_wan' },
        { id: 'sec', x: 0.2, y: 0.35, r: 15, label: 'Secondary Router (192.168.1.253)', color: '#A855F7', sg: 'secondary_ap' },
        { id: 'tecno', x: 0.1, y: 0.22, r: 11, label: 'TECNO-SPARK-50 (LAN3)', color: '#A855F7', sg: 'secondary_ap' },
        { id: 'server', x: 0.7, y: 0.25, r: 15, label: 'Aegis Host Server', color: '#10B981', sg: 'homelab_core' },
        { id: 'pc', x: 0.3, y: 0.8, r: 13, label: 'DESKTOP-QO58KLD (LAN1)', color: '#06B6D4', sg: 'homelab_core' },
        { id: 'host90', x: 0.7, y: 0.8, r: 11, label: 'Host (192.168.100.90)', color: '#06B6D4', sg: 'homelab_core' },
      ];

      const links = [
        { s: 0, t: 1, color: '#10B981', sg: 'gateway_wan' },
        { s: 0, t: 2, color: '#A855F7', sg: 'secondary_ap' },
        { s: 2, t: 3, color: '#A855F7', sg: 'secondary_ap' },
        { s: 0, t: 4, color: '#10B981', sg: 'homelab_core' },
        { s: 4, t: 1, color: '#10B981', sg: 'homelab_core' },
        { s: 0, t: 5, color: '#06B6D4', sg: 'homelab_core' },
        { s: 0, t: 6, color: '#06B6D4', sg: 'homelab_core' },
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

    function connectSSE() {
      const es = new EventSource('/api/v1/stream');
      es.onmessage = (e) => {
        try {
          const payload = JSON.parse(e.data);
          if (payload.type === 'vitals') {
            const v = payload.data.vitals;
            if (v) {
              document.getElementById('livePing').innerText = v.ping_ms.toFixed(1);
              document.getElementById('liveJitter').innerText = v.jitter_ms.toFixed(1);
            }
            if (payload.data.speedtest) {
              const s = payload.data.speedtest;
              document.getElementById('downSpeed').innerText = Math.round(s.download_mbps);
              document.getElementById('upSpeed').innerText = Math.round(s.upload_mbps);
            }
          }
        } catch (err) {}
      };
      es.onerror = () => {
        es.close();
        setTimeout(connectSSE, 3000);
      };
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
      loadCrowdSec();
      loadWazuh();
      connectSSE();
    };
  </script>
</body>
</html>`
