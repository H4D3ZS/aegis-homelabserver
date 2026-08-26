# 🛡️ AEGIS HOMELABSERVER: NATIVE BARE-METAL ALL-IN-ONE SENTINEL

**Repository**: `github.com/H4D3ZS/aegis-homelabserver`  
**Target Host**: x86_64 Ubuntu Server 24.04 LTS (Intel Celeron N4100, 8GB RAM, Teclast F7 Plus)  
**Architecture**: Pure Bare-Metal Systemd Services + Single Embedded Golang Binary (Zero Docker runtime)

---

## 🎯 System Overview & Mission

Aegis HomelabServer unifies enterprise homelab security, DNS privacy, ISP SLA watchdog monitoring, Huawei GPON ONT session automation, and a native Minecraft Forge 1.20.1 game server into a **single compiled 12MB Go binary and native Linux systemd services**.

By eliminating heavy Docker container layers, the entire host idles at **under 25MB RAM** (leaving >7.5GB RAM for the Minecraft server and system cache).

---

## 🏗️ Core Subsystems

### 1. Hardware & Linux OS Hardening (`install.sh`)
- **Display & Thermal Management**: `consoleblank=60` in GRUB turns off LCD backlight after 60s idle (while keeping laptop lid open for keyboard deck thermal convection). Sleep/suspend on lid close is disabled in `/etc/systemd/logind.conf.d/homelab-lid.conf`.
- **Battery Anti-Bloat Protection**: `tlp` and `battery-threshold.service` enforce a 70% max charge threshold (`/sys/class/power_supply/BAT*/charge_control_limit_max`) to prevent lithium battery degradation and swelling.
- **Wi-Fi Stability Over Air**: NetworkManager `wifi.powersave = 2` and `wifi-powersave-off.service` eliminate beacon drop ping jitter spikes.
- **Kernel Networking**: Google BBR TCP congestion control (`net.ipv4.tcp_congestion_control = bbr`, expanded socket buffers, `net.core.default_qdisc = fq`, `net.ipv4.ip_forward = 1`).

### 2. Minecraft Forge Modded Server & Crafty Controller 4
- **Forge 1.20.1 (Forge 47.3.0)**: Native OpenJDK 21 headless execution under user `minecraft` in `/opt/minecraft/server`.
- **Locked Memory Heap & Tuned G1GC**: Configured in `user_jvm_args.txt`:
  ```
  -Xms4G -Xmx4G -XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40 -XX:G1ReservePercent=20 -XX:G1HeapWastePercent=5 -XX:G1MixedGCCountTarget=4 -XX:InitiatingHeapOccupancyPercent=15 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:SurvivorRatio=32 -XX:+PerfDisableSharedMem -XX:MaxTenuringThreshold=1
  ```
- **Crafty Controller 4 GUI**: Native web management dashboard running under `/opt/crafty/` on `https://<HOST_IP>:8443`.
- **Recommended Performance Mods**: `ModernFix`, `FerriteCore`, `Radon` in `/opt/minecraft/server/mods/`.

### 3. DNS Privacy & Threat Mitigation Pipeline
- **Cloudflared DoH (`127.0.0.1:5053`)**: Upstream TLS 1.3 encrypted DNS to Cloudflare (`1.1.1.1`) and Quad9 (`9.9.9.9`).
- **Native Pi-hole FTL**: Upstream locked to `127.0.0.1#5053`, SafeSearch CNAME injection in `/etc/dnsmasq.d/05-safesearch.conf`.
- **CrowdSec Engine + `crowdsec-firewall-bouncer-nftables`**: Host IPS monitoring auth/system logs with automated kernel-level IP dropping.

### 4. Pi-Sentinel Core Go Daemon (`backend/`)
- **Mathematical DGA & Tunneling Detector (`internal/analyzer/`)**: Computes Shannon Entropy $H(X) = -\sum P(x) \log_2 P(x)$ on Second-Level Domains (SLDs). Automatically neutralizes domains with $H(X) \ge 3.85$ or consonant sequence $\ge 5$.
- **Huawei ONT Token Automation (`internal/router/huawei.go`)**: Multi-step challenge session handler (`/asp/GetRandCount.asp` -> `/login.cgi` -> `/html/ssmp/reset/reset.asp` -> `/html/ssmp/reset/set.cgi`).
- **120s Auto-Healer Watchdog**: If local gateway responds but upstream targets fail for 120 continuous seconds, automatically reboots ONT (capped at 2 reboots/hour with 5-minute cooldown).
- **ISP Jitter & Speed Watchdog (`internal/isp/`)**: Continuous RFC 3550 rolling jitter prober and non-blocking `speedtest-cli` runner tracking bandwidth against SLA.
- **1-Click Smart Unbreaker (`internal/pihole/unbreak.go`)**: Scans recent blocked streaming CDNs (`kwik.cx`, `doodstream`, `mp4upload`) and temporarily whitelists them for 15 minutes with automated eviction.

### 5. Minimalist High Data-Ink Dashboard (`frontend/`)
- Solid matte `#09090B` background.
- Clean unboxed numbers for Bandwidth, Latency & Jitter, and Blocked Threats.
- Flat muted action buttons: `Run Speedtest`, `Fix Video Stream`, `Reboot Fiber Router`.
- Soft rose threat badges for anomalies; muted rows for benign queries.
- Next.js static export embedded directly into the Go binary (`//go:embed`) serving on port `3001`.

---

## 🚀 Quickstart & Installation

```bash
# Clone and build
git clone https://github.com/H4D3ZS/aegis-homelabserver.git
cd aegis-homelabserver

# Build static frontend & binary
make build

# Run automated bare-metal installer
sudo bash install.sh
```

---

## 📊 Live Endpoints & Port Map

| Service | Port | Description |
| :--- | :--- | :--- |
| **Aegis Dashboard** | `3001` | Single Embedded Go Web Dashboard & SSE Stream |
| **Pi-hole Admin** | `8080` | Native Pi-hole FTL DNS Management Interface |
| **Cloudflared DoH** | `5053` | Local Encrypted DNS Proxy (`127.0.0.1:5053`) |
| **Crafty Controller 4** | `8443` | Minecraft Web Management GUI (`https://<IP>:8443`) |
| **Minecraft Forge** | `25565` | Native Modded Forge 1.20.1 Game Server |
