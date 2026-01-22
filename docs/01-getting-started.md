# 01 - Getting Started

## 🎯 Mục tiêu

Tài liệu này giúp bạn:
- Cài đặt ProxVN Server và Client.
- Public dịch vụ local (Web, SSH, Game, File) ra Internet.
- Truy cập Dashboard quản lý và sử dụng các tính năng nâng cao.

## 💻 Yêu cầu hệ thống

**Server (nếu tự host):**
- Linux x86_64/arm64.
- RAM: 512MB+.
- Docker (khuyến nghị) hoặc Go 1.21+ (nếu build từ source).
- Domain + SSL (nếu muốn dùng HTTP Subdomain).

**Client:**
- Windows 10/11, Linux, macOS, hoặc Android (Termux).

## 🚀 Cài đặt nhanh

### Cách 1: Sử dụng Docker (Khuyến nghị cho Server)

```bash
git clone https://github.com/hoangtuvungcao/proxvn_tunnel.git
cd proxvn_tunnel
cp .env.server.example .env
# Chỉnh sửa .env nếu cần (HTTP_DOMAIN, DB_DSN...)
docker compose up -d
```

### Cách 2: Chạy Binary (Khuyến nghị cho Client)

Tải binary từ [Releases](https://github.com/hoangtuvungcao/proxvn_tunnel/releases) hoặc build:

```bash
# Build Server & Client
./scripts/build.sh
```

Binary sẽ nằm trong thư mục `bin/`:
- `proxvn-linux-client`, `proxvn.exe` (Client)
- `proxvn-linux-server`, `svproxvn.exe` (Server)

## 🔥 Chạy thử Tunnel

### 1. HTTP Tunnel (Web App)
Public website đang chạy localhost:3000 ra Internet với HTTPS.

```bash
# Client
./bin/proxvn-linux-client --proto http 3000
# Output: https://random-id.vutrungocrong.fun
```

### 2. TCP Tunnel (SSH, RDP, Database)
Public SSH port 22 hoặc Remote Desktop 3389.

```bash
# Client
./bin/proxvn-linux-client 22
# Output: server-ip:10001
```

### 3. UDP Tunnel (Game Server)
Public Minecraft server port 19132.

```bash
# Client
./bin/proxvn-linux-client --proto udp 19132
```

### 4. File Sharing (Mới 🌟)
Chia sẻ thư mục hiện tại thành ổ đĩa mạng (WebDAV) và quản lý qua Web.

```bash
# Client
./bin/proxvn-linux-client --file . --pass 123456 --permissions rwx
```
- **Web UI**: Truy cập URL được cấp, đăng nhập để upload/download/sửa file.
- **WebDAV**: Mount như ổ đĩa mạng trên Windows/macOS.

## 📊 Dashboard & Monitoring

Truy cập Dashboard để xem trạng thái kết nối:
- URL: `http://localhost:8881/dashboard/`
- API: `http://localhost:8881/api`

## ✅ Kiểm tra trạng thái
- Metric trên Dashboard: `Connections`, `Bytes Up/Down`.
- Log terminal của Client/Server.

## ❓ Lỗi thường gặp
- **Lỗi permission denied**: Chạy với `sudo` (Linux) hoặc Administrator (Windows) nếu cần bind port thấp.
- **Lỗi kết nối Server**: Kiểm tra firewall server (8882/tcp, 443/tcp).
- **HTTP 404**: Kiểm tra cấu hình DNS wildcard và biến môi trường `HTTP_DOMAIN`.
