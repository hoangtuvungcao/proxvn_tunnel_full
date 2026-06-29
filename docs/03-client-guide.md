# 03 - Client Guide

ProxVN Client là công cụ dòng lệnh (CLI) giúp bạn kết nối từ máy local tới ProxVN Server để public dịch vụ.

## 📥 Cài đặt

Tải binary phù hợp với hệ điều hành của bạn (thư mục `bin/client/`):
- **Windows**: `proxvn-windows-amd64.exe`
- **Linux**: `proxvn-linux-amd64` / `proxvn-linux-arm64`
- **macOS**: `proxvn-darwin-amd64` (Intel) / `proxvn-darwin-arm64` (Apple Silicon)
- **Android (Termux)**: `proxvn-android-arm64`

> Mặc định client trỏ tới server free `103.77.246.196:8882`. Đổi bằng file
> `proxvn.json` hoặc flag `--server` (xem [02 - Configuration](02-configuration.md)).

> **Flag đặt ở đâu cũng được**: `proxvn --proto http 3000 --ui=false`,
> `proxvn 3000 --proto http` đều hợp lệ (flag có thể đứng trước hoặc sau số port).

## 🕹️ Các Chế Độ Chạy

### 1. HTTP Tunneling (`--proto http`)
Dùng cho Web Server (NodeJS, Python, Apache, Nginx...).
Server sẽ cấp tự động một Subdomain HTTPS (SSL).

```bash
# Cú pháp: proxvn --proto http [LOCAL_PORT]
./proxvn --proto http 8080
```

### 2. TCP Tunneling (`--proto tcp`)
Chế độ mặc định. Dùng cho mọi giao thức TCP (SSH, RDP, MySQL, Redis...).

```bash
# Cú pháp: proxvn [LOCAL_PORT]
./proxvn 22     # SSH
./proxvn 3389   # Remote Desktop
./proxvn 5432   # PostgreSQL
```

### 3. UDP Tunneling (`--proto udp`)
Dùng cho các ứng dụng sử dụng UDP (Game Server, DNS, VoIP...).

```bash
# Cú pháp: proxvn --proto udp [LOCAL_PORT]
./proxvn --proto udp 19132  # Minecraft Bedrock
./proxvn --proto udp 1194   # OpenVPN
```

### 4. File Sharing Mode (`--file`)
Chế độ đặc biệt biến máy tính thành File Server.
Hỗ trợ Web Interface (xem, sửa code, upload) và WebDAV (mount drive).

```bash
# Cú pháp: proxvn --file [PATH] --pass [PASSWORD] [OPTS]
./proxvn --file ./share --pass 123456 --permissions rwx
```

**Tính năng:**
- **Web UI**: Giao diện đẹp, Dark Mode, kéo thả file.
- **Editor**: Sửa file code trực tiếp trên trình duyệt (nhấn icon ✏️).
- **WebDAV**: Tương thích Windows Explorer, Finder, Gnome Files.

## ⚙️ Danh sách Tham số (Flags)

| Flag | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `--server` | (default) | Địa chỉ Server Tunnel (IP:Port). Mặc định server cộng đồng. |
| `--proto` | `tcp` | Giao thức: `tcp`, `udp`, `http`. |
| `--host` | `localhost` | Host local cần forward (VD: 192.168.1.10). |
| `--port` | `80` | Port local (có thể điền trực tiếp không cần flag này). |
| `--id` | (random) | ID định danh client (tùy chọn). |
| `--ui` | `true` | Bật giao diện TUI (`false` để chạy background/service). |
| `--insecure` | `false` | Bỏ qua xác thực SSL (chỉ dùng test). |
| `--file` | `""` | Đường dẫn thư mục để share. |
| `--pass` | `""` | Mật khẩu truy cập file share. |
| `--permissions` | `rw` | Quyền file: `r` (đọc), `rw` (đọc ghi), `rwx` (full). |

## 💡 Mẹo & Thủ Thuật

### Chạy ngầm (Background)
Trên Linux, dùng `nohup` hoặc `systemd`. Tắt UI để log ra file dễ hơn.

```bash
nohup ./proxvn --proto http 3000 --ui=false > client.log 2>&1 &
```

### Kết nối Server Riêng (Self-hosted)
Nếu bạn tự host server ProxVN:

```bash
./proxvn --server YOUR_VPS_IP:8882 --proto http 80
```
*(Nếu server có SSL tự ký, thêm `--insecure` nếu cần)*
