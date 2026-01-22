# 08 - Security Guide

Bảo mật là ưu tiên hàng đầu của ProxVN.

## Các lớp bảo mật

1.  **Transport Encryption (TLS 1.3)**
    - Mọi kết nối Control (Client <-> Server) đều được mã hóa TLS.
    - Tránh bị nghe lén hoặc Man-in-the-middle.

2.  **UDP Encryption (AES-GCM)**
    - Packet UDP tunnel được mã hóa bằng AES-GCM với key xoay vòng mỗi session.

3.  **Authentication (JWT)**
    - Server hỗ trợ xác thực Client bằng JWT Token.
    - Ngăn chặn client trái phép kết nối vào server.

4.  **Rate Limiting**
    - Server tích hợp sẵn Rate Limiter để chống Spam connection và DDoS.
    - Giới hạn số lượng tunnel mỗi IP.

## Khuyến nghị bảo mật cho Admin

1.  **Luôn bật SSL**: Không dùng chế độ `--insecure` trên production.
2.  **Bảo vệ Dashboard**: Không public port 8881 trực tiếp ra Internet nếu không cần thiết. Dùng VPN hoặc IP Whitelist.
3.  **File Sharing**:
    - LUÔN đặt mật khẩu (`--pass`) khi share file.
    - Chỉ dùng quyền `rw` (ghi) cho người tin cậy. Dùng `r` (chỉ đọc) khi public.
4.  **Cập nhật thường xuyên**: Theo dõi GitHub Repo để cập nhật các bản vá lỗi.

## Báo cáo lỗ hổng
Nếu phát hiện lỗ hổng bảo mật, vui lòng email trực tiếp: `trong20843@gmail.com` (Đừng tạo Issue public).
