# Installation Guide 

Hướng dẫn cài đặt ProxVN trên tất cả các nền tảng.

## Method 1: Download from Website (Khuyến nghị)

### Bước 1: Truy cập Website
[https://bacsycay.click](https://bacsycay.click)

### Bước 2: Chọn Platform

#### Windows
- **Client:** `proxvn.exe`
- **Server:** `svproxvn.exe`

#### Linux
- **Client:** `proxvn-linux-client`
- **Server:** `proxvn-linux-server`

#### macOS
- **Apple Silicon (M1/M2):** `proxvn-mac-m1`
- **Intel:** `proxvn-mac-intel`

#### Android
- **Termux:** `proxvn-android`

---

## Method 2: Build from Source

### Requirements
- [Go 1.21+](https://go.dev/dl/)
- Git

### Bước 1: Clone Repository
```bash
git clone https://github.com/hoangtuvungcao/proxvn_tunnel_full.git
cd proxvn_tunnel
```

### Bước 2: Build

**Windows:**
```powershell
cd scripts
.\build.bat
```

**Linux/macOS:**
```bash
cd scripts
chmod +x build.sh
./build.sh
```

### Bước 3: Lấy Binaries
Tất cả file thực thi sẽ nằm trong thư mục `bin/`:
```
bin/
├── proxvn.exe              # Windows Client
├── svproxvn.exe            # Windows Server
├── proxvn-linux-client     # Linux Client
├── proxvn-linux-server     # Linux Server
├── proxvn-mac-m1           # macOS Client (Apple Silicon)
├── proxvn-mac-intel        # macOS Client (Intel)
├── proxvn-android          # Android Client
└── server.tar.gz           # Server package
```

---

## Windows Installation

### Quick Install
1. Tải `proxvn.exe` từ website
2. Đặt vào thư mục bất kỳ (vd: `C:\Tools\`)
3. Mở PowerShell/CMD tại thư mục đó
4. Chạy:
```powershell
.\proxvn.exe --proto http 3000
```

### Thêm vào PATH (Tùy chọn)
Để chạy `proxvn` từ bất kỳ đâu:

1. Copy `proxvn.exe` vào `C:\Windows\System32`
   
   HOẶC
   
2. Thêm thư mục chứa `proxvn.exe` vào System PATH:
   - Right-click **This PC**  **Properties**
   - **Advanced system settings**  **Environment Variables**
   - Trong **System variables**, chọn **Path**  **Edit**
   - **New**  Nhập đường dẫn (vd: `C:\Tools`)
   - **OK**  **OK**

3. Mở PowerShell mới và test:
```powershell
proxvn --help
```

### Launcher Script (Tùy chọn)
Tạo file `run_proxvn.bat`:
```batch
@echo off
proxvn.exe --proto http 3000
pause
```

---

## Linux Installation

### Ubuntu/Debian

#### Quick Install
```bash
# Download
wget https://bacsycay.click/downloads/proxvn-linux-client

# Cấp quyền thực thi
chmod +x proxvn-linux-client

# Chạy
./proxvn-linux-client --proto http 3000
```

#### Install System-wide
```bash
# Copy vào /usr/local/bin
sudo cp proxvn-linux-client /usr/local/bin/proxvn

# Test
proxvn --help
```

#### Desktop Entry (GUI Shortcut)
Tạo file `~/.local/share/applications/proxvn.desktop`:
```ini
[Desktop Entry]
Name=ProxVN Tunnel
Comment=Tunnel localhost to internet
Exec=/usr/local/bin/proxvn --proto http 3000
Icon=network-transmit-receive
Terminal=true
Type=Application
Categories=Network;
```

### CentOS/RHEL
```bash
# Download
curl -O https://bacsycay.click/downloads/proxvn-linux-client

# Cấp quyền
chmod +x proxvn-linux-client

# Move to /usr/local/bin
sudo mv proxvn-linux-client /usr/local/bin/proxvn

# Test
proxvn --help
```

---

## macOS Installation

### Bước 1: Download
**Apple Silicon (M1/M2):**
```bash
curl -O https://bacsycay.click/downloads/proxvn-mac-m1
chmod +x proxvn-mac-m1
```

**Intel:**
```bash
curl -O https://bacsycay.click/downloads/proxvn-mac-intel
chmod +x proxvn-mac-intel
```

### Bước 2: Bypass Gatekeeper
macOS sẽ block app chưa verified. Fix:
```bash
# Allow app
sudo xattr -d com.apple.quarantine proxvn-mac-m1

# HOẶC System Preferences
# Security & Privacy  Allow anyway
```

### Bước 3: Install System-wide
```bash
# Copy to /usr/local/bin
sudo cp proxvn-mac-m1 /usr/local/bin/proxvn

# Test
proxvn --help
```

---

## Android Installation (Termux)

### Bước 1: Install Termux
- Tải Termux từ [F-Droid](https://f-droid.org/en/packages/com.termux/)
- **KHÔNG tải từ Play Store** (outdated)

### Bước 2: Setup Termux
```bash
# Update packages
pkg update && pkg upgrade

# Install required tools
pkg install wget
```

### Bước 3: Download ProxVN
```bash
# Download
wget https://bacsycay.click/downloads/proxvn-android

# Cấp quyền
chmod +x proxvn-android

# Chạy
./proxvn-android --proto http 8080
```

### Bước 4: Access from PC
Share localhost từ Android:
```bash
# Start tunnel
./proxvn-android --proto http 8080

# Bạn sẽ nhận URL như:
# https://abc123.bacsycay.click
```

---

## Verify Installation

### Test Client
```bash
proxvn --help
```

**Expected output:**
```
╔════════════════════════════════════════════════════════════════════════════╗
║                 ProxVN v4.0.0 - Client                                     ║
║            Tunnel Localhost ra Internet - Miễn Phí 100%                   ║
╚════════════════════════════════════════════════════════════════════════════╝
...
```

### Test Connection
```bash
# Start a simple HTTP server (for testing)
# Python 3
python -m http.server 8000

# Python 2
python -m SimpleHTTPServer 8000

# Node.js
npx http-server -p 8000
```

Trong terminal khác:
```bash
proxvn --proto http 8000
```

Truy cập URL public để test!

---

## Common Issues

### Windows: "Windows protected your PC"
**Giải pháp:**
1. Click **More info**
2. Click **Run anyway**

### Linux: "Permission denied"
```bash
chmod +x proxvn-linux-client
```

### macOS: "App can't be opened because it is from an unidentified developer"
```bash
sudo xattr -d com.apple.quarantine proxvn-mac-m1
```

### "Command not found"
File chưa trong PATH. Chạy với `./proxvn` hoặc add to PATH.

---

## Update ProxVN

### Download New Version
1. Truy cập [bacsycay.click](https://bacsycay.click)
2. Tải version mới
3. Replace file cũ

### Check Version
```bash
proxvn --help | head -n 3
```

---

## Next Steps

- [HTTP Tunneling Guide](HTTP-Tunneling) - Sử dụng HTTP mode
- [TCP/UDP Guide](TCP-UDP-Tunneling) - Tunnel nâng cao
- [Troubleshooting](Troubleshooting) - Nếu gặp vấn đề

---

[ Back to Home](Home)
