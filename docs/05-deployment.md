# 05 - Deployment Guide

Hướng dẫn triển khai ProxVN Server lên môi trường Production (VPS, Cloud).

## 1. Chuẩn bị VPS
- Thuê VPS (DigitalOcean, Vultr, AWS...) với OS Linux (Ubuntu 20.04+ khuyến nghị).
- Cấu hình Firewall: Mở port `8881` (tcp), `8882` (tcp), `443` (tcp), `80` (tcp).

## 2. Triển khai với Docker Compose (Nhanh nhất)

Tạo file `docker-compose.yml`:

```yaml
version: '3'
services:
  proxvn-server:
    image: golang:1.21-alpine
    container_name: proxvn-server
    restart: always
    network_mode: host  # Tối ưu hiệu năng mạng
    volumes:
      - ./bin:/app/bin
      - ./data:/app/data
      - ./.env:/app/.env
    working_dir: /app
    command: ./bin/proxvn-linux-server
```

Chạy:
```bash
docker compose up -d
```

## 3. Triển khai thủ công (Systemd)

Tạo file service `/etc/systemd/system/proxvn.service`:

```ini
[Unit]
Description=ProxVN Tunnel Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/proxvn
ExecStart=/opt/proxvn/proxvn-linux-server
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Kích hoạt:
```bash
sudo systemctl daemon-reload
sudo systemctl enable proxvn
sudo systemctl start proxvn
```

## 4. Cấu hình Nginx Reverse Proxy (Tùy chọn)

Nếu bạn muốn Dashboard chạy sau Nginx (VD `https://panel.yoursite.com`):

```nginx
server {
    listen 80;
    server_name panel.yoursite.com;

    location / {
        proxy_pass http://localhost:8881;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## 5. Cấu hình HTTPS Tunneling (Quan trọng)
Để hỗ trợ `https://*.domain.com`, bạn cần:
1. Trỏ DNS Wildcard `*.domain.com` về IP VPS.
2. Có chứng chỉ SSL Wildcard (tạo bằng Let's Encrypt hoặc Cloudflare).
3. Đặt `HTTP_DOMAIN=domain.com` trong `.env`.
