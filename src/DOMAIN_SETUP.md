# Hướng Dẫn Cấu Hình Domain cho HTTP Tunneling

## Tổng Quan

HTTP Tunneling cho phép bạn public website qua subdomain HTTPS (vd: `abc123.yourdomain.com`) thay vì IP:port. Tính năng này yêu cầu:
1. **Domain riêng** (vd: `vutrungocrong.fun`)
2. **Wildcard SSL certificate** cho `*.yourdomain.com`
3. **DNS Wildcard record** trỏ về IP server

> **Lưu ý:** Nếu không cấu hình domain, HTTP mode sẽ tự động fallback về IP:port như TCP/UDP.

---

## Bước 1: Chuẩn Bị Domain

### 1.1. Mua Domain
- Đăng ký domain tại: Namecheap, GoDaddy, Cloudflare, hoặc nhà cung cấp Việt Nam (Mat Bao, iNET, ...)
- **Khuyến nghị:** Dùng Cloudflare (miễn phí DNS management)

### 1.2. Chuyển DNS về Cloudflare (Tùy chọn nhưng khuyến nghị)
1. Đăng ký Cloudflare: https://cloudflare.com
2. Add domain của bạn
3. Đổi nameserver tại nhà đăng ký domain theo hướng dẫn Cloudflare
4. Chờ 24h để DNS propagate

---

## Bước 2: Cấu Hình DNS

### Option A: Sử Dụng Cloudflare (Khuyến nghị)

1. **Login Cloudflare Dashboard**
2. **Chọn domain** của bạn
3. **DNS Records** → **Add Record**

Tạo 2 records sau:

**Record 1: Main Domain (Optional)**
```
Type: A
Name: @
Content: <IP_SERVER_CỦA_BẠN>
Proxy status: DNS only (bật proxy, icon cam)
TTL: Auto
```

**Record 2: Wildcard Subdomain (BẮT BUỘC)**
```
Type: A
Name: *
Content: <IP_SERVER_CỦA_BẠN>
Proxy status: DNS only (QUAN TRỌNG: phải bật proxy!)
TTL: Auto
```


### Option B: Sử Dụng DNS Provider Khác

Với các nhà cung cấp khác (GoDaddy, Namecheap, ...), tạo:

```
Type: A
Host: *
Points to: <IP_SERVER>
TTL: 600
```

### Kiểm Tra DNS

Sau khi cấu hình, test bằng lệnh:

**Windows:**
```powershell
nslookup abc123.yourdomain.com
# Phải trả về IP server của bạn
```

**Linux/Mac:**
```bash
dig abc123.yourdomain.com
# Hoặc
host test.yourdomain.com
```

Kết quả mong đợi: Mọi subdomain đều trỏ về IP server (vd: `103.77.246.206`)

---

## Bước 3: Tạo Wildcard SSL Certificate

### Option A: Let's Encrypt (Miễn phí, tự động renew)
*tự tạo chứng chỉ free tại clf.


## Bước 4: Cấu Hình ProxVN Server

### 4.1. Đặt SSL Cert vào Đúng Path

ProxVN tự động tìm cert ở các vị trí:

**Option 1: Thư mục chạy server (khuyến nghị cho manual SSL)**
```bash
cd /path/to/proxvn
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./wildcard.crt
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./wildcard.key
```

**Option 2: Let's Encrypt path (tự động detect)**
```
/etc/letsencrypt/live/yourdomain.com/fullchain.pem
/etc/letsencrypt/live/yourdomain.com/privkey.pem
```
**Option 3: Manual SSL**
```
etc/bin/svproxvn.exe
```
```
./wildcard.crt
./wildcard.key
```
ở etc và chạy lệnh:
```bash
/bin/svproxvn.exe
```

### 4.2. Set Environment Variables

**Linux (bash/zsh):**
```bash
export HTTP_DOMAIN="yourdomain.com"
export HTTP_PORT=443
```

Hoặc tạo file `.env`:
```bash
# /path/to/proxvn/.env
HTTP_DOMAIN=yourdomain.com
HTTP_PORT=443
```

Load env:
```bash
source .env
./svproxvn
```

**Windows (PowerShell):**
```powershell
$env:HTTP_DOMAIN="yourdomain.com"
$env:HTTP_PORT=443
.\svproxvn.exe
```

Hoặc set permanent:
```powershell
[System.Environment]::SetEnvironmentVariable("HTTP_DOMAIN", "yourdomain.com", "User")
[System.Environment]::SetEnvironmentVariable("HTTP_PORT", "443", "User")
```

### 4.3. Chạy Server

**Linux:**
```bash
sudo ./proxvn-linux-server
```
*(Cần `sudo` để bind port 443)*

**Windows (Run as Administrator):**
```powershell
.\svproxvn.exe
```

**Kiểm tra log:**
```
[http] Certificate: wildcard.crt, Key: wildcard.key
[http] HTTP Domain: *.yourdomain.com
[http] Starting HTTPS proxy server on port 443
```

✅ **Thành công!** Nếu thấy dòng trên, HTTP tunneling đã sẵn sàng.

---

## Bước 5: Sử Dụng HTTP Tunneling

### Client Usage

```bash
# Windows
.\proxvn.exe --proto http 3000

# Linux
./proxvn-linux-client --proto http 3000
```

**Output khi thành công:**
```
✅ HTTP Tunnel Active
🌐 Public URL: https://abc123.yourdomain.com
📍 Forwarding to: localhost:3000
```

### Test

1. **Chạy local HTTP server:**
   ```bash
   python -m http.server 3000
   ```

2. **Truy cập public URL từ browser:**
   ```
   https://abc123.yourdomain.com
   ```

3. **Kết quả:** Thấy nội dung từ `localhost:3000`

---

## Troubleshooting

### ❌ Lỗi: "HTTP_DOMAIN not configured"

**Nguyên nhân:** Chưa set environment variable  
**Giải pháp:**
```bash
export HTTP_DOMAIN="yourdomain.com"
```

**Fallback:** Server tự động disable HTTP mode, clients dùng IP:port

---

### ❌ Lỗi: "Failed to load SSL certificate"

**Nguyên nhân:** Không tìm thấy cert file  
**Giải pháp:**
1. Kiểm tra file tồn tại:
   ```bash
   ls -la wildcard.crt wildcard.key
   ```
2. Copy cert vào thư mục server
3. Hoặc symlink:
   ```bash
   ln -s /etc/letsencrypt/live/yourdomain.com/fullchain.pem wildcard.crt
   ln -s /etc/letsencrypt/live/yourdomain.com/privkey.pem wildcard.key
   ```

---

### ❌ Lỗi: "Tunnel not found for subdomain"

**Nguyên nhân:** DNS chưa propagate hoặc sai config  
**Giải pháp:**
1. Test DNS:
   ```bash
   nslookup abc123.yourdomain.com
   ```
2. Đảm bảo wildcard `*` record tồn tại
3. Chờ DNS propagate (5 phút - 24h)

---

### ❌ Lỗi: "SSL certificate invalid"

**Nguyên nhân:** Cert không phải wildcard hoặc wrong domain  
**Giải pháp:**
1. Verify cert:
   ```bash
   openssl x509 -in wildcard.crt -text -noout | grep "DNS:"
   ```
2. Phải thấy: `DNS:*.yourdomain.com`

---

### ❌ Port 443 đã được dùng

**Nguyên nhân:** Nginx/Apache đang chạy  
**Giải pháp:**

**Option 1: Đổi port HTTP proxy**
```bash
export HTTP_PORT=8443
```

**Option 2: Stop nginx/apache**
```bash
sudo systemctl stop nginx
sudo systemctl stop apache2
```

**Option 3: Reverse proxy qua nginx**
```nginx
server {
    listen 443 ssl;
    server_name *.yourdomain.com;
    
    ssl_certificate /path/to/wildcard.crt;
    ssl_certificate_key /path/to/wildcard.key;
    
    location / {
        proxy_pass https://localhost:8443;
        proxy_set_header Host $host;
    }
}
```

---

## Firewall Rules

Mở port cần thiết:

**Linux (ufw):**
```bash
sudo ufw allow 443/tcp
sudo ufw allow 8882/tcp  # Tunnel control port
sudo ufw reload
```

**Linux (firewalld):**
```bash
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8882/tcp
sudo firewall-cmd --reload
```

**Windows Firewall:**
```powershell
New-NetFirewallRule -DisplayName "ProxVN HTTPS" -Direction Inbound -LocalPort 443 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "ProxVN Tunnel" -Direction Inbound -LocalPort 8882 -Protocol TCP -Action Allow
```

---

## Tóm Tắt Checklist

- [ ] ✅ Mua domain
- [ ] ✅ Cấu hình DNS wildcard (`*.yourdomain.com` → IP server)
- [ ] ✅ Tạo wildcard SSL certificate
- [ ] ✅ Copy cert vào server (`wildcard.crt`, `wildcard.key`)
- [ ] ✅ Set `HTTP_DOMAIN` environment variable
- [ ] ✅ Mở port 443 trong firewall
- [ ] ✅ Chạy server với quyền admin/root
- [ ] ✅ Test với client `--proto http`

---

## Domain Providers Phổ Biến tại Việt Nam

| Nhà cung cấp | Giá (VNĐ/năm) | Ghi chú |
|--------------|---------------|---------|
| **Cloudflare** | 250k - 500k | Khuyến nghị, quản lý DNS tốt |
| **Mat Bao** | 200k - 400k | Hỗ trợ tiếng Việt |
| **iNET.vn** | 180k - 350k | Nhiều khuyến mãi |
| **Namecheap** | $8-12 USD | Quốc tế, dễ dùng |
| **GoDaddy** | $10-15 USD | Phổ biến, UI tiếng Việt |

---

**Tác giả:** TrongDev  
**Phiên bản:** 2.0.0  
**Cập nhật:** 2026-01-10
