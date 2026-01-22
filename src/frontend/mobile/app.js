const API_BASE = `${window.location.origin}/api/v1`;
const WS_ENDPOINT = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/dashboard/ws`;
const THEME_KEY = 'proxvn-mobile-theme';

let ws;
let reconnectTimer;

const elements = {
    themeToggle: document.getElementById('themeToggle'),
    connectionStatus: document.getElementById('connectionStatus'),
    connectionValue: document.querySelector('.connection-value'),
    manualRefresh: document.getElementById('manualRefresh'),
    statActive: document.getElementById('statActiveTunnels'),
    statTotalConnections: document.getElementById('statTotalConnections'),
    statUpload: document.getElementById('statTotalUpload'),
    statDownload: document.getElementById('statTotalDownload'),
    lastUpdated: document.getElementById('lastUpdated'),
    tunnelsList: document.getElementById('tunnelsList'),
    emptyState: document.getElementById('emptyState'),
    toastContainer: document.getElementById('toastContainer'),
    actionDialog: document.getElementById('actionDialog'),
    dialogTitle: document.getElementById('dialogTitle'),
    dialogDescription: document.getElementById('dialogDescription'),
    dialogCode: document.getElementById('dialogCode'),
    dialogClose: document.getElementById('dialogClose'),
    copyAction: document.getElementById('copyAction'),
};

document.addEventListener('DOMContentLoaded', () => {
    applyInitialTheme();
    wireEvents();
    establishWebSocket();
    loadInitialData();
});

function wireEvents() {
    elements.themeToggle?.addEventListener('click', toggleTheme);
    elements.manualRefresh?.addEventListener('click', () => {
        loadInitialData(true);
    });

    document.querySelectorAll('.action-card').forEach((card) => {
        card.addEventListener('click', () => presentAction(card));
    });

    elements.dialogClose?.addEventListener('click', () => closeDialog());
    elements.copyAction?.addEventListener('click', handleCopyAction);

    elements.actionDialog?.addEventListener('close', () => {
        elements.dialogCode.textContent = '';
    });
}

function applyInitialTheme() {
    const saved = localStorage.getItem(THEME_KEY) || 'dark';
    applyTheme(saved);
}

function toggleTheme() {
    const current = document.body.getAttribute('data-theme') || 'dark';
    applyTheme(current === 'dark' ? 'light' : 'dark');
}

function applyTheme(theme) {
    document.body.setAttribute('data-theme', theme);
    localStorage.setItem(THEME_KEY, theme);

    if (elements.themeToggle) {
        elements.themeToggle.textContent = theme === 'dark' ? '🌙' : '☀️';
    }
}

async function loadInitialData(showToastMessage = false) {
    try {
        const [metrics, tunnels] = await Promise.all([
            fetchJSON(`${API_BASE}/metrics`),
            fetchJSON(`${API_BASE}/tunnels`)
        ]);

        if (metrics?.success) updateMetrics(metrics.data);
        if (tunnels?.success) renderTunnels(tunnels.data);

        updateLastUpdated();
        notify('Đã làm mới dữ liệu', 'success', showToastMessage);
    } catch (error) {
        console.error('Initial load failed', error);
        notify('Không thể tải dữ liệu từ server', 'error', showToastMessage);
    }
}

async function fetchJSON(url) {
    const response = await fetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}

function establishWebSocket() {
    try {
        ws = new WebSocket(WS_ENDPOINT);
        ws.addEventListener('open', () => {
            setConnectionState(true);
            notify('Đã kết nối realtime', 'success');
            if (reconnectTimer) {
                clearTimeout(reconnectTimer);
                reconnectTimer = undefined;
            }
        });

        ws.addEventListener('message', ({ data }) => {
            try {
                const payload = JSON.parse(data);
                handleRealtime(payload);
            } catch (err) {
                console.warn('Invalid WS payload', err);
            }
        });

        ws.addEventListener('close', () => {
            setConnectionState(false);
            if (!reconnectTimer) {
                reconnectTimer = setTimeout(establishWebSocket, 4000);
            }
        });

        ws.addEventListener('error', (err) => {
            console.error('WebSocket error', err);
            setConnectionState(false);
            ws.close();
        });
    } catch (error) {
        console.error('WS init error', error);
        setConnectionState(false);
    }
}

function handleRealtime(payload) {
    if (payload.type === 'metrics') {
        updateMetrics(payload.data);
        updateLastUpdated();
    }

    if (payload.type === 'tunnel_update') {
        renderTunnels(payload.data);
    }
}

function updateMetrics(data = {}) {
    const active = data.activeTunnels ?? data.active_tunnels ?? 0;
    const totalConnections = data.totalConnections ?? data.total_connections ?? 0;
    const totalUp = humanBytes(data.totalBytesUp ?? data.total_bytes_up ?? 0);
    const totalDown = humanBytes(data.totalBytesDown ?? data.total_bytes_down ?? 0);

    if (elements.statActive) elements.statActive.textContent = active;
    if (elements.statTotalConnections) elements.statTotalConnections.textContent = totalConnections;
    if (elements.statUpload) elements.statUpload.textContent = totalUp;
    if (elements.statDownload) elements.statDownload.textContent = totalDown;
}

function renderTunnels(tunnels = []) {
    if (!Array.isArray(tunnels) || tunnels.length === 0) {
        elements.tunnelsList.innerHTML = '';
        elements.emptyState.hidden = false;
        return;
    }

    elements.emptyState.hidden = true;
    elements.tunnelsList.innerHTML = tunnels.map(buildTunnelCard).join('');
}

function buildTunnelCard(tunnel) {
    const name = escapeHTML(tunnel.name || tunnel.label || 'Tunnel');
    const protocol = (tunnel.protocol || 'tcp').toLowerCase();
    const status = (tunnel.status || 'active').toLowerCase();
    const publicHost = tunnel.public_host || tunnel.publicHost || `${tunnel.remote_host || '0.0.0.0'}:${tunnel.public_port || '—'}`;
    const local = `${tunnel.local_host || tunnel.localHost || 'localhost'}:${tunnel.local_port || tunnel.localPort || '—'}`;
    const trafficUp = humanBytes(tunnel.bytes_up || tunnel.bytesUp || 0);
    const trafficDown = humanBytes(tunnel.bytes_down || tunnel.bytesDown || 0);
    const heartbeat = tunnel.last_heartbeat || tunnel.lastHeartbeat;

    return `
        <article class="tunnel-card">
            <header>
                <div class="tunnel-name">${name}</div>
                <span class="badge ${protocol}">${protocol.toUpperCase()}</span>
            </header>
            <div class="tunnel-meta">
                <div>
                    <strong>Local</strong>
                    <span>${escapeHTML(local)}</span>
                </div>
                <div>
                    <strong>Public</strong>
                    <span>${escapeHTML(publicHost)}</span>
                </div>
                <div>
                    <strong>Traffic</strong>
                    <span>↑ ${trafficUp} • ↓ ${trafficDown}</span>
                </div>
                <div>
                    <strong>Trạng thái</strong>
                    <span>${status === 'active' ? 'Online' : 'Offline'}</span>
                </div>
                ${heartbeat ? `<div><strong>Heartbeat</strong><span>${formatTime(heartbeat)}</span></div>` : ''}
            </div>
        </article>
    `;
}

function setConnectionState(isOnline) {
    elements.connectionStatus.classList.toggle('is-online', isOnline);
    elements.connectionStatus.classList.toggle('is-offline', !isOnline);

    if (elements.connectionValue) {
        elements.connectionValue.textContent = isOnline ? 'Đang kết nối' : 'Mất kết nối';
    }
}

function updateLastUpdated() {
    if (!elements.lastUpdated) return;
    const now = new Date();
    elements.lastUpdated.textContent = `Cập nhật lúc ${now.toLocaleTimeString('vi-VN', { hour12: false })}`;
}

function notify(message, type = 'info', forced = false) {
    if (!elements.toastContainer || (!forced && type === 'success')) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    elements.toastContainer.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('exit');
        toast.addEventListener('animationend', () => toast.remove(), { once: true });
    }, 2600);
}

function presentAction(card) {
    if (!elements.actionDialog) return;

    const subtitle = card.querySelector('.action-subtitle')?.textContent || '';
    const title = card.querySelector('.action-title')?.textContent || 'Lệnh mẫu';

    elements.dialogTitle.textContent = title;
    elements.dialogDescription.textContent = 'Sao chép và chạy trên client để khởi tạo tunnel tương ứng.';
    elements.dialogCode.textContent = subtitle;

    elements.actionDialog.showModal();
    elements.dialogCode.focus();
}

function closeDialog() {
    elements.actionDialog?.close();
}

async function handleCopyAction() {
    if (!elements.dialogCode?.textContent) return;

    try {
        await navigator.clipboard.writeText(elements.dialogCode.textContent.trim());
        notify('Đã copy vào clipboard', 'success', true);
    } catch (error) {
        console.error('Copy failed', error);
        notify('Không thể copy. Hãy thử thủ công.', 'error', true);
    }
}

function humanBytes(bytes = 0) {
    if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    const exponent = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    const value = bytes / Math.pow(1024, exponent);
    return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[exponent]}`;
}

function escapeHTML(str = '') {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function formatTime(timestamp) {
    const date = new Date(timestamp);
    if (Number.isNaN(date.getTime())) return '—';
    return date.toLocaleString('vi-VN', { hour12: false });
}

window.addEventListener('beforeunload', () => {
    if (ws) ws.close();
});
