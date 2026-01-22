package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"proxvn/backend/cmd/fileserver"
)

// runFileShareMode chạy file sharing mode
func runFileShareMode(path, username, password, perms, serverAddr string, insecure bool) error {
	// 1. Validate và expand path
	expandedPath := fileserver.ExpandPath(path)
	if !fileserver.PathExists(expandedPath) {
		return fmt.Errorf("đường dẫn không tồn tại: %s", path)
	}

	if !fileserver.IsDirectory(expandedPath) {
		return fmt.Errorf("đường dẫn phải là thư mục: %s", path)
	}

	normalizedPath := fileserver.NormalizePath(expandedPath)
	log.Printf("[FileShare] Chia sẻ thư mục: %s", normalizedPath)

	// 2. Parse permissions
	permissions := fileserver.ParsePermissions(perms)
	log.Printf("[FileShare] Quyền hạn: %s (%s)", permissions, permissions.Description())

	// 3. Tạo WebDAV server
	// Fix: Add /webdav prefix so generated XML hrefs are correct for clients like Nautilus
	webdavServer, err := fileserver.NewWebDAVServer(normalizedPath, "/webdav", username, password, permissions)
	if err != nil {
		return fmt.Errorf("không thể tạo WebDAV server: %v", err)
	}

	// 4. Find free port for local server
	localPort, err := findFreePort()
	if err != nil {
		return fmt.Errorf("không tìm được port trống: %v", err)
	}

	// 5. Setup HTTP handlers
	mux := http.NewServeMux()

	// WebDAV endpoint (for mounting as network drive)
	// NOTE: Registering both /webdav and /webdav/ helps avoid redirects that might confuse some clients
	mux.Handle("/webdav", webdavServer)
	mux.Handle("/webdav/", webdavServer)

	// Web UI endpoints
	sessionMgr := webdavServer.GetSessionManager()
	uiHandler := fileserver.NewUIHandler(sessionMgr)
	mux.Handle("/", uiHandler)

	// Start local HTTP server
	localAddr := fmt.Sprintf("localhost:%d", localPort)
	server := &http.Server{
		Addr:    localAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("[FileShare] Đang khởi động local file server trên port %d...", localPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FileShare] Local server lỗi: %v", err)
		}
	}()

	log.Printf("[FileShare] ✅ Local file server đã khởi động")

	// 6. Get client ID
	hostname, _ := os.Hostname()
	clientID := fmt.Sprintf("fileshare-%s", hostname)

	// 7. Create tunnel client
	// Use HTTP protocol to leverage HTTPS proxy (port 443) which has valid SSL certificate
	// Only fall back to TCP if server doesn't support HTTP proxy
	cl := &client{
		serverAddr:         serverAddr,
		localAddr:          localAddr,
		clientID:           clientID,
		protocol:           "http",   // Use HTTP for SSL certificate support
		uiEnabled:          false,    // Disable normal TUI for file sharing
		insecureSkipVerify: insecure, // Skip TLS verification if --insecure
	}

	// 8. Connect tunnel with reconnection loop
	backoff := 3 * time.Second
	maxBackoff := 5 * time.Minute

	for {
		log.Println("[FileShare] 🚀 Đang tạo tunnel...")
		if err := cl.connectControl(); err != nil {
			log.Printf("[FileShare] ❌ Kết nối tunnel thất bại: %v", err)
			log.Printf("[FileShare] Retry sau %v...", backoff)
			time.Sleep(backoff)

			// Exponential backoff
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// Reset backoff on success
		backoff = 3 * time.Second

		// 9. Display access information
		displayFileShareInfo(cl.subdomain, cl.baseDomain, cl.publicHost, username, password, permissions, normalizedPath)

		// 10. Keep tunnel running
		if err := cl.receiveLoop(); err != nil {
			log.Printf("[FileShare] ❌ Tunnel lỗi: %v", err)
		}

		cl.closeControl()
		log.Printf("[FileShare] Thử kết nối lại tunnel sau %v...", backoff)
		time.Sleep(backoff)
	}
}

// findFreePort tìm port trống
func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// displayFileShareInfo hiển thị thông tin file share
func displayFileShareInfo(subdomain, baseDomain, publicHost, username, password string, perms fileserver.Permissions, path string) {
	var webdavURL, webUIURL string

	// Khi dùng HTTP protocol, server sẽ assign subdomain
	// Files sẽ được serve qua HTTPS proxy (port 443) với SSL certificate hợp lệ
	if baseDomain != "" && subdomain != "" {
		// HTTP mode with valid SSL certificate
		webdavURL = fmt.Sprintf("https://%s.%s/webdav", subdomain, baseDomain)
		webUIURL = fmt.Sprintf("https://%s.%s", subdomain, baseDomain)
	} else {
		// TCP mode: Dùng IP:port từ publicHost khi server không có domain
		webdavURL = fmt.Sprintf("http://%s/webdav", publicHost)
		webUIURL = fmt.Sprintf("http://%s", publicHost)
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ FILE SHARING ĐANG HOẠT ĐỘNG                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("📁 Thư mục chia sẻ: %s\n", path)
	fmt.Printf("👤 Username:        %s\n", username)
	fmt.Printf("🔒 Mật khẩu:        %s\n", password)
	fmt.Printf("⚙️  Quyền hạn:       %s (%s)\n", perms, perms.Description())
	fmt.Printf("🌐 URL WebDAV:      %s\n", webdavURL)
	fmt.Printf("🖥️  URL Web UI:      %s\n\n", webUIURL)

	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("📌 CÁCH SỬ DỤNG:")
	fmt.Println("═══════════════════════════════════════════════════════════════════")

	fmt.Println("🪟 WINDOWS - Map Network Drive:")
	fmt.Println("   1. Mở 'This PC' (File Explorer)")
	fmt.Println("   2. Click 'Computer' → 'Map network drive'")
	fmt.Println("	 3. Chọn ổ đĩa (vd: Z:)")
	fmt.Printf("	 4. Folder: %s\n", webdavURL)
	fmt.Println("	 5. Click 'Finish'")
	fmt.Printf("	 6. Username: %s\n", username)
	fmt.Printf("	 7. Password: %s\n", password)
	fmt.Println("	 8. ✅ Ổ đĩa Z:\\ sẽ xuất hiện trong 'This PC'")

	fmt.Println("	 HOẶC dùng Command Line:")
	fmt.Printf("	 net use Z: %s /user:%s %s\n\n", webdavURL, username, password)

	fmt.Println("🐧 LINUX - Mount WebDAV:")
	fmt.Println("   # Cài đặt davfs2 (chỉ cần 1 lần)")
	fmt.Println("   sudo apt install davfs2        # Ubuntu/Debian")
	fmt.Println("   sudo yum install davfs2        # CentOS/RHEL")
	fmt.Println("")
	fmt.Println("	 # Mount folder")
	fmt.Println("	 sudo mkdir -p /mnt/proxvn_share")
	fmt.Printf("	 sudo mount -t davfs %s /mnt/proxvn_share\n", webdavURL)
	fmt.Printf("	 # Khi hỏi username: %s\n", username)
	fmt.Printf("	 # Khi hỏi password: %s\n\n", password)

	fmt.Println("🍎 macOS - Connect to Server:")
	fmt.Println("   1. Mở Finder")
	fmt.Println("   2. Nhấn Cmd+K (hoặc Go → Connect to Server)")
	fmt.Printf("	 3. Server Address: %s\n", webdavURL)
	fmt.Println("	 4. Click 'Connect'")
	fmt.Println("	 5. Chọn 'Registered User'")
	fmt.Printf("	 6. Name: %s\n", username)
	fmt.Printf("	 7. Password: %s\n", password)
	fmt.Println("	 8. ✅ Ổ đĩa sẽ xuất hiện trong Finder sidebar")

	fmt.Println("🌐 TRÌNH DUYỆT WEB (Mọi hệ điều hành):")
	fmt.Printf("   1. Mở trình duyệt, truy cập: %s\n", webUIURL)
	fmt.Printf("   2. Nhập password: %s\n", password)
	fmt.Println("   3. Quản lý file qua giao diện web đẹp mắt")

	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("💡 GỢI Ý:")
	fmt.Println("   • Giữ cửa sổ này mở để duy trì kết nối")
	fmt.Println("   • Nhấn Ctrl+C để ngắt kết nối và dừng chia sẻ")
	fmt.Println("   • Không chia sẻ password với người không tin tưởng")

	if perms == fileserver.PermRead {
		fmt.Println("   • ⚠️  Chế độ chỉ đọc: Người khác chỉ xem và download")
	} else if perms == fileserver.PermReadWrite {
		fmt.Println("   • ✏️  Chế độ đọc-ghi: Người khác có thể upload, xóa, sửa file")
	}

	fmt.Println("═══════════════════════════════════════════════════════════════════")
}
