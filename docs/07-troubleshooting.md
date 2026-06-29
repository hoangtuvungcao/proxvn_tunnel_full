# 07 - Troubleshooting Guide

Tổng hợp các lỗi thường gặp và cách khắc phục.

## Lỗi Client

### 1. "connection refused" / "reconnect..."
- **Nguyên nhân**: Không kết nối được đến Server (IP sai, Port sai, Server chết, Firewall chặn).
- **Khắc phục**:
    - Kiểm tra IP/Port server (`--server`).
    - Telnet thử: `telnet SERVER_IP 8882`.
    - Kiểm tra firewall server.

### 2. "subdomain not found" / "404 Not Found"
- **Nguyên nhân**: DNS chưa trỏ đúng hoặc sai cấu hình `HTTP_DOMAIN`.
- **Khắc phục**:
    - Ping thử subdomain: `ping abc.YOURDOMAIN`. Nó phải ra IP Server (hoặc IP Cloudflare nếu proxied).
    - Kiểm tra biến môi trường `HTTP_DOMAIN` trên server và bản ghi DNS wildcard `*`.

### 2b. Cloudflare "Error 521" khi mở domain
- **Nguyên nhân**: Cloudflare không kết nối được tới origin ở cổng nó gọi. Hay gặp khi
  SSL mode = **Flexible** (CF gọi origin cổng **80**) nhưng origin chỉ mở 443.
- **Khắc phục**:
    - ProxVN server đã tự lắng nghe cả `:80` và `:443` → mở firewall cho `80/tcp`, `443/tcp`.
    - Khuyến nghị: SSL/TLS mode = **Full** và đảm bảo origin mở `:443` với cert wildcard.
    - Kiểm tra bản ghi A của apex/wildcard trỏ đúng IP VPS.

### 2c. "subdomain" trả về 502
- **Nguyên nhân**: Routing tới server OK nhưng **không có tunnel** nào đang giữ subdomain đó
  (client đã ngắt). Đây là hành vi đúng.
- **Khắc phục**: Chạy lại client; kiểm tra client còn `HTTP Tunnel Active`.

### 3. "permission denied" (Client)
- **Nguyên nhân**: Client cố bind vào port hệ thống (< 1024) mà không có quyền root/admin.
- **Khắc phục**: Chạy với `sudo` hoặc `Run as Administrator`.

## Lỗi Server

### 1. "bind: address already in use"
- **Nguyên nhân**: Port 8881 hoặc 8882 đã bị chiếm dụng.
- **Khắc phục**: Kill process cũ (`lsof -i :8881`) hoặc đổi port trong `.env`.

### 2. "too many open files"
- **Nguyên nhân**: Server quá tải connection, vượt giới hạn OS.
- **Khắc phục**: Tăng `ulimit -n 65535` trên Linux.

## Lỗi File Sharing

### 1. Không lưu được file / Access Denied
- **Nguyên nhân**: User chạy client không có quyền ghi vào thư mục share, hoặc chạy client với quyền thấp.
- **Khắc phục**: `chmod 777` thư mục share hoặc chạy client với quyền cao hơn.

### 2. WebDAV trên Windows chậm
- **Khắc phục**: Vào Internet Options -> Connections -> LAN Settings -> Bỏ chọn "Automatically detect settings".
