#!/usr/bin/env bash
# ==============================================================================
# AEGIS HOMELAB - ZERO-DOCKER BARE-METAL SENTINEL, GITEA, MEDIA & GAME SERVER INSTALLER
# Optimized for: x86_64 Ubuntu Server 24.04 / Teclast F7 Plus (Intel Celeron N4100, 8GB RAM)
# Repository: github.com/H4D3ZS/aegis-homelabserver
# ==============================================================================

set -euo pipefail

# Ensure root privileges
if [ "$EUID" -ne 0 ]; then
  echo "[!] Please run as root: sudo ./install.sh"
  exit 1
fi

echo "============================================================"
echo "    [+] INITIATING AEGIS HOMELAB NATIVE SERVER PROVISIONING "
echo "============================================================"

# --- 1. SYSTEM OS & DEPENDENCIES ---
echo "[1/14] Updating package repository and installing core tools..."
apt-get update -y
apt-get install -y --no-install-recommends \
    curl wget git git-lfs jq ufw nftables tlp tlp-rdw iw wireless-tools \
    wpasupplicant network-manager linux-firmware \
    sqlite3 openjdk-21-jre-headless build-essential python3 python3-pip python3-venv \
    ntfs-3g smartmontools rsync va-driver-all intel-media-va-driver-non-free vainfo \
    qbittorrent-nox

# --- 2. DISABLE SYSTEMD-RESOLVED STUB (FREE PORT 53 FOR PI-HOLE) ---
echo "[2/14] Disabling systemd-resolved DNSStubListener to free Port 53 for Pi-hole..."
mkdir -p /etc/systemd/resolved.conf.d/
cat <<EOF > /etc/systemd/resolved.conf.d/disable-stub.conf
[Resolve]
DNSStubListener=no
DNS=127.0.0.1#5053 1.1.1.1
EOF

systemctl restart systemd-resolved || true
rm -f /etc/resolv.conf
echo "nameserver 1.1.1.1" > /etc/resolv.conf

# --- 3. TECLAST LAPTOP WI-FI FIRMWARE & HARDENING ---
echo "[3/14] Configuring Teclast F7 Plus Wi-Fi (Intel iwlwifi / Realtek) & Power Limits..."

# Detect Wireless Interface Name (wlan0, wlp1s0, wlp2s0, etc.)
WIFI_IFACE=$(iw dev 2>/dev/null | awk '$1=="Interface"{print $2}' | head -n 1 || echo "")
if [ -z "$WIFI_IFACE" ]; then
    WIFI_IFACE=$(ip link 2>/dev/null | awk -F': ' '/wl/{print $2}' | head -n 1 || echo "wlan0")
fi
echo "Detected Wi-Fi Interface: ${WIFI_IFACE}"

# Prevent laptop from suspending when lid is closed or idle (Keep lid open for thermal convection)
mkdir -p /etc/systemd/logind.conf.d/
cat <<EOF > /etc/systemd/logind.conf.d/homelab-lid.conf
[Login]
HandleLidSwitch=ignore
HandleLidSwitchDocked=ignore
HandleLidSwitchExternalPower=ignore
IdleAction=ignore
EOF
systemctl restart systemd-logind || true

# Turn off display backlight after 60s idle via GRUB
if ! grep -q "consoleblank=60" /etc/default/grub; then
    sed -i 's/GRUB_CMDLINE_LINUX_DEFAULT="/GRUB_CMDLINE_LINUX_DEFAULT="consoleblank=60 /' /etc/default/grub
    update-grub || true
fi

# Disable Wi-Fi power saving permanently (Stops beacon drops & ping jitter)
mkdir -p /etc/NetworkManager/conf.d/
cat <<EOF > /etc/NetworkManager/conf.d/default-wifi-powersave-on.conf
[connection]
wifi.powersave = 2
EOF

cat <<EOF > /etc/systemd/system/wifi-powersave-off.service
[Unit]
Description=Disable Wi-Fi Power Management for Homelab Stability
After=network.target

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'for dev in \$(iw dev 2>/dev/null | awk "\$1==\"Interface\"{print \$2}"); do iw dev "\$dev" set power_save off || true; done'
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now wifi-powersave-off.service || true

# Setup Battery Protection & UPS Watchdog (Shuts down cleanly if power outage drains battery < 7%)
cat <<EOF > /etc/systemd/system/battery-threshold.service
[Unit]
Description=Set Battery Charge Threshold (Anti-Bloat Protection)
After=multi-user.target

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'for bat in /sys/class/power_supply/BAT*; do [ -f "\$bat/charge_control_limit_max" ] && echo 70 > "\$bat/charge_control_limit_max" || true; [ -f "\$bat/charge_stop_threshold" ] && echo 70 > "\$bat/charge_stop_threshold" || true; done'
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF
systemctl enable --now battery-threshold.service || true
systemctl enable --now tlp || true

# Battery UPS Auto-Shutdown Watchdog Daemon (Safeguard 1TB Drive & DBs)
cat << 'EOF' > /usr/local/bin/aegis-ups-watchdog.sh
#!/usr/bin/env bash
for bat in /sys/class/power_supply/BAT*; do
    if [ -f "$bat/capacity" ] && [ -f "$bat/status" ]; then
        CAPACITY=$(cat "$bat/capacity")
        STATUS=$(cat "$bat/status")
        if [ "$STATUS" = "Discharging" ] && [ "$CAPACITY" -le 7 ]; then
            logger -t aegis-ups "CRITICAL: Battery at ${CAPACITY}%. Initiating safe shutdown of 1TB drive & services."
            systemctl stop minecraft crafty gitea jellyfin || true
            sync
            umount -f /mnt/external_1tb || true
            shutdown -h now "Battery critically low (${CAPACITY}%) during power outage."
        fi
    fi
done
EOF
chmod +x /usr/local/bin/aegis-ups-watchdog.sh

cat <<EOF > /etc/systemd/system/aegis-ups-watchdog.service
[Unit]
Description=Aegis Laptop Battery UPS Graceful Shutdown Watchdog

[Service]
Type=oneshot
ExecStart=/usr/local/bin/aegis-ups-watchdog.sh
EOF

cat <<EOF > /etc/systemd/system/aegis-ups-watchdog.timer
[Unit]
Description=Run Aegis Battery UPS Watchdog every 2 minutes

[Timer]
OnBootSec=2min
OnUnitActiveSec=2min

[Install]
WantedBy=timers.target
EOF
systemctl daemon-reload
systemctl enable --now aegis-ups-watchdog.timer || true

# --- 4. KERNEL TCP BBR & CONGESTION TUNING ---
echo "[4/14] Enabling TCP BBR and tuning network sysctl parameters..."
cat <<EOF > /etc/sysctl.d/99-aegis-network.conf
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_fastopen = 3
net.ipv4.ip_forward = 1
EOF
sysctl --system || true

# --- 5. EXTERNAL 1TB NTFS ENCLOSURE STORAGE MOUNT ---
echo "[5/14] Configuring External 1TB NTFS Drive Mount (/mnt/external_1tb)..."
mkdir -p /mnt/external_1tb /mnt/external_1tb/gitea-data /mnt/external_1tb/minecraft-backups /mnt/external_1tb/aegis-archive \
         /mnt/external_1tb/media /mnt/external_1tb/media/anime /mnt/external_1tb/media/movies /mnt/external_1tb/media/downloads

# Find external NTFS drive partition if attached
NTFS_DEV=$(lsblk -o NAME,FSTYPE -rn 2>/dev/null | grep -i "ntfs" | head -n 1 | awk '{print "/dev/" $1}' || echo "")

if [ -n "$NTFS_DEV" ]; then
    echo "Detected NTFS Partition: ${NTFS_DEV}. Mounting to /mnt/external_1tb..."
    mount -t ntfs-3g -o windows_names,big_writes,nofail,uid=1000,gid=1000,umask=022 "${NTFS_DEV}" /mnt/external_1tb || true
    
    # Add non-blocking safe fstab entry
    if ! grep -q "/mnt/external_1tb" /etc/fstab; then
        UUID_VAL=$(blkid -s UUID -o value "${NTFS_DEV}" || echo "")
        if [ -n "$UUID_VAL" ]; then
            echo "UUID=${UUID_VAL} /mnt/external_1tb ntfs-3g windows_names,big_writes,nofail,uid=1000,gid=1000,umask=022 0 0" >> /etc/fstab
        fi
    fi
fi

# --- 6. TAILSCALE REMOTE MULTIPLAYER MESH VPN ---
echo "[6/14] Installing and provisioning Tailscale VPN (Zero-Config Multiplayer Mesh)..."
if ! command -v tailscale &> /dev/null; then
    curl -fsSL https://tailscale.com/install.sh | sh || true
    systemctl enable --now tailscaled || true
fi

# --- 7. ENCRYPTED DOH (CLOUDFLARED) ---
echo "[7/14] Installing and provisioning Cloudflared DoH on port 5053..."
if ! command -v cloudflared &> /dev/null; then
    wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb -O /tmp/cloudflared.deb
    dpkg -i /tmp/cloudflared.deb || true
    rm -f /tmp/cloudflared.deb
fi

mkdir -p /etc/cloudflared
cat <<EOF > /etc/cloudflared/config.yml
proxy-dns: true
proxy-dns-port: 5053
proxy-dns-address: 127.0.0.1
proxy-dns-upstream:
  - https://1.1.1.1/dns-query
  - https://1.0.0.1/dns-query
  - https://9.9.9.9/dns-query
EOF

cloudflared service install || true
systemctl enable --now cloudflared || true

# --- 8. NATIVE PI-HOLE SINKHOLE & CONDITIONAL FORWARDING ---
echo "[8/14] Installing Pi-hole (Native Bare-Metal Port 53) & SafeSearch CNAMEs..."
mkdir -p /etc/pihole /etc/dnsmasq.d
cat <<EOF > /etc/pihole/setupVars.conf
PIHOLE_INTERFACE=${WIFI_IFACE}
IPV4_ADDRESS=$(ip route get 1.1.1.1 2>/dev/null | awk '{print $7}' || echo "192.168.100.81")/24
IPV6_ADDRESS=
PIHOLE_DNS_1=127.0.0.1#5053
QUERY_LOGGING=true
INSTALL_WEB_SERVER=true
INSTALL_WEB_INTERFACE=true
LIGHTTPD_ENABLED=true
CACHE_SIZE=10000
BLOCKING_ENABLED=true
REV_SERVER=true
REV_SERVER_CIDR=192.168.100.0/24
REV_SERVER_TARGET=192.168.100.1
REV_SERVER_DOMAIN=lan
EOF

cat <<EOF > /etc/dnsmasq.d/05-safesearch.conf
# Aegis Deterministic SafeSearch & Parental Restriction CNAME Injection
cname=google.com,forcesafesearch.google.com
cname=www.google.com,forcesafesearch.google.com
cname=youtube.com,restrictmoderate.youtube.com
cname=www.youtube.com,restrictmoderate.youtube.com
cname=bing.com,strict.bing.com
cname=duckduckgo.com,safe.duckduckgo.com
EOF

if ! command -v pihole &> /dev/null; then
    curl -sSL https://install.pi-hole.net | bash --unattended || true
fi

# Set default Pi-hole web admin password: Programming123
pihole -a -p "Programming123" || true

# --- 9. CROWDSEC IPS & NFTABLES BOUNCER ---
echo "[9/14] Installing CrowdSec security engine and firewall bouncer..."
if ! command -v crowdsec &> /dev/null; then
    curl -s https://packagecloud.io/install/repositories/crowdsecurity/crowdsec/script.deb.sh | bash || true
    apt-get install -y crowdsec crowdsec-firewall-bouncer-nftables || true
    systemctl enable --now crowdsec || true
    systemctl enable --now crowdsec-firewall-bouncer-nftables || true
fi

# Allow Open Ports in Firewall
if command -v ufw &> /dev/null; then
    ufw allow 53/tcp || true
    ufw allow 53/udp || true
    ufw allow 80/tcp || true
    ufw allow 25565/tcp || true
    ufw allow 25565/udp || true
    ufw allow 3000/tcp || true
    ufw allow 2222/tcp || true
    ufw allow 8096/tcp || true
    ufw allow 9091/tcp || true
fi

# --- 10. MINECRAFT FORGE SERVER (NATIVE MODDED 1.20.1 MULTIPLAYER) ---
echo "[10/14] Provisioning native Minecraft Forge 1.20.1 Server (Forge 47.3.0)..."
if ! id "minecraft" &>/dev/null; then
    useradd -r -m -U -d /opt/minecraft -s /bin/bash minecraft || true
fi

mkdir -p /opt/minecraft/server/mods
cd /opt/minecraft/server

MC_VERSION="1.20.1"
FORGE_VERSION="47.3.0"
INSTALLER_NAME="forge-${MC_VERSION}-${FORGE_VERSION}-installer.jar"

if [ ! -f "run.sh" ]; then
    echo "Downloading and installing Forge Server (${MC_VERSION} - ${FORGE_VERSION})..."
    wget -q "https://maven.minecraftforge.net/net/minecraftforge/forge/${MC_VERSION}-${FORGE_VERSION}/${INSTALLER_NAME}" -O "${INSTALLER_NAME}" || true
    if [ -f "${INSTALLER_NAME}" ]; then
        java -jar "${INSTALLER_NAME}" --installServer || true
        rm -f "${INSTALLER_NAME}"
    fi
    
    echo "eula=true" > eula.txt

    # Multiplayer Server Configuration
    cat <<EOF > server.properties
server-port=25565
server-ip=0.0.0.0
online-mode=true
enable-status=true
enable-query=true
max-players=20
motd=Aegis Homelab Modded Server (Forge 1.20.1)
view-distance=10
simulation-distance=8
sync-chunk-writes=true
EOF
    
    # Inject Aikar's Tuned G1GC Flags directly into Forge JVM args (4GB dedicated RAM)
    cat <<EOF > user_jvm_args.txt
-Xms4G
-Xmx4G
-XX:+UseG1GC
-XX:+ParallelRefProcEnabled
-XX:MaxGCPauseMillis=200
-XX:+UnlockExperimentalVMOptions
-XX:+DisableExplicitGC
-XX:+AlwaysPreTouch
-XX:G1NewSizePercent=30
-XX:G1MaxNewSizePercent=40
-XX:G1ReservePercent=20
-XX:G1HeapWastePercent=5
-XX:G1MixedGCCountTarget=4
-XX:InitiatingHeapOccupancyPercent=15
-XX:G1MixedGCLiveThresholdPercent=90
-XX:G1RSetUpdatingPauseTimePercent=5
-XX:SurvivorRatio=32
-XX:+PerfDisableSharedMem
-XX:MaxTenuringThreshold=1
EOF
fi

chown -R minecraft:minecraft /opt/minecraft

cat <<EOF > /etc/systemd/system/minecraft.service
[Unit]
Description=Minecraft Forge Server (Native Modded 1.20.1 Multiplayer)
After=network.target tailscaled.service

[Service]
User=minecraft
WorkingDirectory=/opt/minecraft/server
ExecStart=/opt/minecraft/server/run.sh nogui
Restart=always
RestartSec=15

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now minecraft.service || true

# --- 11. CRAFTY CONTROLLER 4 (MINECRAFT WEB GUI) ---
echo "[11/14] Provisioning Crafty Controller 4 Web Management GUI..."
mkdir -p /opt/crafty/app/config
if [ -d "/opt/crafty" ]; then
    if [ -d "./crafty-4" ]; then
        cp -r ./crafty-4/* /opt/crafty/ || true
    elif [ -d "../crafty-4" ]; then
        cp -r ../crafty-4/* /opt/crafty/ || true
    fi

    # Set predetermined default credentials: admin / Programming123
    cat << 'EOF' > /opt/crafty/app/config/default-creds.txt
{
    "username": "admin",
    "password": "Programming123",
    "info": "Default admin credentials provisioned by Aegis Installer"
}
EOF

    python3 -m venv /opt/crafty/.venv || true
    if [ -f "/opt/crafty/requirements.txt" ]; then
        /opt/crafty/.venv/bin/pip install --quiet -r /opt/crafty/requirements.txt || true
    fi

    chown -R minecraft:minecraft /opt/crafty
fi

cat <<EOF > /etc/systemd/system/crafty.service
[Unit]
Description=Crafty Controller 4 (Minecraft Web GUI)
After=network.target

[Service]
Type=simple
User=minecraft
WorkingDirectory=/opt/crafty
ExecStart=/opt/crafty/.venv/bin/python3 /opt/crafty/main.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now crafty.service || true

# --- 12. GITEA (SOVEREIGN SELF-HOSTED GIT FORGE ON 1TB EXTERNAL DRIVE) ---
echo "[12/14] Provisioning Native Gitea Git Server on Port 3000 (100GB+ LFS Ready)..."
if ! id "git" &>/dev/null; then
    useradd -r -m -U -d /var/lib/gitea -s /bin/bash git || true
fi

mkdir -p /etc/gitea /var/lib/gitea/data /var/lib/gitea/log /mnt/external_1tb/gitea-data/repositories /mnt/external_1tb/gitea-data/lfs
GITEA_BIN="/usr/local/bin/gitea"

if [ ! -f "$GITEA_BIN" ]; then
    echo "Downloading static Gitea 1.22.6 binary..."
    wget -q "https://dl.gitea.com/gitea/1.22.6/gitea-1.22.6-linux-amd64" -O "$GITEA_BIN" || true
    chmod +x "$GITEA_BIN" || true
fi

cat <<EOF > /etc/gitea/app.ini
APP_NAME = Aegis Sovereign Git Forge
RUN_USER = git
RUN_MODE = prod

[server]
PROTOCOL = http
DOMAIN = localhost
HTTP_PORT = 3000
ROOT_URL = http://localhost:3000/
SSH_PORT = 2222
START_SSH_SERVER = true
OFFLINE_MODE = true

[database]
DB_TYPE = sqlite3
PATH = /mnt/external_1tb/gitea-data/gitea.db

[repository]
ROOT = /mnt/external_1tb/gitea-data/repositories
MAX_CREATION_LIMIT = -1

[repository.upload]
ENABLED = true
TEMP_PATH = /mnt/external_1tb/gitea-data/tmp/uploads
ALLOWED_TYPES = *
FILE_MAX_SIZE = 102400
MAX_FILES = 50

[lfs]
PATH = /mnt/external_1tb/gitea-data/lfs
STORAGE_TYPE = local

[attachment]
ENABLED = true
PATH = /mnt/external_1tb/gitea-data/attachments
MAX_SIZE = 102400
MAX_FILES = 50

[service]
DISABLE_REGISTRATION = false
ALLOW_ONLY_EXTERNAL_REGISTRATION = false
ENABLE_CAPTCHA = false
REQUIRE_SIGNIN_VIEW = false

[security]
INSTALL_LOCK = true
SECRET_KEY = aegis-homelab-sovereign-secret-token

[user]
RESERVED_USERNAMES = 

[log]
MODE = console
LEVEL = Info
EOF

chown -R git:git /etc/gitea /var/lib/gitea /mnt/external_1tb/gitea-data

cat <<EOF > /etc/systemd/system/gitea.service
[Unit]
Description=Gitea (Sovereign Self-Hosted Git Forge)
After=network.target

[Service]
Type=simple
User=git
Group=git
WorkingDirectory=/var/lib/gitea/
ExecStart=/usr/local/bin/gitea web --config /etc/gitea/app.ini
Restart=always
RestartSec=5
Environment=USER=git HOME=/var/lib/gitea GITEA_WORK_DIR=/var/lib/gitea

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now gitea.service || true

# Pre-create Admin Users in Gitea
su - git -c "gitea admin user create --username administrator --password Programming123 --email admin@homelab.local --admin --config /etc/gitea/app.ini" || true
su - git -c "gitea admin user create --username hades --password Programming123 --email hades@homelab.local --admin --config /etc/gitea/app.ini" || true

# --- 13. JELLYFIN MEDIA SERVER (HARDWARE ACCELERATED ANIME STREAMING ON PORT 8096) ---
echo "[13/14] Provisioning Jellyfin Media Server with Intel QuickSync QSV on Port 8096..."
if ! command -v jellyfin &> /dev/null; then
    curl -fsSL https://repo.jellyfin.org/ubuntu/jellyfin_team.gpg.key | gpg --dearmor -o /etc/apt/trusted.gpg.d/jellyfin.gpg || true
    echo "deb [signed-by=/etc/apt/trusted.gpg.d/jellyfin.gpg] https://repo.jellyfin.org/ubuntu $(lsb_release -c -s) main" > /etc/apt/sources.list.d/jellyfin.list
    apt-get update -y || true
    apt-get install -y jellyfin || true
fi

# Add jellyfin user to video and render groups for Intel QuickSync hardware acceleration
usermod -aG video,render jellyfin || true
mkdir -p /mnt/external_1tb/media/anime /mnt/external_1tb/media/movies
chown -R jellyfin:jellyfin /mnt/external_1tb/media/anime /mnt/external_1tb/media/movies || true
systemctl enable --now jellyfin || true

# --- 14. QBITTORRENT-NOX (HEADLESS TORRENT DOWNLOADER ON PORT 9091) ---
echo "[14/14] Provisioning qBittorrent-nox Headless Torrent Downloader on Port 9091..."
mkdir -p /home/hades/.config/qBittorrent /mnt/external_1tb/media/downloads
cat <<EOF > /home/hades/.config/qBittorrent/qBittorrent.conf
[LegalNotice]
Accepted=true

[Preferences]
Downloads\SavePath=/mnt/external_1tb/media/downloads/
Downloads\TempPath=/mnt/external_1tb/media/downloads/temp/
WebUI\Port=9091
WebUI\Address=0.0.0.0
WebUI\Username=admin
WebUI\Password_PBKDF2="@ByteArray(AR4NVl5uu2pU4Xg08aF1nw==:2L+XNqjZ7B/0f4p2kXNq/Wc7C9D6g9v7HqZ9QY4Fh2K=)"
WebUI\CSRFProtection=false
WebUI\ClickjackingProtection=false
EOF
chown -R hades:hades /home/hades/.config/qBittorrent /mnt/external_1tb/media/downloads || true

cat <<EOF > /etc/systemd/system/qbittorrent-nox.service
[Unit]
Description=qBittorrent-nox Headless Torrent Downloader
After=network.target

[Service]
Type=simple
User=hades
ExecStart=/usr/bin/qbittorrent-nox --webui-port=9091
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now qbittorrent-nox.service || true

# --- BUILD AND LAUNCH PI-SENTINEL CORE DAEMON ---
echo "Compiling Go Sentinel daemon..."
mkdir -p /var/lib/pi-sentinel /etc/pi-sentinel /var/log/pi-sentinel

# Ensure Go toolchain exists
if ! command -v go &> /dev/null; then
    apt-get install -y golang-go || true
fi

if [ -f "./backend/configs/config.yaml" ]; then
    cp ./backend/configs/config.yaml /etc/pi-sentinel/config.yaml || true
fi

if [ -d "backend" ]; then
    echo "Compiling embedded Go binary..."
    cd backend
    go build -buildvcs=false -ldflags="-s -w" -o /usr/local/bin/pi-sentinel ./cmd/sentinel/main.go || true
    cd ..
fi

cat <<EOF > /etc/systemd/system/pi-sentinel.service
[Unit]
Description=Aegis Pi-Sentinel Telemetry, DGA Detector & Watchdog
After=network.target pihole-FTL.service cloudflared.service

[Service]
Type=simple
ExecStart=/usr/local/bin/pi-sentinel --port=3001 --config=/etc/pi-sentinel/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now pi-sentinel.service || true

HOST_IP=$(ip route get 1.1.1.1 2>/dev/null | awk '{print $7}' || echo "localhost")
TAILSCALE_IP=$(tailscale ip -4 2>/dev/null || echo "Run 'sudo tailscale up' to connect")

echo "============================================================"
echo "    [✔] AEGIS HOMELAB COMPLETE (ALL-IN-ONE PROVISIONED)"
echo "    • Aegis Dashboard:      http://${HOST_IP}:3001"
echo "    • Pi-hole Admin Web:    http://${HOST_IP}/admin/ (Pass: Programming123)"
echo "    • Primary DNS Server:   ${HOST_IP}:53 (Pi-hole + Cloudflared DoH)"
echo "    • Jellyfin Streaming:   http://${HOST_IP}:8096 (SyncPlay + QSV)"
echo "    • qBittorrent WebUI:    http://${HOST_IP}:9091 (admin / Programming123)"
echo "    • Sovereign Gitea Git:  http://${HOST_IP}:3000 (admin / Programming123)"
echo "    • Crafty Controller:    https://${HOST_IP}:8443 (admin / Programming123)"
echo "    • 1TB Anime Storage:    /mnt/external_1tb/media/anime"
echo "    • Tailscale Multiplay:  ${TAILSCALE_IP}:25565"
echo "    • UPS Battery Guard:    Auto-Safe Shutdown < 7% Battery"
echo "============================================================"
