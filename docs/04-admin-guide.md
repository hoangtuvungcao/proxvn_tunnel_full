# 04 - Admin Guide

## Dashboard Quản Trị

ProxVN cung cấp giao diện Web Dashboard trực quan để giám sát toàn bộ hệ thống.

**Truy cập:** `http://localhost:8881/dashboard/` (hoặc IP Server)

### Các tính năng chính

1.  **Overview (Tổng quan)**
    - Xem số lượng Client đang kết nối.
    - Xem lưu lượng mạng (Bytes Up/Down) realtime.
    - Biểu đồ traffic hệ thống.

2.  **Tunnels (Danh sách Tunnel)**
    - Liệt kê tất cả tunnel đang hoạt động.
    - Thông tin chi tiết: Client ID, Subdomain/Port, Protocol, IP Client.
    - **Hành động**: Ngắt kết nối (Kick) client vi phạm.

3.  **Users (Quản lý User)**
    - *(Chỉ khả dụng khi bật chế độ Database Login)*
    - Tạo, sửa, xóa người dùng.
    - Phân quyền (Admin/User).

4.  **Settings (Cài đặt)**
    - Xem thông tin phiên bản Server.
    - Cấu hình hệ thống (nếu có quyền Admin).

## Quản trị dòng lệnh (CLI)

Hiện tại server chạy hoàn toàn tự động. Các tác vụ quản trị chủ yếu thực hiện qua Dashboard hoặc thao tác trực tiếp với Database (SQLite).

## Backup & Restore

**Database SQLite:**
Copy file `proxvn.db` để sao lưu toàn bộ dữ liệu user và lịch sử.

**Cấu hình:**
Sao lưu file `.env` và các chứng chỉ SSL (`*.crt`, `*.key`).
