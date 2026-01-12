// ProxVN Dashboard - Real-time Backend Integration
// by TrongDev - 2025

const API_BASE = window.location.origin + '/api/v1';
let chart = null;
let ws = null;
let trafficData = {
    labels: [],
    upload: [],
    download: []
};

// Initialize Dashboard
document.addEventListener('DOMContentLoaded', () => {
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
        const response = await fetch(`${API_BASE}/metrics`);
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
        const response = await fetch(`${API_BASE}/tunnels`);
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
    tunnelsList.innerHTML = tunnels.map((tunnel, index) => `
        <div class="tunnel-card" style="animation-delay: ${index * 0.1}s">
            <div class="tunnel-header">
                <div class="tunnel-name">${escapeHtml(tunnel.name || 'Tunnel #' + (index + 1))}</div>
                <div class="status-badge status-${tunnel.status || 'inactive'}">
                    ${tunnel.status === 'active' ? '🟢 ACTIVE' : '⚫ INACTIVE'}
                </div>
            </div>
            <div class="tunnel-info">
                <div class="info-item">
                    <span class="info-icon">📡</span>
                    <span class="info-text">
                        <span class="info-label">Protocol:</span>
                        <span class="info-value">${(tunnel.protocol || 'tcp').toUpperCase()}</span>
                    </span>
                </div>
                <div class="info-item">
                    <span class="info-icon">🔗</span>
                    <span class="info-text">
                        <span class="info-label">Local:</span>
                        <span class="info-value">${tunnel.local_host || 'localhost'}:${tunnel.local_port || tunnel.localPort || 'N/A'}</span>
                    </span>
                </div>
                <div class="info-item">
                    <span class="info-icon">🌐</span>
                    <span class="info-text">
                        <span class="info-label">Public:</span>
                        <span class="info-value">${tunnel.public_host || tunnel.publicHost || `Port ${tunnel.public_port || tunnel.publicPort || 'N/A'}`}</span>
                    </span>
                </div>
                <div class="info-item">
                    <span class="info-icon">📊</span>
                    <span class="info-text">
                        <span class="info-label">Traffic:</span>
                        <span class="info-value">↑${formatBytes(tunnel.bytes_up || tunnel.bytesUp || 0)} ↓${formatBytes(tunnel.bytes_down || tunnel.bytesDown || 0)}</span>
                    </span>
                </div>
            </div>
        </div>
    `).join('');
}

function showNoTunnels() {
    document.getElementById('tunnelsList').innerHTML = '';
    document.getElementById('noTunnels').style.display = 'block';
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
    // TODO: Implement theme switching
    showToast('Theme toggle coming soon!', 'info');
}

// Auto Refresh
function startAutoRefresh() {
    setInterval(() => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            fetchMetrics();
            fetchTunnels();
        }
    }, 5000); // Refresh every 5 seconds if WebSocket is not connected
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
    if (connected) {
        status.innerHTML = '<span class="pulse-dot"></span><span>Connected</span>';
        status.style.background = 'rgba(16, 185, 129, 0.1)';
        status.style.borderColor = 'rgba(16, 185, 129, 0.3)';
        status.style.color = 'var(--success)';
    } else {
        status.innerHTML = '<span class="pulse-dot" style="background: var(--danger)"></span><span>Disconnected</span>';
        status.style.background = 'rgba(239, 68, 68, 0.1)';
        status.style.borderColor = 'rgba(239, 68, 68, 0.3)';
        status.style.color = 'var(--danger)';
    }
}

function showToast(message, type = 'info') {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = 'toast toast-' + type;
    toast.textContent = message;
    
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease-in forwards';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
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
