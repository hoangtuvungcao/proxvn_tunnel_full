# 05 - Deployment Guide (Docker + Cloudflare)

Hướng dẫn triển khai ProxVN Server lên Production bằng **Docker Compose**.
Kiến trúc khuyến nghị: **1 container duy nhất** chạy `network_mode: host`, không cần nginx
(server Go đã tự phục vụ landing + dashboard proxy + subdomain tunneling trên cổng 443).

> Ví dụ thực tế dưới đây dùng VPS Ubuntu 22.04 (`103.77.246.196`) và domain `bacsycay.click`.

## Kiến trúc cổng

| Cổng | Vai trò |
| :--- | :--- |
| `8881` | Dashboard + API (Gin, nội bộ) |
| `8882` | Tunnel control (TCP-TLS) + UDP control |
| `443` | HTTPS: landing page + proxy `/dashboard` + `*.<domain>` tunnels |
| `10000-20000` | Cổng public cho TCP/UDP tunnels |

## 1. Chuẩn bị VPS

```bash
ssh root@<VPS_IP>

# Cài Docker
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# Firewall (UFW) — giữ SSH, mở các cổng tunnel
ufw allow 22/tcp
ufw allow 80/tcp 443/tcp
ufw allow 8881/tcp 8882/tcp 8882/udp
ufw allow 10000:20000/tcp
ufw allow 10000:20000/udp
ufw --force enable
```

## 2. Lấy mã nguồn

```bash
mkdir -p /opt/proxvn && cd /opt/proxvn
# git clone <repo> .   (hoặc rsync mã nguồn lên /opt/proxvn)
mkdir -p data backups logs ssl
```

## 3. Cấu hình `.env`

```bash
cp .env.server.example .env
nano .env
```

Bắt buộc đặt cho production:

```env
PUBLIC_HOST=<VPS_IP>          # IP công khai cho TCP/UDP (client kết nối trực tiếp, không qua Cloudflare)
HTTP_DOMAIN=bacsycay.click    # bật HTTP subdomain tunneling
HTTP_PORT=443
DB_PATH=/data/proxvn.db
BACKUP_DIR=/backups
JWT_SECRET=<chuỗi ngẫu nhiên 32+ ký tự>     # openssl rand -hex 32
ADMIN_USERNAME=admin
ADMIN_PASSWORD=<mật khẩu mạnh>              # KHÔNG để admin123
```

## 4. DNS Cloudflare

Trỏ domain về IP VPS (API hoặc Dashboard Cloudflare):

| Type | Name | Content | Proxy |
| :--- | :--- | :--- | :--- |
| A | `@` (apex) | `<VPS_IP>` | DNS only (hoặc Proxied) |
| A | `*` (wildcard) | `<VPS_IP>` | **DNS only** |

> Wildcard `*` nên để **DNS only** (mây xám). Tunnel control (8882) và cổng public
> (10000-20000) là non-HTTP nên không đi qua Cloudflare proxy được.

## 5. Chứng chỉ Wildcard TLS (Let's Encrypt DNS-01)

HTTP subdomain (`https://<sub>.bacsycay.click`) cần chứng chỉ wildcard. Dùng certbot
với plugin Cloudflare (DNS-01 hỗ trợ wildcard):

```bash
# Tạo API token Cloudflare (quyền Zone:DNS:Edit cho domain) rồi:
mkdir -p ~/.secrets
cat > ~/.secrets/cloudflare.ini <<EOF
dns_cloudflare_api_token = <CLOUDFLARE_API_TOKEN>
EOF
chmod 600 ~/.secrets/cloudflare.ini

docker run --rm \
  -v /etc/letsencrypt:/etc/letsencrypt \
  -v ~/.secrets:/secrets \
  certbot/dns-cloudflare certonly \
  --dns-cloudflare --dns-cloudflare-credentials /secrets/cloudflare.ini \
  -d 'bacsycay.click' -d '*.bacsycay.click' \
  --agree-tos -m admin@bacsycay.click --no-eff-email

# Đưa cert vào ssl/ (container đọc qua bind mount, chạy uid 1000)
cp /etc/letsencrypt/live/bacsycay.click/fullchain.pem ssl/wildcard.crt
cp /etc/letsencrypt/live/bacsycay.click/privkey.pem   ssl/wildcard.key
chown 1000:1000 ssl/wildcard.crt ssl/wildcard.key
chmod 644 ssl/wildcard.crt && chmod 640 ssl/wildcard.key
```

> Script tiện ích: `scripts/get-cert.sh` (Linux) / `scripts/get-cert.ps1` (Windows).

## 6. Khởi chạy

```bash
cd /opt/proxvn
docker compose build      # build Go + CGO (sqlite3) trong container
docker compose up -d
docker compose logs -f proxvn-server
```

Kiểm tra nhanh:

```bash
curl -s http://127.0.0.1:8881/health                            # {"status":"ok",...}
curl -sk -H 'Host: bacsycay.click' https://127.0.0.1/ | head    # landing page
```

> Container chạy non-root (uid 1000); binary được cấp `cap_net_bind_service` để bind 443.
> Thư mục bind-mount (`data/ backups/ logs/ ssl/`) phải thuộc uid 1000:
> `chown -R 1000:1000 data backups logs ssl`.

## 7. Certificate Pinning (tùy chọn)

Cổng tunnel 8882 dùng self-signed cert (`ssl/server.crt`, được persist nên pin ổn định
qua các lần redeploy). Lấy pin để client xác thực server:

```bash
echo | openssl s_client -connect <VPS_IP>:8882 2>/dev/null \
  | openssl x509 -outform DER | sha256sum
# Client dùng:  proxvn --cert-pin <hash> ...
```

Giá trị hiện tại lưu trong `cert-pin.txt`.

## 8. Monitoring (tùy chọn)

```bash
docker compose --profile monitoring up -d   # Prometheus :9090, Grafana :3000
```

## 9. Vận hành

```bash
docker compose logs -f proxvn-server    # log
docker compose restart proxvn-server    # restart
docker compose pull && docker compose up -d --build   # nâng cấp
```

Gia hạn chứng chỉ: lặp lại bước 5 (hoặc cron `certbot renew`) rồi
`docker compose restart proxvn-server`.
