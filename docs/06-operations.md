# 06 - Operations Guide

Hướng dẫn vận hành, giám sát và bảo trì hệ thống ProxVN.

## Giám sát (Monitoring)

### 1. Log theo dõi
- **Docker**: `docker compose logs -f proxvn-server`
- **Systemd**: `journalctl -u proxvn -f`

Client log:
- Theo dõi log output trên terminal client. Tìm các dòng lỗi `[error]` hoặc `reconnect`.

### 2. Metrics
ProxVN Server expose Prometheus-compatible metrics (nếu được cấu hình):
- Endpoint: `/api/v1/metrics`
- Các chỉ số quan trọng: `active_tunnels`, `bytes_transferred`, `uptime`.

## Nâng cấp (Upgrade)

### Server
1. Pull code mới: `git pull`
2. Build lại: `./scripts/build.sh`
3. Restart service: `docker compose restart` hoặc `systemctl restart proxvn`

### Client
Tải binary mới từ Dashboard (nếu server có host file) hoặc từ GitHub Releases, sau đó thay thế file cũ.

## Bảo trì định kỳ
- Kiểm tra dung lượng disk (database SQLite có thể phình to nếu log nhiều).
- Rotate SSL certificates (nếu dùng Let's Encrypt thủ công).
- Restart service định kỳ (không bắt buộc, nhưng tốt để refresh in-memory cache).
