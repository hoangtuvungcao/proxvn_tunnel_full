#!/bin/bash
set -e

echo "⚙️  Tuning system for high-scale ProxVN deployment..."

# Create sysctl config
sudo tee /etc/sysctl.d/99-proxvn.conf > /dev/null <<'EOF'
# Network optimization for thousands of connections
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_max_tw_buckets = 1440000

# File descriptors for many open connections
fs.file-max = 2097152
fs.nr_open = 2097152

# Connection tracking
net.netfilter.nf_conntrack_max = 1048576

# Buffer sizes
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
EOF

# Apply settings
sudo sysctl -p /etc/sysctl.d/99-proxvn.conf

# Update systemd service limits
if [ -f /etc/systemd/system/proxvn-server.service ]; then
    echo "Updating systemd service limits..."
    sudo mkdir -p /etc/systemd/system/proxvn-server.service.d
    
    sudo tee /etc/systemd/system/proxvn-server.service.d/limits.conf > /dev/null <<'EOF'
[Service]
LimitNOFILE=1048576
LimitNPROC=512
CPUQuota=80%
MemoryMax=512M
EOF
    
    sudo systemctl daemon-reload
fi

echo "✅ System tuned for high performance!"
echo "   Max connections: 10,000+"
echo "   File descriptors: 1,048,576"
echo "   Memory limit: 512MB"
