# ProxVN Tunnel - Complete Protocol & Setup Guide

## 📑 Mục Lục
- [Tổng Quan Protocols](#tổng-quan-protocols)
- [Server Setup Chi Tiết](#server-setup-chi-tiết)
- [Client Setup Chi Tiết](#client-setup-chi-tiết)
- [Tất Cả Commands](#tất-cả-commands)
- [Configuration Reference](#configuration-reference)
- [Troubleshooting](#troubleshooting)

---

## 🔌 Tổng Quan Protocols

### TCP Tunneling
**Mô tả:** Tunnel TCP connections qua internet  
**Use cases:** SSH, RDP, databases, game servers  
**Ports:** Server tự động assign từ pool (10000-20000)

**Flow:**
```
Client Local Service (port X) 
  ↓ 
ProxVN Client 
  ↓ (TLS 1.3 encrypted tunnel)
ProxVN Server (port 8882)
  ↓
Public Port (e.g. 10500)
  ↓
External Users
```

### UDP Tunneling  
**Mô tả:** Tunnel UDP packets cho real-time apps  
**Use cases:** Gaming (Minecraft, Palworld), VoIP, streaming  
**Encryption:** Custom UDP encryption với shared secret

**Flow:**
```
Client Local Service (UDP port X)
  ↓
ProxVN Client (UDP control channel)
  ↓ (Encrypted UDP packets)
ProxVN Server (port 8882 UDP)
  ↓  
Public UDP Port
  ↓
External Users
```

### HTTP Tunneling
**Mô tả:** Tạo subdomain HTTPS tự động  
**Use cases:** Web dev, webhooks, APIs  
**Requirements:** Server phải có HTTP_DOMAIN configured

**Flow:**
```
Client Local HTTP Server (port 3000)
  ↓
ProxVN Client (--proto http)
  ↓
ProxVN HTTP Proxy (port 443)
  ↓
https://random-id.yourdomain.com
  ↓
External Users
```

### File Sharing (WebDAV)
**Mô tả:** Share folders qua HTTPS  
**Features:** Upload, download, password protection  
**Protocols:** WebDAV, HTTP

**Flow:**
```
Local Directory
  ↓
ProxVN Client (--file ~/folder)
  ↓
ProxVN File Server
  ↓
https://file-id.yourdomain.com/browse
  ↓
Browser Access (password protected)
```

---

## 🖥️ Server Setup Chi Tiết

### Phương Pháp 1: Docker (Khuyến Nghị)

#### Bước 1: Chuẩn Bị
```bash
# Clone repository
git clone https://github.com/hoangtuvungcao/proxvn_tunnel.git
cd proxvn_tunnel

# Tạo directories
mkdir -p data backups logs
```

#### Bước 2: Cấu Hình
```bash
# Copy file env mẫu
cp .env.server.example .env

# Edit cấu hình (BẮT BUỘC thay đổi các giá trị sau)
nano .env
```

**Cấu hình tối thiểu (.env):**
```bash
# === BẮT BUỘC ===
SERVER_PORT=8882
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_very_strong_password_here_change_this
JWT_SECRET=your_super_secret_jwt_key_minimum_32_characters

# === TÙY CHỌN (cho HTTP tunneling) ===
HTTP_DOMAIN=yourdomain.com
TLS_CERT_FILE=./wildcard.crt
TLS_KEY_FILE=./wildcard.key

# === Performance ===
MAX_CONNECTIONS=10000
ENABLE_COMPRESSION=true
ENABLE_HTTP2=true
WORKER_POOL_SIZE=100

# === Security ===
TLS_MIN_VERSION=1.3
RATE_LIMIT_RPS=10
ENABLE_DDOS_PROTECTION=true
```

#### Bước 3: Start Server
```bash
# Start với docker-compose
docker-compose up -d

# Check logs
docker-compose logs -f proxvn-server

# Check status
docker-compose ps
```

#### Bước 4: Verify
```bash
# Test health endpoint
curl http://localhost:8881/health

# Expected response:
# {"status":"ok","server":"ProxVN by TrongDev","version":"7.1.0"}

# Access dashboard
# Browser: http://YOUR_SERVER_IP:8881/dashboard/
```

---

### Phương Pháp 2: Binary (Linux)

#### Bước 1: Download Binary
```bash
# Tạo directory
sudo mkdir -p /opt/proxvn
cd /opt/proxvn

# Download latest release
wget https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-server-linux-amd64

# Make executable
chmod +x proxvn-server-linux-amd64

# Create symlink
sudo ln -s /opt/proxvn/proxvn-server-linux-amd64 /usr/local/bin/proxvn-server
```

#### Bước 2: Tạo .env File
```bash
sudo nano /opt/proxvn/.env

# Paste cấu hình (xem mẫu ở trên)
```

#### Bước 3: Tạo Systemd Service
```bash
sudo nano /etc/systemd/system/proxvn.service
```

**Nội dung file:**
```ini
[Unit]
Description=ProxVN Tunnel Server
After=network.target

[Service]
Type=simple
User=proxvn
Group=proxvn
WorkingDirectory=/opt/proxvn
EnvironmentFile=/opt/proxvn/.env
ExecStart=/usr/local/bin/proxvn-server
Restart=always
RestartSec=10
StandardOutput=append:/var/log/proxvn/server.log
StandardError=append:/var/log/proxvn/error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/proxvn/data /opt/proxvn/backups

[Install]
WantedBy=multi-user.target
```

#### Bước 4: Tạo User và Start Service
```bash
# Tạo user riêng
sudo useradd -r -s /bin/false proxvn

# Tạo log directory
sudo mkdir -p /var/log/proxvn
sudo chown proxvn:proxvn /var/log/proxvn

# Set permissions
sudo chown -R proxvn:proxvn /opt/proxvn

# Enable và start service
sudo systemctl daemon-reload
sudo systemctl enable proxvn
sudo systemctl start proxvn

# Check status
sudo systemctl status proxvn

# View logs
sudo journalctl -u proxvn -f
```

---

### Phương Pháp 3: Build từ Source

```bash
# Prerequisites
sudo apt install golang-go build-essential sqlite3

# Clone
git clone https://github.com/hoangtuvungcao/proxvn_tunnel.git
cd proxvn_tunnel

# Build
cd src/backend
go build -o ../../bin/proxvn-server ./cmd/server

# Verify
../../bin/proxvn-server --version
```

---

## 💻 Client Setup Chi Tiết

### Linux

#### Install
```bash
# Download
wget https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-linux-amd64

# Make executable
chmod +x proxvn-linux-amd64

# Move to PATH
sudo mv proxvn-linux-amd64 /usr/local/bin/proxvn

# Verify
proxvn --version
```

#### Use
```bash
# TCP tunnel (SSH example)
proxvn 22

# HTTP tunnel
proxvn --proto http 3000

# UDP tunnel (Minecraft)
proxvn --proto udp 19132

# File share
proxvn --file ~/Documents --pass mypassword
```

---

### Windows

#### Install
```powershell
# Download từ GitHub Releases
# https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-windows-amd64.exe

# Lưu vào C:\Program Files\ProxVN\
# Thêm vào PATH (System Properties > Environment Variables)
```

#### Use (PowerShell/CMD)
```powershell
# TCP tunnel
proxvn.exe 3389

# HTTP tunnel
proxvn.exe --proto http 8000

# UDP tunnel
proxvn.exe --proto udp 25565
```

---

### macOS

#### Install
```bash
# Intel Mac
curl -LO https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-darwin-amd64
chmod +x proxvn-darwin-amd64
sudo mv proxvn-darwin-amd64 /usr/local/bin/proxvn

# M1/M2/M3 Mac
curl -LO https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-darwin-arm64
chmod +x proxvn-darwin-arm64
sudo mv proxvn-darwin-arm64 /usr/local/bin/proxvn
```

---

## 📝 Tất Cả Commands

### Server Commands

```bash
# Start server
./proxvn-server

# Start với custom port
./proxvn-server -port 8881

# Start với .env file
# (Tự động load .env nếu có trong cùng directory)
./proxvn-server

# Check version
./proxvn-server --version

# Help
./proxvn-server --help
```

### Client Commands - Syntax

```bash
proxvn [OPTIONS] [LOCAL_PORT]
```

### TCP Tunneling Commands

```bash
# Basic - tunnel local port 8080
proxvn 8080

# SSH server
proxvn 22

# MySQL
proxvn 3306

# PostgreSQL  
proxvn 5432

# Redis
proxvn 6379

# MongoDB
proxvn 27017

# RDP (Windows Remote Desktop)
proxvn 3389

# VNC
proxvn 5900

# Custom server
proxvn --server your-server.com:8882 8080

# Với specific remote port
proxvn --remote-port 10500 8080
```

### HTTP Tunneling Commands

```bash
# Tunnel local HTTP server port 3000
proxvn --proto http 3000

# React/Vite dev server
proxvn --proto http 5173

# Next.js
proxvn --proto http 3000

# Laravel
proxvn --proto http 8000

# Django
proxvn --proto http 8000

# Custom subdomain
proxvn --proto http --subdomain myapp 3000
# → https://myapp.yourdomain.com

# Với custom server
proxvn --server your-server.com:8882 --proto http 3000
```

### UDP Tunneling Commands

```bash
# Minecraft Bedrock
proxvn --proto udp 19132

# Minecraft Java
proxvn --proto udp 25565

# Palworld
proxvn --proto udp 8211

# Rust Game Server
proxvn --proto udp 28015

# Counter-Strike
proxvn --proto udp 27015

# Voice chat
proxvn --proto udp 50000
```

### File Sharing Commands

```bash
# Share folder read-only
proxvn --file ~/Documents --pass doc123

# Share với upload permission
proxvn --file ~/Shared --pass upload123 --permissions rw

# Share với full permissions
proxvn --file ~/Projects --pass dev2024 --permissions rwx

# Share specific directory
proxvn --file /var/www/html --pass web123 --permissions r
```

### Advanced Options

```bash
# Custom buffer size (for high bandwidth)
proxvn --buffer-size 65536 --proto tcp 8080

# Disable compression (for pre-compressed data)
proxvn --compression=false 9000

# Custom timeout
proxvn --timeout 60s --proto http 3000

# Max reconnect attempts
proxvn --max-reconnect 20 --reconnect-delay 3s 8080

# Verbose logging
proxvn --log-level debug --proto http 3000

# No UI (for scripts)
proxvn --no-ui --quiet 8080 > tunnel.log 2>&1 &

# Multiple options combined
proxvn --proto http \
  --subdomain myapp \
  --server your-server.com:8882 \
  --buffer-size 65536 \
  --log-level info \
  3000
```

---

## ⚙️ Configuration Reference

### Server Environment Variables (.env)

| Category | Variable | Default | Description |
|----------|----------|---------|-------------|
| **Server** | `SERVER_HOST` | 0.0.0.0 | Bind address |
| | `SERVER_PORT` | 8882 | Server port (tunnel runs on PORT+1) |
| | `PUBLIC_PORT_START` | 10000 | Public port range start |
| | `PUBLIC_PORT_END` | 20000 | Public port range end |
| **Admin** | `ADMIN_USERNAME` | admin | Admin dashboard username |
| | `ADMIN_PASSWORD` | - | **REQUIRED** Admin password |
| | `JWT_SECRET` | - | **REQUIRED** JWT signing key |
| **Database** | `DB_PATH` | ./proxvn.db | SQLite database file |
| **Performance** | `MAX_CONNECTIONS` | 10000 | Max concurrent connections |
| | `BUFFER_SIZE` | 32768 | Buffer size (bytes) |
| | `ENABLE_COMPRESSION` | true | Enable gzip/zstd |
| | `COMPRESSION_LEVEL` | 6 | 1-9, higher = better compression |
| | `ENABLE_HTTP2` | true | Enable HTTP/2 |
| | `ENABLE_CACHE` | true | Enable response caching |
| | `CACHE_SIZE_MB` | 256 | Cache size limit |
| | `WORKER_POOL_SIZE` | 100 | Goroutine pool size |
| **Security** | `TLS_MIN_VERSION` | 1.3 | Minimum TLS version |
| | `RATE_LIMIT_RPS` | 10 | Requests per second limit |
| | `RATE_LIMIT_BURST` | 20 | Burst capacity |
| | `ENABLE_DDOS_PROTECTION` | true | Auto-block excessive requests |
| **HTTP Tunnel** | `HTTP_DOMAIN` | - | Base domain (e.g. example.com) |
| | `TLS_CERT_FILE` | ./wildcard.crt | SSL certificate for HTTPS |
| | `TLS_KEY_FILE` | ./wildcard.key | SSL private key |
| **Backup** | `AUTO_BACKUP` | true | Enable auto backup |
| | `BACKUP_INTERVAL` | 24h | Backup frequency |
| | `BACKUP_DIR` | ./backups | Backup directory |
| | `BACKUP_RETENTION_DAYS` | 7 | Days to keep backups |

Xem [.env.server.example](.env.server.example) cho tất cả 150+ options!

---

## 🔍 Troubleshooting

### Server Issues

#### Port Already in Use
```bash
# Check what's using port 8882
sudo lsof -i :8882
# Or
sudo netstat -tlnp | grep 8882

# Kill process
sudo kill -9 <PID>

# Or change port in .env
SERVER_PORT=8883
```

#### Database Locked
```bash
# Stop server
sudo systemctl stop proxvn

# Backup database
cp data/proxvn.db data/proxvn.db.backup

# Remove WAL files
rm data/proxvn.db-shm data/proxvn.db-wal

# Restart
sudo systemctl start proxvn
```

#### High Memory Usage
```bash
# Check memory
docker stats proxvn-server

# Reduce cache size in .env
CACHE_SIZE_MB=128
MAX_CONNECTIONS=5000

# Restart
docker-compose restart proxvn-server
```

### Client Issues

#### Connection Failed
```bash
# Test server connectivity
telnet your-server.com 8882

# Test HTTPS
curl https://your-server.com:8881/health

# Try insecure mode (testing only)
proxvn --insecure --server your-server.com:8882 3000
```

#### Certificate Errors
```bash
# Download server cert
openssl s_client -connect your-server.com:8882 -showcerts

# Use cert pinning
proxvn --cert-pin <SHA256_HASH> 3000
```

#### Slow Performance
```bash
# Increase buffer
proxvn --buffer-size 131072 8080

# Disable compression for binary data
proxvn --compression=false 9000

# Check network
ping your-server.com
traceroute your-server.com
```

---

## 📚 Tài Liệu Khác

- [CLIENT_GUIDE.md](CLIENT_GUIDE.md) - 50+ ví dụ chi tiết
- [DEPLOYMENT.md](DEPLOYMENT.md) - Production deployment  
- [README.md](README.md) - Project overview
- [.env.server.example](.env.server.example) - Full config reference

---

## 🆘 Support

- 📧 Email: trong20843@gmail.com
- 💬 Telegram: [t.me/proxvn](https://t.me/proxvn)
- 🐛 Issues: [GitHub Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)

---

**Made with ❤️ by TrongDev**  
**Version 7.1.0**
