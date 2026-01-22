# ProxVN Client - Hướng Dẫn Sử Dụng Chi Tiết

## 📖 Mục Lục
- [Cài Đặt](#cài-đặt)
- [Cú Pháp Cơ Bản](#cú-pháp-cơ-bản)
- [Tất Cả Options](#tất-cả-options)
- [Ví Dụ Sử Dụng](#ví-dụ-sử-dụng)
- [Troubleshooting](#troubleshooting)

---

## 🚀 Cài Đặt

### Windows
```powershell
# Download từ GitHub Releases
curl -LO https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-windows-amd64.exe
```

### Linux
```bash
curl -LO https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-linux-amd64
chmod +x proxvn-linux-amd64
sudo mv proxvn-linux-amd64 /usr/local/bin/proxvn
```

### macOS
```bash
# Intel
curl -LO https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-darwin-amd64

# Apple Silicon (M1/M2/M3)
curl -LO https://github.com/hoangtuvungcao/proxvn_tunnel/releases/latest/download/proxvn-darwin-arm64

chmod +x proxvn-darwin-*
sudo mv proxvn-darwin-* /usr/local/bin/proxvn
```

---

## 📝 Cú Pháp Cơ Bản

```bash
proxvn [OPTIONS] [LOCAL_PORT]
```

**Port mặc định:** Nếu không chỉ định protocol, mặc định là TCP tunneling

---

## ⚙️ Tất Cả Options

### 🌐 Server Connection

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|-------|
| `--server <địa_chỉ>` | Địa chỉ server (IP:Port hoặc domain:Port) | `vutrungocrong.fun:8882` | `--server localhost:8882` |
| `--insecure` | Bỏ qua xác thực TLS certificate (CHỈ cho testing) | `false` | `--insecure` |
| `--cert-pin <fingerprint>` | Certificate pinning bằng SHA256 fingerprint | - | `--cert-pin ABC123...` |

### 🔌 Tunnel Configuration

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|-------|
| `--proto <protocol>` | Protocol: `http`, `tcp`, `udp` | `tcp` | `--proto http` |
| `--local <địa_chỉ>` | Địa chỉ local để tunnel | `localhost:PORT` | `--local 127.0.0.1:3000` |
| `--subdomain <tên>` | Subdomain cho HTTP tunnel | Auto-generated | `--subdomain myapp` |
| `--remote-port <port>` | Port cụ thể trên server (TCP/UDP) | Auto-assigned | `--remote-port 10500` |

### 📁 File Sharing Mode

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|---------|
| `--file <đường_dẫn>` | Bật file sharing mode với thư mục | - | `--file ~/Documents` |
| `--user <username>` | Username cho WebDAV authentication | `proxvn` | `--user myusername` |
| `--pass <mật_khẩu>` | Mật khẩu bảo vệ file share | - | `--pass mypassword` |
| `--permissions <rwx>` | Quyền: r(read), w(write), x(execute) | `r` | `--permissions rw` |

### 🔐 Authentication

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|-------|
| `--key <client_key>` | Client key để xác thực | Auto-generated | `--key my-unique-key` |
| `--api-key <key>` | API key (nếu server yêu cầu) | - | `--api-key sk_abc123` |

### ⚡ Performance Settings

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|-------|
| `--buffer-size <bytes>` | Kích thước buffer (bytes) | `32768` | `--buffer-size 65536` |
| `--compression` | Bật compression | `true` | `--compression=false` |
| `--timeout <duration>` | Connection timeout | `30s` | `--timeout 60s` |
| `--max-reconnect <số>` | Số lần retry khi mất kết nối | `10` | `--max-reconnect 5` |
| `--reconnect-delay <duration>` | Delay giữa các lần retry | `5s` | `--reconnect-delay 10s` |

### 📊 Display & Logging

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|-------|
| `--no-ui` | Tắt UI, chỉ log text | `false` | `--no-ui` |
| `--log-level <level>` | Log level: debug, info, warn, error | `info` | `--log-level debug` |
| `--color` | Bật màu cho output | `true` | `--color=false` |
| `--quiet` | Chế độ im lặng, chỉ hiện lỗi | `false` | `--quiet` |

### 🔧 Advanced Settings

| Option | Mô Tả | Giá Trị Mặc Định | Ví Dụ |
|--------|-------|------------------|-------|
| `--heartbeat <duration>` | Interval gửi heartbeat | `30s` | `--heartbeat 15s` |
| `--udp-control-interval <duration>` | UDP control packet interval | `3s` | `--udp-control-interval 5s` |
| `--udp-control-timeout <duration>` | UDP control timeout | `6s` | `--udp-control-timeout 10s` |

### ℹ️ Help & Version

| Option | Mô Tả |
|--------|-------|
| `--help`, `-h` | Hiển thị help message |
| `--version`, `-v` | Hiển thị version |

---

## 💡 Ví Dụ Sử Dụng

### 1. HTTP Tunneling - Web Development

#### React/Vue/Angular Development
```bash
# Tunnel local dev server (port 3000)
proxvn --proto http 3000

# Với subdomain tùy chỉnh
proxvn --proto http --subdomain myapp 3000
# → https://myapp.vutrungocrong.fun
```

#### Django/Flask/Laravel
```bash
# Django dev server
proxvn --proto http 8000

# Flask
proxvn --proto http 5000

# Laravel
proxvn --proto http --subdomain laravel-app 8000
```

### 2. TCP Tunneling - Remote Access

#### SSH Server
```bash
# Expose SSH server
proxvn 22
# Server sẽ báo: "Tunnel active: 103.77.246.206:10500"

# Kết nối từ máy khác:
ssh -p 10500 user@103.77.246.206
```

#### MySQL/PostgreSQL Database
```bash
# MySQL
proxvn 3306

# PostgreSQL
proxvn 5432

# MongoDB
proxvn 27017
```

#### Remote Desktop (RDP)
```bash
# Windows RDP
proxvn 3389

# VNC
proxvn 5900
```

### 3. UDP Tunneling - Game Servers

#### Minecraft Bedrock
```bash
proxvn --proto udp 19132
```

#### Palworld Dedicated Server
```bash
proxvn --proto udp 8211
```

#### Counter-Strike Server
```bash
proxvn --proto udp 27015
```

#### Voice Chat (Discord Bot)
```bash
proxvn --proto udp 50000
```

### 4. File Sharing

#### Chia Sẻ Thư Mục Read-Only
```bash
proxvn --file ~/Documents --pass doc2024 --permissions r
# → https://xyz789.vutrungocrong.fun/browse
```

#### Chia Sẻ với Quyền Upload
```bash
proxvn --file ~/Shared --pass upload123 --permissions rw
```

#### Chia Sẻ Project Folder
```bash
proxvn --file ~/projects/myapp --user devteam --pass team2024 --permissions rwx
# Khi mount WebDAV, dùng username: devteam, password: team2024
```

#### Sử Dụng Username Mặc Định
```bash
# Không cần --user, mặc định là "proxvn"
proxvn --file ~/Shared --pass upload123 --permissions rw
# Mount WebDAV: username=proxvn, password=upload123
```

### 5. Webhooks Testing

#### Stripe Webhook
```bash
# Tunnel webhook endpoint
proxvn --proto http 4242
# Update Stripe webhook URL: https://abc123.vutrungocrong.fun/webhook
```

#### GitHub Webhook
```bash
proxvn --proto http --subdomain github-bot 8080
```

### 6. Custom Server

#### Sử Dụng Server Riêng
```bash
proxvn --server myserver.com:8882 --proto http 3000
```

#### Testing với Server Local
```bash
proxvn --server localhost:8882 --insecure --proto http 3000
```

### 7. Performance Tuning

#### Băng Thông Cao - Buffer Lớn
```bash
proxvn --buffer-size 131072 --proto tcp 8080
# 128KB buffer cho throughput cao
```

#### Kết Nối Không Ổn Định
```bash
proxvn --max-reconnect 20 --reconnect-delay 3s --proto http 3000
# Retry nhiều hơn, delay ngắn hơn
```

#### Tắt Compression cho File Lớn
```bash
proxvn --compression=false --buffer-size 65536 --proto tcp 9000
```

### 8. Production Deployment

#### Background Service (Linux)
```bash
# Chạy trong background với nohup
nohup proxvn --proto http --subdomain prod-api 8000 \
  --log-level info \
  --no-ui \
  > proxvn.log 2>&1 &

# Hoặc dùng systemd service
```

#### Docker Container
```bash
# Tunnel container port
proxvn --local 172.17.0.2:80 --proto http
```

### 9. Multiple Tunnels

#### Chạy Nhiều Tunnel Song Song
```bash
# Terminal 1: HTTP tunnel
proxvn --proto http --subdomain web 3000

# Terminal 2: TCP tunnel
proxvn --remote-port 10500 22

# Terminal 3: File sharing
proxvn --file ~/Shared --pass files123
```

### 10. Debug & Troubleshooting

#### Debug Mode
```bash
proxvn --log-level debug --proto http 3000
```

#### No UI Mode (cho script/automation)
```bash
proxvn --no-ui --quiet --proto http 8000 > tunnel.log
```

---

## 🔍 Troubleshooting

### ❌ Connection Failed

```bash
# Kiểm tra server có hoạt động không
proxvn --server vutrungocrong.fun:8882 --log-level debug 3000

# Thử với insecure mode (chỉ testing)
proxvn --insecure --proto http 3000
```

### ❌ Port Already in Use

```bash
# Kiểm tra port đang dùng
# Linux/macOS
lsof -i :3000

# Windows
netstat -ano | findstr :3000

# Dùng port khác
proxvn --proto http 3001
```

### ❌ Certificate Error

```bash
# Skip certificate verification (không nên dùng production)
proxvn --insecure --proto http 3000

# Hoặc dùng certificate pinning
proxvn --cert-pin SHA256_FINGERPRINT --proto http 3000
```

### ❌ Slow Performance

```bash
# Tăng buffer size
proxvn --buffer-size 131072 --proto tcp 8080

# Tắt compression nếu file đã nén sẵn
proxvn --compression=false --proto http 9000
```

### ❌ Frequent Disconnects

```bash
# Tăng retry và giảm delay
proxvn --max-reconnect 30 --reconnect-delay 2s --proto http 3000

# Tăng heartbeat frequency
proxvn --heartbeat 15s --proto http 3000
```

---

## 📚 Tips & Best Practices

### ✅ Security
- **KHÔNG dùng** `--insecure` trong production
- Luôn dùng mật khẩu mạnh cho `--pass`
- Dùng `--cert-pin` cho bảo mật cao
- Giới hạn `--permissions` phù hợp khi share file

### ✅ Performance
- Dùng buffer lớn hơn cho file/streaming: `--buffer-size 131072`
- Tắt compression cho file đã nén: `--compression=false`
- Dùng TCP cho độ tin cậy cao, UDP cho realtime/gaming

### ✅ Debugging
- Luôn dùng `--log-level debug` khi gặp vấn đề
- Dùng `--no-ui` khi chạy trong script/cronjob
- Check log file khi chạy background

### ✅ Production
- Dùng systemd/supervisor để auto-restart
- Monitor với `--log-level info`
- Setup multiple tunnels cho high availability

---

## 🆘 Cần Hỗ Trợ?

- 📧 Email: trong20843@gmail.com
- 💬 Telegram: [t.me/proxvn](https://t.me/proxvn)
- 🐛 Issues: [GitHub Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)
- 📖 Docs: [Full Documentation](https://github.com/hoangtuvungcao/proxvn_tunnel/tree/main/docs)

---

**Made with ❤️ by TrongDev**
