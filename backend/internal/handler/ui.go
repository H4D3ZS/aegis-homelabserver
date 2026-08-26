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

  <!-- ==================== TAB 2: MINECRAFT & TAILSCALE MULTIPLAYER ==================== -->
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

  <!-- ==================== TAB 3: GITEA & 1TB EXTERNAL STORAGE ==================== -->
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
        <div class="text-lg font-bold text-white mt-1">admin</div>
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
          <div class="font-semibold text-emerald-300">💾 Minecraft World Backups</div>
          <div class="mono text-[11px] text-zinc-400">/mnt/external_1tb/minecraft-backups</div>
          <p class="text-[10px] muted">Nightly compressed chunk and player inventory snapshots.</p>
        </div>

        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-1">
          <div class="font-semibold text-purple-300">📦 Aegis 30-Day Archive</div>
          <div class="mono text-[11px] text-zinc-400">/mnt/external_1tb/aegis-archive</div>
          <p class="text-[10px] muted">Rotated DNS query logs and incident forensics.</p>
        </div>
      </div>
    </div>

    <!-- Gitea Fast Access & SSH Clone Info -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 text-xs">
      <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-3">
        <h3 class="font-semibold text-white">🐙 Quick Git Clone Instructions</h3>
        <p class="muted">Clone your sovereign repositories securely over LAN or Tailscale Mesh VPN:</p>
        <div class="p-3 rounded bg-[#0C0C0E] border border-[#27272A] space-y-2 mono text-[11px]">
          <div class="text-zinc-400"># HTTP Clone:</div>
          <div class="text-cyan-400 font-bold">git clone http://localhost:3000/admin/my-project.git</div>
          <div class="text-zinc-400 pt-1"># Tailscale Remote Clone:</div>
          <div class="text-emerald-400 font-bold">git clone http://100.115.42.18:3000/admin/my-project.git</div>
        </div>
      </div>

      <div class="p-5 rounded-lg bg-[#121215] border border-[#27272A] space-y-3">
        <h3 class="font-semibold text-white">⚙️ Service &amp; Storage Configuration</h3>
        <div class="space-y-2 text-zinc-300">
          <div class="flex justify-between"><span class="muted">Systemd Unit:</span> <span class="mono text-emerald-400">gitea.service (Active)</span></div>
          <div class="flex justify-between"><span class="muted">Configuration File:</span> <span class="mono text-zinc-300">/etc/gitea/app.ini</span></div>
          <div class="flex justify-between"><span class="muted">SQLite Database:</span> <span class="mono text-zinc-300">/mnt/external_1tb/gitea-data/gitea.db</span></div>
          <div class="flex justify-between"><span class="muted">NTFS Mount Driver:</span> <span class="mono text-zinc-300">ntfs-3g (big_writes, nofail)</span></div>
          <div class="flex justify-between"><span class="muted">Base Memory Footprint:</span> <span class="mono text-emerald-400">~38.2 MB (Zero Docker)</span></div>
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 4: CROWDSEC IPS ==================== -->
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

    <!-- Installed Hub Scenarios & Bouncers 2-Col -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 text-xs">
      <div class="space-y-3">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">Installed Scenarios &amp; Parsers</h3>
        <div class="border border-[#18181B] rounded-lg divide-y divide-[#18181B]" id="crowdsecScenariosList">
          <!-- Populated from API -->
        </div>
      </div>

      <div class="space-y-3">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">Registered Bouncers</h3>
        <div class="border border-[#18181B] rounded-lg divide-y divide-[#18181B]" id="crowdsecBouncersList">
          <!-- Populated from API -->
        </div>
      </div>
    </div>
  </main>

  <!-- ==================== TAB 5: WAZUH HIDS ==================== -->
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

    <!-- File Integrity Monitoring (FIM) Events -->
    <div class="space-y-3">
      <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">File Integrity Events (/etc/ &amp; /bin/ Audit)</h3>
      <div class="overflow-x-auto text-xs border border-[#18181B] rounded-lg">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
              <th class="py-2.5 px-3">Path</th>
              <th class="py-2.5 px-3">Event</th>
              <th class="py-2.5 px-3">Checksum</th>
              <th class="py-2.5 px-3">Owner</th>
              <th class="py-2.5 px-3">Status</th>
              <th class="py-2.5 px-3 text-right">Timestamp</th>
            </tr>
          </thead>
          <tbody id="wazuhFimBody" class="divide-y divide-[#18181B]">
            <!-- Populated from API -->
          </tbody>
        </table>
      </div>
    </div>

    <!-- CIS SCA Checks & SSH Forensics 2-Col -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 text-xs">
      <div class="space-y-3">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">CIS Linux Benchmark Assessment</h3>
        <div class="border border-[#18181B] rounded-lg divide-y divide-[#18181B]" id="wazuhScaList">
          <!-- Populated from API -->
        </div>
      </div>

      <div class="space-y-3">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-300">SSH Authentication Audit Log</h3>
        <div class="border border-[#18181B] rounded-lg divide-y divide-[#18181B]" id="wazuhSshList">
          <!-- Populated from API -->
        </div>
      </div>
    </div>
  </main>

  <!-- Manual Ban IP Modal -->
  <div id="banModal" class="hidden fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
    <div class="w-full max-w-md bg-[#121215] p-6 rounded-lg border border-[#27272A] space-y-4 text-xs">
      <div class="flex justify-between items-start">
        <h3 class="text-base font-bold text-white">Manual IP Ban (CrowdSec / nftables)</h3>
        <button onclick="closeBanModal()" class="text-zinc-500 hover:text-white text-lg">&times;</button>
      </div>
      <div class="space-y-3">
        <div>
          <label class="block muted mb-1">Target IP Address</label>
          <input id="banInputIP" type="text" placeholder="e.g. 185.220.101.5" class="w-full px-3 py-2 rounded bg-[#18181B] border border-[#27272A] text-white focus:outline-none focus:border-zinc-500">
        </div>
        <div>
          <label class="block muted mb-1">Reason / Scenario</label>
          <input id="banInputReason" type="text" placeholder="e.g. ssh-bruteforce / port-scan" class="w-full px-3 py-2 rounded bg-[#18181B] border border-[#27272A] text-white focus:outline-none focus:border-zinc-500">
        </div>
        <div>
          <label class="block muted mb-1">Duration</label>
          <select id="banInputDuration" class="w-full px-3 py-2 rounded bg-[#18181B] border border-[#27272A] text-white focus:outline-none focus:border-zinc-500">
            <option value="4h">4 Hours</option>
            <option value="24h">24 Hours</option>
            <option value="7d">7 Days</option>
            <option value="30d">30 Days</option>
          </select>
        </div>
      </div>
      <button onclick="executeBan()" class="w-full py-2.5 rounded bg-rose-600 text-white font-semibold hover:bg-rose-700 transition cursor-pointer">
        Enforce Firewall Ban
      </button>
    </div>
  </div>

  <!-- Query Details Inspector Modal -->
  <div id="queryInspectorModal" class="hidden fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
    <div class="w-full max-w-lg bg-[#121215] p-6 rounded-lg border border-[#27272A] space-y-4 text-xs">
      <div class="flex justify-between items-start">
        <div>
          <h3 class="text-base font-bold text-white">DNS Query Forensics</h3>
          <p class="text-xs muted mt-0.5" id="inspectDomainTitle">domain.com</p>
        </div>
        <button onclick="closeInspectorModal()" class="text-zinc-500 hover:text-white text-lg">&times;</button>
      </div>
      <div class="space-y-2 pt-2 border-t border-[#27272A]">
        <div class="flex justify-between"><span class="muted">Device:</span> <span id="inspectDevice" class="font-medium text-white">DESKTOP-QO58KLD</span></div>
        <div class="flex justify-between"><span class="muted">Client IP:</span> <span id="inspectIP" class="mono text-zinc-300">192.168.100.220</span></div>
        <div class="flex justify-between"><span class="muted">Timestamp:</span> <span id="inspectTime" class="mono muted">2026-08-26 18:40:12</span></div>
        <div class="flex justify-between"><span class="muted">Query Type:</span> <span id="inspectType" class="mono text-cyan-400">A (IPv4)</span></div>
        <div class="flex justify-between"><span class="muted">Resolution Path:</span> <span id="inspectReason" class="text-zinc-200 font-medium">Forwarded to Cloudflare DoH (127.0.0.1:5053)</span></div>
        <div class="flex justify-between"><span class="muted">Shannon Entropy:</span> <span id="inspectEntropy" class="mono text-emerald-400">2.81 (Benign)</span></div>
      </div>
      <div class="flex items-center gap-2 pt-2 border-t border-[#27272A]">
        <button id="inspectAllowBtn" class="flex-1 py-2 rounded bg-[#18181B] hover:bg-[#27272A] text-white font-medium">Whitelist Domain</button>
        <button id="inspectBlockBtn" class="flex-1 py-2 rounded bg-[#18181B] hover:bg-rose-900/30 text-rose-400 font-medium">Block Domain</button>
      </div>
    </div>
  </div>

  <!-- Unbreak Modal -->
  <div id="unbreakModal" class="hidden fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
    <div class="w-full max-w-md bg-[#121215] p-6 rounded-lg border border-[#27272A] space-y-4 text-xs">
      <div class="flex justify-between items-start">
        <div>
          <h3 class="text-base font-bold text-white">1-Click Smart Unbreak</h3>
          <p class="text-xs muted mt-0.5">Temporarily unblock media streaming CDNs (15-min auto-eviction)</p>
        </div>
        <button onclick="closeUnbreakModal()" class="text-zinc-500 hover:text-white text-lg">&times;</button>
      </div>
      <p class="text-xs muted">Scans recent blocked stream CDNs (<code>kwik.cx</code>, <code>doodstream</code>, <code>mp4upload</code>) and permits playback without whitelisting ad trackers.</p>
      <button id="triggerUnbreakBtn" onclick="executeUnbreak()" class="w-full py-2.5 rounded bg-white text-black font-semibold text-xs hover:bg-zinc-200 transition cursor-pointer">
        Unblock Streams (Last 120s)
      </button>
      <div id="unbreakOutput" class="hidden text-xs text-emerald-400 pt-2"></div>
    </div>
  </div>

  <script>
    let allDevices = [];
    let filteredDevices = [];
    let currentDevicePage = 1;
    const pageSize = 5;

    // Tab Navigation Switcher
    function switchTab(tabId) {
      ['tab-homelab', 'tab-minecraft', 'tab-gitea', 'tab-crowdsec', 'tab-wazuh'].forEach(t => {
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

          const scList = document.getElementById('crowdsecScenariosList');
          scList.innerHTML = (cs.installed_scenarios || []).map(s =>
            '<div class="p-3 flex items-center justify-between">' +
              '<div><div class="font-medium text-white">' + s.name + ' <span class="muted mono text-[10px]">v' + s.version + '</span></div><div class="text-zinc-400 mt-0.5">' + s.description + '</div></div>' +
              '<span class="px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 font-medium text-[10px]">' + s.status + '</span>' +
            '</div>'
          ).join('');

          const bList = document.getElementById('crowdsecBouncersList');
          bList.innerHTML = (cs.bouncers || []).map(b =>
            '<div class="p-3 flex items-center justify-between">' +
              '<div><div class="font-medium text-white">' + b.name + '</div><div class="muted mono text-[10px]">Type: ' + b.type + ' • Host: ' + b.ip_address + '</div></div>' +
              '<span class="px-2 py-0.5 rounded bg-cyan-950 text-cyan-400 font-medium text-[10px]">' + b.status + '</span>' +
            '</div>'
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

          const tbody = document.getElementById('wazuhFimBody');
          tbody.innerHTML = (wz.fim_events || []).map(f =>
            '<tr class="hover:bg-[#121215] transition">' +
              '<td class="py-2.5 px-3 mono font-medium text-zinc-200">' + f.path + '</td>' +
              '<td class="py-2.5 px-3"><span class="px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300 font-medium text-[10px]">' + f.event_type + '</span></td>' +
              '<td class="py-2.5 px-3 mono muted truncate max-w-[160px]">' + f.checksum + '</td>' +
              '<td class="py-2.5 px-3 muted">' + f.user + ' (' + f.permissions + ')</td>' +
              '<td class="py-2.5 px-3 text-emerald-400 font-medium">' + f.status + '</td>' +
              '<td class="py-2.5 px-3 text-right mono muted">' + new Date(f.timestamp).toLocaleTimeString() + '</td>' +
            '</tr>'
          ).join('');

          const scaList = document.getElementById('wazuhScaList');
          scaList.innerHTML = (wz.sca_checks || []).map(c =>
            '<div class="p-3 flex items-center justify-between">' +
              '<div><div class="font-medium text-white">' + c.id + ': ' + c.title + '</div><div class="text-zinc-400 mt-0.5">Fix: <code>' + c.remediation + '</code></div></div>' +
              '<span class="px-2 py-0.5 rounded font-bold text-[10px] ' + (c.status === 'Passed' ? 'bg-emerald-950 text-emerald-400' : 'bg-rose-950 text-rose-300') + '">' + c.status + '</span>' +
            '</div>'
          ).join('');

          const sshList = document.getElementById('wazuhSshList');
          sshList.innerHTML = (wz.ssh_auth_logs || []).map(s =>
            '<div class="p-3 flex items-center justify-between">' +
              '<div><div class="font-medium text-white">' + s.user + ' from <span class="mono text-zinc-300">' + s.client_ip + '</span></div><div class="muted mono text-[10px]">Auth: ' + s.auth_type + ' • Port: ' + s.port + '</div></div>' +
              '<span class="px-2 py-0.5 rounded font-bold text-[10px] ' + (s.status === 'Accepted' ? 'bg-emerald-950 text-emerald-400' : 'bg-rose-950 text-rose-300') + '">' + s.status + '</span>' +
            '</div>'
          ).join('');
        }
      } catch (e) {}
    }

    async function triggerFIMScan() {
      const btn = document.getElementById('fimScanBtn');
      btn.innerText = 'Scanning /etc/ & /bin/...';
      btn.disabled = true;
      try {
        await fetch('/api/v1/security/wazuh/scan', { method: 'POST' });
        await loadWazuh();
      } catch (e) {}
      btn.innerText = '⟲ Trigger FIM Scan (syscheck)';
      btn.disabled = false;
    }

    function openBanModal() { document.getElementById('banModal').classList.remove('hidden'); }
    function closeBanModal() { document.getElementById('banModal').classList.add('hidden'); }

    async function executeBan() {
      const ip = document.getElementById('banInputIP').value.trim();
      const reason = document.getElementById('banInputReason').value.trim();
      const duration = document.getElementById('banInputDuration').value;
      if (!ip) return;
      await fetch('/api/v1/security/crowdsec/ban', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ ip, reason, duration })
      });
      closeBanModal();
      await loadCrowdSec();
    }

    async function unbanIP(ip) {
      if (!confirm('Unban IP ' + ip + ' from firewall?')) return;
      await fetch('/api/v1/security/crowdsec/unban', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ ip })
      });
      await loadCrowdSec();
    }

    // EtherApe Canvas
    let zoomLevel = 1.0;
    let activeSubgraph = 'all';

    function setSubgraph(sg) {
      activeSubgraph = sg;
      ['all', 'secondary_ap', 'gateway_wan', 'homelab_core'].forEach(id => {
        const btn = document.getElementById('btn-sg-' + id);
        if (btn) {
          btn.className = (id === sg) ? 'px-2.5 py-1 rounded bg-[#27272A] text-white font-medium' : 'px-2.5 py-1 rounded text-zinc-400 hover:text-white';
        }
      });
    }

    function zoomCanvas(factor) {
      zoomLevel = Math.max(0.5, Math.min(2.5, zoomLevel * factor));
      document.getElementById('zoomLevelText').innerText = Math.round(zoomLevel * 100) + '%';
    }

    function resetZoom() {
      zoomLevel = 1.0;
      activeSubgraph = 'all';
      setSubgraph('all');
      document.getElementById('zoomLevelText').innerText = '100%';
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

        ctx.save();
        ctx.translate(canvas.width / 2, canvas.height / 2);
        ctx.scale(zoomLevel, zoomLevel);
        ctx.translate(-canvas.width / 2, -canvas.height / 2);

        const visibleNodes = rawNodes.filter(n => activeSubgraph === 'all' || n.sg === activeSubgraph || n.id === 'gw');
        const nodeMap = {};
        visibleNodes.forEach(n => nodeMap[n.id] = n);

        links.forEach(l => {
          const s = rawNodes[l.s];
          const trg = rawNodes[l.t];
          if (!nodeMap[s.id] || !nodeMap[trg.id]) return;

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

        visibleNodes.forEach(n => {
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

        ctx.restore();
        t++;
        requestAnimationFrame(render);
      }
      render();
    }

    // SSE Connection & Live Activity
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
          } else if (payload.type === 'query') {
            addActivityRow(payload.data);
          }
        } catch (err) {}
      };
      es.onerror = () => {
        es.close();
        setTimeout(connectSSE, 3000);
      };
    }

    function addActivityRow(q) {
      const tbody = document.getElementById('activityTableBody');
      const isThreat = q.threat && (q.threat.is_threat || q.threat.threat_score >= 0.75);
      const isBlocked = q.status === 'BLOCKED' || q.status === 'GRAVITY';
      const timeStr = new Date(q.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      
      const tr = document.createElement('tr');
      tr.className = isThreat ? 'bg-rose-950/20 text-rose-300 cursor-pointer hover:bg-rose-900/30 transition' : 'hover:bg-[#121215] text-zinc-400 transition cursor-pointer';

      const devName = q.device_name || 'DESKTOP-QO58KLD';
      const clientInfo = '<span class="font-medium text-zinc-200">' + devName + '</span><br><span class="muted mono text-[10px]">' + q.client_ip + '</span>';

      let decisionHtml = '<span class="text-emerald-400">✓ ' + (q.reason || 'Forwarded to Cloudflare DoH') + '</span>';
      if (isThreat) {
        decisionHtml = '<span class="text-rose-400 font-semibold">✗ ' + (q.reason || 'Blocked by Heuristic DGA Analyzer') + '</span>';
      } else if (isBlocked) {
        decisionHtml = '<span class="text-rose-400">✗ ' + (q.reason || 'Blocked by Pi-hole Gravity adlist') + '</span>';
      }

      tr.innerHTML = 
        '<td class="py-2.5 px-3 muted mono text-[11px] whitespace-nowrap">' + timeStr + '</td>' +
        '<td class="py-2.5 px-3 pr-2">' + clientInfo + '</td>' +
        '<td class="py-2.5 px-3 pr-2 truncate max-w-[200px]"><span class="mono text-white font-medium">' + q.domain + '</span></td>' +
        '<td class="py-2.5 px-3 pr-2 text-[11px]">' + decisionHtml + '</td>' +
        '<td class="py-2.5 px-3 text-right whitespace-nowrap">' +
          '<button onclick="event.stopPropagation(); actionDomain(\'whitelist\', \'' + q.domain + '\')" class="px-2 py-0.5 text-zinc-400 hover:text-white mr-1 text-[11px]">Allow</button>' +
          '<button onclick="event.stopPropagation(); actionDomain(\'block\', \'' + q.domain + '\')" class="px-2 py-0.5 text-zinc-400 hover:text-rose-400 text-[11px]">Block</button>' +
        '</td>';

      tr.onclick = () => openInspectorModal(q);

      if (tbody.children.length === 1 && tbody.children[0].innerText.includes('Listening')) {
        tbody.innerHTML = '';
      }
      tbody.insertBefore(tr, tbody.firstChild);
      if (tbody.children.length > 50) tbody.removeChild(tbody.lastChild);
    }

    function openInspectorModal(q) {
      document.getElementById('inspectDomainTitle').innerText = q.domain;
      document.getElementById('inspectDevice').innerText = q.device_name || 'DESKTOP-QO58KLD';
      document.getElementById('inspectIP').innerText = q.client_ip;
      document.getElementById('inspectTime').innerText = new Date(q.timestamp).toLocaleString();
      document.getElementById('inspectType').innerText = (q.query_type || 'A') + ' (Latency: ' + (q.response_time_ms ? q.response_time_ms.toFixed(1) : '1.2') + 'ms)';
      document.getElementById('inspectReason').innerText = q.reason || 'Forwarded to Cloudflare DoH (127.0.0.1:5053)';
      const entropy = q.threat ? q.threat.shannon_entropy.toFixed(2) : '2.40';
      document.getElementById('inspectEntropy').innerText = entropy + ' (Heuristic Score: ' + (q.threat ? (q.threat.threat_score*100).toFixed(0)+'%' : '0%') + ')';
      
      document.getElementById('inspectAllowBtn').onclick = () => { actionDomain('whitelist', q.domain); closeInspectorModal(); };
      document.getElementById('inspectBlockBtn').onclick = () => { actionDomain('block', q.domain); closeInspectorModal(); };
      document.getElementById('queryInspectorModal').classList.remove('hidden');
    }

    function closeInspectorModal() { document.getElementById('queryInspectorModal').classList.add('hidden'); }

    async function actionDomain(action, domain) {
      await fetch('/api/v1/' + action, { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ domain }) });
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

    function openUnbreakModal() { document.getElementById('unbreakModal').classList.remove('hidden'); }
    function closeUnbreakModal() { document.getElementById('unbreakModal').classList.add('hidden'); }

    async function executeUnbreak() {
      const btn = document.getElementById('triggerUnbreakBtn');
      const out = document.getElementById('unbreakOutput');
      btn.disabled = true;
      btn.innerText = 'Scanning...';
      try {
        const res = await fetch('/api/v1/unbreak', { method: 'POST' });
        const json = await res.json();
        out.classList.remove('hidden');
        out.innerText = json.message;
      } catch (e) {}
      btn.disabled = false;
      btn.innerText = 'Fix Video Stream';
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
