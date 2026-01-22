# 02 - Configuration Guide

## Cấu hình Server (ProxVN Server)

Server có thể được cấu hình qua biến môi trường (Environment Variables) hoặc file `.env`.

### Biến môi trường quan trọng

| Biến | Mặc định | Mô tả |
| :--- | :--- | :--- |
| `SERVER_PORT` | `8881` | Port chính cho Dashboard và API. |
| `HTTP_DOMAIN` | `""` | Domain chính để cấp phát Subdomain (VD: `proxvn.com`). Bắt buộc nếu dùng HTTP Tunnel. |
| `DB_DSN` | `proxvn.db` | Đường dẫn file SQLite hoặc connection string PostgreSQL. |
| `JWT_SECRET` | (random) | Chuỗi bí mật để ký Token đăng nhập. Nên đặt cố định để không bị logout khi restart. |
| `ADMIN_PASSWORD`| `""` | (Tùy chọn) Mật khẩu Admin mặc định khởi tạo. |

### Cấu hình SSL (Cho HTTP Tunnel)
Đặt file `server.crt` và `server.key` (SSL của server tunnel) và `wildcard.crt`/`wildcard.key` (SSL cho subdomain) vào thư mục chạy server.

## Cấu hình Client (ProxVN Client)

Client chủ yếu cấu hình qua flags (xem [03-Client Guide](03-client-guide.md)). Tuy nhiên, có thể dùng biến môi trường trong một số trường hợp automation.

| Biến | Mô tả |
| :--- | :--- |
| `PROXVN_SERVER` | Địa chỉ server mặc định (thay thế `--server`). |
| `PROXVN_TOKEN` | Token xác thực (nếu server yêu cầu Auth). |

## File `.env` mẫu (Server)

```env
SERVER_PORT=8881
HTTP_DOMAIN=vutrungocrong.fun

# Database (SQLite)
DB_DSN=proxvn.db

# Security
JWT_SECRET=super-secret-key-change-me
TOKEN_EXPIRY=24h
```
