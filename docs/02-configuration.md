# 02 - Configuration Guide

## Cấu hình Server (ProxVN Server)

Server có thể được cấu hình qua biến môi trường (Environment Variables) hoặc file `.env`.

### Biến môi trường quan trọng

| Biến | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `SERVER_PORT` | `8882` | Port điều khiển tunnel (TCP-TLS + UDP). Dashboard/API chạy ở port `8881`. |
| `HTTP_PORT` | `443` | Port HTTPS phục vụ landing + dashboard proxy + `*.HTTP_DOMAIN`. |
| `HTTP_DOMAIN` | `""` | Domain chính để cấp phát Subdomain (VD: `bacsycay.click`). Bắt buộc nếu dùng HTTP Tunnel. |
| `PUBLIC_HOST` | (auto) | IP công khai quảng bá cho client TCP/UDP. Để trống = tự phát hiện IP outbound. |
| `DB_PATH` | `./proxvn.db` | Đường dẫn file SQLite3. Trong Docker dùng `/data/proxvn.db`. |
| `JWT_SECRET` | (random) | Chuỗi bí mật để ký Token đăng nhập. Nên đặt cố định để không bị logout khi restart. |
| `ADMIN_USERNAME` | `admin` | Tên đăng nhập admin khởi tạo (chỉ tạo khi DB rỗng). |
| `ADMIN_PASSWORD`| `admin123` | Mật khẩu admin khởi tạo. **Bắt buộc đổi trên production.** |

> Danh sách đầy đủ các biến (rate limit, backup, monitoring, file server...) xem trong `.env.server.example`.

### Cấu hình SSL (Cho HTTP Tunnel)
Đặt cặp file `wildcard.crt` / `wildcard.key` (chứng chỉ wildcard cho `*.HTTP_DOMAIN`) vào thư mục chạy server (hoặc mount vào `/app/` trong Docker). Server cũng tự dò chứng chỉ Let's Encrypt tại `/etc/letsencrypt/live/<HTTP_DOMAIN>/`. Nếu không tìm thấy chứng chỉ, HTTP tunneling tự tắt và client fallback sang chế độ IP:Port.

## Cấu hình Client (ProxVN Client)

Client có thể cấu hình theo 3 cấp, ưu tiên từ cao xuống thấp:

**Flag dòng lệnh  >  File cấu hình `proxvn.json`  >  Giá trị mặc định build-in**

Nhờ vậy bạn đổi server/domain mà **không cần sửa code hay build lại** — chỉ sửa file `proxvn.json`.

### File cấu hình `proxvn.json`

Client tự tìm file theo thứ tự: `--config <path>` → biến môi trường `PROXVN_CONFIG`
→ `proxvn.json` cạnh binary → `proxvn.json` / `config.json` ở thư mục hiện tại
→ `~/.proxvn/config.json`. Mẫu: [`bin/proxvn.json.example`](../bin/proxvn.json.example).

```json
{
  "server": "103.77.246.196:8882",
  "host": "localhost",
  "port": 80,
  "proto": "tcp",
  "ui": true,
  "cert_pin": "",
  "insecure": false
}
```

```bash
# Dùng file cấu hình mặc định
./proxvn-linux-amd64                       # đọc proxvn.json nếu có

# Chỉ định file cấu hình
./proxvn-linux-amd64 --config /etc/proxvn/client.json

# Flag luôn ghi đè file cấu hình
./proxvn-linux-amd64 --server my-vps:8882 --proto http 3000
```

## File `.env` mẫu (Server)

```env
SERVER_PORT=8882
HTTP_DOMAIN=bacsycay.click
HTTP_PORT=443
PUBLIC_HOST=103.77.246.196

# Database (SQLite3)
DB_PATH=/data/proxvn.db

# Security
JWT_SECRET=super-secret-key-change-me
TOKEN_EXPIRY=24h
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-this-strong-password
```
