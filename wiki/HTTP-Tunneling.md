# HTTP Tunneling Guide 🌐

HTTP Tunneling là tính năng **mới nhất của ProxVN v4.0**, cho phép bạn nhận subdomain HTTPS đẹp ngay lập tức - giống như ngrok!

## 🎯 What is HTTP Tunneling?

HTTP Tunneling biến localhost của bạn thành URL HTTPS công khai:

```
localhost:3000  →  https://abc123.bacsycay.click
```

### Ưu điểm
- ✅ **URL đẹp** - Dễ chia sẻ với client/team
- ✅ **HTTPS miễn phí** - SSL certificate tự động
- ✅ **Zero config** - Không cần port forwarding
- ✅ **Share ngay** - Chỉ 1 command

---

## ⚡ Quick Start

### Basic Usage
```bash
proxvn --proto http [PORT]
```

### Examples

#### Share React App (port 3000)
```bash
proxvn --proto http 3000
```

**Output:**
```
✅ HTTP Tunnel Active
🌐 Public URL: https://a1b2c3.bacsycay.click
📍 Local: localhost:3000
🔒 Security: TLS Encrypted
```

#### Share Node.js API (port 8080)
```bash
proxvn --proto http 8080
```

#### Share XAMPP/Apache (port 80)
```bash
proxvn --proto http 80
```

#### Share Local HTTPS site (port 443)
```bash
proxvn --proto http 443
```

---

## 🔧 Advanced Usage

### Custom Server
Kết nối tới VPS riêng:
```bash
proxvn --server YOUR_VPS:8882 --proto http 3000
```

### Custom Local Host
Tunnel từ host khác localhost:
```bash
proxvn --host 192.168.1.100 --proto http 8080
```

### Custom Client ID
Set ID để dễ quản lý:
```bash
proxvn --id my-laptop --proto http 3000
```

---

## 💡 Common Use Cases

### 1. Web Development

#### Share Frontend với Client
```bash
# React/Vue/Next.js dev server
proxvn --proto http 3000
```

#### Test Webhook từ Services
```bash
# Expose API endpoint cho Stripe/PayPal webhooks
proxvn --proto http 8000
```

#### Demo Website cho Team
```bash
# Share staging site
proxvn --proto http 4000
```

### 2. Mobile App Development

#### Test API Callbacks
```bash
# iOS/Android app calling localhost API
proxvn --proto http 5000
```

#### Deep Link Testing
```bash
# Test OAuth redirects
proxvn --proto http 3000
```

### 3. IoT & Smart Home

#### Expose Home Assistant
```bash
proxvn --proto http 8123
```

#### Test Smart Device APIs
```bash
proxvn --proto http 8080
```

### 4. File Sharing

#### Share Files qua HTTP
```bash
# Start HTTP server
python -m http.server 8000

# Tunnel it
proxvn --proto http 8000

# Share URL với friends
```

---

## 🔐 Security Best Practices

### 1. Don't Share Sensitive Data
- ❌ Không tunnel database admin panels
- ❌ Không public credentials/API keys
- ✅ Chỉ tunnel cho development/testing

### 2. Use Authentication
Thêm basic auth vào app của bạn:

#### Express.js Example
```javascript
const auth = require('express-basic-auth');

app.use(auth({
    users: { 'admin': 'supersecret' },
    challenge: true
}));
```

#### Python Flask Example
```python
from flask_httpauth import HTTPBasicAuth

auth = HTTPBasicAuth()

@auth.verify_password
def verify_password(username, password):
    if username == 'admin' and password == 'secret':
        return username
```

### 3. Monitor Traffic
Check client TUI để xem requests:
```
📊 Traffic: ⬆️ 1.2 KB/s ⬇️ 450 B/s
🔌 Sessions: active 2 | total 15
```

---

## 🎨 Working with Frameworks

### React/Vite
```bash
# Dev server thường chạy port 5173
npm run dev

# Trong terminal khác
proxvn --proto http 5173
```

### Next.js
```bash
npm run dev  # Port 3000

# Tunnel
proxvn --proto http 3000
```

### Vue.js
```bash
npm run serve  # Port 8080

# Tunnel
proxvn --proto http 8080
```

### Django
```bash
python manage.py runserver 0.0.0.0:8000

# Tunnel
proxvn --proto http 8000
```

### Flask
```bash
flask run --host=0.0.0.0 --port=5000

# Tunnel
proxvn --proto http 5000
```

### Rails
```bash
rails server -b 0.0.0.0 -p 3000

# Tunnel
proxvn --proto http 3000
```

---

## 🌍 Testing from Different Locations

### Test from Mobile Device
1. Start tunnel:
```bash
proxvn --proto http 3000
```

2. Copy public URL: `https://abc123.bacsycay.click`

3. Open trên điện thoại (4G/5G để test thật)

### Test from Client Location
Share URL với client ở bất kỳ đâu:
- Client ở Mỹ vẫn truy cập được
- Không cần VPN
- Tốc độ phụ thuộc server location

---

## ⚠️ Limitations & Notes

### Subdomain is Ephemeral
- 🔄 **Reconnect:** Giữ subdomain cũ (trong vài phút)
- 🆕 **Restart:** Subdomain mới
- ❌ **Server restart:** Mất tất cả subdomain

### Not for Production
- ⚠️ ProxVN là development tool
- ⚠️ Không dùng cho production deployment
- ✅ Dùng cho: dev, demo, testing, sharing

### Performance
- ⚡ Latency: +20-50ms (qua tunnel)
- 📊 Bandwidth: Unlimited (nhưng phụ thuộc VPS)
- 🔌 Concurrent: Support nhiều clients

---

## 🔄 Auto Reconnect

ProxVN tự động reconnect khi mất mạng:

```
[INFO] Connection lost. Reconnecting...
[INFO] Reconnected successfully!
[INFO] Subdomain preserved: abc123.bacsycay.click
```

**Lưu ý:** Chỉ giữ subdomain nếu reconnect trong **5 phút**.

---

## 🎯 Troubleshooting

### "Connection refused"
**Nguyên nhân:** App chưa chạy hoặc sai port.

**Giải pháp:**
```bash
# Check app đã chạy chưa
netstat -an | grep :3000  # Linux/macOS
netstat -an | findstr :3000  # Windows

# Đảm bảo app bind 0.0.0.0 hoặc localhost
```

### "SSL Certificate Error" trên browser
**Nguyên nhân:** Cloudflare Proxy chưa bật.

**Giải pháp:** Báo admin bật Cloudflare Proxy cho wildcard domain.

### Subdomain bị đổi liên tục
**Nguyên nhân:** Client restart hoặc server restart.

**Giải pháp:** Ephemeral là design. Nếu cần fixed, dùng custom domain.

### "Too many requests"
**Nguyên nhân:** Server rate limit.

**Giải pháp:** Đợi 1 phút hoặc self-host server riêng.

---

## 📊 Monitoring

### Client TUI
Client hiển thị real-time stats:
```
╔══════════════════════════════════════════════════════
║  🟢 Status   : ACTIVE
║  🔗 Local     : localhost:3000
║  🌐 Public    : https://abc123.bacsycay.click
║  📡 Protocol  : HTTP
╠══════════════════════════════════════════════════════
║  📊 Traffic  : ⬆️  1.2 KB/s ⬇️  450 B/s
║  📈 Total    : 15.3 MB ↑  8.7 MB ↓
║  🔌 Sessions : active 2 | total 47
║  🏓 Ping     : 21 ms [||||]
╚══════════════════════════════════════════════════════
```

### Server Dashboard
Truy cập `http://VPS_IP:8881/dashboard/` để xem:
- Connected clients
- Traffic graphs
- Session history

---

## ➡️ Next Steps

- 🔌 [TCP & UDP Tunneling](TCP-UDP-Tunneling) - Tunnel nâng cao
- 🖥️ [Server Setup](Server-Setup) - Self-host server riêng
- 🛠️ [Troubleshooting](Troubleshooting) - Xử lý sự cố

---

[🏠 Back to Home](Home) | [📖 All Guides](Home#-documentation-structure)
