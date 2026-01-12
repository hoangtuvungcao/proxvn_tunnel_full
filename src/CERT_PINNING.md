# 🔐 Certificate Pinning Guide

**Bảo mật nâng cao cho ProxVN với Certificate Pinning**

---

## 📖 Certificate Pinning Là Gì?

Certificate Pinning là kỹ thuật bảo mật bằng cách "ghim" (pin) **SHA256 fingerprint** của server certificate. 

### 🎯 Tại Sao Cần?

**Kịch bản tấn công MITM:**
```
[Bạn] ←→ [Kẻ Tấn Công] ←→ [Server ProxVN]
           ↑
    Giả mạo certificate
    (dùng CA bị hack/lừa)
```

**Với Certificate Pinning:**
```
[Bạn] ←→ [Kẻ Tấn Công] ✗ Connection Rejected!
           ↑
    Fingerprint không khớp
    → Tấn công thất bại
```

---

## 🔑 Fingerprint Server ProxVN

### Production Server

**Server:** `103.77.246.206:8882`  
**Domain:** `*.vutrungocrong.fun`

**SHA256 Fingerprint:**
```
5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

**Certificate:** Cloudflare Origin Certificate  
**Valid Until:** ~2041 (15 years)  
**Issuer:** Cloudflare Inc ECC CA-3

---

## 💻 Cách Sử Dụng

### 1. Copy Fingerprint Ở Trên

```bash
CERT_PIN=5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

### 2. Chạy ProxVN Với `--cert-pin`

```bash
# HTTP Tunnel
proxvn --proto http 3000 --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6

# TCP Tunnel  
proxvn 22 --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6

# UDP Tunnel
proxvn --proto udp 19132 --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

### 3. Hoặc Dùng Script (Tự Động)

```powershell
# Windows
cd scripts
.\run_client.bat
# Script đã include cert-pin sẵn!
```

---

## 🔍 Verify Certificate (Tùy Chọn)

### Trên Windows (PowerShell):

```powershell
# Get certificate fingerprint
$cert = (New-Object System.Net.Sockets.TcpClient("103.77.246.206", 8882)).GetStream()
$sslStream = New-Object System.Net.Security.SslStream($cert, $false, {$true})
$sslStream.AuthenticateAsClient("103.77.246.206")
$remoteCert = $sslStream.RemoteCertificate
$hash = [System.Security.Cryptography.SHA256]::Create()
$certHash = $hash.ComputeHash($remoteCert.Export([System.Security.Cryptography.X509Certificates.X509ContentType]::Cert))
$fingerprint = -join ($certHash | ForEach-Object { $_.ToString("x2") })
Write-Host "Fingerprint: $fingerprint"
$sslStream.Close()
$cert.Close()
```

**Expected Output:**
```
Fingerprint: 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6
```

### Trên Linux/macOS:

```bash
echo | openssl s_client -connect 103.77.246.206:8882 2>/dev/null | \
  openssl x509 -fingerprint -sha256 -noout | \
  cut -d'=' -f2 | tr -d ':' | tr '[:upper:]' '[:lower:]'
```

---

## ⚠️ Lỗi Thường Gặp

### 1. "certificate fingerprint mismatch"

**Lỗi:**
```
Error: certificate fingerprint mismatch: 
expected 8ff1f269..., got abc123...
```

**Nguyên nhân:**
- Server certificate đã thay đổi (renew/rotate)
- Bạn dùng sai fingerprint
- **Hiếm:** Có kẻ tấn công MITM

**Giải pháp:**
1. Verify fingerprint lại bằng PowerShell/OpenSSL
2. Nếu match với `8ff1f269...` → OK, dùng tiếp
3. Nếu khác → Liên hệ admin: trong20843@gmail.com

### 2. Connection Works Without `--cert-pin` But Fails With It

**Nguyên nhân:**  
Bạn đang ở mạng có MITM proxy (corporate/school network)

**Giải pháp:**
```bash
# Bỏ cert-pin trong môi trường này
proxvn --proto http 3000

# Hoặc dùng VPN để bypass proxy
```

---

## 🎯 Khi Nào Dùng Certificate Pinning?

### ✅ NÊN DÙNG:

- **Production applications** cần bảo mật cao
- **Sensitive data** (financial, healthcare)
- **Corporate environments** có risk MITM
- **Long-running tunnels** (24/7 services)
- **API webhooks** từ third-party services

### ❌ KHÔNG CẦN:

- **Dev/testing** quick demos
- **Short-lived tunnels** (< 1 hour)
- **Public demos** không có data nhạy cảm
- **Let's Encrypt servers** (cert đổi mỗi 90 ngày)

---

## 📊 So Sánh Bảo Mật

| Mode | TLS | Cert Validation | MITM Protection | Use Case |
|------|-----|-----------------|-----------------|----------|
| **Default** | ✅ Yes | ⚠️ Auto-fallback | ⚠️ Basic | Dev/Test |
| **`--cert-pin`** | ✅ Yes | ✅ Strict | ✅ Maximum | Production |
| **`--insecure`** | ✅ Yes | ❌ Disabled | ❌ None | Debug only |

---

## 🔄 Cert Lifecycle

### Production Server (Current)

```
Certificate: Cloudflare Origin Certificate
Created:     ~2026
Expires:     ~2041 (15 years)
Fingerprint: 8ff1f269... (unchanged for 15 years)
```

**✅ Kết luận:** Fingerprint `8ff1f269...` sẽ **valid cho ~15 năm**, không cần update thường xuyên.

### Nếu Certificate Renew

**Khi nào?** ~2041 (còn 15 năm nữa)

**Làm gì?**
1. Admin sẽ public fingerprint mới
2. Update `--cert-pin` với giá trị mới
3. Hoặc update script `run_client.bat`

---

## 🛠️ Script Automation

### Auto-Update Certificate (Advanced)

```bash
#!/bin/bash
# auto-pin.sh

SERVER="103.77.246.206:8882"
EXPECTED="5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6"

# Get current fingerprint
CURRENT=$(echo | openssl s_client -connect $SERVER 2>/dev/null | \
  openssl x509 -fingerprint -sha256 -noout | \
  cut -d'=' -f2 | tr -d ':' | tr '[:upper:]' '[:lower:]')

if [ "$CURRENT" != "$EXPECTED" ]; then
    echo "⚠️  WARNING: Certificate fingerprint mismatch!"
    echo "Expected: $EXPECTED"
    echo "Current:  $CURRENT"
    exit 1
fi

echo "✅ Certificate verified!"
./proxvn --proto http 3000 --cert-pin $EXPECTED
```

---

## 📚 Tài Liệu Liên Quan

- 🚀 **[Quick Start](QUICKSTART.md)** - Bắt đầu nhanh
- 📖 **[README](README.md)** - Tài liệu đầy đủ
- 🏠 **[Self-Hosting](DOMAIN_SETUP.md)** - Tự host server
- ❓ **[FAQ](wiki/FAQ.md)** - Câu hỏi thường gặp

---

## 🆘 Support

- 💬 **GitHub:** [Discussions](https://github.com/hoangtuvungcao/proxvn_tunnel/discussions)
- 🐛 **Issues:** [Bug Report](https://github.com/hoangtuvungcao/proxvn_tunnel/issues)
- 📧 **Email:** trong20843@gmail.com

---

<div align="center">

**Certificate Pinning = Maximum Security! 🔐**

[⬆ Back to Top](#-certificate-pinning-guide)

</div>

========================================
Subject:     CN=*.vutrungocrong.fun
Issuer:      CN=Cloudflare Inc ECC CA-3
Valid From:  1/1/2026
Valid To:    12/31/2026

SHA256 Fingerprint:
5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6

========================================
Usage:
========================================
proxvn.exe --server 103.77.246.206:8882 `
           --cert-pin 5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6 `
           --proto http 3000
```

### Trên Linux/macOS:

```bash
# Lấy fingerprint
echo | openssl s_client -connect 103.77.246.206:8882 2>/dev/null | \
  openssl x509 -fingerprint -sha256 -noout | \
  cut -d'=' -f2 | tr -d ':' | tr '[:upper:]' '[:lower:]'
```

---

## 2. Sử Dụng Certificate Pinning

### Cơ Bản:
```bash
# Copy fingerprint từ output ở trên
proxvn.exe --server 103.77.246.206:8882 \
           --cert-pin a1b2c3d4... \
           --proto http 3000
```

### Lưu Vào Biến:
```powershell
# PowerShell
$FINGERPRINT = "a1b2c3d4e5f6..."
proxvn.exe --server 103.77.246.206:8882 `
           --cert-pin $FINGERPRINT `
           --proto http 3000
```

---

## 3. Khi Nào Dùng?

### ✅ NÊN DÙNG Certificate Pinning:

- **Production deployment** với server certificate cố định
- **Corporate networks** có MITM proxy/firewall
- **High-security applications** cần chống MITM
- **Self-hosted servers** mà bạn control certificate

### ❌ KHÔNG CẦN Dùng:

- **Dev/testing** với self-signed cert thường xuyên đổi
- **Default connection** - client tự động xử lý
- **Let's Encrypt** auto-renew (fingerprint đổi mỗi 90 ngày)

---

## 4. Lỗi Thường Gặp

### Lỗi: "certificate fingerprint mismatch"

**Nguyên nhân:**
- Server certificate đã đổi (renew/rotate)
- Bạn dùng sai fingerprint
- Có kẻ tấn công MITM (hiếm)

**Giải pháp:**
```powershell
# Lấy lại fingerprint mới
cd scripts
.\get-cert-fingerprint.ps1
```

### Lỗi: "Connection refused"

**Nguyên nhân:**
- Server không chạy
- Firewall block port 8882

**Giải pháp:**
```bash
# Test connection trước
Test-NetConnection -ComputerName 103.77.246.206 -Port 8882
```

---

## 5. Bảo Mật

### Fingerprint Có An Toàn Không?

✅ **CÓ** - Fingerprint là thông tin public, không bí mật
- Không giống password
- An toàn để commit vào Git
- An toàn để share qua email/chat

### Tại Sao Pinning An Toàn Hơn?

```
Normal TLS:
[Client] ←→ [Trusted CA] ←→ [Server]
           ↑ Nếu CA bị hack/lừa → Kẻ tấn công có thể giả mạo

Certificate Pinning:
[Client] ←→ [Exact Certificate Only] ←→ [Server]
           ↑ Chỉ chấp nhận ĐÚNG certificate, không tin CA
```

---

## 6. Automation

### Script Tự Động Lấy + Connect:

```powershell
# auto-connect-pinned.ps1
$SERVER = "103.77.246.206:8882"
$PROTO = "http"
$PORT = 3000

Write-Host "Getting certificate fingerprint..." -ForegroundColor Yellow
$FINGERPRINT = & .\scripts\get-cert-fingerprint.ps1 | Select-String -Pattern '^[a-f0-9]{64}$'

Write-Host "Connecting with certificate pinning..." -ForegroundColor Green
& .\bin\proxvn.exe --server $SERVER `
                   --cert-pin $FINGERPRINT.Line `
                   --proto $PROTO $PORT
```

---

**Liên Hệ:**
- GitHub: https://github.com/hoangtuvungcao/proxvn_tunnel
- Email: trong20843@gmail.com
