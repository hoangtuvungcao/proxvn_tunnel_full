// ProxVN Dashboard - Real-time Backend Integration
// by TrongDev - 2025

const API_BASE = window.location.origin + '/api';
const THEME_STORAGE_KEY = 'proxvn-theme';
const DEFAULT_THEME = 'dark';

// Check Auth first
const token = localStorage.getItem('token');
if (!token) {
    window.location.href = '/dashboard/login.html';
}

// Global Auth Header Helper
function authHeader() {
    return {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
    };
}

let chart = null;
let ws = null;
let trafficData = {
    labels: [],
    upload: [],
    download: []
};

// Initialize Dashboard
document.addEventListener('DOMContentLoaded', () => {
    initializeTheme();
    initChart();
    connectWebSocket();
    loadInitialData();
    setupEventListeners();
    startAutoRefresh();
});

// WebSocket Connection
function connectWebSocket() {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/api/v1/dashboard/ws`;

    try {
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('✅ WebSocket connected');
            updateConnectionStatus(true);
            showToast('Kết nối thành công!', 'success');
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                handleWebSocketMessage(data);
            } catch (error) {
                console.error('WebSocket message error:', error);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            updateConnectionStatus(false);
        };

        ws.onclose = () => {
            console.log('WebSocket disconnected');
            updateConnectionStatus(false);
            // Reconnect after 5 seconds
            setTimeout(connectWebSocket, 5000);
        };
    } catch (error) {
        console.error('WebSocket connection failed:', error);
        updateConnectionStatus(false);
    }
}

function handleWebSocketMessage(data) {
    if (data.type === 'tunnel_update') {
        renderTunnels(data.data);
    } else if (data.type === 'metrics') {
        updateStats(data.data);
    }
}

// Load Initial Data
async function loadInitialData() {
    try {
        await Promise.all([
            fetchMetrics(),
            fetchTunnels()
        ]);
    } catch (error) {
        console.error('Failed to load initial data:', error);
        showToast('Không thể tải dữ liệu. Đang dùng chế độ demo.', 'warning');
        loadDemoData();
    }
}

// Fetch Metrics
async function fetchMetrics() {
    try {
        const response = await fetch(`${API_BASE}/metrics`, { headers: authHeader() });
        if (response.status === 401) {
            logout();
            return;
        }
        if (!response.ok) throw new Error('Metrics fetch failed');

        const result = await response.json();
        if (result.success && result.data) {
            updateStats(result.data);
        }
    } catch (error) {
        console.error('Fetch metrics error:', error);
        // Fallback to demo data
    }
}

// Fetch Tunnels
async function fetchTunnels() {
    try {
        const response = await fetch(`${API_BASE}/tunnels`, { headers: authHeader() });
        if (response.status === 401) {
            logout();
            return;
        }
        if (!response.ok) throw new Error('Tunnels fetch failed');

        const result = await response.json();
        if (result.success && result.data) {
            renderTunnels(result.data);
        } else {
            showNoTunnels();
        }
    } catch (error) {
        console.error('Fetch tunnels error:', error);
        showNoTunnels();
    }
}

// Update Stats
function updateStats(metrics) {
    // Update counters with animation
    animateCounter('activeTunnels', metrics.activeTunnels || metrics.active_tunnels || 0);
    animateCounter('totalConnections', metrics.totalConnections || metrics.total_connections || 0);

    // Update data sizes
    const uploadMB = (metrics.totalBytesUp || metrics.total_bytes_up || 0) / (1024 * 1024);
    const downloadMB = (metrics.totalBytesDown || metrics.total_bytes_down || 0) / (1024 * 1024);

    document.getElementById('totalUpload').querySelector('.data-size').textContent = formatBytes(uploadMB * 1024 * 1024);
    document.getElementById('totalDownload').querySelector('.data-size').textContent = formatBytes(downloadMB * 1024 * 1024);

    // Update chart
    updateChart(uploadMB, downloadMB);
}

// Render Tunnels
function renderTunnels(tunnels) {
    const tunnelsList = document.getElementById('tunnelsList');
    const noTunnels = document.getElementById('noTunnels');

    if (!tunnels || tunnels.length === 0) {
        showNoTunnels();
        return;
    }

    noTunnels.style.display = 'none';
    const cardMarkup = tunnels.map((tunnel) => {
        const name = escapeHtml(tunnel.name || tunnel.label || 'Tunnel');
        const protocol = (tunnel.protocol || 'tcp').toLowerCase();
        const protocolLabel = protocol.toUpperCase();
        const status = (tunnel.status || 'inactive').toLowerCase();
        const localHost = escapeHtml(`${tunnel.local_host || tunnel.localHost || 'localhost'}:${tunnel.local_port || tunnel.localPort || 'N/A'}`);
        const publicEndpoint = escapeHtml(
            tunnel.public_host ||
            tunnel.publicHost ||
            (tunnel.remote_host || tunnel.remoteHost ? `${tunnel.remote_host || tunnel.remoteHost}:${tunnel.public_port || tunnel.publicPort || 'N/A'}` : `Port ${tunnel.public_port || tunnel.publicPort || 'N/A'}`)
        );
        const bytesUp = formatBytes(tunnel.bytes_up || tunnel.bytesUp || 0);
        const bytesDown = formatBytes(tunnel.bytes_down || tunnel.bytesDown || 0);
        const remotePort = escapeHtml(String(tunnel.remote_port || tunnel.remotePort || tunnel.public_port || tunnel.publicPort || '—'));
        const createdAt = tunnel.created_at || tunnel.createdAt;
        const lastHeartbeat = tunnel.last_heartbeat || tunnel.lastHeartbeat;

        const badgeClass = `badge-${protocol}`;
        const statusClass = status === 'active' ? 'status-active' : 'status-inactive';

        return `
            <article class="tunnel-card">
                <header class="tunnel-card-header">
                    <div class="tunnel-card-heading">
                        <span class="tunnel-name">${name}</span>
                        <span class="badge ${badgeClass}">${protocolLabel}</span>
                    </div>
                    <span class="status-chip ${statusClass}">${status === 'active' ? 'Đang chạy' : 'Tạm dừng'}</span>
                </header>
                <dl class="tunnel-card-grid">
                    <div class="tunnel-card-item">
                        <dt>Local</dt>
                        <dd>${localHost}</dd>
                    </div>
                    <div class="tunnel-card-item">
                        <dt>Public</dt>
                        <dd>${publicEndpoint}</dd>
                    </div>
                    <div class="tunnel-card-item">
                        <dt>Remote Port</dt>
                        <dd>${remotePort}</dd>
                    </div>
                    <div class="tunnel-card-item">
                        <dt>Traffic</dt>
                        <dd>↑ ${bytesUp} · ↓ ${bytesDown}</dd>
                    </div>
                    ${createdAt ? `<div class="tunnel-card-item"><dt>Tạo lúc</dt><dd>${formatTimestamp(createdAt)}</dd></div>` : ''}
                    ${lastHeartbeat ? `<div class="tunnel-card-item"><dt>Heartbeat</dt><dd>${formatTimestamp(lastHeartbeat)}</dd></div>` : ''}
                </dl>
            </article>
        `;
    }).join('');

    tunnelsList.innerHTML = cardMarkup;
}

function showNoTunnels() {
    const tunnelsList = document.getElementById('tunnelsList');
    const noTunnels = document.getElementById('noTunnels');

    if (tunnelsList) {
        tunnelsList.innerHTML = '';
    }

    if (noTunnels) {
        noTunnels.style.display = 'grid';
    }
}

// Initialize Chart
function initChart() {
    const ctx = document.getElementById('trafficChart').getContext('2d');
    chart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: trafficData.labels,
            datasets: [
                {
                    label: 'Upload (MB)',
                    data: trafficData.upload,
                    borderColor: 'rgb(79, 172, 254)',
                    backgroundColor: 'rgba(79, 172, 254, 0.1)',
                    borderWidth: 3,
                    tension: 0.4,
                    fill: true
                },
                {
                    label: 'Download (MB)',
                    data: trafficData.download,
                    borderColor: 'rgb(16, 185, 129)',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    borderWidth: 3,
                    tension: 0.4,
                    fill: true
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    labels: { color: '#9ca3af', font: { size: 14, weight: '600' } }
                },
                tooltip: {
                    backgroundColor: 'rgba(30, 41, 59, 0.9)',
                    titleColor: '#fff',
                    bodyColor: '#9ca3af',
                    borderColor: 'rgba(102, 126, 234, 0.3)',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: true
                }
            },
            scales: {
                x: {
                    grid: { color: 'rgba(255, 255, 255, 0.05)' },
                    ticks: { color: '#6b7280' }
                },
                y: {
                    grid: { color: 'rgba(255, 255, 255, 0.05)' },
                    ticks: { color: '#6b7280' },
                    beginAtZero: true
                }
            }
        }
    });
}

function updateChart(upload, download) {
    const now = new Date().toLocaleTimeString();

    trafficData.labels.push(now);
    trafficData.upload.push(upload);
    trafficData.download.push(download);

    // Keep only last 10 data points
    if (trafficData.labels.length > 10) {
        trafficData.labels.shift();
        trafficData.upload.shift();
        trafficData.download.shift();
    }

    chart.update();
}

// Event Listeners
function setupEventListeners() {
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadInitialData();
        showToast('Đã làm mới!', 'success');
    });

    document.getElementById('themeToggle').addEventListener('click', toggleTheme);
}

function toggleTheme() {
    const currentTheme = document.body.getAttribute('data-theme') || DEFAULT_THEME;
    const nextTheme = currentTheme === 'dark' ? 'light' : 'dark';
    applyTheme(nextTheme);
}

// Auto Refresh
function startAutoRefresh() {
    setInterval(() => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            fetchMetrics();
            fetchTunnels();
        }
    }, 2000); // Faster Refresh
}

// Utility Functions
function animateCounter(elementId, target) {
    const element = document.getElementById(elementId).querySelector('.counter');
    if (!element) {
        document.getElementById(elementId).textContent = target;
        return;
    }

    const current = parseInt(element.textContent) || 0;
    const increment = (target - current) / 20;
    let count = current;

    const timer = setInterval(() => {
        count += increment;
        if ((increment > 0 && count >= target) || (increment < 0 && count <= target)) {
            count = target;
            clearInterval(timer);
        }
        element.textContent = Math.floor(count);
    }, 50);
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function updateConnectionStatus(connected) {
    const status = document.getElementById('connectionStatus');
    if (!status) return;

    const indicator = status.querySelector('.pill-indicator');
    const label = status.querySelector('.pill-label');

    status.classList.remove('is-online', 'is-offline');
    status.classList.add(connected ? 'is-online' : 'is-offline');

    if (label) {
        label.textContent = connected ? 'Đã kết nối' : 'Mất kết nối';
    }

    if (indicator) {
        indicator.setAttribute('aria-hidden', 'true');
    }
}

function showToast(message, type = 'info') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;

    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('is-leaving');
        toast.addEventListener('animationend', () => toast.remove(), { once: true });
    }, 2800);
}

// Demo Data (fallback)
function loadDemoData() {
    const demoMetrics = {
        activeTunnels: 2,
        totalConnections: 15,
        totalBytesUp: 128000000,
        totalBytesDown: 256000000
    };

    const demoTunnels = [
        {
            name: 'Web Server',
            status: 'active',
            protocol: 'tcp',
            local_host: 'localhost',
            local_port: 80,
            public_port: 10001,
            public_host: '103.78.0.204:10001',
            bytes_up: 64000000,
            bytes_down: 128000000
        },
        {
            name: 'API Server',
            status: 'active',
            protocol: 'tcp',
            local_host: 'localhost',
            local_port: 3000,
            public_port: 10002,
            public_host: '103.78.0.204:10002',
            bytes_up: 32000000,
            bytes_down: 64000000
        }
    ];

    updateStats(demoMetrics);
    renderTunnels(demoTunnels);
}

// Handle page visibility
document.addEventListener('visibilitychange', () => {
    if (!document.hidden && (!ws || ws.readyState !== WebSocket.OPEN)) {
        connectWebSocket();
        loadInitialData();
    }
});

console.log('🚀 ProxVN Dashboard initialized by TrongDev');

// Theme Utilities
function initializeTheme() {
    const storedTheme = localStorage.getItem(THEME_STORAGE_KEY);
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const theme = storedTheme || (prefersDark ? 'dark' : DEFAULT_THEME);
    applyTheme(theme);
}

function applyTheme(theme) {
    document.body.setAttribute('data-theme', theme);
    localStorage.setItem(THEME_STORAGE_KEY, theme);

    const toggle = document.getElementById('themeToggle');
    if (toggle) {
        const icon = toggle.querySelector('.theme-toggle-icon');
        const label = toggle.querySelector('.theme-toggle-label');
        if (icon) {
            icon.textContent = theme === 'dark' ? '🌙' : '☀️';
        }
        if (label) {
            label.textContent = theme === 'dark' ? 'Dark' : 'Light';
        }
    }

    updateChartThemeColors();
}

function updateChartThemeColors() {
    if (!chart) return;

    const getColor = (name, fallback) => {
        const value = getComputedStyle(document.body).getPropertyValue(name).trim();
        return value || fallback;
    };

    const accentCyan = getColor('--accent-cyan', 'rgba(14, 165, 233, 1)');
    const accentEmerald = getColor('--accent-emerald', 'rgba(16, 185, 129, 1)');
    const textMuted = getColor('--color-text-muted', '#94a3b8');
    const gridColor = getColor('--color-border', 'rgba(148, 163, 184, 0.2)');
    const tooltipBg = getColor('--color-card', 'rgba(15, 23, 42, 0.9)');
    const tooltipBorder = getColor('--color-border-strong', 'rgba(148, 163, 184, 0.3)');

    chart.data.datasets[0].borderColor = accentCyan;
    chart.data.datasets[0].backgroundColor = hexToRgba(accentCyan, 0.15);
    chart.data.datasets[1].borderColor = accentEmerald;
    chart.data.datasets[1].backgroundColor = hexToRgba(accentEmerald, 0.15);

    chart.options.scales.x.ticks.color = textMuted;
    chart.options.scales.y.ticks.color = textMuted;
    chart.options.scales.x.grid.color = hexToRgba(gridColor, 0.4);
    chart.options.scales.y.grid.color = hexToRgba(gridColor, 0.4);

    chart.options.plugins.legend.labels.color = textMuted;
    chart.options.plugins.tooltip.backgroundColor = tooltipBg;
    chart.options.plugins.tooltip.borderColor = tooltipBorder;
    chart.options.plugins.tooltip.titleColor = getColor('--color-text-primary', '#ffffff');
    chart.options.plugins.tooltip.bodyColor = textMuted;

    chart.update('none');
}

function hexToRgba(input, alpha) {
    if (!input) return `rgba(148, 163, 184, ${alpha})`;
    const hex = input.replace('#', '').trim();
    if (hex.startsWith('rgb')) {
        return input.replace(')', `, ${alpha})`).replace('rgb', 'rgba');
    }
    if (hex.length === 3) {
        const [r, g, b] = hex.split('').map((char) => parseInt(char + char, 16));
        return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    }
    if (hex.length === 6) {
        const r = parseInt(hex.substring(0, 2), 16);
        const g = parseInt(hex.substring(2, 4), 16);
        const b = parseInt(hex.substring(4, 6), 16);
        return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    }

    function logout() {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/dashboard/login.html';
    }
    return `rgba(148, 163, 184, ${alpha})`;
}

function formatTimestamp(input) {
    if (!input) return '';
    try {
        const date = new Date(input);
        if (Number.isNaN(date.getTime())) return escapeHtml(String(input));
        return date.toLocaleString();
    } catch (error) {
        return escapeHtml(String(input));
    }
}
