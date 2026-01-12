# Welcome to ProxVN Wiki 🚀

> **Version 4.0.0** - Complete HTTP Tunneling Solution

ProxVN là công cụ tunnel mạnh mẽ, **100% miễn phí** và **không giới hạn**, giúp bạn đưa localhost lên Internet chỉ với một câu lệnh - giống như ngrok nhưng hoàn toàn free!

## 🌟 Quick Links

- 🏠 [**Home**](Home) - Bạn đang ở đây
- 📥 [**Installation**](Installation) - Hướng dẫn cài đặt
- 🌐 [**HTTP Tunneling**](HTTP-Tunneling) - Sử dụng HTTP mode với subdomain
- 🔌 [**TCP & UDP Tunneling**](TCP-UDP-Tunneling) - Tunnel TCP/UDP truyền thống
- 🖥️ [**Server Setup**](Server-Setup) - Self-host server riêng
- 🔐 [**Domain Configuration**](Domain-Configuration) - Cấu hình SSL và DNS
- 🛠️ [**Troubleshooting**](Troubleshooting) - Xử lý sự cố
- ❓ [**FAQ**](FAQ) - Câu hỏi thường gặp

## ⚡ Quick Start

### 1. Tải về từ Website

Truy cập **[vutrungocrong.fun](https://vutrungocrong.fun)** và tải file cho hệ điều hành của bạn.

### 2. Chạy ngay - Không cần cài đặt!

**Windows:**
```powershell
.\proxvn.exe --proto http 3000
```

**Linux/macOS:**
```bash
chmod +x proxvn-linux-client
./proxvn-linux-client --proto http 3000
```

**Kết quả:**
```
✅ HTTP Tunnel Active
🌐 Public URL: https://abc123.vutrungocrong.fun
📍 Local: localhost:3000
```

Đơn giản như vậy! 🎉

## 🌟 Key Features

### 🌐 HTTP Tunneling (MỚI v4.0!)
Nhận subdomain HTTPS đẹp ngay lập tức:
- ✅ `https://abc123.domain.com` - URL dễ chia sẻ
- ✅ SSL tự động - không cần Let's Encrypt
- ✅ Zero config - chỉ 1 command

### 🔌 TCP Tunneling
Public bất kỳ service TCP nào:
- SSH (port 22)
- RDP (port 3389)
- Database (MySQL, PostgreSQL...)
- Web server (HTTP/HTTPS)

### 🎮 UDP Tunneling
Cho game server và real-time apps:
- Minecraft PE (port 19132)
- CS:GO, Palworld
- Voice chat, video streaming

### 🚫 NO LIMITS!
- ∞ **Không giới hạn băng thông**
- ∞ **Không giới hạn thời gian** (24/7 nếu muốn)
- ∞ **Không giới hạn số tunnel**
- 💰 **100% Miễn phí** - không phí ẩn

### 🔒 Security
- TLS encryption cho tất cả kết nối
- Auto-reconnect khi mất mạng
- Secure by default

### 💻 Cross-Platform
- Windows (10/11)
- Linux (Ubuntu, Debian, CentOS...)
- macOS (Apple Silicon & Intel)
- Android (Termux)

## 📊 So Sánh Với Ngrok

| Tính Năng | ProxVN | Ngrok |
|-----------|--------|-------|
| HTTP Tunneling | ✅ Free | ✅ Free |
| TCP Tunneling | ✅ Free | 💰 $8/tháng |
| UDP Tunneling | ✅ Free | 💰 $20/tháng |
| Custom Domain | ✅ Free (Self-hosted) | 💰 Paid |
| Không giới hạn băng thông | ✅ | ❌ |
| Không giới hạn thời gian | ✅ | ❌ (2h) |
| Self-Hosted | ✅ | ❌ |
| Open Source | ✅ | ❌ |

## 🎯 Use Cases

### Web Development
Share localhost với client/team:
```bash
proxvn --proto http 3000  # Share React/Next.js app
```

### Mobile App Testing
Test webhook/API callbacks:
```bash
proxvn --proto http 8080  # Expose API endpoint
```

### Game Server
Host Minecraft cho bạn bè:
```bash
proxvn --proto udp 19132  # Minecraft PE
```

### Remote Access
Truy cập máy tính từ xa:
```bash
proxvn 3389  # Remote Desktop (RDP)
proxvn 22    # SSH
```

### IoT & Smart Home
Expose local IoT dashboard:
```bash
proxvn --proto http 8123  # Home Assistant
```

## 📚 Documentation Structure

### Beginner
1. [Installation](Installation) - Cài đặt trên Windows/Linux/macOS
2. [HTTP Tunneling](HTTP-Tunneling) - Bắt đầu với HTTP mode
3. [FAQ](FAQ) - Câu hỏi thường gặp

### Intermediate
1. [TCP & UDP Tunneling](TCP-UDP-Tunneling) - Tunnel nâng cao
2. [Troubleshooting](Troubleshooting) - Xử lý sự cố

### Advanced
1. [Server Setup](Server-Setup) - Self-host server riêng
2. [Domain Configuration](Domain-Configuration) - Cấu hình domain & SSL

## 🤝 Community & Support

- 💬 [GitHub Discussions](https://github.com/hoangtuvungcao/proxvn_tunnel/discussions)
- 🐛 [Report Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)
- 📧 Email: trong20843@gmail.com
- 🌐 Website: [vutrungocrong.fun](https://vutrungocrong.fun)

## 📝 Contributing

ProxVN là open-source project! Chúng tôi welcome contributions:

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push và tạo Pull Request

Chi tiết xem [CONTRIBUTING.md](https://github.com/hoangtuvungcao/proxvn_tunnel/blob/main/CONTRIBUTING.md)

## ⚖️ License

**FREE TO USE - NON-COMMERCIAL ONLY**

✅ Download, sử dụng, modify cho cá nhân  
❌ Không được bán hoặc kinh doanh  

Commercial license cần liên hệ tác giả.

---

© 2026 **ProxVN** • Developed by **TrongDev**

[🏠 Back to Top](#welcome-to-proxvn-wiki-)
