# Sửa Lỗi SSL Certificate - Cloudflare Origin Certificate

## 🔴 Vấn Đề Bạn Đang Gặp

Từ server logs, bạn thấy lỗi này:
```
http: TLS handshake error from 162.158.88.130: remote error: tls: unknown certificate authority
```

**Nguyên nhân:** Server đang dùng self-signed certificate (`server.crt`) mà Cloudflare không tin tưởng.

**Các IP này là Cloudflare servers:**
- `162.158.x.x`
- `172.68.x.x`, `172.70.x.x`, `172.71.x.x`
- `104.23.x.x`

---

## ✅ Giải Pháp: Cloudflare Origin Certificate

### Bước 1: Lấy Certificate Từ Cloudflare

1. **Đăng nhập Cloudflare Dashboard:**
   - Truy cập: https://dash.cloudflare.com
   - Chọn domain: `vutrungocrong.fun`

2. **Tạo Origin Certificate:**
   - Menu bên trái → `SSL/TLS` → `Origin Server`
   - Click nút **"Create Certificate"**

3. **Cấu hình Certificate:**
   ```
   ┌────────────────────────────────────────────┐
   │ Generate private key and CSR with CF       │  ← Chọn cái này
   ├────────────────────────────────────────────┤
   │ Hostnames:                                 │
   │   *.vutrungocrong.fun                      │  ← Nhập wildcard
   │   vutrungocrong.fun                        │  ← Nhập root domain
   ├────────────────────────────────────────────┤
   │ Certificate Validity: 15 years             │  ← Để 15 năm
   └────────────────────────────────────────────┘
   ```

4. **Download Certificate:**
   - Cloudflare sẽ hiển thị 2 box:
     - **Origin Certificate** (file `.pem`)
     - **Private Key** (file `.key`)
   - Click **"OK"** để confirm

5. **Lưu certificates:**
   - Copy nội dung **Origin Certificate** → lưu vào file `wildcard.crt`
   - Copy nội dung **Private Key** → lưu vào file `wildcard.key`

### Bước 2: Upload Lên Server

**Trên máy local (nơi bạn vừa tạo file):**

```bash
# Giả sử server IP: 103.77.246.206
# Thay YOUR_SERVER_IP bằng IP thật của bạn

scp wildcard.crt root@103.77.246.206:/root/proxvn_tunnel_bakup/
scp wildcard.key root@103.77.246.206:/root/proxvn_tunnel_bakup/
```

**Hoặc copy thủ công:**
1. Mở file `wildcard.crt` và `wildcard.key` trên máy local
2. SSH vào server
3. Tạo file và paste nội dung:

```bash
cd /path/to/proxvn_tunnel_bakup
nano wildcard.crt  # Paste nội dung Origin Certificate
nano wildcard.key  # Paste nội dung Private Key
chmod 600 wildcard.key  # Bảo mật private key
```

### Bước 3: Cấu Hình Server Dùng Certificate Mới

**Kiểm tra file certificate:**
```bash
ls -lh wildcard.crt wildcard.key
```

Output mong đợi:
```
-rw-r--r-- 1 root root 1.6K Jan 20 15:40 wildcard.crt
-rw------- 1 root root 1.7K Jan 20 15:40 wildcard.key
```

**Restart server:**
```bash
# Dừng server hiện tại (Ctrl+C)
# Hoặc kill process
pkill svproxvn

# Chạy lại server
cd /path/to/proxvn_tunnel_bakup
./svproxvn
```

Server sẽ tự động load `wildcard.crt` và `wildcard.key`.

### Bước 4: Verify Certificate Đã Hoạt Động

**Test 1: Kiểm tra server logs**

Sau khi restart, bạn KHÔNG nên thấy lỗi này nữa:
```
http: TLS handshake error from 162.158.x.x: remote error: tls: unknown certificate authority
```

**Test 2: Test HTTPS connection**

```bash
# Từ máy local
openssl s_client -connect vutrungocrong.fun:443 -servername vutrungocrong.fun

# Kiểm tra output:
# - Verify return code: 0 (ok)  ← Phải là 0
# - NOT "unable to verify"
```

**Test 3: Truy cập qua browser**

1. Tạo file share mới với client
2. Access URL: `https://xxxxx.vutrungocrong.fun`
3. **KHÔNG còn lỗi SSL error 526**
4. File download thành công

---

## 🔧 Cloudflare SSL Settings (Bắt Buộc)

Sau khi cài certificate, kiểm tra SSL mode trên Cloudflare:

1. Cloudflare Dashboard → `SSL/TLS` → `Overview`
2. **SSL/TLS encryption mode** phải là:
   - ✅ **Full (strict)** ← KHUYẾN NGHỊ
   - ⚠️  **Full** ← Cũng OK nhưng kém bảo mật
   - ❌ **Flexible** ← SẼ KHÔNG HOẠT ĐỘNG với Origin Cert

---

## 📝 Troubleshooting

### Vẫn thấy lỗi "unknown certificate authority"?

1. **Kiểm tra certificate files có đúng không:**
   ```bash
   openssl x509 -in wildcard.crt -text -noout | grep "Issuer"
   ```
   Should see: `Issuer: C = US, O = Cloudflare, Inc.`

2. **Kiểm tra private key match với certificate:**
   ```bash
   openssl x509 -noout -modulus -in wildcard.crt | openssl md5
   openssl rsa -noout -modulus -in wildcard.key | openssl md5
   ```
   Hai MD5 hash phải GIỐNG NHAU.

3. **Kiểm tra permissions:**
   ```bash
   chmod 600 wildcard.key
   chmod 644 wildcard.crt
   ```

### Certificate đã đúng nhưng vẫn lỗi?

- Kiểm tra Cloudflare DNS settings
- Đảm bảo Proxy status = **Proxied** (☁️ màu cam)
- Clear Cloudflare cache

---

## 🎯 Kết Quả Mong Đợi

**Server logs sau khi fix:**
```
[http] Loaded certificate from: wildcard.crt
[http] Starting HTTPS proxy server on port 443
[http] Base domain: *.vutrungocrong.fun
```

**KHÔNG còn thấy:**
```
http: TLS handshake error from 162.158.x.x: remote error: tls: unknown certificate authority
```

**Browser:**
- ✅ HTTPS connection thành công
- ✅ Valid certificate (ổ khóa xanh)
- ✅ File sharing hoạt động
- ✅ No error 526
