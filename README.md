<div align="center">
ProxVN - Phiên bản 5.0.0


![ProxVN Logo](https://img.shields.io/badge/ProxVN-v4.0.1-blue?style=for-the-badge)
![License](https://img.shields.io/badge/License-Non--Commercial-orange?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS%20%7C%20Android-green?style=for-the-badge)

**Công cụ tunnel miễn phí, mạnh mẽ - Đưa localhost lên Internet trong 30 giây**

[🚀 Quick Start](#-quick-start-30-giây) • [📥 Download](#-download) • [📖 Wiki](https://github.com/hoangtuvungcao/proxvn_tunnel/wiki) • [🐛 Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)

</div>

---
## CERT-PIN 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
## 🆕 What's New in v4.0.1 (2026-01-12)

### 🔧 Bug Fixes
- **Fixed:** Lỗi `x509: certificate signed by unknown authority` đã được sửa
- **Changed:** Tự động bỏ qua kiểm tra certificate theo mặc định (dễ sử dụng hơn với self-signed certs)
- **Removed:** Cờ `--insecure` (không còn cần thiết)
- **Improved:** Đơn giản hóa logic kết nối TLS

### 📌 Migration Guide
**Trước (v4.0.0):**
```bash
proxvn --proto tcp --port 3389 --insecure  # ❌ Lỗi: flag không tồn tại
```

**Bây giờ (v4.0.1):**
```bash
proxvn --proto tcp --port 3389  # ✅ Hoạt động ngay lập tức, không cần cờ gì
```

**Bảo mật cao (khuyến nghị cho production):**
```bash
proxvn --proto tcp --port 3389 --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

---

## 🎯 ProxVN Là Gì?

ProxVN giúp bạn **đưa dịch vụ chạy ở localhost lên Internet** mà không cần:
- ❌ Public IP
- ❌ Port forwarding
- ❌ Cấu hình phức tạp
- ❌ Kiến thức networking

**Chỉ cần 1 dòng lệnh:** `proxvn --proto http 3000` 🚀

---

## ✨ Tính Năng Nổi Bật

### 🌐 HTTP Tunneling với HTTPS Tự Động
```bash
proxvn --proto http 3000
# → Nhận ngay: https://abc123.vutrungocrong.fun
```
- ✅ HTTPS tự động (SSL/TLS certificate có sẵn)
- ✅ Subdomain ngẫu nhiên an toàn
- ✅ Không giới hạn băng thông

### 🔌 TCP Tunneling
```bash
proxvn 22
# → SSH: ssh user@103.77.246.206 -p 10001
```
- ✅ Mọi giao thức TCP: SSH, RDP, MySQL, PostgreSQL...
- ✅ Port mapping tự động

### 🎮 UDP Tunneling với AES-256 Encryption
```bash
proxvn --proto udp 19132
# → Minecraft PE, VoIP, game multiplayer
```
- ✅ **MỚI v4.0:** Mã hóa AES-GCM 256-bit
- ✅ An toàn tuyệt đối cho game/voice traffic

### 🛡️ Security Features
- ✅ TLS 1.2+ cho mọi kết nối
- ✅ Certificate pinning (chống MITM)
- ✅ Rate limiting (chống DoS)
- ✅ End-to-end encryption cho UDP

---

## 🚀 Quick Start (30 Giây)

### Windows:
```powershell
# 1. Chạy (ví dụ: web server port 3000)
.\proxvn.exe --proto http 3000

# ✅ Nhận ngay URL: https://xyz789.vutrungocrong.fun
```

### Linux/macOS:
```bash
# 1. Chạy
./proxvn-linux-client --proto http 8080

# ✅ Done!
```

### 🎬 Hoặc Dùng Script (Windows):
```powershell
.\run_windows.bat
# → Nhập Host, Port, Protocol → Done!
```

---

## 📥 Download

### 📦 Pre-built Binaries

| Platform | Download | SHA256 |
|----------|----------|--------|
| **Windows** | [proxvn.exe](https://github.com/hoangtuvungcao/proxvn_tunnel/releases/download/v5.0.0/windows.zip) | `sha256:53ecbae0afc41f076218010bf462929c8e267f7f60b3855617eedf7475663014` |
| **Linux** | [proxvn-linux-client](https://github.com/hoangtuvungcao/proxvn_tunnel/releases/download/v5.0.0/linux.zip) | `sha256:7dff6cbfecf9b63255838dba109d79cfeea9b20aff5c24ab2841f3b60daf0c95` |
| **macOS (M1)** | [proxvn-mac-m1](https://github.com/hoangtuvungcao/proxvn_tunnel/releases/download/v5.0.0/mac-m1.zip) | `sha256:dbd23b7bb888b925797efd3151684b0164cbb785ed2d1f922cd7f5a69fd113c4` |
| **macOS (Intel)** | [proxvn-mac-intel](https://github.com/hoangtuvungcao/proxvn_tunnel/releases/download/v5.0.0/mac-intel.zip) | `sha256:6a6bb45a5447fa6f9d6aa16f2f8b102d40b35ad6cdfcbba2f416f1a3bd2eadac` |
| **Android (Termux)** | [proxvn-android](https://github.com/hoangtuvungcao/proxvn_tunnel/releases/download/v5.0.0/android.zip) | `sha256:888235024237ac8c7b5f87430205d83ad769eaa8dfb5866d3206595c2ae93acb` |

### 🏗️ Build Từ Source
```bash
git clone https://github.com/hoangtuvungcao/proxvn_tunnel
cd proxvn_tunnel
cd scripts && ./build.bat  # Windows
```

---

## 📖 Sử Dụng Chi Tiết

### 1. HTTP Tunneling (Web Development)

```bash
# React/Next.js (port 3000)
proxvn --proto http 3000

# Node.js/Express (port 8080)
proxvn --proto http 8080

# Python Flask (port 5000)
proxvn --proto http 5000

# HTTPS local app (port 443)
proxvn --proto http 443
```

**Kết quả:**
```
✓ Đã kết nối tới ProxVN Server
✓ Public URL: https://a1b2c3.vutrungocrong.fun
  → Forwards to: http://localhost:3000
```

### 2. TCP Tunneling (Remote Access)

```bash
# SSH Server
proxvn 22
# Connect: ssh user@103.77.246.206 -p 10001

# Windows RDP
proxvn 3389
# Connect: mstsc /v:103.77.246.206:10002

# MySQL Database
proxvn 3306
# Connect: mysql -h 103.77.246.206 -P 10003 -u root
```

### 3. UDP Tunneling (Gaming)

```bash
# Minecraft Bedrock Edition
proxvn --proto udp 19132

# VoIP (Voice Chat)
proxvn --proto udp 5060

# Game Server
proxvn --proto udp 27015
```

---

## 🔐 Certificate Pinning (Production)

Để bảo mật tối đa, dùng certificate pinning:

```bash
# Với cert fingerprint cố định
proxvn --proto http 3000 \
       --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

**Chi tiết:** Xem [CERT_PINNING.md](src/CERT_PINNING.md)

---

- 📘 **[Quick Start Guide](src/QUICKSTART.md)** - Bắt đầu nhanh với các ví dụ cụ thể
- 🔐 **[Certificate Pinning](src/CERT_PINNING.md)** - Bảo mật nâng cao
- 🏠 **[Self-Hosting Guide](src/DOMAIN_SETUP.md)** - Tự host server riêng
- 📖 **[GitHub Wiki](https://github.com/hoangtuvungcao/proxvn_tunnel/wiki)** - Tài liệu chi tiết

---

## 🎯 Use Cases

### 👨‍💻 Development
- Preview website cho client (không cần deploy)
- Test webhook từ GitHub, Stripe, PayPal...
- Share localhost với team remote

### 🏠 Homelab
- Remote access Home Assistant, Plex, NAS
- Tránh CGNAT khi ISP không cho public IP
- Không cần mở port forwarding (an toàn hơn)

### 🎮 Gaming
- Host Minecraft server từ nhà
- Chơi LAN games qua Internet
- Voice chat servers

### 🤖 IoT & Devices
- Remote access Raspberry Pi, Arduino
- Monitor cameras, sensors từ xa
- Control home automation

---

## 🆚 So Sánh Với Ngrok

| Tính Năng | ProxVN | Ngrok |
|-----------|--------|-------|
| HTTP Tunneling | ✅ Free | ✅ Free |
| HTTPS Auto | ✅ Free | ✅ Free |
| TCP Tunneling | ✅ **Free** | 💰 $8/tháng |
| UDP Tunneling | ✅ **Free + Encrypted** | 💰 $20/tháng |
| Bandwidth | ✅ Unlimited | ❌ Limited |
| Time Limit | ✅ None | ❌ 2 hours |
| Open Source | ✅ Yes | ❌ No |
| Self-Hostable | ✅ Yes | ❌ No |
| Vietnamese Support | ✅ Yes | ❌ No |

---

## ⚙️ Advanced Configuration

### Custom Server
```bash
# Dùng server tự host
proxvn --server your-server.com:8882 --proto http 3000
```

### Custom Host/Port
```bash
# Forward custom host:port
proxvn --host 192.168.1.100 --port 8080 --proto http
```

### Disable TUI
```bash
# Chạy ở background không có UI
proxvn --ui=false --proto http 3000
```

---

## 🛠️ Server Information

### Default Server (Public)
- **Address:** `103.77.246.206:8882`
- **Domain:** `*.vutrungocrong.fun`
- **Location:** Vietnam
- **Status:** [Check Status](https://vutrungocrong.fun)

### Certificate Fingerprint
```
5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```
**Valid Until:** ~2041 (Cloudflare Origin Certificate)

---

## 🐛 Troubleshooting

### "Connection refused"
```bash
# 1. Check server status
ping 103.77.246.206

# 2. Check firewall
# Windows: Disable Windows Firewall tạm thời
# Linux: sudo ufw allow 8882

# 3. Test với telnet
telnet 103.77.246.206 8882
```

### Chi tiết: [FAQ](src/wiki/FAQ.md)

---

## 🤝 Contributing

Contributions are welcome!

1. Fork repo
2. Create feature branch: `git checkout -b feature/amazing`
3. Commit changes: `git commit -am 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing`
5. Submit Pull Request

---

## 📜 License

**Non-Commercial License**

- ✅ Personal use: FREE
- ✅ Educational: FREE
- ✅ Open source projects: FREE
- ❌ Commercial use: Contact for license

**Commercial License:** trong20843@gmail.com

---

## 📞 Support & Contact

- 🌐 **Website:** [vutrungocrong.fun](https://vutrungocrong.fun)
- 💬 **GitHub Discussions:** [Discussions](https://github.com/hoangtuvungcao/proxvn_tunnel/discussions)
- 🐛 **Bug Reports:** [Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)
- 📧 **Email:** trong20843@gmail.com
- 📖 **Wiki:** [Documentation](https://github.com/hoangtuvungcao/proxvn_tunnel/wiki)

---

## 🙏 Credits

- Built with ❤️ in Vietnam
- Powered by Go
- SSL by Cloudflare
- Server hosting by [AIVPS.online](https://aivps.online) 🚀
- Inspired by ngrok, frp, and localtunnel

---

## ⭐ Star History

If you find ProxVN useful, please consider giving it a star! ⭐

---

<div align="center">

**[⬆ Back to Top](#proxvn---phiên-bản-401)**

Made with ❤️ by [Hoàng Tử Vùng Cao](https://github.com/hoangtuvungcao)  
Server powered by [AIVPS.online](https://aivps.online)

</div>
```
