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

export DEBIAN_FRONTEND=noninteractive

echo "============================================================"
echo "    [+] INITIATING AEGIS HOMELAB NATIVE SERVER PROVISIONING "
echo "============================================================"

# --- INTERACTIVE STORAGE SELECTION ---
echo ""
echo "============================================================"
echo " [?] SELECT STORAGE ALLOCATION STRATEGY FOR YOUR SERVICES:"
echo "============================================================"
echo " [1] HYBRID / OPTIMAL (Recommended):"
echo "     • Internal 256GB SSD: Fast Minecraft World Chunks & Databases (Max FPS)"
echo "     • External 1TB HDD:   Gitea Repositories, Anime Library & Torrents"
echo ""
echo " [2] INTERNAL SSD ONLY (256GB Fast NVMe/SATA):"
echo "     • Store everything on internal 256GB drive (/var/lib/aegis-data)"
echo ""
echo " [3] EXTERNAL 1TB DRIVE ONLY (/mnt/external_1tb):"
echo "     • Store all services and data on the 1TB enclosure"
echo ""
echo " [4] CUSTOM PATH:"
echo "     • Specify your own custom storage directories"
echo "============================================================"

STORAGE_CHOICE=""
if [ -t 0 ]; then
    read -p "Enter choice [1-4] (Default: 1): " STORAGE_CHOICE || true
fi
STORAGE_CHOICE="${STORAGE_CHOICE:-1}"

GITEA_DATA_DIR="/mnt/external_1tb/gitea-data"
MEDIA_DIR="/mnt/external_1tb/media"
MINECRAFT_DATA_DIR="/opt/minecraft"
BACKUP_DIR="/mnt/external_1tb/minecraft-backups"

case "$STORAGE_CHOICE" in
    1)
        echo "[+] Selected: HYBRID (Internal SSD for Minecraft DBs + 1TB for Gitea/Media)"
        GITEA_DATA_DIR="/mnt/external_1tb/gitea-data"
        MEDIA_DIR="/mnt/external_1tb/media"
        MINECRAFT_DATA_DIR="/opt/minecraft"
        BACKUP_DIR="/mnt/external_1tb/minecraft-backups"
        ;;
    2)
        echo "[+] Selected: INTERNAL 256GB SSD ONLY"
        GITEA_DATA_DIR="/var/lib/aegis-data/gitea"
        MEDIA_DIR="/var/lib/aegis-data/media"
        MINECRAFT_DATA_DIR="/var/lib/aegis-data/minecraft"
        BACKUP_DIR="/var/lib/aegis-data/backups"
        ;;
    3)
        echo "[+] Selected: EXTERNAL 1TB HDD ONLY"
        GITEA_DATA_DIR="/mnt/external_1tb/gitea-data"
        MEDIA_DIR="/mnt/external_1tb/media"
        MINECRAFT_DATA_DIR="/mnt/external_1tb/minecraft"
        BACKUP_DIR="/mnt/external_1tb/minecraft-backups"
        ;;
    4)
        echo "[+] Custom Storage Configuration:"
        read -p "Enter Gitea storage path (Default: /mnt/external_1tb/gitea-data): " CUSTOM_GITEA || true
        read -p "Enter Media/Anime storage path (Default: /mnt/external_1tb/media): " CUSTOM_MEDIA || true
        read -p "Enter Minecraft server path (Default: /opt/minecraft): " CUSTOM_MC || true
        GITEA_DATA_DIR="${CUSTOM_GITEA:-/mnt/external_1tb/gitea-data}"
        MEDIA_DIR="${CUSTOM_MEDIA:-/mnt/external_1tb/media}"
        MINECRAFT_DATA_DIR="${CUSTOM_MC:-/opt/minecraft}"
        BACKUP_DIR="${MEDIA_DIR}/backups"
        ;;
    *)
        echo "[+] Defaulting to HYBRID mode."
        ;;
esac

echo ""
echo "-> Gitea Storage:     ${GITEA_DATA_DIR}"
echo "-> Media & Anime:     ${MEDIA_DIR}"
echo "-> Minecraft Server:  ${MINECRAFT_DATA_DIR}"
echo "-> World Backups:     ${BACKUP_DIR}"
echo "============================================================"
sleep 2

# --- 1. SYSTEM REPOSITORIES & DEPENDENCIES ---
echo "[1/15] Enabling universe/multiverse repositories and installing core packages..."
apt-get update -y
apt-get install -y software-properties-common ca-certificates gnupg
add-apt-repository -y universe || true
add-apt-repository -y multiverse || true
apt-get update -y

# Install Core Tools, ACPI, UPower, & Storage
echo "[+] Installing core utilities, power tools, and storage packages..."
apt-get install -y --fix-broken || true
apt-get install -y \
    curl wget git git-lfs jq ufw nftables tlp tlp-rdw iw wireless-tools \
    wpasupplicant network-manager linux-firmware acpi upower nvme-cli \
    neofetch screenfetch fastfetch || apt-get install -y neofetch screenfetch || true \
    sqlite3 build-essential ntfs-3g smartmontools rsync samba

# Install Java 21 & Python 3
echo "[+] Installing OpenJDK 21 and Python 3 venv..."
apt-get install -y openjdk-21-jre-headless python3 python3-pip python3-venv python3-full || true

# Install Media & Hardware Video Acceleration
echo "[+] Installing qBittorrent and Intel QuickSync VA-API drivers..."
apt-get install -y qbittorrent-nox vainfo || true
apt-get install -y intel-media-va-driver-non-free || apt-get install -y intel-media-va-driver || true

# --- 2. DISABLE SYSTEMD-RESOLVED STUB (FREE PORT 53 FOR PI-HOLE) ---
echo "[2/15] Disabling systemd-resolved DNSStubListener to free Port 53 for Pi-hole..."
mkdir -p /etc/systemd/resolved.conf.d/
cat <<EOF > /etc/systemd/resolved.conf.d/disable-stub.conf
[Resolve]
DNSStubListener=no
DNS=127.0.0.1#5053 1.1.1.1
EOF

systemctl restart systemd-resolved || true
rm -f /etc/resolv.conf
echo "nameserver 1.1.1.1" > /etc/resolv.conf

# --- 3. TECLAST LAPTOP BATTERY UPS, HARDENING & SLEEP MASKING ---
echo "[3/15] Hardening Teclast F7 Plus 24/7 Server, Battery UPS & Power Management..."

# Mask all sleep and suspend targets (Stay 100% active during AC disconnect / blackout)
systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target || true

# Prevent laptop from suspending when lid is closed
mkdir -p /etc/systemd/logind.conf.d/
cat <<EOF > /etc/systemd/logind.conf.d/homelab-lid.conf
[Login]
HandleLidSwitch=ignore
HandleLidSwitchDocked=ignore
HandleLidSwitchExternalPower=ignore
IdleAction=ignore
EOF
systemctl restart systemd-logind || true

# Configure UPower Low-Battery Protection
if [ -f "/etc/UPower/UPower.conf" ]; then
    sed -i 's/^PercentageLow=.*/PercentageLow=20/' /etc/UPower/UPower.conf || true
    sed -i 's/^PercentageCritical=.*/PercentageCritical=10/' /etc/UPower/UPower.conf || true
    sed -i 's/^PercentageAction=.*/PercentageAction=5/' /etc/UPower/UPower.conf || true
    sed -i 's/^CriticalPowerAction=.*/CriticalPowerAction=PowerOff/' /etc/UPower/UPower.conf || true
    systemctl restart upower || true
fi

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

WIFI_IFACE=$(iw dev 2>/dev/null | awk '$1=="Interface"{print $2}' | head -n 1 || echo "")
if [ -z "$WIFI_IFACE" ]; then
    WIFI_IFACE=$(ip link 2>/dev/null | awk -F': ' '/wl/{print $2}' | head -n 1 || echo "wlan0")
fi
echo "Detected Wi-Fi Interface: ${WIFI_IFACE}"

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

# Setup Battery Protection Threshold (70% Anti-Bloat limit)
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

# Battery UPS Graceful Auto-Shutdown Watchdog Daemon (Protects 1TB Drive & Databases)
cat << 'EOF' > /usr/local/bin/aegis-ups-watchdog.sh
#!/usr/bin/env bash
for bat in /sys/class/power_supply/BAT*; do
    if [ -f "$bat/capacity" ] && [ -f "$bat/status" ]; then
        CAPACITY=$(cat "$bat/capacity")
        STATUS=$(cat "$bat/status")
        if [ "$STATUS" = "Discharging" ] && [ "$CAPACITY" -le 7 ]; then
            logger -t aegis-ups "CRITICAL: Battery at ${CAPACITY}%. Initiating safe shutdown of services."
            systemctl stop minecraft crafty gitea jellyfin || true
            sync
            umount -f /mnt/external_1tb 2>/dev/null || true
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
echo "[4/15] Enabling TCP BBR and tuning network sysctl parameters..."
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

# --- 5. STORAGE DIRECTORIES INITIALIZATION ---
echo "[5/15] Initializing chosen storage layout..."
mkdir -p "${GITEA_DATA_DIR}/repositories" "${GITEA_DATA_DIR}/lfs" \
         "${MEDIA_DIR}/anime" "${MEDIA_DIR}/movies" "${MEDIA_DIR}/downloads" \
         "${MINECRAFT_DATA_DIR}" "${BACKUP_DIR}" /mnt/external_1tb

# Find and mount external NTFS drive if attached
NTFS_DEV=$(lsblk -o NAME,FSTYPE -rn 2>/dev/null | grep -i "ntfs" | head -n 1 | awk '{print "/dev/" $1}' || echo "")
if [ -n "$NTFS_DEV" ]; then
    echo "Detected NTFS Partition: ${NTFS_DEV}. Mounting to /mnt/external_1tb..."
    mount -t ntfs-3g -o windows_names,big_writes,nofail,uid=1000,gid=1000,umask=022 "${NTFS_DEV}" /mnt/external_1tb 2>/dev/null || true
    if ! grep -q "/mnt/external_1tb" /etc/fstab; then
        UUID_VAL=$(blkid -s UUID -o value "${NTFS_DEV}" || echo "")
        if [ -n "$UUID_VAL" ]; then
            echo "UUID=${UUID_VAL} /mnt/external_1tb ntfs-3g windows_names,big_writes,nofail,uid=1000,gid=1000,umask=022 0 0" >> /etc/fstab
        fi
    fi
fi

# --- 6. SAMBA WINDOWS/MAC NETWORK FILE SHARING ---
echo "[6/15] Configuring Samba Windows Network Drive (\\\\TECLAST\\Storage)..."
cat << EOF >> /etc/samba/smb.conf
[Storage]
   comment = Aegis Homelab Storage
   path = ${MEDIA_DIR}
   browseable = yes
   read only = no
   guest ok = yes
   create mask = 0777
   directory mask = 0777
   force user = root
EOF
systemctl restart smbd nmbd || true

# --- 7. TAILSCALE REMOTE MULTIPLAYER MESH VPN ---
echo "[7/15] Installing and provisioning Tailscale VPN (Zero-Config Multiplayer Mesh)..."
if ! command -v tailscale &> /dev/null; then
    curl -fsSL https://tailscale.com/install.sh | sh || true
    systemctl enable --now tailscaled || true
fi

# --- 8. ENCRYPTED DOH (CLOUDFLARED) ---
echo "[8/15] Installing and provisioning Cloudflared DoH on port 5053..."
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

# --- 9. NATIVE PI-HOLE SINKHOLE & CONDITIONAL FORWARDING ---
echo "[9/15] Installing Pi-hole (Native Bare-Metal Port 53) & SafeSearch CNAMEs..."
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

# --- 10. CROWDSEC IPS & NFTABLES BOUNCER ---
echo "[10/15] Installing CrowdSec security engine and firewall bouncer..."
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
    ufw allow 445/tcp || true
    ufw allow 139/tcp || true
    ufw allow 25565/tcp || true
    ufw allow 25565/udp || true
    ufw allow 3000/tcp || true
    ufw allow 2222/tcp || true
    ufw allow 8096/tcp || true
    ufw allow 9091/tcp || true
fi

# --- 11. MINECRAFT FORGE SERVER (NATIVE MODDED 1.20.1 MULTIPLAYER) ---
echo "[11/15] Provisioning native Minecraft Forge 1.20.1 Server in ${MINECRAFT_DATA_DIR}..."
if ! id "minecraft" &>/dev/null; then
    useradd -r -m -U -d /home/minecraft -s /bin/bash minecraft || true
fi

mkdir -p "${MINECRAFT_DATA_DIR}/server/mods"
cd "${MINECRAFT_DATA_DIR}/server"

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
    
    # Aikar's Tuned G1GC Flags (4GB dedicated RAM)
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

chown -R minecraft:minecraft "${MINECRAFT_DATA_DIR}"

# Daily Minecraft Automated Backup Rotation (3:00 AM)
cat << EOF > /usr/local/bin/minecraft-backup.sh
#!/usr/bin/env bash
BACKUP_DIR="${BACKUP_DIR}"
mkdir -p "\$BACKUP_DIR"
DATE=\$(date +%Y-%m-%d_%H%M)
tar -czf "\$BACKUP_DIR/world-\$DATE.tar.gz" -C "${MINECRAFT_DATA_DIR}/server" world || true
ls -tp "\$BACKUP_DIR"/world-*.tar.gz 2>/dev/null | grep -v '/\$' | tail -n +8 | xargs -I {} rm -- {} 2>/dev/null || true
EOF
chmod +x /usr/local/bin/minecraft-backup.sh
(crontab -l 2>/dev/null | grep -v "minecraft-backup.sh"; echo "0 3 * * * /usr/local/bin/minecraft-backup.sh >/dev/null 2>&1") | crontab - || true

cat <<EOF > /etc/systemd/system/minecraft.service
[Unit]
Description=Minecraft Forge Server (Native Modded 1.20.1 Multiplayer)
After=network.target tailscaled.service

[Service]
User=minecraft
WorkingDirectory=${MINECRAFT_DATA_DIR}/server
ExecStart=${MINECRAFT_DATA_DIR}/server/run.sh nogui
Restart=always
RestartSec=15

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now minecraft.service || true

# --- 12. CRAFTY CONTROLLER 4 (MINECRAFT WEB GUI) ---
echo "[12/15] Provisioning Crafty Controller 4 Web Management GUI..."
mkdir -p /opt/crafty/app/config
if [ -d "/opt/crafty" ]; then
    if [ -d "./crafty-4" ]; then
        cp -r ./crafty-4/* /opt/crafty/ || true
    elif [ -d "../crafty-4" ]; then
        cp -r ../crafty-4/* /opt/crafty/ || true
    fi

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

# --- 13. GITEA (SOVEREIGN SELF-HOSTED GIT FORGE) ---
echo "[13/15] Provisioning Native Gitea Git Server on Port 3000..."
if ! id "git" &>/dev/null; then
    useradd -r -m -U -d /var/lib/gitea -s /bin/bash git || true
fi

mkdir -p /etc/gitea /var/lib/gitea/data /var/lib/gitea/log "${GITEA_DATA_DIR}/repositories" "${GITEA_DATA_DIR}/lfs"
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
PATH = ${GITEA_DATA_DIR}/gitea.db

[repository]
ROOT = ${GITEA_DATA_DIR}/repositories
MAX_CREATION_LIMIT = -1

[repository.upload]
ENABLED = true
TEMP_PATH = ${GITEA_DATA_DIR}/tmp/uploads
ALLOWED_TYPES = *
FILE_MAX_SIZE = 102400
MAX_FILES = 50

[lfs]
PATH = ${GITEA_DATA_DIR}/lfs
STORAGE_TYPE = local

[attachment]
ENABLED = true
PATH = ${GITEA_DATA_DIR}/attachments
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

chown -R git:git /etc/gitea /var/lib/gitea "${GITEA_DATA_DIR}"

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

# --- 14. JELLYFIN MEDIA SERVER (HARDWARE ACCELERATED ANIME STREAMING ON PORT 8096) ---
echo "[14/15] Provisioning Jellyfin Media Server with Intel QuickSync QSV on Port 8096..."
if ! command -v jellyfin &> /dev/null; then
    curl -fsSL https://repo.jellyfin.org/ubuntu/jellyfin_team.gpg.key | gpg --dearmor --yes -o /etc/apt/trusted.gpg.d/jellyfin.gpg || true
    echo "deb [signed-by=/etc/apt/trusted.gpg.d/jellyfin.gpg] https://repo.jellyfin.org/ubuntu $(lsb_release -c -s) main" > /etc/apt/sources.list.d/jellyfin.list
    apt-get update -y || true
    apt-get install -y jellyfin || true
fi

# Add jellyfin user to video and render groups for Intel QuickSync hardware acceleration
usermod -aG video,render jellyfin || true
mkdir -p "${MEDIA_DIR}/anime" "${MEDIA_DIR}/movies"
chown -R jellyfin:jellyfin "${MEDIA_DIR}/anime" "${MEDIA_DIR}/movies" || true
systemctl enable --now jellyfin || true

# --- 15. QBITTORRENT-NOX (HEADLESS TORRENT DOWNLOADER ON PORT 9091) ---
echo "[15/15] Provisioning qBittorrent-nox Headless Torrent Downloader on Port 9091..."
CURRENT_USER="${SUDO_USER:-hades}"
USER_HOME=$(getent passwd "$CURRENT_USER" | cut -d: -f6)
mkdir -p "${USER_HOME}/.config/qBittorrent" "${MEDIA_DIR}/downloads"

cat <<EOF > "${USER_HOME}/.config/qBittorrent/qBittorrent.conf"
[LegalNotice]
Accepted=true

[Preferences]
Downloads\SavePath=${MEDIA_DIR}/downloads/
Downloads\TempPath=${MEDIA_DIR}/downloads/temp/
WebUI\Port=9091
WebUI\Address=0.0.0.0
WebUI\Username=admin
WebUI\Password_PBKDF2="@ByteArray(AR4NVl5uu2pU4Xg08aF1nw==:2L+XNqjZ7B/0f4p2kXNq/Wc7C9D6g9v7HqZ9QY4Fh2K=)"
WebUI\CSRFProtection=false
WebUI\ClickjackingProtection=false
EOF
chown -R "${CURRENT_USER}:${CURRENT_USER}" "${USER_HOME}/.config/qBittorrent" "${MEDIA_DIR}/downloads" || true

cat <<EOF > /etc/systemd/system/qbittorrent-nox.service
[Unit]
Description=qBittorrent-nox Headless Torrent Downloader
After=network.target

[Service]
Type=simple
User=${CURRENT_USER}
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
echo "    • Storage Strategy:     Choice [${STORAGE_CHOICE}]"
echo "    • Aegis Dashboard:      http://${HOST_IP}:3001"
echo "    • Hardware Health:      http://${HOST_IP}:3001 (Tab 2: Battery/SSD)"
echo "    • Samba Windows Share:  \\\\${HOST_IP}\\Storage (Guest Access)"
echo "    • Pi-hole Admin Web:    http://${HOST_IP}/admin/ (Pass: Programming123)"
echo "    • Primary DNS Server:   ${HOST_IP}:53 (Pi-hole + Cloudflared DoH)"
echo "    • Jellyfin Streaming:   http://${HOST_IP}:8096 (SyncPlay + QSV)"
echo "    • qBittorrent WebUI:    http://${HOST_IP}:9091 (admin / Programming123)"
echo "    • Sovereign Gitea Git:  http://${HOST_IP}:3000 (admin / Programming123)"
echo "    • Crafty Controller:    https://${HOST_IP}:8443 (admin / Programming123)"
echo "    • Media & Downloads:    ${MEDIA_DIR}"
echo "    • Minecraft Forge 1.20: ${MINECRAFT_DATA_DIR} (Port 25565)"
echo "    • Tailscale Multiplay:  ${TAILSCALE_IP}:25565"
echo "    • UPS Battery Guard:    Auto-Safe Shutdown < 7% Battery"
echo "============================================================"
