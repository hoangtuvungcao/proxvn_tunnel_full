# FAQ - Frequently Asked Questions ❓

Câu hỏi thường gặp về ProxVN.

## 🌐 General

### ProxVN là gì?
ProxVN là công cụ tunneling **miễn phí 100%**, giúp bạn đưa localhost lên Internet. Giống như ngrok nhưng:
- ✅ Hoàn toàn miễn phí
- ✅ Không giới hạn băng thông
- ✅ Không giới hạn thời gian
- ✅ Open source

### ProxVN có miễn phí không?
**Có!** ProxVN hoàn toàn miễn phí cho non-commercial use. Không có:
- ❌ Phí ẩn
- ❌ Premium tier
- ❌ Giới hạn băng thông/thời gian
- ❌ Ads

### ProxVN có an toàn không?
**Có**, ProxVN:
- ✅ Mã hóa TLS end-to-end
- ✅ Open source - code public trên GitHub
- ✅ Không lưu data của bạn
- ⚠️ NHƯNG: Đừng tunnel sensitive data (production database, API keys...)

### ProxVN vs Ngrok?
| Tính năng | ProxVN | Ngrok |
|-----------|--------|-------|
| HTTP Tunneling | ✅ Free | ✅ Free |
| TCP Tunneling | ✅ Free | 💰 $8/tháng |
| UDP Tunneling | ✅ Free | 💰 $20/tháng |
| Không giới hạn | ✅ | ❌ |
| Self-hosted | ✅ | ❌ |
| Open source | ✅ | ❌ |

---

## 🌐 HTTP Tunneling

### Làm sao để share website của tôi?
```bash
# Chạy app ở port 3000
npm run dev

# Tunnel
proxvn --proto http 3000

# Bạn sẽ nhận URL: https://abc123.vutrungocrong.fun
```

### Subdomain có thay đổi không?
**Có**. Subdomain là **ephemeral** (tạm thời):
- 🔄 **Reconnect** (mất mạng): Giữ subdomain cũ (5 phút)
- 🆕 **Restart** (tắt app): Subdomain mới
- ❌ **Server restart**: Tất cả subdomain bị reset

### Có thể tự chọn subdomain không?
**Hiện tại chưa**. Subdomain được random để tránh conflict. Tính năng custom subdomain sẽ có trong version sau.

**Workaround:** Self-host server + custom domain.

### Tại sao browser báo lỗi SSL?
**Nguyên nhân:** Cloudflare Proxy chưa bật.

**Giải pháp:** Liên hệ admin để bật Cloudflare Proxy cho wildcard domain.

### Có thể dùng custom domain không?
**Có!** Self-host server riêng:
1. Setup server trên VPS
2. Point domain của bạn tới VPS
3. Cấu hình SSL certificate
4. Client connect tới server của bạn

Chi tiết: [Server Setup Guide](Server-Setup)

---

## 🔌 TCP/UDP Tunneling

### Làm sao để public SSH server?
```bash
proxvn 22
```
Bạn sẽ nhận: `103.77.246.206:10000`

Kết nối:
```bash
ssh user@103.77.246.206 -p 10000
```

### Làm sao để host Minecraft server?
```bash
# Minecraft PE (UDP port 19132)
proxvn --proto udp 19132
```
✅ **An Toàn:** Từ phiên bản v4.0.0, ProxVN đã hỗ trợ **Mã hóa AES-GCM 256-bit** cho toàn bộ traffic UDP. Dữ liệu game/voice của bạn được bảo vệ an toàn khỏi việc nghe lén trên đường truyền Internet.

Bạn bè connect vào: `103.77.246.206:10000`

### TCP mode có SSL không?
**Có**, tất cả tunnel connections đều mã hóa TLS, kể cả TCP mode.

---

## 🛠️ Technical

### ProxVN hoạt động như thế nào?
```
[Your App] ←→ [ProxVN Client] ←TLS→ [ProxVN Server] ←→ [Internet]
           localhost           Encrypted Tunnel        Public
```

1. Client kết nối tới Server qua TLS
2. Server cấp public endpoint (port hoặc subdomain)
3. Requests từ Internet → Server → Client → Your App
4. Response ngược lại

### Port nào được sử dụng?
**Server:**
- `8881` - Dashboard/API
- `8882` - Tunnel connections
- `443` - HTTPS proxy (HTTP mode)

**Client:**
- Dynamic - Kết nối ra port 8882 của server

### Có cần mở firewall không?
**Server (VPS):**
```bash
# Required
sudo ufw allow 8882/tcp  # Tunnel
sudo ufw allow 443/tcp   # HTTPS (HTTP mode)
sudo ufw allow 8881/tcp  # Dashboard (optional)
```

**Client (Local):**
Không cần mở port. Client chỉ kết nối ra (outbound).

### Bandwidth limit là bao nhiêu?
**Không giới hạn!** Nhưng phụ thuộc:
- VPS bandwidth của server
- Network connection của bạn
- Server load

### Có thể chạy nhiều tunnel cùng lúc không?
**Có!** Mỗi tunnel cần 1 client instance:

Terminal 1:
```bash
proxvn --proto http 3000
```

Terminal 2:
```bash
proxvn --proto http 8080
```

Terminal 3:
```bash
proxvn 22
```

---

## 💻 Platform-Specific

### Windows Defender báo virus?
**Đây là false positive** do:
- Binary được pack với UPX
- Tunneling behavior giống malware

**Giải pháp:**
1. Add folder vào Windows Defender Exclusions
2. Hoặc build từ source

### macOS block app "unidentified developer"?
```bash
sudo xattr -d com.apple.quarantine proxvn-mac-m1
```

Hoặc: System Preferences → Security → Allow anyway

### Linux: "Permission denied"?
```bash
chmod +x proxvn-linux-client
```

### Android trong Termux không chạy?
```bash
# Ensure downloaded binary is ARM
file proxvn-android
# Output should show: ARM aarch64

# If wrong arch, download correct one
```

---

## 🔥 Troubleshooting

### "Connection refused" khi chạy client?
**Check:**
1. Server có đang chạy không?
2. Firewall có block port 8882?
3. Internet connection OK?

```bash
# Test connection to server
telnet 103.77.246.206 8882
```

### Local app không nhận traffic?
**Check:**
1. App có đang chạy không?
2. App bind đúng port?
3. App listen trên `0.0.0.0` hoặc `localhost`?

```bash
# Check port
netstat -an | grep :3000  # Linux/macOS
netstat -an | findstr :3000  # Windows
```

### "Too many requests" error?
**Nguyên nhân:** Server rate limit.

**Giải pháp:**
- Đợi 1 phút
- Hoặc self-host server riêng

### Tunnel bị disconnect liên tục?
**Nguyên nhân:**
- Network không ổn định
- Firewall block keep-alive packets

**Giải pháp:**
- Check network stability
- Try different network (4G/5G)
- Self-host server gần hơn

---

## 🖥️ Self-Hosting

### Tôi có thể host server riêng không?
**Có!** ProxVN support self-hosting:
1. Thuê VPS (recommend 1GB RAM+)
2. Build/Download server binary
3. Setup domain & SSL
4. Run server

Chi tiết: [Server Setup Guide](Server-Setup)

### Server requirements?
**Minimum:**
- 512MB RAM
- 1 CPU core
- 10GB storage
- Public IP

**Recommended:**
- 1GB+ RAM
- 2 CPU cores
- 20GB storage
- 100Mbps+ bandwidth

### VPS nào tốt cho ProxVN?
**Cheap options:**
- Contabo ($4/tháng)
- Hetzner ($4.5/tháng)
- DigitalOcean ($5/tháng)
- Vultr ($5/tháng)

**Vietnam hosting:**
- Azdigi (từ 50k/tháng)
- Matbao
- Nhanhoa

---

## 🔐 Privacy & Security

### ProxVN có lưu data của tôi không?
**Không**. ProxVN:
- ❌ Không log request content
- ❌ Không lưu credentials
- ✅ Chỉ log metadata (IP, port, timestamp) cho debug

### Có thể trust public server không?
**For development: Có**
**For production: Không**

Best practice:
- ✅ Development/testing: OK
- ✅ Demo websites: OK
- ❌ Production apps: KHÔNG
- ❌ Sensitive data: KHÔNG

### Làm sao để secure tunnel?
1. **Add authentication** vào app
2. **Whitelist IPs** nếu có thể
3. **Monitor traffic** qua client TUI
4. **Use HTTPS** app khi có thể
5. **Self-host server** cho sensitive apps

---

## 📱 Mobile & IoT

### Có thể tunnel từ Android không?
**Có!** Dùng Termux:
```bash
# In Termux
wget https://vutrungocrong.fun/downloads/proxvn-android
chmod +x proxvn-android
./proxvn-android --proto http 8080
```

### Có thể tunnel từ Raspberry Pi không?
**Có!** Dùng Linux client:
```bash
wget https://vutrungocrong.fun/downloads/proxvn-linux-client
chmod +x proxvn-linux-client
./proxvn-linux-client --proto http 8123  # Home Assistant
```

### Có iOS client không?
**Chưa**. Nhưng bạn có thể:
- Dùng web browser
- SSH vào server và chạy client
- Request iOS client trên GitHub Issues

---

## 🚀 Performance

### Latency bao nhiêu?
**Trung bình:** +20-50ms qua tunnel.

**Phụ thuộc:**
- Server location
- Network path
- Server load

### Max concurrent connections?
**Unlimited** (Lý thuyết)

**Thực tế:**
- Phụ thuộc server specs
- Public server: ~100-500 concurrent
- Self-hosted: Tùy VPS specs

### Có cache requests không?
**Không**. Mọi request đều forward real-time.

---

## 💰 Commercial Use

### Có thể dùng cho business không?
**Cần license thương mại**. ProxVN free là "NON-COMMERCIAL ONLY".

**Liên hệ:**
- Email: trong20843@gmail.com
- Subject: "ProxVN Commercial License"

### Có thể resell ProxVN không?
**Không** được phép resell hoặc rebrand.

**Nhưng:**
- ✅ Offer ProxVN setup service
- ✅ Include in paid tutorials/courses
- ✅ Use for client projects (non-commercial)

---

## 🤝 Contributing

### Làm sao để contribute?
1. Fork [GitHub repo](https://github.com/hoangtuvungcao/proxvn_tunnel)
2. Create feature branch
3. Code + test
4. Submit Pull Request

Chi tiết: [CONTRIBUTING.md](https://github.com/hoangtuvungcao/proxvn_tunnel/blob/main/CONTRIBUTING.md)

### Tìm thấy bug, làm gì?
[Report trên GitHub Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)

Include:
- OS & version
- ProxVN version
- Steps to reproduce
- Error messages/logs

### Feature request?
[GitHub Discussions](https://github.com/hoangtuvungcao/proxvn_tunnel/discussions)

---

## 📞 Support

### Câu hỏi chưa được trả lời?
- 💬 [GitHub Discussions](https://github.com/hoangtuvungcao/proxvn_tunnel/discussions)
- 🐛 [GitHub Issues](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)
- 📧 Email: trong20843@gmail.com
- 🌐 Website: [vutrungocrong.fun](https://vutrungocrong.fun)

---

[🏠 Back to Home](Home) | [📖 All Guides](Home#-documentation-structure)
