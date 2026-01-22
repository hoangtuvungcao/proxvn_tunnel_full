# ProxVN Deployment Guide

## 📋 Table of Contents
- [Quick Start with Docker](#quick-start-with-docker)
- [Manual Deployment](#manual-deployment)
- [Nginx Configuration](#nginx-configuration)
- [Environment Variables](#environment-variables)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

---

## 🐳 Quick Start with Docker

### Prerequisites
- Docker 20.10+
- Docker Compose 7.5+

### 1. Clone Repository
```bash
git clone https://github.com/hoangtuvungcao/proxvn_tunnel.git
cd proxvn_tunnel
```

### 2. Configure Environment
```bash
# Copy example env file
cp .env.server.example .env

# Edit configuration
nano .env
```

### 3. Start Services
```bash
# Start ProxVN server only
docker-compose up -d proxvn-server

# Start with monitoring stack
docker-compose --profile monitoring up -d
```

### 4. Check Status
```bash
docker-compose ps
docker-compose logs -f proxvn-server
```

### 5. Access Dashboard
Open browser: `http://localhost:8881`

---

## 🔧 Manual Deployment

### Linux Server

#### 1. Install Dependencies
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y build-essential sqlite3

# CentOS/RHEL
sudo yum install -y gcc sqlite-devel
```

#### 2. Build from Source
```bash
cd src/backend
go build -o ../../bin/proxvn-server ./cmd/server
```

#### 3. Create Systemd Service
```bash
sudo nano /etc/systemd/system/proxvn.service
```

```ini
[Unit]
Description=ProxVN Tunnel Server
After=network.target

[Service]
Type=simple
User=proxvn
WorkingDirectory=/opt/proxvn
EnvironmentFile=/opt/proxvn/.env
ExecStart=/opt/proxvn/bin/proxvn-server
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### 4. Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable proxvn
sudo systemctl start proxvn
sudo systemctl status proxvn
```

---

## 🌐 Nginx Configuration

### Reverse Proxy for Admin Dashboard

```nginx
# /etc/nginx/sites-available/proxvn-admin
server {
    listen 80;
    server_name admin.yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;

    location / {
        proxy_pass http://localhost:8881;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://localhost:8881;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Wildcard Subdomain for HTTP Tunnels

```nginx
# /etc/nginx/sites-available/proxvn-tunnels
server {
    listen 80;
    listen 443 ssl http2;
    server_name *.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8881;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Enable Configuration
```bash
sudo ln -s /etc/nginx/sites-available/proxvn-admin /etc/nginx/sites-enabled/
sudo ln -s /etc/nginx/sites-available/proxvn-tunnels /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## ⚙️ Environment Variables

### Essential Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | 0.0.0.0 | No |
| `SERVER_PORT` | Server port | 8882 | No |
| `HTTP_DOMAIN` | Base domain for HTTP tunnels | - | Yes (for HTTP) |
| `JWT_SECRET` | Secret key for JWT | - | **Yes** |
| `ADMIN_USERNAME` | Admin username | admin | No |
| `ADMIN_PASSWORD` | Admin password | - | **Yes** |

### Performance Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MAX_CONNECTIONS` | Max concurrent connections | 10000 |
| `BUFFER_SIZE` | Buffer size in bytes | 32768 |
| `ENABLE_COMPRESSION` | Enable gzip/zstd compression | true |
| `COMPRESSION_LEVEL` | Compression level (1-9) | 6 |
| `ENABLE_HTTP2` | Enable HTTP/2 | true |

See [.env.server.example](../.env.server.example) for full list.

---

## 📊 Monitoring

### Prometheus Metrics

ProxVN exposes Prometheus metrics at `/metrics` endpoint (default port 9090).

**Example prometheus.yml:**
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'proxvn'
    static_configs:
      - targets: ['proxvn-server:9090']
```

### Grafana Dashboard

Import the ProxVN dashboard (coming soon) or create custom panels:

**Key Metrics:**
- Active tunnels count
- Total connections
- Bandwidth usage (bytes up/down)
- Request rate
- Error rate
- Database connection pool stats

---

## 🔍 Troubleshooting

### Server Won't Start

```bash
# Check logs
docker-compose logs proxvn-server

# Or for systemd
journalctl -u proxvn -f
```

**Common Issues:**
- Port already in use: `lsof -i :8882`
- Permission denied: Check file permissions
- Database corruption: Restore from backup

### High Memory Usage

```bash
# Check current usage
docker stats proxvn-server

# Adjust limits in docker-compose.yml
services:
  proxvn-server:
    mem_limit: 512m
    memswap_limit: 1g
```

### Slow Performance

1. **Enable compression:**
   ```env
   ENABLE_COMPRESSION=true
   COMPRESSION_LEVEL=6
   ```

2. **Increase buffer size:**
   ```env
   BUFFER_SIZE=65536
   ```

3. **Optimize database:**
   - Run `VACUUM` regularly
   - Increase `CACHE_SIZE_MB`

### Connection Timeouts

```env
# Increase timeouts
READ_TIMEOUT=60s
WRITE_TIMEOUT=60s
IDLE_TIMEOUT=120s
```

---

## 🔐 Security Best Practices

1. **Change default credentials:**
   ```env
   ADMIN_USERNAME=your_admin
   ADMIN_PASSWORD=strong_password_here
   JWT_SECRET=generate_long_random_string
   ```

2. **Use TLS 1.3 only:**
   ```env
   TLS_MIN_VERSION=1.3
   ```

3. **Enable rate limiting:**
   ```env
   RATE_LIMIT_RPS=10
   RATE_LIMIT_BURST=20
   ENABLE_DDOS_PROTECTION=true
   ```

4. **Set up firewall:**
   ```bash
   # Allow only necessary ports
   sudo ufw allow 8882/tcp
   sudo ufw allow 10000:20000/tcp
   sudo ufw enable
   ```

---

## 📦 Backup & Restore

### Automatic Backups

Enabled by default:
```env
AUTO_BACKUP=true
BACKUP_INTERVAL=24h
BACKUP_DIR=./backups
BACKUP_RETENTION_DAYS=7
```

### Manual Backup

```bash
# Backup database
docker-compose exec proxvn-server sqlite3 /data/proxvn.db ".backup /backups/manual_backup_$(date +%Y%m%d).db"
```

### Restore from Backup

```bash
# Stop server
docker-compose stop proxvn-server

# Restore database
cp backups/manual_backup_20260120.db data/proxvn.db

# Start server
docker-compose start proxvn-server
```

---

## 🆘 Support

- 📧 Email: trong20843@gmail.com
- 💬 Telegram: [t.me/proxvn](https://t.me/proxvn)
- 🐛 Issues: [GitHub Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)

---

**Made with ❤️ in Vietnam**
