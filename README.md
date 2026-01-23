# ProxVN Tunnel Platform

**ProxVN Tunnel** là giải pháp tunneling mạnh mẽ, an toàn và dễ sử dụng, cho phép bạn đưa các dịch vụ local (localhost) ra Internet ngay lập tức. Được xây dựng bằng Golang với hiệu năng cao, hỗ trợ đa nền tảng và đầy đủ các tính năng nâng cao.

<p align="left">
  <a href="https://go.dev/" target="_blank"><img src="https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square" alt="Go" /></a>
  <a href="#documentation"><img src="https://img.shields.io/badge/Docs-Complete-success?style=flat-square" alt="Docs" /></a>
  <a href="#license"><img src="https://img.shields.io/badge/License-Free%20for%20Non--Commercial-blue?style=flat-square" alt="License" /></a>
</p>

---

## 🌟 Tính Năng Nổi Bật

*   **Đa Giao Thức (Multi-Protocol)**:
    *   **HTTP/HTTPS**: Tự động cấp Subdomain HTTPS (SSL) cho web app.
    *   **TCP**: Forwarding port cho SSH, RDP, Database, v.v.
    *   **UDP**: Hỗ trợ Game Server (Minecraft, CS:GO, Palworld...) và các ứng dụng realtime.
    *   **File Sharing**: Chia sẻ file/folder an toàn như Google Drive hoặc ổ đĩa mạng (WebDAV).
*   **Bảo Mật Cao (Zero-Trust Security)**:
    *   Mã hóa toàn diện (TLS 1.3) cho mọi kết nối Control & Data.
    *   Hỗ trợ xác thực JWT, Rate Limiting, chống DDoS.
    *   Certificate Pinning để đảm bảo kết nối đến đúng server.
    *   Chế độ "Private" với Password bảo vệ.
*   **Quản Lý Toàn Diện**:
    *   **Web Dashboard**: Giao diện trực quan xem trạng thái, băng thông, connections.
    *   **In-Browser Editor**: Sửa code/text trực tiếp trên trình duyệt mà không cần tải về.
*   **Hiệu Năng Cao**: Viết bằng Go, tối ưu RAM/CPU, hỗ trợ hàng vạn kết nối đồng thời.
*   **Dễ Dàng Triển Khai**: Hỗ trợ Docker, Systemd, Windows Service. Binary chạy ngay không cần cài đặt.

---

## 🚀 Cài Đặt & Chạy Nhanh

### 1. Tải về (Download)
Tải binary mới nhất từ [Releases](https://github.com/hoangtuvungcao/proxvn_tunnel_full/releases) hoặc build từ source:

```bash
# Build (yêu cầu Go 1.21+)
./build-all.sh
```

### 2. Chạy Client (Cơ bản)

**Public Web Server port 3000:**
```bash
./bin/client/proxvn-linux-amd64 --proto http 3000
# Output: https://random-id.vutrungocrong.fun
```

**Public SSH port 22:**
```bash
./bin/client/proxvn-linux-amd64 --proto tcp 22
# Output: 103.77.246.206:10001
```

---

## 📖 Hướng Dẫn Sử Dụng Chi Tiết (Client)

Binary client: `proxvn-linux-amd64` (Linux), `proxvn-windows-amd64.exe` (Windows), `proxvn-darwin-amd64` (macOS Intel), `proxvn-darwin-arm64` (macOS M1/M2).

### Cú pháp chung
```bash
proxvn [OPTIONS] [LOCAL_PORT]
```

### Các Tùy Chọn (Flags)

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `--proto` | `tcp` | Giao thức tunnel: `tcp`, `udp`, `http`. |
| `--server` | `103.77.246.206:8882` | Địa chỉ server tunnel (IP:Port). Mặc định trỏ về server cộng đồng. |
| `--host` | `localhost` | Địa chỉ local service (VD: 192.168.1.10). |
| `--port` | `80` | Port local service (có thể điền trực tiếp cuối lệnh). |
| `--id` | (random) | Custom Client ID (để nhận diện trong Dashboard). |
| `--ui` | `true` | Bật/tắt giao diện TUI đẹp mắt (`true`/`false`). |
| `--cert-pin` | (none) | SHA256 fingerprint của server certificate để verify (bảo mật cao). |
| `--insecure` | `false` | Bỏ qua xác thực SSL server (dùng cho dev/test). |

#### File Sharing Flags

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `--file` | - | Đường dẫn thư mục cần share (VD: `./share`, `C:\Docs`). |
| `--user` | `proxvn` | Username để xác thực WebDAV. |
| `--pass` | - | Mật khẩu bảo vệ truy cập (bắt buộc với `--file`). |
| `--permissions` | `rw` | Quyền hạn: `r` (chỉ đọc), `rw` (đọc-ghi), `rwx` (full quyền). |

### 🔐 Certificate Pinning (Bảo mật cao)

Để đảm bảo client chỉ kết nối đến đúng server của bạn (tránh MITM attack), sử dụng Certificate Pinning:

```bash
# Kết nối với cert-pin verification
proxvn --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 --proto http 3000
```

**Cert-pin cho server chính thức:**
```
5D21642F9C2AC2AEF414ECB27B54CDB5D53CB6D554BBF965DE19D2C8652F47C6
```

**Lưu ý:** Fingerprint này phải khớp với certificate của server. Nếu không khớp, client sẽ từ chối kết nối.

---

### Các Chế Độ Chạy (Modes)

#### 1. HTTP Tunneling (`--proto http`)
Dùng cho Web Application. Server sẽ cấp subdomain HTTPS.

```bash
# Public port 8080 local ra Internet
proxvn --proto http 8080

# Public Service ở máy khác trong mạng LAN (VD: Camera IP)
proxvn --proto http --host 192.168.1.50 80

# Với cert-pin security
proxvn --proto http --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 3000
```

**Kết quả:**
```
✅ HTTP Tunnel Active
🌐 Public URL: https://abc123.vutrungocrong.fun
📍 Forwarding to: localhost:3000
```

#### 2. TCP Tunneling (`--proto tcp`)
Dùng cho SSH, RDP, MySQL, PostgreSQL, v.v.

```bash
# Public SSH (mặc định port 22)
proxvn 22

# Public SSH với bảo mật cao
proxvn --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 22

# Public Remote Desktop (Windows)
proxvn 3389

# Public MySQL Database
proxvn 3306

# Kết nối tới server riêng của bạn
proxvn --server YOUR_VPS_IP:8882 22
```

**Kết quả:**
```
Public Address: 103.77.246.206:10001
```

#### 3. UDP Tunneling (`--proto udp`)
Dùng cho Game Server hoặc ứng dụng UDP.

```bash
# Minecraft Bedrock
proxvn --proto udp 19132

# Minecraft Java Edition
proxvn --proto udp 25565

# Palworld Server
proxvn --proto udp 8211

# CS:GO Server
proxvn --proto udp 27015

# Với cert-pin security
proxvn --proto udp --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 19132
```

#### 4. File Sharing Mode (`--file`)
Biến máy tính thành Cloud Storage mini. Hỗ trợ Web Interface và WebDAV.

```bash
# Share thư mục hiện tại, quyền full (username mặc định: proxvn)
proxvn --file . --pass 123456 --permissions rwx

# Share với custom username
proxvn --file /home/user/Movies --user media --pass secret --permissions r
# Khi mount WebDAV: username=media, password=secret

# Share folder Windows
proxvn --file "C:\Projects" --pass abc123 --permissions rw

# Share với bảo mật cao
proxvn --file ~/Documents --pass mypassword --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

**Tính năng File Share:**
*   **Web UI**: Truy cập qua trình duyệt, xem/tải/upload file, **Sửa code trực tiếp (Editor)**.
*   **WebDAV**: Mount thành ổ đĩa mạng trên Windows (Z:), macOS (Finder), Linux.

**Mount WebDAV trên các hệ điều hành:**

*Windows:*
```cmd
net use Z: https://abc123.vutrungocrong.fun /user:proxvn yourpassword
```

*macOS:*
```
Finder → Go → Connect to Server
Server: https://abc123.vutrungocrong.fun
Username: proxvn
Password: yourpassword
```

*Linux:*
```bash
sudo apt install davfs2
sudo mount -t davfs https://abc123.vutrungocrong.fun /mnt/proxvn
# Username: proxvn
# Password: yourpassword
```

---

## 🛠️ Hướng Dẫn Vận Hành Server

Binary server: `proxvn-server-linux-amd64`.

### Cú pháp
```bash
./bin/server/proxvn-server-linux-amd64 [OPTIONS]
```

### Các Tùy Chọn (Server Flags)

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `-port` | `8881` | Port cho Dashboard quản lý và API. |

*Lưu ý: Tunnel Port sẽ luôn là `Dashboard Port + 1` (VD: 8882).*

### Biến Môi Trường (Environment Variables)

Thay vì dùng flag, bạn nên dùng file `.env` hoặc set biến môi trường. Copy file `.env.server.example` thành `.env` và tùy chỉnh:

```bash
cp .env.server.example .env
nano .env
```

#### Các biến môi trường quan trọng:

**Server Settings:**
```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8882
PUBLIC_PORT_START=10000
PUBLIC_PORT_END=20000
```

**HTTP Domain (cho HTTP Tunneling):**
```bash
# Cấu hình Domain cho HTTP Tunneling (Bắt buộc nếu muốn dùng tính năng này)
HTTP_DOMAIN=yourdomain.com
HTTP_PORT=443
```

**Database:**
```bash
# SQLite3 Database
DB_PATH=./proxvn.db
```

**Bảo mật:**
```bash
JWT_SECRET=your-super-secret-jwt-key-change-this
TOKEN_EXPIRY=24h

# Admin Account mặc định
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

**TLS/SSL:**
```bash
TLS_MIN_VERSION=1.3
AUTO_TLS=true
TLS_CERT_FILE=./server.crt
TLS_KEY_FILE=./server.key
```

**Performance:**
```bash
MAX_CONNECTIONS=10000
BUFFER_SIZE=32768
ENABLE_COMPRESSION=true
COMPRESSION_LEVEL=6
ENABLE_HTTP2=true
```

**Rate Limiting:**
```bash
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
ENABLE_DDOS_PROTECTION=true
```

**Resource Management:**
```bash
MAX_UPLOAD_SIZE=1000
USER_STORAGE_QUOTA=10000
BANDWIDTH_LIMIT=0
```

**Monitoring:**
```bash
MONITORING_ENABLED=true
MONITORING_PORT=9090
DEBUG_MODE=false
LOG_LEVEL=info
```

**File Server & WebDAV:**
```bash
FILE_SERVER_ENABLED=true
FILE_SERVER_PORT=8080
WEBDAV_ENABLED=true
WEBDAV_PATH=/webdav
```

**Cache:**
```bash
ENABLE_CACHE=true
CACHE_SIZE_MB=256
CACHE_TTL=3600s
```

Xem file `.env.server.example` để có danh sách đầy đủ các biến môi trường.

### Triển Khai Server Riêng

Để chạy server riêng hỗ trợ HTTPS Subdomain, bạn cần:

1.  **Một tên miền** (VD: vutrungocrong.fun) trỏ về IP VPS.
2.  **Chứng chỉ SSL Wildcard** (`*.vutrungocrong.fun`).
3.  Đặt file `server.crt` và `server.key` (SSL của server tunnel) và wildcard cert (cho HTTP proxy) vào thư mục chạy.

#### Cách 1: Dùng Cloudflare Origin Certificate (Khuyến nghị)

```bash
# 1. Tạo Origin Certificate trên Cloudflare
#    Cloudflare Dashboard → SSL/TLS → Origin Server → Create Certificate
#    Lưu file: wildcard.crt và wildcard.key

# 2. Đặt file vào thư mục server
cp wildcard.crt /path/to/server/
cp wildcard.key /path/to/server/

# 3. Cấu hình DNS trên Cloudflare
#    A     @    YOUR_VPS_IP    (Proxied: ON)
#    CNAME *    yourdomain.com (Proxied: ON)

# 4. SSL Mode: Full (strict)

# 5. Chạy server
export HTTP_DOMAIN="yourdomain.com"
./bin/server/proxvn-server-linux-amd64
```

#### Cách 2: Dùng Let's Encrypt

```bash
sudo apt install python3-certbot-dns-cloudflare
sudo certbot certonly --dns-cloudflare \
  --dns-cloudflare-credentials /root/.secrets/cloudflare.ini \
  -d '*.yourdomain.com' -d 'yourdomain.com'

# Copy cert
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem wildcard.crt
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem wildcard.key

# Chạy server
export HTTP_DOMAIN="yourdomain.com"
./bin/server/proxvn-server-linux-amd64
```

#### Mở Firewall:

```bash
# Linux (ufw)
sudo ufw allow 8881/tcp  # Dashboard
sudo ufw allow 8882/tcp  # Tunnel
sudo ufw allow 443/tcp   # HTTPS (HTTP Tunneling)

# Windows: Mở Windows Firewall → Inbound Rules → New Rule
```

### Chạy Server

**Chạy trực tiếp:**
```bash
./bin/server/proxvn-server-linux-amd64
```

**Hoặc dùng script helper:**
```bash
./bin/run-server.sh
```

**Dashboard Access:**
```
http://localhost:8881/dashboard/
http://YOUR_VPS_IP:8881/dashboard/
```

**Default Admin Credentials:**
```
Username: admin
Password: admin123
```

⚠️ **Lưu ý:** Đổi mật khẩu ngay sau lần đăng nhập đầu tiên!

---

## 🔧 Build từ Source

### Yêu cầu
- Go 1.21 hoặc cao hơn
- Git

### Build All Platforms

```bash
# Clone repository
git clone https://github.com/hoangtuvungcao/proxvn_tunnel.git
cd proxvn_tunnel

# Build tất cả platforms (Linux, Windows, macOS, Android)
./build-all.sh
```

Script sẽ tạo ra các binary sau:

**Client binaries** (trong `bin/client/`):
- `proxvn-windows-amd64.exe` - Windows 64-bit
- `proxvn-linux-amd64` - Linux 64-bit
- `proxvn-linux-arm64` - Linux ARM64
- `proxvn-darwin-amd64` - macOS Intel
- `proxvn-darwin-arm64` - macOS M1/M2
- `proxvn-android-arm64` - Android ARM64

**Server binaries** (trong `bin/server/`):
- `proxvn-server-windows-amd64.exe` - Windows Server
- `proxvn-server-linux-amd64` - Linux Server
- `proxvn-server-linux-arm64` - Linux ARM64 Server
- `proxvn-server-darwin-amd64` - macOS Server Intel
- `proxvn-server-darwin-arm64` - macOS Server M1/M2

**Checksums:**
- `bin/SHA256SUMS-client.txt`
- `bin/SHA256SUMS-server.txt`

### Build Manual (cho một platform cụ thể)

```bash
cd src/backend

# Build client Linux
GOOS=linux GOARCH=amd64 go build -o ../../bin/client/proxvn-linux-amd64 ./cmd/client

# Build client Windows
GOOS=windows GOARCH=amd64 go build -o ../../bin/client/proxvn-windows-amd64.exe ./cmd/client

# Build server Linux
GOOS=linux GOARCH=amd64 go build -o ../../bin/server/proxvn-server-linux-amd64 ./cmd/server
```

---

## 📂 Cấu Trúc Dự Án

```
proxvn_tunnel/
├── bin/                        # Binary executables
│   ├── client/                 # Client binaries
│   │   ├── proxvn-linux-amd64
│   │   ├── proxvn-windows-amd64.exe
│   │   ├── proxvn-darwin-amd64
│   │   └── ...
│   ├── server/                 # Server binaries
│   │   ├── proxvn-server-linux-amd64
│   │   └── ...
│   ├── run-client.sh          # Client helper script (Linux/Mac)
│   ├── run-client.bat         # Client helper script (Windows)
│   ├── run-server.sh          # Server helper script (Linux/Mac)
│   └── run-server.bat         # Server helper script (Windows)
├── src/
│   ├── backend/               # Go source code
│   │   ├── cmd/
│   │   │   ├── client/        # Client main.go
│   │   │   ├── server/        # Server main.go
│   │   │   └── fileserver/    # File server module
│   │   └── internal/          # Internal packages
│   │       ├── api/           # REST API handlers
│   │       ├── auth/          # Authentication service
│   │       ├── config/        # Configuration management
│   │       ├── database/      # Database layer (SQLite3)
│   │       ├── http/          # HTTP proxy server
│   │       ├── middleware/    # HTTP middlewares
│   │       ├── models/        # Data models
│   │       └── tunnel/        # Tunnel protocol
│   └── frontend/              # Web Dashboard & Landing Page
│       ├── dashboard/         # Admin Dashboard
│       └── landing/           # Landing Page
├── docs/                      # Documentation
│   ├── 01-getting-started.md
│   ├── 02-configuration.md
│   ├── 03-client-guide.md
│   ├── 04-admin-guide.md
│   ├── 05-deployment.md
│   ├── 06-operations.md
│   ├── 07-troubleshooting.md
│   └── 08-security.md
├── scripts/                   # Build & deployment scripts
├── wiki/                      # Additional documentation
├── .env.server.example        # Server configuration template
├── cert-pin.txt              # Certificate pinning fingerprint
├── build-all.sh              # Build script
├── Dockerfile                # Docker configuration
├── docker-compose.yml        # Docker Compose
└── README.md                 # This file
```

---

## 📚 Tài Liệu Chi Tiết

Tài liệu đầy đủ có trong thư mục `docs/`:

- [01 - Getting Started](docs/01-getting-started.md) - Hướng dẫn bắt đầu
- [02 - Configuration](docs/02-configuration.md) - Cấu hình chi tiết
- [03 - Client Guide](docs/03-client-guide.md) - Hướng dẫn client
- [04 - Admin Guide](docs/04-admin-guide.md) - Hướng dẫn quản trị
- [05 - Deployment](docs/05-deployment.md) - Triển khai production
- [06 - Operations](docs/06-operations.md) - Vận hành hệ thống
- [07 - Troubleshooting](docs/07-troubleshooting.md) - Xử lý sự cố
- [08 - Security](docs/08-security.md) - Bảo mật

---

## 🐳 Docker Deployment

### Docker Compose (Recommended)

```bash
# Clone repository
git clone https://github.com/hoangtuvungcao/proxvn_tunnel.git
cd proxvn_tunnel

# Copy và chỉnh sửa .env
cp .env.server.example .env
nano .env

# Start server
docker-compose up -d

# View logs
docker-compose logs -f

# Stop server
docker-compose down
```

### Docker Manual

```bash
# Build image
docker build -t proxvn-server .

# Run server
docker run -d \
  -p 8881:8881 \
  -p 8882:8882 \
  -p 443:443 \
  -e HTTP_DOMAIN=yourdomain.com \
  --name proxvn-server \
  proxvn-server
```

---

## 🔧 Troubleshooting

### Client không kết nối được

```bash
# Kiểm tra kết nối tới server
telnet 103.77.246.206 8882

# Chạy với insecure mode để test
proxvn --insecure --proto http 3000

# Check logs
proxvn --proto http 3000 2>&1 | tee client.log
```

### Server không start

```bash
# Check port đã sử dụng chưa
sudo netstat -tlnp | grep 8881
sudo netstat -tlnp | grep 8882

# Kill process đang dùng port
sudo kill -9 PID

# Check logs
./bin/server/proxvn-server-linux-amd64 2>&1 | tee server.log
```

### Certificate Pinning Error

Nếu gặp lỗi cert-pin không khớp:

```bash
# Lấy cert fingerprint của server
openssl s_client -connect 103.77.246.206:8882 < /dev/null 2>/dev/null | \
  openssl x509 -fingerprint -sha256 -noout -in /dev/stdin

# Hoặc chạy client không có cert-pin để xem fingerprint
proxvn --proto http 3000
```

### File Sharing không mount được WebDAV

**Windows:**
```cmd
# Enable WebClient service
sc config WebClient start=auto
net start WebClient

# Mount với username/password
net use Z: https://subdomain.vutrungocrong.fun /user:proxvn yourpassword
```

**Linux:**
```bash
# Install davfs2
sudo apt install davfs2

# Mount
sudo mount -t davfs https://subdomain.vutrungocrong.fun /mnt/proxvn
```

---

## 🔐 Security Best Practices

1. **Sử dụng Certificate Pinning:**
   ```bash
   proxvn --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 --proto http 3000
   ```

2. **Đặt mật khẩu mạnh cho File Sharing:**
   ```bash
   proxvn --file ~/Documents --pass "MyStr0ng!P@ssw0rd#2024"
   ```

3. **Đổi mật khẩu admin mặc định ngay:**
   - Login vào Dashboard
   - Settings → Change Password

4. **Giới hạn quyền File Sharing:**
   ```bash
   # Chỉ đọc
   proxvn --file ~/Public --pass secret --permissions r
   
   # Đọc-ghi
   proxvn --file ~/Share --pass secret --permissions rw
   ```

5. **Enable Rate Limiting trên server:**
   ```bash
   # Trong .env
   RATE_LIMIT_RPS=10
   RATE_LIMIT_BURST=20
   ENABLE_DDOS_PROTECTION=true
   ```

6. **Sử dụng TLS 1.3:**
   ```bash
   # Trong .env
   TLS_MIN_VERSION=1.3
   ```

---

## 📊 Performance Tips

1. **Tăng buffer size cho throughput cao:**
   ```bash
   # Trong .env
   BUFFER_SIZE=65536  # 64KB
   ```

2. **Enable compression:**
   ```bash
   # Trong .env
   ENABLE_COMPRESSION=true
   COMPRESSION_LEVEL=6
   ```

3. **Tăng connection pool:**
   ```bash
   # Trong .env
   MAX_CONNECTIONS=20000
   ```

4. **Enable HTTP/2:**
   ```bash
   # Trong .env
   ENABLE_HTTP2=true
   ```

5. **Optimize timeout:**
   ```bash
   # Trong .env
   READ_TIMEOUT=30s
   WRITE_TIMEOUT=30s
   IDLE_TIMEOUT=60s
   ```

---

## 🤝 Support & Community

*   📧 **Email**: trong20843@gmail.com
*   💬 **Telegram**: [t.me/ZzTLINHzZ](https://t.me/ZzTLINHzZ)
*   🐛 **Báo lỗi**: [GitHub Issues](https://github.com/hoangtuvungcao/proxvn_tunnel_full/issues)
*   🌐 **Website**: [https://vutrungocrong.fun](https://vutrungocrong.fun)
*   📖 **Documentation**: [https://github.com/hoangtuvungcao/proxvn_tunnel_full/tree/main/docs](https://github.com/hoangtuvungcao/proxvn_tunnel_full/tree/main/docs)

---

## 📝 License

**FREE TO USE - NON-COMMERCIAL ONLY**

ProxVN Tunnel được cung cấp miễn phí cho mục đích phi thương mại. Nếu bạn muốn sử dụng cho mục đích thương mại, vui lòng liên hệ qua email.

---

## 🎯 Roadmap

- [x] HTTP/HTTPS Tunneling với auto SSL
- [x] TCP Tunneling
- [x] UDP Tunneling
- [x] File Sharing với WebDAV
- [x] Web Dashboard
- [x] In-Browser Code Editor
- [x] Certificate Pinning
- [x] Rate Limiting & DDoS Protection
- [x] Docker Support
- [ ] Mobile App (iOS/Android)
- [ ] Load Balancing
- [ ] Custom Domain Support
- [ ] Bandwidth Analytics
- [ ] API Webhooks
- [ ] Multi-User Management

---

## 🙏 Acknowledgments

Cảm ơn tất cả những người đã đóng góp và hỗ trợ dự án ProxVN!

**Made with ❤️ in Vietnam by TrongDev**

---

## 📌 Quick Reference Card

### Client Commands Cheatsheet

```bash
# HTTP Tunnel
proxvn --proto http 3000
proxvn --proto http --cert-pin 5d21...47c6 3000

# TCP Tunnel
proxvn 22
proxvn --cert-pin 5d21...47c6 3389

# UDP Tunnel  
proxvn --proto udp 19132
proxvn --proto udp --cert-pin 5d21...47c6 25565

# File Sharing
proxvn --file ~/Documents --pass secret
proxvn --file . --pass 123 --permissions rwx --cert-pin 5d21...47c6

# Custom Server
proxvn --server YOUR_IP:8882 --proto http 3000
```

### Server Commands Cheatsheet

```bash
# Start Server (default port 8881)
./bin/server/proxvn-server-linux-amd64

# Custom Port
./bin/server/proxvn-server-linux-amd64 -port 9000

# With Environment Variables
export HTTP_DOMAIN="yourdomain.com"
export JWT_SECRET="your-secret"
./bin/server/proxvn-server-linux-amd64

# Using .env file
cp .env.server.example .env
# Edit .env
./bin/server/proxvn-server-linux-amd64
```

### Certificate Pinning

**Official Server Cert-Pin:**
```
5D21642F9C2AC2AEF414ECB27B54CDB5D53CB6D554BBF965DE19D2C8652F47C6
```

**Usage:**
```bash
proxvn --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 [other-flags]
```

## Acknowledgments

Dự án này được phát triển dựa trên mã nguồn mở từ:
* [tunnel](https://github.com/kami2k1/tunnel) của tác giả [kami2k1](https://github.com/kami2k1).
* Copyright (c) 2026 kami2k1.
