# 🚀 ProxVN Quick Start Guide

**Bắt đầu sử dụng ProxVN trong 2 phút!**

---

## 📦 Bước 1: Download

### Windows:
```powershell
# Mở PowerShell
cd Downloads
Invoke-WebRequest -Uri "https://vutrungocrong.fun/downloads/proxvn.exe" -OutFile "proxvn.exe"
```

### Linux:
```bash
wget https://vutrungocrong.fun/downloads/proxvn-linux-client
chmod +x proxvn-linux-client
```

### macOS:
```bash
# M1/M2/M3
curl -O https://vutrungocrong.fun/downloads/proxvn-mac-m1
chmod +x proxvn-mac-m1

# Intel
curl -O https://vutrungocrong.fun/downloads/proxvn-mac-intel
chmod +x proxvn-mac-intel
```

---

## 🎯 Bước 2: Chạy (Chọn Theo Use Case)

### 💻 Web Development

#### Next.js / React (port 3000)
```bash
# Start Next.js
npm run dev

# Terminal mới - chạy ProxVN
proxvn --proto http 3000
```

**✅ Kết quả:**
```
✓ Public URL: https://a1b2c3.vutrungocrong.fun
  → Forwards to http://localhost:3000
```

#### Node.js / Express (port 8080)
```bash
proxvn --proto http 8080
```

#### Python Flask (port 5000)
```bash
proxvn --proto http 5000
```

#### Laravel / PHP (port 8000)
```bash
proxvn --proto http 8000
```

---

### 🖥️ Remote Desktop

#### Windows RDP
```bash
# Enable Remote Desktop trên Windows
# Settings → System → Remote Desktop → ON

# Chạy ProxVN
proxvn 3389
```

**Kết nối từ máy khác:**
```
Windows: mstsc /v:103.77.246.206:10001
macOS:   Microsoft Remote Desktop → 103.77.246.206:10001
```

#### SSH Server
```bash
# Linux/macOS
proxvn 22

# Windows (có OpenSSH Server)
proxvn 22
```

**Kết nối:**
```bash
ssh your-username@103.77.246.206 -p 10002
```

---

### 🎮 Gaming

#### Minecraft Java Edition
```bash
# Server chạy port 25565
proxvn 25565
```

**Bạn bè connect:**
```
Server Address: 103.77.246.206:10003
```

#### Minecraft Bedrock Edition (PE)
```bash
# Server chạy port 19132 (UDP)
proxvn --proto udp 19132
```

**Bạn bè connect:**
```
Server: 103.77.246.206
Port: 10004
```

#### Counter-Strike / Source Games
```bash
proxvn --proto udp 27015
```

---

### 🏠 Homelab

#### Home Assistant
```bash
# Home Assistant chạy port 8123
proxvn --proto http 8123
```

**Truy cập từ xa:**
```
https://xyz789.vutrungocrong.fun
```

#### Plex Media Server
```bash
proxvn --proto http 32400
```

#### Synology NAS / DSM
```bash
proxvn --proto http 5000
```

---

### 🗄️ Database

#### MySQL / MariaDB
```bash
proxvn 3306
```

**Kết nối:**
```bash
mysql -h 103.77.246.206 -P 10005 -u root -p
```

#### PostgreSQL
```bash
proxvn 5432
```

**Kết nối:**
```bash
psql -h 103.77.246.206 -p 10006 -U postgres
```

#### MongoDB
```bash
proxvn 27017
```

---

## 🔐 Bước 3: Bảo Mật (Production)

### Dùng Certificate Pinning

```bash
# Với fingerprint server chính thức
proxvn --proto http 3000 \
       --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

**Lợi ích:**
- ✅ Chống Man-in-the-Middle attack
- ✅ Đảm bảo connect đúng server ProxVN
- ✅ Khuyến nghị cho production apps

**Chi tiết:** [CERT_PINNING.md](CERT_PINNING.md)

---

## 🎬 Sử Dụng Script (Dễ Hơn)

### Windows - Interactive Launcher

```powershell
cd scripts
.\run_client.bat
```

**Script sẽ hỏi:**
```
➤ Host   [127.0.0.1]:       ← Enter để dùng localhost
➤ Port   [vd: 3389 / 80]:   ← Nhập port (vd: 3000)
➤ Proto  [tcp / udp /http]: ← Chọn protocol (vd: http)
```

✅ **Xong!** ProxVN sẽ chạy với certificate pinning tự động.

---

## 📊 Hiểu Output

### Khi Chạy HTTP Mode:
```
[client] ⚠️  Certificate verification failed, retrying in INSECURE mode...
[client] ⚠️  This is normal for self-signed certificates in dev/test
✓ Đã kết nối tới ProxVN Server
✓ HTTP Tunnel: https://a1b2c3.vutrungocrong.fun
  → Forwards to: http://localhost:3000
  
Traffic:
  ↑ Upload:   0 B
  ↓ Download: 0 B
```

**Giải thích:**
- ⚠️ **Certificate warning:** Bình thường cho dev/test (server dùng self-signed cert)
- ✓ **Public URL:** Đây là URL để chia sẻ
- **Traffic:** Real-time bandwidth monitor

### Khi Chạy TCP Mode:
```
✓ Đã kết nối tới ProxVN Server
✓ Public Endpoint: 103.77.246.206:10001
  → Forwards to: localhost:3389
  
Active Sessions: 0
Total Sessions:  0
```

---

## 🛑 Dừng ProxVN

```bash
# Nhấn Ctrl+C trong terminal
^C
[client] Shutting down gracefully...
```

---

## 🔄 Use Cases Nâng Cao

### 1. Custom Server (Self-Hosted)
```bash
proxvn --server your-domain.com:8882 --proto http 3000
```

### 2. Forward Custom IP
```bash
# Forward server khác trong LAN
proxvn --host 192.168.1.100 --port 8080 --proto http
```

### 3. Run in Background (No UI)
```bash
# Linux/macOS
nohup ./proxvn-linux-client --proto http 3000 > proxvn.log 2>&1 &

# Windows
start /B proxvn.exe --ui=false --proto http 3000
```

---

## ❓ FAQ Nhanh

### Q: URL có đổi không khi restart?
**A:** Có. Subdomain ngẫu nhiên mỗi lần connect.

### Q: Có giới hạn băng thông không?
**A:** Không! Unlimited bandwidth.

### Q: Có giới hạn thời gian không?
**A:** Không! Chạy được 24/7.

### Q: Có an toàn không?
**A:** 
- ✅ An toàn cho dev/demo
- ⚠️ KHÔNG dùng cho production data nhạy cảm qua public server
- ✅ Hoặc self-host server riêng

### Q: Tôi bị lỗi "connection refused"?
**A:** 
1. Check internet: `ping 103.77.246.206`
2. Check firewall: Tắt tạm thời
3. Xem [Troubleshooting](README.md#-troubleshooting)

---

## 📚 Đọc Thêm

- 📖 **[README.md](README.md)** - Tài liệu đầy đủ
- 🔐 **[CERT_PINNING.md](src/CERT_PINNING.md)** - Certificate pinning chi tiết
- 🏠 **[DOMAIN_SETUP.md](src/DOMAIN_SETUP.md)** - Self-hosting guide
- 📖 **[GitHub Wiki](https://github.com/hoangtuvungcao/proxvn_tunnel/wiki)** - Advanced guides

---

## 🆘 Cần Giúp?

- 💬 **GitHub Discussions:** [Ask Questions](https://github.com/hoangtuvungcao/proxvn_tunnel/discussions)
- 🐛 **Bug Report:** [Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)
- 📧 **Email:** trong20843@gmail.com

---

<div align="center">

**Happy Tunneling! 🚀**

[⬆ Back to Top](#-proxvn-quick-start-guide)

</div>
