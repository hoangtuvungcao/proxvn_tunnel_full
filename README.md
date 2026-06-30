
# ProxVN Tunnel is a free ngrok alternative for exposing localhost to the Internet. It supports HTTP, HTTPS, TCP, UDP, WebDAV, includes a free public tunnel server, and can be self-hosted with no artificial session limits.




ProxVN Tunnel là giải pháp tunneling viết bằng Go, cho phép đưa dịch vụ chạy trên
máy local (localhost) ra Internet ngay lập tức — tương tự ngrok nhưng tự host được.
Hỗ trợ HTTP/HTTPS, TCP, UDP và chia sẻ file (WebDAV), chạy đa nền tảng, một binary
duy nhất không cần cài đặt.


- Tunnel server công cộng miễn phí: `103.77.246.196:8882`
- Domain: `bacsycay.click`
- Dashboard: https://bacsycay.click/dashboard/

---

## Mục lục

1. [Tính năng](#tính-năng)
2. [Server công cộng miễn phí](#server-công-cộng-miễn-phí)
3. [Cài đặt nhanh](#cài-đặt-nhanh)
4. [Hướng dẫn Client](#hướng-dẫn-client)
5. [Certificate Pinning](#certificate-pinning)
6. [Chia sẻ file (WebDAV)](#chia-sẻ-file-webdav)
7. [Vận hành Server](#vận-hành-server)
8. [Triển khai Server riêng](#triển-khai-server-riêng)
9. [Docker](#docker)
10. [Build từ source](#build-từ-source)
11. [Cấu trúc dự án](#cấu-trúc-dự-án)
12. [Tài liệu](#tài-liệu)
13. [Troubleshooting](#troubleshooting)
14. [Bản tham chiếu nhanh](#bản-tham-chiếu-nhanh)
15. [License](#license)

---

## Tính năng

- Đa giao thức:
  - HTTP/HTTPS: tự động cấp subdomain HTTPS `https://<id>.bacsycay.click` cho web app.
  - TCP: forward port cho SSH, RDP, MySQL, PostgreSQL...
  - UDP: cho game server (Minecraft, Palworld, CS...) và ứng dụng realtime.
  - File sharing: chia sẻ thư mục qua Web UI và WebDAV (mount thành ổ đĩa mạng).
- Bảo mật:
  - Mã hóa TLS 1.3 cho kết nối control và data.
  - Certificate pinning để chống tấn công MITM.
  - Xác thực JWT, rate limiting, chống DDoS ở phía server.
  - Chế độ private có mật khẩu cho file sharing.
- Quản lý:
  - Web Dashboard realtime: trạng thái tunnel, băng thông, số kết nối.
  - Trình soạn thảo file trên trình duyệt.
- Hiệu năng cao, tối ưu RAM/CPU, xử lý nhiều kết nối đồng thời.
- Triển khai dễ: Docker, systemd, Windows Service; binary chạy ngay.

---

## Server công cộng miễn phí

Không cần tự dựng server — đã có sẵn một server ProxVN miễn phí dùng chung:

| Thành phần | Giá trị |
| :--- | :--- |
| Tunnel server | `103.77.246.196:8882` (mặc định của client) |
| Domain | `bacsycay.click` |
| Dashboard | https://bacsycay.click/dashboard/ |
| Cert-pin | `29e1546abeb0e1d27adc57362422670b5347a0f19a847c5e9dda8fa7cd99c6d8` |

Client đã trỏ sẵn về server này, chỉ cần tải binary và chạy. Muốn dùng server riêng
thì sửa file `proxvn.json` hoặc dùng flag `--server` (xem [Cấu hình](docs/02-configuration.md)).

---

## Cài đặt nhanh

### 1. Tải binary

Tải trực tiếp từ trang chủ (qua server công cộng):

| Nền tảng | Client | Server |
| :--- | :--- | :--- |
| Linux x86-64 | https://bacsycay.click/bin/client/proxvn-linux-amd64 | .../bin/server/proxvn-server-linux-amd64 |
| Linux ARM64 | .../bin/client/proxvn-linux-arm64 | .../bin/server/proxvn-server-linux-arm64 |
| Windows x86-64 | .../bin/client/proxvn-windows-amd64.exe | .../bin/server/proxvn-server-windows-amd64.exe |
| macOS Intel | .../bin/client/proxvn-darwin-amd64 | .../bin/server/proxvn-server-darwin-amd64 |
| macOS Apple Silicon | .../bin/client/proxvn-darwin-arm64 | .../bin/server/proxvn-server-darwin-arm64 |
| Android ARM64 | .../bin/client/proxvn-android-arm64 | — |

Hoặc build từ source (cần Go 1.21+):

```bash
./build-all.sh
```

Trên Linux/macOS, nhớ cấp quyền chạy: `chmod +x proxvn-linux-amd64`.

### 2. Chạy Client

Public web server đang chạy ở port 3000:

```bash
./proxvn-linux-amd64 --proto http 3000
# => https://<id>.bacsycay.click
```

Public SSH (port 22):

```bash
./proxvn-linux-amd64 --proto tcp 22
# => 103.77.246.196:<port>
```

Public game server UDP (ví dụ Minecraft Bedrock 19132):

```bash
./proxvn-linux-amd64 --proto udp 19132
# => 103.77.246.196:<port> (UDP)
```

### 3. Cấu hình bằng file (không cần sửa code)

Sao chép [`bin/proxvn.json.example`](bin/proxvn.json.example) thành `proxvn.json` (đặt
cạnh binary), chỉnh `server`/`proto`/`port`/`cert_pin`… rồi chạy. Thứ tự ưu tiên:

```
flag dòng lệnh  >  proxvn.json  >  giá trị mặc định build-in
```

---

## Hướng dẫn Client

### Cú pháp

```
proxvn [OPTIONS] [LOCAL_PORT]
```

`LOCAL_PORT` có thể đặt ở cuối lệnh thay cho flag `--port`.

### Các flag

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `--proto` | `tcp` | Giao thức tunnel: `tcp`, `udp`, `http`. |
| `--server` | `103.77.246.196:8882` | Địa chỉ server tunnel (IP:Port). |
| `--host` | `localhost` | Địa chỉ service local (vd `192.168.1.10`). |
| `--port` | `80` | Port service local (điền trực tiếp cuối lệnh được). |
| `--id` | (ngẫu nhiên) | Client ID tùy chọn, dùng để nhận diện trong Dashboard. |
| `--ui` | `true` | Bật/tắt giao diện terminal (`true`/`false`). |
| `--cert-pin` | (none) | SHA256 fingerprint của cert server để xác thực. |
| `--insecure` | `false` | Bỏ qua xác thực TLS server (chỉ dùng cho dev/test). |
| `--config` | (auto) | Đường dẫn file cấu hình client. |

Flag cho chế độ chia sẻ file:

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `--file` | — | Thư mục cần chia sẻ (vd `./share`, `C:\Docs`). |
| `--user` | `proxvn` | Username để xác thực WebDAV. |
| `--pass` | — | Mật khẩu truy cập (bắt buộc khi dùng `--file`). |
| `--permissions` | `rw` | Quyền: `r` (chỉ đọc), `rw` (đọc-ghi), `rwx` (đầy đủ). |

### HTTP Tunneling (`--proto http`)

Dùng cho web app; server cấp subdomain HTTPS.

```bash
# Public port 8080
proxvn --proto http 8080

# Public service ở máy khác trong LAN (vd camera IP)
proxvn --proto http --host 192.168.1.50 80
```

Kết quả:

```
Public URL: https://abc123.bacsycay.click
Forwarding: localhost:3000
```

### TCP Tunneling (`--proto tcp`)

Dùng cho SSH, RDP, MySQL, PostgreSQL...

```bash
proxvn 22            # SSH (mặc định proto tcp)
proxvn 3389          # Remote Desktop (Windows)
proxvn 3306          # MySQL
proxvn --server YOUR_VPS_IP:8882 22   # dùng server riêng
```

Kết quả:

```
Public Address: 103.77.246.196:<port>
```

### UDP Tunneling (`--proto udp`)

Dùng cho game server và ứng dụng UDP.

```bash
proxvn --proto udp 19132   # Minecraft Bedrock
proxvn --proto udp 25565   # Minecraft Java
proxvn --proto udp 8211    # Palworld
proxvn --proto udp 27015   # CS / Source engine
```

---

## Certificate Pinning

Để client chỉ kết nối đến đúng server (chống MITM), dùng `--cert-pin` với fingerprint
SHA256 của certificate server:

```bash
proxvn --cert-pin 29e1546abeb0e1d27adc57362422670b5347a0f19a847c5e9dda8fa7cd99c6d8 --proto http 3000
```

Cert-pin của server công cộng (cũng lưu trong [`cert-pin.txt`](cert-pin.txt)):

```
29e1546abeb0e1d27adc57362422670b5347a0f19a847c5e9dda8fa7cd99c6d8
```

Lấy lại fingerprint từ một server bất kỳ:

```bash
echo | openssl s_client -connect <IP>:8882 2>/dev/null \
  | openssl x509 -outform DER | sha256sum
```

Nếu fingerprint không khớp, client sẽ từ chối kết nối.

---

## Chia sẻ file (WebDAV)

Biến máy tính thành kho lưu trữ mini với Web UI và WebDAV.

```bash
# Chia sẻ thư mục hiện tại, quyền đầy đủ (user mặc định: proxvn)
proxvn --file . --pass 123456 --permissions rwx

# Chia sẻ với username tùy chỉnh, chỉ đọc
proxvn --file /home/user/Movies --user media --pass secret --permissions r

# Chia sẻ thư mục Windows
proxvn --file "C:\Projects" --pass abc123 --permissions rw
```

Mở URL `https://<id>.bacsycay.click` để xem/tải/upload file và sửa file trực tiếp trên
trình duyệt, hoặc mount qua WebDAV:

Windows:

```cmd
net use Z: https://<id>.bacsycay.click /user:proxvn yourpassword
```

macOS: Finder → Go → Connect to Server → `https://<id>.bacsycay.click`

Linux:

```bash
sudo apt install davfs2
sudo mount -t davfs https://<id>.bacsycay.click /mnt/proxvn
```

---

## Vận hành Server

Binary server: `proxvn-server-linux-amd64` (và các nền tảng khác).

### Flag

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `-port` | `8881` | Port Dashboard và API. Tunnel port = Dashboard port + 1 (vd 8882). |

### Cấu hình bằng biến môi trường

Server cấu hình hoàn toàn qua biến môi trường hoặc file `.env`. Chỉ cần đổi vài biến,
phần còn lại đều có giá trị mặc định hợp lý.

```bash
cp .env.server.example .env
nano .env
```

Các biến quan trọng nhất:

| Biến | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `SERVER_PORT` | `8882` | Port tunnel control (TCP-TLS) + UDP. Dashboard/API ở `8881`. |
| `HTTP_PORT` | `443` | Port HTTPS phục vụ landing + dashboard + subdomain tunnel. |
| `HTTP_DOMAIN` | (trống) | Domain cấp subdomain (vd `bacsycay.click`). Bắt buộc cho HTTP tunnel. |
| `PUBLIC_HOST` | (auto) | IP công khai quảng bá cho client TCP/UDP. Trống = tự dò. |
| `PUBLIC_PORT_START` / `_END` | `10000` / `20000` | Dải port công khai cho tunnel TCP/UDP. |
| `DB_PATH` | `./proxvn.db` | File SQLite3 (Docker: `/data/proxvn.db`). |
| `JWT_SECRET` | (random) | Khóa ký token. Nên đặt cố định để không bị logout khi restart. |
| `ADMIN_USERNAME` | `admin` | Tài khoản admin khởi tạo (chỉ tạo khi DB rỗng). |
| `ADMIN_PASSWORD` | `admin123` | Mật khẩu admin khởi tạo. Bắt buộc đổi trên production. |

Danh sách đầy đủ (rate limit, backup, monitoring, file server, cache...) xem trong
[`.env.server.example`](.env.server.example).

### Chạy server

```bash
./bin/server/proxvn-server-linux-amd64
```

Dashboard: `http://YOUR_VPS_IP:8881/dashboard/` — đăng nhập bằng `admin` / `admin123`
(đổi mật khẩu ngay sau lần đăng nhập đầu tiên).

---

## Triển khai Server riêng

Để chạy server riêng có HTTPS subdomain, cần:

1. Một tên miền (vd `example.com`) trỏ về IP VPS.
2. Chứng chỉ wildcard cho `*.example.com` (HTTP proxy).
3. Cặp `server.crt` / `server.key` cho tunnel control (server tự sinh self-signed nếu
   bật `AUTO_TLS=true`; persist file này để cert-pin ổn định qua các lần redeploy).

### Cách 1: Cloudflare Origin Certificate (khuyến nghị)

```bash
# 1. Cloudflare Dashboard -> SSL/TLS -> Origin Server -> Create Certificate
#    Lưu thành wildcard.crt và wildcard.key, đặt cạnh binary server.
# 2. DNS trên Cloudflare:
#    A      @   YOUR_VPS_IP   (Proxied: ON)
#    CNAME  *   example.com   (Proxied: ON)
# 3. SSL/TLS mode: Full (strict)
# 4. Chạy server:
export HTTP_DOMAIN="example.com"
./bin/server/proxvn-server-linux-amd64
```

### Cách 2: Let's Encrypt (DNS-01)

```bash
sudo apt install python3-certbot-dns-cloudflare
sudo certbot certonly --dns-cloudflare \
  --dns-cloudflare-credentials /root/.secrets/cloudflare.ini \
  -d '*.example.com' -d 'example.com'

sudo cp /etc/letsencrypt/live/example.com/fullchain.pem wildcard.crt
sudo cp /etc/letsencrypt/live/example.com/privkey.pem wildcard.key

export HTTP_DOMAIN="example.com"
./bin/server/proxvn-server-linux-amd64
```

### Mở firewall

```bash
sudo ufw allow 22/tcp            # SSH
sudo ufw allow 8881/tcp          # Dashboard / API
sudo ufw allow 8882/tcp          # Tunnel control (TCP-TLS) + UDP
sudo ufw allow 443/tcp           # HTTPS landing + subdomain tunnel
sudo ufw allow 10000:20000/tcp   # Dải port công khai TCP
sudo ufw allow 10000:20000/udp   # Dải port công khai UDP
```

Chi tiết triển khai production xem [docs/05-deployment.md](docs/05-deployment.md).

---

## Docker

Cách triển khai khuyến nghị. `docker-compose.yml` chạy một container host-networked
xử lý tất cả: dashboard (8881), tunnel (8882), HTTPS (443), và dải port 10000-20000.

```bash
cp .env.server.example .env
nano .env                 # đặt HTTP_DOMAIN, PUBLIC_HOST, JWT_SECRET, ADMIN_PASSWORD

docker compose up -d      # khởi động
docker compose logs -f    # xem log
docker compose down       # dừng
```

Đặt chứng chỉ vào thư mục `ssl/` trước khi chạy (xem docker-compose.yml để biết các
file được mount: `ssl/wildcard.*`, `ssl/server.*`). Binary tải về được mount read-only
từ `./bin`, nên có thể cập nhật binary mà không cần build lại image; còn `src/frontend`
được COPY vào image nên khi đổi giao diện cần `docker compose build` lại.

---

## Build từ source

Yêu cầu: Go 1.21+, Git.

```bash
git clone https://github.com/hoangtuvungcao/proxvn_tunnel_full.git
cd proxvn_tunnel_full
./build-all.sh
```

`build-all.sh` cross-compile toàn bộ binary vào `bin/client/` và `bin/server/`:

Client (`bin/client/`):

- `proxvn-linux-amd64`, `proxvn-linux-arm64`
- `proxvn-windows-amd64.exe`
- `proxvn-darwin-amd64`, `proxvn-darwin-arm64`
- `proxvn-android-arm64`

Server (`bin/server/`):

- `proxvn-server-linux-amd64`, `proxvn-server-linux-arm64`
- `proxvn-server-windows-amd64.exe`
- `proxvn-server-darwin-amd64`, `proxvn-server-darwin-arm64`

Checksum: `bin/SHA256SUMS-client.txt`, `bin/SHA256SUMS-server.txt`.

Build một nền tảng cụ thể:

```bash
cd src/backend
GOOS=linux   GOARCH=amd64 go build -o ../../bin/client/proxvn-linux-amd64       ./cmd/client
GOOS=windows GOARCH=amd64 go build -o ../../bin/client/proxvn-windows-amd64.exe ./cmd/client
GOOS=linux   GOARCH=amd64 go build -o ../../bin/server/proxvn-server-linux-amd64 ./cmd/server
```

---

## Cấu trúc dự án

```
proxvn_tunnel_full/
├── bin/                         # Binary đã build + script helper
│   ├── client/                  # 6 binary client (linux/windows/darwin/android)
│   ├── server/                  # 5 binary server
│   ├── proxvn.json.example      # Mẫu cấu hình client
│   ├── run-client.sh / .bat     # Script chạy client
│   ├── run-server.sh / .bat     # Script chạy server
│   └── SHA256SUMS-*.txt         # Checksum
├── src/
│   ├── backend/                 # Mã nguồn Go
│   │   ├── cmd/
│   │   │   ├── client/          # main.go của client
│   │   │   ├── server/          # main.go của server
│   │   │   └── fileserver/      # Module file server / WebDAV
│   │   └── internal/
│   │       ├── api/             # REST API handlers
│   │       ├── auth/            # Xác thực JWT
│   │       ├── config/          # Quản lý cấu hình
│   │       ├── database/        # Tầng SQLite3
│   │       ├── http/            # HTTP proxy (landing + subdomain)
│   │       ├── middleware/      # Middleware HTTP
│   │       ├── models/          # Data models
│   │       └── tunnel/          # Giao thức tunnel
│   └── frontend/                # Web UI
│       ├── home.html            # Landing page (phục vụ tại /)
│       ├── dashboard.html       # Dashboard admin
│       ├── login.html           # Trang đăng nhập
│       ├── users.html           # Quản lý user
│       ├── css/ js/             # Tài nguyên dùng chung
│       ├── docs/                # Tài liệu HTML (phục vụ tại /docs/)
│       └── mobile/              # Giao diện mobile
├── docs/                        # Tài liệu Markdown (01-08 + PRODUCTION)
├── scripts/                     # Script build / SSL / tuning
├── .env.server.example          # Mẫu cấu hình server đầy đủ
├── cert-pin.txt                 # Fingerprint cert server công cộng
├── build-all.sh                 # Script build đa nền tảng
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## Tài liệu

Tài liệu chi tiết trong thư mục [`docs/`](docs/):

- [01 - Getting Started](docs/01-getting-started.md) — bắt đầu nhanh
- [02 - Configuration](docs/02-configuration.md) — cấu hình server & client
- [03 - Client Guide](docs/03-client-guide.md) — chi tiết client
- [04 - Admin Guide](docs/04-admin-guide.md) — quản trị Dashboard
- [05 - Deployment](docs/05-deployment.md) — triển khai production
- [06 - Operations](docs/06-operations.md) — vận hành, giám sát
- [07 - Troubleshooting](docs/07-troubleshooting.md) — xử lý sự cố
- [08 - Security](docs/08-security.md) — bảo mật

---

## Troubleshooting

Client không kết nối được:

```bash
# Kiểm tra kết nối tới server
nc -vz 103.77.246.196 8882

# Test bỏ qua xác thực TLS (chỉ để chẩn đoán)
proxvn --insecure --proto http 3000
```

Server không start (port đã bị chiếm):

```bash
sudo ss -tlnp | grep -E '8881|8882'
```

Lỗi cert-pin không khớp — lấy lại fingerprint đúng:

```bash
echo | openssl s_client -connect 103.77.246.196:8882 2>/dev/null \
  | openssl x509 -outform DER | sha256sum
```

WebDAV không mount được trên Windows:

```cmd
sc config WebClient start=auto
net start WebClient
net use Z: https://<id>.bacsycay.click /user:proxvn yourpassword
```

Thêm tình huống trong [docs/07-troubleshooting.md](docs/07-troubleshooting.md).

---

## Bản tham chiếu nhanh

Client:

```bash
proxvn --proto http 3000                         # HTTP tunnel
proxvn 22                                         # TCP tunnel (SSH)
proxvn --proto udp 19132                          # UDP tunnel
proxvn --file ~/Documents --pass secret           # Chia sẻ file
proxvn --server YOUR_IP:8882 --proto http 3000    # Dùng server riêng
proxvn --cert-pin 29e1546abeb0e1d27adc57362422670b5347a0f19a847c5e9dda8fa7cd99c6d8 --proto http 3000
```

Server:

```bash
./bin/server/proxvn-server-linux-amd64            # chạy (Dashboard port 8881)
./bin/server/proxvn-server-linux-amd64 -port 9000 # đổi port Dashboard
cp .env.server.example .env && nano .env          # cấu hình bằng .env
```

Cert-pin server công cộng:

```
29e1546abeb0e1d27adc57362422670b5347a0f19a847c5e9dda8fa7cd99c6d8
```

---

## License

FREE TO USE — NON-COMMERCIAL ONLY. ProxVN Tunnel miễn phí cho mục đích phi thương mại;
liên hệ qua email nếu dùng cho mục đích thương mại.

## Liên hệ

- Email: trong20843@gmail.com
- Telegram: https://t.me/ZzTLINHzZ
- Issues: https://github.com/hoangtuvungcao/proxvn_tunnel_full/issues
- Website: https://bacsycay.click

## Acknowledgments

Dự án phát triển dựa trên mã nguồn mở [tunnel](https://github.com/kami2k1/tunnel)
của [kami2k1](https://github.com/kami2k1). Copyright (c) 2026 kami2k1.
