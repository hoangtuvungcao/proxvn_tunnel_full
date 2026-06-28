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

Client chủ yếu cấu hình qua flags (xem [03-Client Guide](03-client-guide.md)). Tuy nhiên, có thể dùng biến môi trường trong một số trường hợp automation.

| Biến | Mô tả |
| :--- | :--- |
| `PROXVN_SERVER` | Địa chỉ server mặc định (thay thế `--server`). |
| `PROXVN_TOKEN` | Token xác thực (nếu server yêu cầu Auth). |

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
