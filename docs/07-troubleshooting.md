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
    - Ping thử subdomain: `ping abc.YOURDOMAIN`. Nó phải ra IP Server.
    - Kiểm tra biến môi trường `HTTP_DOMAIN` trên server.

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
