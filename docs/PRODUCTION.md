# ProxVN Production Deployment & Operations Guide

This guide describes operational steps, kernel tunings, backup tasks, and recovery flows for running ProxVN in a high-availability environment.

---

## 1. Operating System Tuning

For servers managing thousands of concurrent TCP/UDP/HTTP tunnels, the host OS default constraints must be increased.

### File Descriptor Limits
By default, Linux limits file descriptors per process to 1024, which limits concurrent client connections.

1. **Temporary adjustment**:
   ```bash
   ulimit -n 65535
   ```
2. **Persistent configuration**:
   Edit `/etc/security/limits.conf` and append:
   ```text
   * soft nofile 65535
   * hard nofile 65535
   ```
3. **Verify configuration**:
   ```bash
   ulimit -n
   ```

### Kernel TCP Tweaks
Create/edit `/etc/sysctl.d/99-proxvn.conf`:
```text
# Enable fast recycling of TIME_WAIT sockets
net.ipv4.tcp_tw_reuse = 1

# Max local port range
net.ipv4.ip_local_port_range = 10240 65535

# Increase connection queue size
net.core.somaxconn = 4096
net.ipv4.tcp_max_syn_backlog = 4096
```
Apply the changes:
```bash
sudo sysctl --system
```

---

## 2. Docker Deployment

Deploying ProxVN is streamlined via `docker-compose`.

### Launching the Stack
1. Ensure `nginx.conf`, `prometheus.yml`, and `docker-compose.yml` are in your root deployment directory.
2. Place a valid wildcard SSL certificate inside the `./ssl/` folder (`server.crt` and `server.key` matching your wildcard `*.yourdomain.com`).
3. Run:
   ```bash
   docker-compose up -d --build
   ```

---

## 3. Database Maintenance & Backups

ProxVN uses SQLite 3 with Write-Ahead Logging (WAL) enabled, which allows safe hot-backups while the server is active.

### Manual Backup
To make a transaction-safe backup of the SQLite database:
```bash
sqlite3 /data/proxvn.db ".backup '/backups/backup-$(date +%F).db'"
```

### Automated Nightly Backup Cron
Add the following job to `crontab -e`:
```text
0 2 * * * docker exec proxvn-server sqlite3 /data/proxvn.db ".backup '/backups/backup-nightly.db'"
```

---

## 4. Monitoring & Metrics

ProxVN exposes Prometheus metrics at `/metrics` (Default port `8881`).

### Prometheus Targets
Prometheus is configured in `prometheus.yml` to query metrics from the server container.

Key Metrics to track:
- `active_tunnels`: Gauge tracking active client connections.
- `active_connections`: Gauge tracking public client data plane connections.
- `total_connections`: Counter showing connection attempts.
- `total_bytes_up` / `total_bytes_down`: Traffic throughput counters.

---

## 5. Troubleshooting Checklist

### Port Conflict Issues
If the server fails to start, verify if ports `8881`, `8882`, or Nginx ports `80`/`443` are in use:
```bash
netstat -tulpn | grep -E '8881|8882|80|443'
```

### Client Takeover Logs
Verify that generation increments are working correctly if client drops connections:
```bash
docker logs -f proxvn-server | grep "takeover"
```

### Reset Database Migration
If schema migrations get corrupted, you can force migration by dropping the migration version table:
```sql
sqlite3 /data/proxvn.db "DROP TABLE IF EXISTS schema_migrations;"
```
On next startup, ProxVN will verify and rebuild the tables automatically.
