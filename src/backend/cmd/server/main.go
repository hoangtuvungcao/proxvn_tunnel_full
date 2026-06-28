package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"runtime"
	"runtime/debug"
	"time"

	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"proxvn/backend/internal/api"
	"proxvn/backend/internal/auth"
	"proxvn/backend/internal/config"
	"proxvn/backend/internal/database"
	httpproxy "proxvn/backend/internal/http"
	"proxvn/backend/internal/middleware"
	"proxvn/backend/internal/models"
	"proxvn/backend/internal/pool"
	"proxvn/backend/internal/tunnel"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

const (
	defaultListenPort  = 8881
	publicPortStart    = 10000
	publicPortEnd      = 20000
	heartbeatInterval  = 5 * time.Second
	clientIdleTimeout  = 30 * time.Second // Faster cleanup (was 60)
	udpControlInterval = 2 * time.Second
	udpControlTimeout  = 6 * time.Second
	backendIdleTimeout = 5 * time.Second
	backendIdleRetries = 3

	// Performance limits
	maxConnections = 10000
	bufferSize     = 32768 // 32KB buffers
)

const (
	udpMsgHandshake byte = 1
	udpMsgData      byte = 2
	udpMsgClose     byte = 3
	udpMsgPing      byte = 4
	udpMsgPong      byte = 5
)

// debugServerUDP enables verbose logging of the inbound UDP data path.
const debugServerUDP = false

type server struct {
	listenPort int
	publicHost string // Public host/IP advertised to clients for TCP/UDP tunnels
	clients    map[string]*clientSession
	clientsMu  sync.RWMutex

	// Port pool management
	availablePorts []int
	usedPorts      map[int]bool
	portMu         sync.Mutex

	// Reservation store for reconnecting clients (SQLite or Memory)
	reservations ReservationStore
	listener     net.Listener

	udpServer    *net.UDPConn
	udpMu        sync.Mutex
	udpSessions  map[string]*udpServerSession
	// Last known UDP control address of each client, keyed by client key.
	// Needed so the server can push inbound (public→client) UDP packets even
	// before the client has sent any data for a given session.
	udpClientAddrs map[string]*net.UDPAddr
	httpServer   *http.Server
	proxyWaiting map[string]chan net.Conn
	proxyMu      sync.Mutex
	httpProxy    *httpproxy.HTTPProxyServer
	httpRequests map[string]chan *httpproxy.HTTPResponse
	httpReqMu    sync.Mutex

	// Rate limiting
	rateLimiters   map[string]*rateLimiter
	rateLimitersMu sync.Mutex

	// Connection limiting
	connSemaphore chan struct{}

	runtimeStart      time.Time
	totalConnections  uint64
	activeConnections int64
	totalBytesUp      uint64
	totalBytesDown    uint64

	throttleMu    sync.Mutex
	throttledLogs map[string]time.Time
}

type rateLimiter struct {
	registrations *rate.Limiter
	httpRequests  *rate.Limiter
	udpSessions   *rate.Limiter
	lastSeen      time.Time
}

type clientSession struct {
	server         *server // Reference to parent server for HTTP response handling
	conn           net.Conn
	enc            *jsonWriter
	dec            *jsonReader
	clientID       string
	key            string
	target         string
	protocol       string
	publicPort     int
	subdomain      string // For HTTP tunneling
	generation     int64  // Client generation ID to prevent ghost sessions
	state          *tunnel.StateMachine
	lastSeen       time.Time
	closeOnce      sync.Once
	done           chan struct{}
	mu             sync.Mutex
	bytesUp        uint64
	bytesDown      uint64
	remoteIP       string
	udpSecret      []byte // Key for UDP encryption
	publicListener net.Listener
	publicUDPConn  *net.UDPConn // Public UDP listener for inbound (proto=udp) tunnels

	activeConnections int64
	totalConnections  uint64
}

type udpServerSession struct {
	id         string
	clientKey  string
	udpSecret  []byte // Key for UDP encryption
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
	clientAddr *net.UDPAddr // Client's public UDP address
	closeOnce  sync.Once
	closed     chan struct{}
	timer      *time.Timer
	idleCount  int

	// Inbound (public → client) session fields. When inbound is true the
	// session bridges an external UDP peer (extAddr) on the tunnel's public
	// port (publicConn) to the client; conn/remoteAddr are unused.
	inbound    bool
	publicConn *net.UDPConn
	extAddr    *net.UDPAddr
	lastActive time.Time
}

type jsonWriter struct {
	enc *json.Encoder
	mu  sync.Mutex
}

type jsonReader struct {
	dec *json.Decoder
	mu  sync.Mutex
}

func (w *jsonWriter) Encode(msg tunnel.Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.enc.Encode(msg)
}

func (r *jsonReader) Decode(msg *tunnel.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.dec.Decode(msg)
}

func main() {
	// Custom usage message with setup guide
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
╔════════════════════════════════════════════════════════════════════════════╗
║                 ProxVN v%s - Server                                   ║
║          Tunnel Server - Hỗ trợ TCP, UDP và HTTP Tunneling                 ║
╚════════════════════════════════════════════════════════════════════════════╝

🌟 TÍNH NĂNG SERVER:
  • TCP/UDP Tunneling:  Hỗ trợ tunnel protocols truyền thống
  • HTTP Tunneling:     Cấp subdomain HTTPS tự động cho clients
  • Web Dashboard:      Quản lý clients qua giao diện web
  • Auto SSL:           Tự động load SSL cert từ nhiều nguồn
  • Cross-Platform:     Windows & Linux server support

📖 CÚ PHÁP:
  svproxvn [OPTIONS]

⚙️  CÁC THAM SỐ:
`, tunnel.Version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
💡 CÁCH SỬ DỤNG:

▶ Chạy server cơ bản (TCP/UDP only):
  svproxvn
  svproxvn -port 8881

▶ Chạy server với HTTP Tunneling (cần SSL cert):
  # Linux
  export HTTP_DOMAIN="yourdomain.com"
  ./svproxvn

  # Windows
  set HTTP_DOMAIN=yourdomain.com
  svproxvn.exe

🔧 CẤU HÌNH HTTP TUNNELING:

1️⃣  Chuẩn bị Domain & SSL Certificate:
  
  Cách 1: Dùng Cloudflare Origin Certificate (Khuyến nghị)
    • Vào Cloudflare Dashboard → SSL/TLS → Origin Server
    • Tạo Origin Certificate
    • Lưu file: wildcard.crt và wildcard.key
    • Đặt 2 file vào cùng thư mục với svproxvn

  Cách 2: Dùng Let's Encrypt
    sudo apt install python3-certbot-dns-cloudflare
    sudo certbot certonly --dns-cloudflare \
      --dns-cloudflare-credentials /root/.secrets/cloudflare.ini \
      -d '*.yourdomain.com' -d 'yourdomain.com'
    
    # Copy cert
    sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem wildcard.crt
    sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem wildcard.key

2️⃣  Cấu hình DNS trên Cloudflare:
  
  Tạo 2 bản ghi DNS:
  ┌──────┬──────┬─────────────────┬──────────────┐
  │ Type │ Name │ Content         │ Proxy Status │
  ├──────┼──────┼─────────────────┼──────────────┤
  │ A    │ @    │ YOUR_VPS_IP     │ 🟠 Proxied  │
  │ CNAME│ *    │ yourdomain.com  │ 🟠 Proxied  │
  └──────┴──────┴─────────────────┴──────────────┘
  
  ⚠️  QUAN TRỌNG: Phải bật Cloudflare Proxy (đám mây màu cam)!

3️⃣  Cấu hình SSL Mode:
  
  Cloudflare Dashboard → SSL/TLS → Overview
  Chọn: Full (strict)

4️⃣  Mở Firewall (nếu cần):
  
  # Linux (ufw)
  sudo ufw allow 8881/tcp  # Dashboard
  sudo ufw allow 8882/tcp  # Tunnel
  sudo ufw allow 443/tcp   # HTTPS (HTTP Tunneling)
  
  # Windows: Mở Windows Firewall → Inbound Rules → New Rule

🌐 TRUY CẬP DASHBOARD:
  http://localhost:8881/dashboard/
  http://YOUR_VPS_IP:8881/dashboard/

📊 PORTS:
  • Dashboard/API: 8881 (hoặc port bạn chọn)
  • Tunnel:        8882 (Dashboard Port + 1)
  • HTTPS Proxy:   443  (nếu bật HTTP Tunneling)

🔗 THÔNG TIN:
  • Website:        https://bacsycay.click
  • Documentation:  https://github.com/proxvn/docs
  • Setup Guide:    DOMAIN_SETUP.md

© 2026 ProxVN - Developed by TrongDev
Licensed under FREE TO USE - NON-COMMERCIAL ONLY

`)
	}

	portFlag := flag.Int("port", defaultListenPort, "Port cho Dashboard & API (Tunnel port = Port + 1)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// CPU/RAM Optimizations
	runtime.GOMAXPROCS(runtime.NumCPU())    // Use all CPUs
	debug.SetGCPercent(50)                  // Aggressive GC for low RAM
	debug.SetMemoryLimit(400 * 1024 * 1024) // 400MB soft limit

	// Load configuration (optional)
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[server] Using defaults: %v", err)
		cfg = &config.Config{
			Server: config.ServerConfig{Port: *portFlag},
		}
	}

	// ✅ FIX: Initialize database OUTSIDE goroutine với proper cleanup
	var db *database.Database
	dbDSN := cfg.GetDatabaseDSN()
	if dbDSN != "" {
		db, err = database.NewDatabase(dbDSN)
		if err != nil {
			log.Printf("[database] Failed to init: %v (running without database)", err)
		} else {
			defer db.Close() // ✅ This WILL run when main() exits
			log.Printf("[database] SQLite3 initialized successfully")

			// Seed default admin if no users exist, using configured credentials
			users, _ := db.GetAllUsers()
			if len(users) == 0 {
				adminUser := cfg.Auth.AdminUsername
				if adminUser == "" {
					adminUser = "admin"
				}
				adminPass := cfg.Auth.AdminPassword
				if adminPass == "" {
					adminPass = "admin123"
				}
				log.Printf("[database] No users found. Creating default admin '%s'...", adminUser)
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
				admin := &models.User{
					ID:       uuid.New(),
					Username: adminUser,
					Email:    "admin@proxvn.com",
					Password: string(hashedPassword),
					Role:     models.UserRoleAdmin,
					APIKey:   uuid.New().String(),
				}
				if err := db.CreateUser(admin); err != nil {
					log.Printf("[database] Failed to create default admin: %v", err)
				} else if adminPass == "admin123" {
					log.Printf("[database] ⚠️  Created default admin '%s' with INSECURE default password. Set ADMIN_PASSWORD!", adminUser)
				} else {
					log.Printf("[database] ✅ Created admin '%s' from ADMIN_PASSWORD", adminUser)
				}
			}
		}
	}

	var resStore ReservationStore
	if db != nil {
		resStore = NewSQLiteReservationStore(db)
		log.Printf("[server] Initialized SQLite Reservation Store")
	} else {
		resStore = NewMemoryReservationStore()
		log.Printf("[server] Initialized Memory Reservation Store (stateless)")
	}

	srv := &server{
		listenPort:     *portFlag,
		publicHost:     resolvePublicHost(cfg.Server.PublicHost),
		clients:        make(map[string]*clientSession),
		availablePorts: make([]int, 0, publicPortEnd-publicPortStart+1),
		usedPorts:      make(map[int]bool),
		reservations:   resStore,
		udpSessions:    make(map[string]*udpServerSession),
		udpClientAddrs: make(map[string]*net.UDPAddr),
		proxyWaiting:   make(map[string]chan net.Conn),
		httpRequests:   make(map[string]chan *httpproxy.HTTPResponse),
		rateLimiters:   make(map[string]*rateLimiter),
		connSemaphore:  make(chan struct{}, maxConnections),
	}

	srv.runtimeStart = time.Now()
	srv.throttledLogs = make(map[string]time.Time)

	// Initialize port pool
	for port := publicPortStart; port <= publicPortEnd; port++ {
		srv.availablePorts = append(srv.availablePorts, port)
	}

	// Start rate limiter cleanup goroutine
	go srv.cleanupRateLimiters()

	// Start reservation cleanup ticker
	startReservationCleanupTicker(resStore, srv)

	// Start HTTP/API/Dashboard server (✅ pass db as param)
	go srv.startHTTPServer(cfg, db)

	// Initialize HTTP proxy for HTTP tunneling (if SSL cert available)
	// Landing page will be served on main domain (bacsycay.click) via HTTPS
	go srv.initHTTPProxy(cfg)

	// Catch OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		srv.Shutdown(db)
	}()

	// Run tunnel server
	if err := srv.run(); err != nil {
		log.Fatalf("[server] fatal error: %v", err)
	}
}

// startHTTPServer starts the HTTP server with API and dashboard
// NOTE: db is passed as parameter to avoid leak (defer in goroutine never runs)
func (s *server) startHTTPServer(cfg *config.Config, db *database.Database) {
	// Initialize handlers if database available
	var handlers *api.Handler
	var authService *auth.AuthService

	if db != nil {
		authService = auth.NewAuthService(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)
		handlers = api.NewHandler(db, authService)
		log.Printf("[api] Database connected")
	} else {
		log.Printf("[api] No database (tunnel-only mode)")
	}

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.New()
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestSizeLimitMiddleware(10)) // 10MB Limit
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.TimestampValidationMiddleware())
	router.Use(middleware.GzipAndCacheMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"server":  "ProxVN by TrongDev",
			"version": "7.5.0",
		})
	})

	// Serve web dashboard
	dashboardDir := "./frontend"
	if _, err := os.Stat("frontend"); err == nil {
		dashboardDir = "./frontend"
	}

	log.Printf("[http] Serving dashboard from: %s", dashboardDir)
	router.Static("/dashboard", dashboardDir)

	// Serve static files for Landing Page
	landingDir := filepath.Join(dashboardDir, "landing")
	router.Static("/assets", landingDir) // Helper for assets if needed, but we use root

	// Serve Landing Page files directly at root to work with relative links
	router.StaticFile("/", filepath.Join(landingDir, "index.html"))

	// Force correct MIME types for CSS/JS
	router.GET("/style.css", func(c *gin.Context) {
		c.Header("Content-Type", "text/css")
		c.File(filepath.Join(landingDir, "style.css"))
	})

	router.GET("/script.js", func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript")
		c.File(filepath.Join(landingDir, "script.js"))
	})

	// Serve Downloads (Map virtual paths to actual binaries)
	// We assume 'bin' is in CWD or parent
	binDir := "bin"
	if _, err := os.Stat("bin"); os.IsNotExist(err) {
		binDir = "." // If running inside bin
	}

	router.StaticFile("/downloads/proxvn-windows.zip", filepath.Join(binDir, "proxvn.exe")) // Map zip to exe for now or just generic
	router.StaticFile("/downloads/proxvn.exe", filepath.Join(binDir, "proxvn.exe"))
	router.StaticFile("/downloads/proxvn-linux-client", filepath.Join(binDir, "proxvn-linux-client"))
	router.StaticFile("/downloads/proxvn-mac-intel", filepath.Join(binDir, "proxvn-mac-intel"))
	router.StaticFile("/downloads/proxvn-mac-m1", filepath.Join(binDir, "proxvn-mac-m1"))
	router.StaticFile("/downloads/proxvn-android", filepath.Join(binDir, "proxvn-android"))
	router.StaticFile("/downloads/proxvn-linux-server", filepath.Join(binDir, "proxvn-linux-server"))
	router.StaticFile("/downloads/svproxvn.exe", filepath.Join(binDir, "svproxvn.exe"))

	// Explicitly redirect /dashboard/ to /dashboard/index.html if needed,
	// or ensure main route hits it.

	// Simple metrics endpoint removed - now handled by handlers.GetMetrics() if DB available

	// Simple tunnels list endpoint removed - now handled by handlers.GetAllTunnels() if DB available

	// Public WebSocket endpoint for dashboard (no auth required)
	router.GET("/api/v1/dashboard/ws", func(c *gin.Context) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://localhost") {
					return true
				}
				httpDomain := os.Getenv("HTTP_DOMAIN")
				if httpDomain != "" {
					cleanDomain := strings.TrimPrefix(httpDomain, ".")
					if strings.HasSuffix(origin, cleanDomain) {
						return true
					}
				}
				return false
			},
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send initial update
		if err := s.sendDashboardUpdate(conn); err != nil {
			return
		}

		// Stream updates every 2 seconds
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.sendDashboardUpdate(conn); err != nil {
				return
			}
		}
	})

	// Old Unprotected WS route removed

	// Prometheus metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		var out strings.Builder
		s.clientsMu.RLock()
		activeTunnels := len(s.clients)
		s.clientsMu.RUnlock()

		out.WriteString(fmt.Sprintf("# HELP active_tunnels Number of active tunnels\n# TYPE active_tunnels gauge\nactive_tunnels %d\n", activeTunnels))
		out.WriteString(fmt.Sprintf("# HELP active_connections Number of active connections\n# TYPE active_connections gauge\nactive_connections %d\n", atomic.LoadInt64(&s.activeConnections)))
		out.WriteString(fmt.Sprintf("# HELP total_connections Total connections routed\n# TYPE total_connections counter\ntotal_connections %d\n", atomic.LoadUint64(&s.totalConnections)))
		out.WriteString(fmt.Sprintf("# HELP total_bytes_up Total bytes uploaded\n# TYPE total_bytes_up counter\ntotal_bytes_up %d\n", atomic.LoadUint64(&s.totalBytesUp)))
		out.WriteString(fmt.Sprintf("# HELP total_bytes_down Total bytes downloaded\n# TYPE total_bytes_down counter\ntotal_bytes_down %d\n", atomic.LoadUint64(&s.totalBytesDown)))
		c.String(http.StatusOK, out.String())
	})

	// API routes (if database available)
	if handlers != nil {
		apiRouter := router.Group("/api")
		{
			// Public endpoints
			apiRouter.POST("/auth/login", middleware.RateLimitMiddleware(3, 5), middleware.BruteForceProtectionMiddleware(), handlers.Login)
			apiRouter.POST("/auth/register", middleware.RateLimitMiddleware(3, 5), handlers.Register)
			apiRouter.GET("/metrics", handlers.GetMetrics)
			apiRouter.GET("/health", handlers.Health)

			// Protected user endpoints
			protected := apiRouter.Group("")
			protected.Use(middleware.AuthMiddleware(authService))
			{
				protected.GET("/profile", handlers.GetProfile)
				protected.GET("/tunnels", handlers.GetTunnels)
				protected.GET("/ws", handlers.HandleWebSocket)
			}

			// Admin-only endpoints
			admin := apiRouter.Group("/admin")
			admin.Use(middleware.AuthMiddleware(authService))
			admin.Use(middleware.AdminMiddleware())
			{
				// User Management
				admin.GET("/users", handlers.GetAllUsers)
				admin.POST("/users", handlers.CreateUserByAdmin)
				admin.DELETE("/users/:id", handlers.DeleteUser)

				// Tunnel Management
				admin.GET("/tunnels", handlers.GetAllTunnels)
				admin.DELETE("/tunnels/:id", handlers.DeleteTunnelByAdmin)

				// System Stats
				admin.GET("/stats", handlers.GetSystemStats)
			}
		}
	}

	// Start HTTP server
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", s.listenPort),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 8192, // 8KB max header size
	}

	log.Printf("[http] Starting on port %d", s.listenPort)
	log.Printf("[http] Dashboard: http://localhost:%d/dashboard/", s.listenPort)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("[http] Server error: %v", err)
	}
}

func (s *server) run() error {
	// Start UDP server on tunnel port
	tunnelPort := s.listenPort + 1 // 8882 for tunnel control
	if err := s.startUDPServer(tunnelPort); err != nil {
		log.Printf("[tunnel] Failed to start UDP server: %v", err)
	}

	// Start TCP tunnel server on separate port (8882)
	// Enable TLS
	certFile := "server.crt"
	keyFile := "server.key"
	if err := generateSelfSignedCert(certFile, keyFile); err != nil {
		log.Printf("[server] Failed to generate certs: %v, falling back to plain TCP (NOT SECURE)", err)
		// Fallback code (optional, but better to fail securely)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load key pair: %w", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", tunnelPort), tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to listen on tunnel port %d: %w", tunnelPort, err)
	}
	s.listener = listener
	defer listener.Close()

	log.Printf("[tunnel] Tunnel server listening on port %d (TLS Enabled)", tunnelPort)
	log.Printf("[tunnel] Client should connect to: localhost:%d", tunnelPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[server] accept error: %v", err)
			continue
		}

		// Connection limiting with semaphore
		select {
		case s.connSemaphore <- struct{}{}:
			// Slot available, handle connection
			go func() {
				defer func() { <-s.connSemaphore }()
				s.handleConnection(conn)
			}()
		default:
			// No slots available, reject connection
			conn.Close()
			log.Printf("[server] ⚠️  Connection limit reached, rejected new connection")
		}
	}
}

func generateSelfSignedCert(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil // Files exist
		}
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ProxVN Tunnel"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	log.Printf("[server] Generated self-signed certificate: %s, %s", certFile, keyFile)
	return nil
}

func (s *server) startUDPServer(port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	s.udpServer = conn
	_ = conn.SetReadBuffer(4 * 1024 * 1024)
	_ = conn.SetWriteBuffer(4 * 1024 * 1024)

	go s.readUDPControl()
	log.Printf("[tunnel] UDP server listening on port %d", port)
	return nil
}

func (s *server) handleConnection(conn net.Conn) {
	// Don't close immediately here, responsibility passed to handlers

	br := bufio.NewReader(conn)
	// Peek to see if it's empty or closed
	if _, err := br.Peek(1); err != nil {
		if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
			_ = err
		}
		conn.Close()
		return
	}

	dec := tunnel.NewDecoder(br)

	var msg tunnel.Message
	if err := dec.Decode(&msg); err != nil {
		if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
			log.Printf("[server] failed to decode handshake: %v", err)
		}
		conn.Close()
		return
	}

	if msg.Type == "register" {
		// New client session
		session := &clientSession{
			server:   s, // Set parent server reference for HTTP response handling
			conn:     conn,
			enc:      &jsonWriter{enc: tunnel.NewEncoder(conn)},
			dec:      &jsonReader{dec: dec}, // Pass the decoder with existing buffer state
			state:    tunnel.NewStateMachine(),
			lastSeen: time.Now(),
			done:     make(chan struct{}),
		}

		// Capture client IP (strip port)
		host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		session.remoteIP = host

		if err := s.handleClient(session, msg); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				log.Printf("[server] client error: %v", err)
			}
			session.conn.Close()
		}
		s.removeClient(session)
		return
	}

	if msg.Type == "proxy" {
		// Proxy data connection from Client
		s.dispatchProxyConnection(conn, msg.ID)
		return
	}

	log.Printf("[server] unknown handshake type: %s", msg.Type)
	conn.Close()
}

func (s *server) handleClient(session *clientSession, msg tunnel.Message) error {
	// Check rate limit for registration
	if !s.checkRegistrationRateLimit(session.remoteIP) {
		return fmt.Errorf("rate limit exceeded for registration from %s", session.remoteIP)
	}

	if msg.Type != "register" {
		return fmt.Errorf("expected register message, got: %s", msg.Type)
	}

	_ = session.state.TransitionTo(tunnel.StateConnecting)
	_ = session.state.TransitionTo(tunnel.StateTLSHandshake)
	_ = session.state.TransitionTo(tunnel.StateAuthenticating)

	clientID := strings.TrimSpace(msg.ClientID)
	if clientID == "" {
		if strings.TrimSpace(msg.Key) != "" {
			clientID = fmt.Sprintf("client-%s", strings.TrimSpace(msg.Key)[:8])
		} else {
			clientID = "client-unknown"
		}
	}

	// Force disconnect old session of same clientID if exists and newer generation ID is received
	s.clientsMu.Lock()
	oldSession, exists := s.clients[clientID]
	s.clientsMu.Unlock()

	if exists && oldSession != nil {
		if msg.Generation <= oldSession.generation {
			log.Printf("[server] Rejecting registration from client %s with older/same generation %d <= %d", clientID, msg.Generation, oldSession.generation)
			return fmt.Errorf("stale generation ID %d", msg.Generation)
		}

		log.Printf("[server] Client ID %s is reconnecting with newer generation %d (closing old session generation %d)", clientID, msg.Generation, oldSession.generation)

		s.unregisterHTTPClient(oldSession)
		_ = oldSession.state.TransitionTo(tunnel.StateClosed)
		oldSession.Close()

		if oldSession.publicPort > 0 {
			_ = s.reservations.ReservePort(oldSession.key, oldSession.publicPort, 5*time.Minute)
			s.releasePort(oldSession.publicPort)
		}
		if oldSession.protocol == "http" && oldSession.subdomain != "" {
			_ = s.reservations.ReserveSubdomain(oldSession.key, oldSession.subdomain, 5*time.Minute)
		}

		time.Sleep(50 * time.Millisecond)
	}

	_ = session.state.TransitionTo(tunnel.StateRegistering)

	// Generate key if not provided
	key := strings.TrimSpace(msg.Key)
	if key == "" {
		var err error
		key, err = tunnel.GenerateID()
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}
	}

	// Assign public port (try to honor requested port for reconnecting clients)
	publicPort := s.getNextPublicPort(key, msg.RequestedPort)

	session.clientID = clientID
	session.key = key
	session.target = msg.Target
	session.protocol = strings.ToLower(strings.TrimSpace(msg.Protocol))
	if session.protocol == "" {
		session.protocol = "tcp"
	}
	session.publicPort = publicPort
	session.generation = msg.Generation

	// Register client
	s.addClient(session)

	// For HTTP protocol, assign subdomain
	var baseDomain string
	if session.protocol == "http" {
		if err := s.registerHTTPClient(session); err != nil {
			log.Printf("[server] Failed to register HTTP client: %v", err)
			return fmt.Errorf("HTTP tunneling unavailable: %w", err)
		}
		if s.httpProxy != nil {
			baseDomain = s.httpProxy.GetBaseDomain()
		}
	}

	// Generate UDP encryption key
	udpSecret, err := tunnel.GenerateKey()
	if err != nil {
		log.Printf("[server] Failed to generate UDP key: %v", err)
		// Fallback to plain text if key generation fails (should not happen)
	} else {
		session.udpSecret = udpSecret
	}

	// Send registration response
	resp := tunnel.Message{
		Type:       "registered",
		Key:        key,
		ClientID:   session.clientID,
		RemotePort: publicPort,
		Protocol:   session.protocol,
		Version:    tunnel.Version,
		Subdomain:  session.subdomain, // Include subdomain for HTTP mode
		BaseDomain: baseDomain,
	}

	if udpSecret != nil {
		resp.UDPSecret = base64.StdEncoding.EncodeToString(udpSecret)
	}

	if err := session.enc.Encode(resp); err != nil {
		return fmt.Errorf("failed to send registration response: %w", err)
	}

	if session.protocol == "http" {
		logDomain := baseDomain
		if logDomain == "" {
			logDomain = "bacsycay.click"
		}
		log.Printf("[server] client %s registered, HTTP mode, subdomain: %s.%s, target %s",
			session.clientID, session.subdomain, logDomain, session.target)
	} else {
		log.Printf("[server] client %s registered, public port %d, protocol %s, target %s",
			session.clientID, publicPort, session.protocol, session.target)
	}

	_ = session.state.TransitionTo(tunnel.StateActive)
	// Start heartbeat checker
	go s.heartbeatChecker(session)

	atomic.AddInt64(&s.activeConnections, 1)
	s.logOnce(fmt.Sprintf("[server] active sessions: %d", atomic.LoadInt64(&s.activeConnections)), "active_sessions")
	defer atomic.AddInt64(&s.activeConnections, -1)

	// Start public listener for TCP / UDP
	if session.protocol == "tcp" {
		go s.startPublicListener(session)
	} else if session.protocol == "udp" {
		go s.startPublicUDPListener(session)
	}

	// Handle control messages
	return s.controlLoop(session)
}

func (s *server) controlLoop(session *clientSession) error {
	for {
		msg := tunnel.Message{}
		if err := session.dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}

		session.mu.Lock()
		session.lastSeen = time.Now()
		session.mu.Unlock()

		switch msg.Type {
		case "ping":
			if err := session.enc.Encode(tunnel.Message{Type: "pong"}); err != nil {
				return err
			}
		case "proxy":
			go s.handleProxyRequest(session, msg.ID)
		case "udp_open":
			go s.handleUDPOpen(session, msg)
		case "udp_close":
			s.handleUDPClose(msg.ID)
		case "udp_idle":
			s.handleUDPClose(msg.ID)
		case "proxy_error":
			// Client failed to connect to local target
			s.cancelProxyConnection(msg.ID)
		case "http_response":
			// Handle HTTP response from client
			go s.handleHTTPResponse(msg)
		default:
			log.Printf("[server] unknown message type: %s", msg.Type)
		}
	}
}

func (s *server) startPublicListener(session *clientSession) {
	listenAddr := fmt.Sprintf(":%d", session.publicPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("[server] failed to listen on public port %d: %v", session.publicPort, err)
		return
	}

	// Store listener safely
	session.mu.Lock()
	select {
	case <-session.done:
		// Session closed while we were setting up
		session.mu.Unlock()
		listener.Close()
		return
	default:
		session.publicListener = listener
	}
	session.mu.Unlock()

	defer listener.Close()

	log.Printf("[server] public listener started on port %d for client %s", session.publicPort, session.clientID)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			log.Printf("[server] public listener error: %v", err)

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return
		}

		go s.handlePublicConnection(session, conn)
	}
}

// startPublicUDPListener opens a public UDP socket on the tunnel's allocated
// public port and bridges external UDP peers to the client. Each distinct
// external peer (source addr) becomes a udpServerSession: the server tells the
// client to dial its local backend (udp_open), relays peer→client packets via
// the UDP control channel, and writes client→peer replies back out this socket.
func (s *server) startPublicUDPListener(session *clientSession) {
	listenAddr := fmt.Sprintf(":%d", session.publicPort)
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Printf("[server] invalid UDP public addr %s: %v", listenAddr, err)
		return
	}
	pc, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("[server] failed to listen on public UDP port %d: %v", session.publicPort, err)
		return
	}
	_ = pc.SetReadBuffer(4 * 1024 * 1024)
	_ = pc.SetWriteBuffer(4 * 1024 * 1024)

	session.mu.Lock()
	select {
	case <-session.done:
		session.mu.Unlock()
		pc.Close()
		return
	default:
		session.publicUDPConn = pc
	}
	session.mu.Unlock()

	defer pc.Close()
	log.Printf("[server] public UDP listener started on port %d for client %s", session.publicPort, session.clientID)

	// extAddr string -> session ID, owned by this goroutine + the GC ticker.
	peers := make(map[string]string)
	var peersMu sync.Mutex

	// Idle GC: drop external peers that have gone quiet so sessions don't leak.
	gcStop := make(chan struct{})
	defer close(gcStop)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gcStop:
				return
			case <-session.done:
				return
			case <-ticker.C:
				now := time.Now()
				peersMu.Lock()
				for ext, id := range peers {
					s.udpMu.Lock()
					sess := s.udpSessions[id]
					s.udpMu.Unlock()
					if sess == nil || now.Sub(sess.lastActive) > clientIdleTimeout {
						delete(peers, ext)
						s.sendUDPClose(session.key, id)
						s.handleUDPClose(id)
					}
				}
				peersMu.Unlock()
			}
		}
	}()

	buf := make([]byte, 65535)
	for {
		n, extAddr, err := pc.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("[server] public UDP read error on port %d: %v", session.publicPort, err)
			return
		}
		if n == 0 {
			continue
		}

		key := extAddr.String()
		peersMu.Lock()
		sessID, known := peers[key]
		peersMu.Unlock()

		if debugServerUDP {
			log.Printf("[server] inbound UDP %d bytes from peer %s on port %d (known=%v)", n, key, session.publicPort, known)
		}
		if !known {
			if !s.checkUDPSessionRateLimit(session.remoteIP) {
				continue
			}
			id, gerr := tunnel.GenerateID()
			if gerr != nil {
				continue
			}
			sessID = id
			us := &udpServerSession{
				id:         sessID,
				clientKey:  session.key,
				udpSecret:  session.udpSecret,
				inbound:    true,
				publicConn: pc,
				extAddr:    extAddr,
				lastActive: time.Now(),
				closed:     make(chan struct{}),
			}
			s.udpMu.Lock()
			s.udpSessions[sessID] = us
			s.udpMu.Unlock()

			peersMu.Lock()
			peers[key] = sessID
			peersMu.Unlock()

			// Tell the client to open a backend UDP connection for this peer.
			if debugServerUDP {
				log.Printf("[server] inbound UDP new session %s -> sending udp_open to client", sessID)
			}
			if err := session.enc.Encode(tunnel.Message{Type: "udp_open", ID: sessID, Protocol: "udp"}); err != nil {
				s.handleUDPClose(sessID)
				peersMu.Lock()
				delete(peers, key)
				peersMu.Unlock()
				continue
			}
		}

		// Refresh activity so the idle GC keeps live peers alive.
		s.udpMu.Lock()
		if us := s.udpSessions[sessID]; us != nil {
			us.lastActive = time.Now()
		}
		s.udpMu.Unlock()

		payload := make([]byte, n)
		copy(payload, buf[:n])
		if err := s.sendUDPData(session.key, sessID, payload); err != nil && debugServerUDP {
			log.Printf("[server] inbound UDP forward to client failed: %v", err)
		}
	}
}

func (s *server) handlePublicConnection(session *clientSession, publicConn net.Conn) {
	defer publicConn.Close()

	// Generate proxy ID
	proxyID, err := tunnel.GenerateID()
	if err != nil {
		log.Printf("[server] failed to generate proxy ID: %v", err)
		return
	}

	// Register waiter
	waitCh := make(chan net.Conn, 1)
	s.proxyMu.Lock()
	s.proxyWaiting[proxyID] = waitCh
	s.proxyMu.Unlock()

	// Ensure cleanup if ignored
	defer func() {
		s.proxyMu.Lock()
		delete(s.proxyWaiting, proxyID)
		s.proxyMu.Unlock()
	}()

	// Send proxy request to client
	proxyMsg := tunnel.Message{
		Type:     "proxy",
		Key:      session.key,
		ClientID: session.clientID,
		ID:       proxyID,
	}

	if err := session.enc.Encode(proxyMsg); err != nil {
		log.Printf("[server] failed to send proxy request: %v", err)
		return
	}

	// Wait for client to connect back
	select {
	case clientConn := <-waitCh:
		if clientConn == nil {
			s.logOnce(fmt.Sprintf("[server] client refused proxy connection %s", proxyID), "proxy_refused")
			return
		}

		s.handleProxyStream(session, publicConn, clientConn)

	case <-time.After(10 * time.Second):
		s.logOnce(fmt.Sprintf("[server] timeout waiting for client proxy connection %s", proxyID), "proxy_timeout", 15*time.Second)
	}
}

func (s *server) handleProxyStream(session *clientSession, publicConn, clientConn net.Conn) {
	atomic.AddInt64(&session.activeConnections, 1)
	atomic.AddUint64(&session.totalConnections, 1)
	atomic.AddUint64(&s.totalConnections, 1)
	defer atomic.AddInt64(&session.activeConnections, -1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		proxyCopy(publicConn, clientConn, &session.bytesUp, &s.totalBytesUp)
	}()

	proxyCopy(clientConn, publicConn, &session.bytesDown, &s.totalBytesDown)
	wg.Wait()
}

func proxyCopy(dst, src net.Conn, counter *uint64, totalCounter *uint64) {
	defer dst.Close()
	defer src.Close()

	buf := pool.GlobalBufferPool.Get(32 * 1024)
	defer pool.GlobalBufferPool.Put(buf)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			atomic.AddUint64(counter, uint64(nr))
			if totalCounter != nil {
				atomic.AddUint64(totalCounter, uint64(nr))
			}
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			if ew != nil {
				return
			}
			if nr != nw {
				return // short write
			}
		}
		if er != nil {
			return
		}
	}
}

func (s *server) dispatchProxyConnection(conn net.Conn, proxyID string) {
	s.proxyMu.Lock()
	ch, ok := s.proxyWaiting[proxyID]
	if ok {
		delete(s.proxyWaiting, proxyID)
	}
	s.proxyMu.Unlock()

	if !ok {
		log.Printf("[server] unexpected proxy connection for ID %s", proxyID)
		conn.Close()
		return
	}

	// Send to waiting public handler
	select {
	case ch <- conn:
	case <-time.After(10 * time.Second):
		log.Printf("[server] timeout waiting for public handler to accept proxy connection %s", proxyID)
		conn.Close()
	}
}

func (s *server) cancelProxyConnection(proxyID string) {
	s.proxyMu.Lock()
	ch, ok := s.proxyWaiting[proxyID]
	if ok {
		delete(s.proxyWaiting, proxyID)
	}
	s.proxyMu.Unlock()

	if ok {
		// Signal cancellation by sending nil
		select {
		case ch <- nil:
		default:
		}
	}
}

func (s *server) handleProxyRequest(session *clientSession, proxyID string) {
	// This is just a notification log if needed, logic is in handlePublicConnection
}

func (s *server) handleUDPOpen(session *clientSession, msg tunnel.Message) {
	if session.protocol != "udp" {
		return
	}

	// Check rate limit for UDP session creation
	if !s.checkUDPSessionRateLimit(session.remoteIP) {
		log.Printf("[server] rate limit exceeded for UDP session creation from %s", session.remoteIP)
		return
	}

	// Parse remote address
	remoteAddr := strings.TrimSpace(msg.RemoteAddr)
	if remoteAddr == "" {
		log.Printf("[server] UDP open missing remote address")
		return
	}

	addr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		log.Printf("[server] invalid UDP remote address %s: %v", remoteAddr, err)
		return
	}

	// Validate address to prevent SSRF (Simple check for private ranges)
	// In production, use a more robust library to check against all private/multicast ranges.
	udpIP := addr.IP
	if udpIP.IsLoopback() || udpIP.IsPrivate() || udpIP.IsMulticast() {
		log.Printf("[server] blocked UDP attempt to restricted address %s", remoteAddr)
		return
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("[server] failed to create UDP connection to %s: %v", remoteAddr, err)
		return
	}

	udpSession := &udpServerSession{
		id:         msg.ID,
		clientKey:  session.key,
		udpSecret:  session.udpSecret,
		conn:       conn,
		remoteAddr: addr,
		closed:     make(chan struct{}),
	}

	s.udpMu.Lock()
	s.udpSessions[msg.ID] = udpSession
	s.udpMu.Unlock()

	go s.readFromUDPRemote(udpSession)
	log.Printf("[server] UDP session %s opened for %s", msg.ID, remoteAddr)
}

func (s *server) handleUDPClose(sessionID string) {
	s.udpMu.Lock()
	session := s.udpSessions[sessionID]
	if session != nil {
		delete(s.udpSessions, sessionID)
	}
	s.udpMu.Unlock()

	if session != nil {
		session.Close()
		log.Printf("[server] UDP session %s closed", sessionID)
	}
}

func (s *server) readFromUDPRemote(session *udpServerSession) {
	defer s.handleUDPClose(session.id)

	buf := make([]byte, 65535)
	for {
		n, err := session.conn.Read(buf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("[server] UDP read error for session %s: %v", session.id, err)
			}
			return
		}

		if n == 0 {
			continue
		}

		payload := make([]byte, n)
		copy(payload, buf[:n])

		// Send to client via UDP control
		if err := s.sendUDPData(session.clientKey, session.id, payload); err != nil {
			log.Printf("[server] failed to send UDP data to client: %v", err)
			return
		}
	}
}

func (s *server) readUDPControl() {
	if s.udpServer == nil {
		return
	}

	buf := make([]byte, 65535)
	for {
		n, addr, err := s.udpServer.ReadFromUDP(buf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("[server] UDP control read error: %v", err)
			}
			return
		}

		if n == 0 {
			continue
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])
		go s.handleUDPControlPacket(packet, addr)
	}
}

func (s *server) handleUDPControlPacket(packet []byte, addr *net.UDPAddr) {
	if len(packet) < 3 {
		return
	}

	msgType := packet[0]
	key, idx, ok := decodeUDPField(packet, 1)
	if !ok || key == "" {
		return
	}

	// Remember this client's UDP control address so inbound (public→client)
	// sessions can reach it even before the client has sent session data.
	s.udpMu.Lock()
	s.udpClientAddrs[key] = addr
	s.udpMu.Unlock()

	switch msgType {
	case udpMsgHandshake:
		_ = s.sendUDPResponse(addr, udpMsgHandshake, key, "", nil)
	case udpMsgData:
		id, next, ok := decodeUDPField(packet, idx)
		if !ok || id == "" {
			return
		}

		payload := make([]byte, len(packet)-next)
		copy(payload, packet[next:])

		s.handleUDPDataFromClient(key, id, payload, addr)
	case udpMsgClose:
		id, _, ok := decodeUDPField(packet, idx)
		if !ok || id == "" {
			return
		}
		s.handleUDPClose(id)
	case udpMsgPing:
		payload := make([]byte, len(packet)-idx)
		copy(payload, packet[idx:])
		_ = s.sendUDPResponse(addr, udpMsgPong, key, "", payload)
	}
}

func (s *server) handleUDPDataFromClient(clientKey, sessionID string, payload []byte, clientAddr *net.UDPAddr) {
	s.udpMu.Lock()
	session := s.udpSessions[sessionID]
	s.udpMu.Unlock()

	if session == nil || session.clientKey != clientKey {
		log.Printf("[server] UDP data for unknown or mismatched session %s", sessionID)
		return
	}

	// Update client address for return traffic
	if session.clientAddr == nil || session.clientAddr.String() != clientAddr.String() {
		session.clientAddr = clientAddr
	}
	session.lastActive = time.Now()

	// Decrypt if secret is available
	if session.udpSecret != nil {
		decrypted, err := tunnel.DecryptUDP(session.udpSecret, payload)
		if err != nil {
			log.Printf("[server] UDP decryption failed for session %s: %v", sessionID, err)
			return
		}
		payload = decrypted
	}

	// Inbound session: reply goes back out the public socket to the external peer.
	if session.inbound {
		if session.publicConn == nil || session.extAddr == nil {
			return
		}
		if _, err := session.publicConn.WriteToUDP(payload, session.extAddr); err != nil {
			if debugServerUDP {
				log.Printf("[server] failed to write UDP to peer for session %s: %v", sessionID, err)
			}
			s.handleUDPClose(sessionID)
		}
		return
	}

	// Outbound session: forward to the dialed remote.
	if session.conn == nil {
		return
	}
	if _, err := session.conn.Write(payload); err != nil {
		log.Printf("[server] failed to write UDP to remote for session %s: %v", sessionID, err)
		s.handleUDPClose(sessionID)
	}
}

func (s *server) sendUDPData(clientKey, sessionID string, payload []byte) error {
	s.udpMu.Lock()
	session := s.udpSessions[sessionID]
	s.udpMu.Unlock()

	if session == nil {
		return errors.New("udp session not found")
	}

	// Encrypt if secret is available
	if session.udpSecret != nil {
		encrypted, err := tunnel.EncryptUDP(session.udpSecret, payload)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}
		payload = encrypted
	}

	// Resolve where to send: the session's learned client addr, or the last
	// known UDP control addr for this client (needed for inbound sessions
	// before the client has replied on this particular session).
	dst := session.clientAddr
	if dst == nil {
		s.udpMu.Lock()
		dst = s.udpClientAddrs[clientKey]
		s.udpMu.Unlock()
	}
	if debugServerUDP {
		log.Printf("[server] sendUDPData session=%s clientKey=%q dst=%v sessionAddr=%v", sessionID, clientKey, dst, session.clientAddr)
	}
	if dst == nil {
		// Don't know where the client is yet; drop.
		return nil
	}

	return s.writeUDP(udpMsgData, clientKey, sessionID, payload, dst)
}

// sendUDPClose notifies the client that a UDP session is finished so it can
// tear down the matching backend connection.
func (s *server) sendUDPClose(clientKey, sessionID string) {
	s.udpMu.Lock()
	dst := s.udpClientAddrs[clientKey]
	s.udpMu.Unlock()
	if dst == nil {
		return
	}
	_ = s.writeUDP(udpMsgClose, clientKey, sessionID, nil, dst)
}

func (s *server) sendUDPResponse(addr *net.UDPAddr, msgType byte, key, id string, payload []byte) error {
	// Pings/Handshakes are not encrypted (for now, or use separate secret?)
	// Handshake doesn't have secret yet.
	// Ping payload is random bytes, less critical.
	// Ideally encrypt pings too if key established.
	return s.writeUDP(msgType, key, id, payload, addr)
}

func (s *server) writeUDP(msgType byte, key, id string, payload []byte, addr *net.UDPAddr) error {
	if s.udpServer == nil {
		return errors.New("UDP server not available")
	}

	buf := buildUDPMessage(msgType, key, id, payload)
	_, err := s.udpServer.WriteToUDP(buf, addr)
	return err
}

func (s *server) heartbeatChecker(session *clientSession) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			session.mu.Lock()
			idle := time.Since(session.lastSeen)
			session.mu.Unlock()

			if idle > clientIdleTimeout {
				log.Printf("[server] client %s idle timeout, disconnecting", session.clientID)
				session.Close()
				return
			}
		case <-session.done:
			return
		}
	}
}

func (s *server) addClient(session *clientSession) {
	s.clientsMu.Lock()
	s.clients[session.clientID] = session
	s.clientsMu.Unlock()
}

func (s *server) removeClient(session *clientSession) {
	_ = session.state.TransitionTo(tunnel.StateClosed)
	s.clientsMu.Lock()
	existingSession, exists := s.clients[session.clientID]
	if exists && existingSession == session {
		delete(s.clients, session.clientID)

		// Create port & subdomain reservation instead of releasing immediately
		if session.publicPort > 0 && session.key != "" {
			_ = s.reservations.ReservePort(session.key, session.publicPort, 5*time.Minute)
			log.Printf("[server] Reserved port %d for client %s (key: %s) for 5 minutes",
				session.publicPort, session.clientID, session.key)
		}
		if session.protocol == "http" && session.subdomain != "" && session.key != "" {
			_ = s.reservations.ReserveSubdomain(session.key, session.subdomain, 5*time.Minute)
			log.Printf("[server] Reserved subdomain %s for client %s (key: %s) for 5 minutes",
				session.subdomain, session.clientID, session.key)
		}
	}
	s.clientsMu.Unlock()

	// Always ensure this specific session is closed
	// Unregister from HTTP proxy if applicable
	if existingSession == session {
		s.unregisterHTTPClient(session)
	}
	session.Close()
}

// Send dashboard update with real-time tunnel info and metrics
func (s *server) sendDashboardUpdate(conn *websocket.Conn) error {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	tunnels := make([]gin.H, 0, len(s.clients))
	var totalUp, totalDown uint64

	// Get base domain for HTTP tunnels
	baseDomain := ""
	if s.httpProxy != nil {
		baseDomain = s.httpProxy.GetBaseDomain()
	}

	var activeProxyConnections int64
	for _, session := range s.clients {
		host, port, _ := net.SplitHostPort(session.target)
		if host == "" || host == "localhost" || host == "127.0.0.1" || host == "::1" {
			host = session.remoteIP
		}
		if port == "" {
			port = session.target
		}

		// Display subdomain for HTTP tunnels, IP:port for others
		publicHost := fmt.Sprintf("%s:%d", s.publicHost, session.publicPort)
		if session.protocol == "http" && session.subdomain != "" {
			publicHost = fmt.Sprintf("https://%s.%s", session.subdomain, baseDomain)
		}

		up := atomic.LoadUint64(&session.bytesUp)
		down := atomic.LoadUint64(&session.bytesDown)
		totalUp += up
		totalDown += down
		activeProxyConnections += atomic.LoadInt64(&session.activeConnections)

		tunnels = append(tunnels, gin.H{
			"name":        session.clientID,
			"status":      "active",
			"protocol":    session.protocol,
			"local_host":  host,
			"local_port":  port,
			"public_port": session.publicPort,
			"public_host": publicHost,
			"bytes_up":    up,
			"bytes_down":  down,
		})
	}

	// Send tunnel update
	if err := conn.WriteJSON(gin.H{
		"type": "tunnel_update",
		"data": tunnels,
	}); err != nil {
		return err
	}

	// Send metrics
	metrics := gin.H{
		"activeTunnels":          len(s.clients),
		"activeProxyConnections": activeProxyConnections,
		"activeBytesUp":          totalUp,
		"activeBytesDown":        totalDown,
		"totalConnections":       atomic.LoadUint64(&s.totalConnections),
		"totalBytesUp":           atomic.LoadUint64(&s.totalBytesUp),
		"totalBytesDown":         atomic.LoadUint64(&s.totalBytesDown),
		"uptimeSeconds":          time.Since(s.runtimeStart).Seconds(),
	}

	return conn.WriteJSON(gin.H{
		"type": "metrics",
		"data": metrics,
	})
}

func (s *server) logOnce(message, key string, cooldown ...time.Duration) {
	if message == "" || key == "" {
		return
	}

	interval := 30 * time.Second
	if len(cooldown) > 0 {
		interval = cooldown[0]
	}

	now := time.Now()
	s.throttleMu.Lock()
	last, ok := s.throttledLogs[key]
	if ok && now.Sub(last) < interval {
		s.throttleMu.Unlock()
		return
	}
	s.throttledLogs[key] = now
	s.throttleMu.Unlock()

	log.Println(message)
}

func (s *server) isPortInActiveUse(port int) bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	for _, sess := range s.clients {
		if sess.publicPort == port {
			return true
		}
	}
	return false
}

func (s *server) getNextPublicPort(clientKey string, requestedPort int) int {
	s.portMu.Lock()
	defer s.portMu.Unlock()

	// Check if requested port is reserved by another client
	if requestedPort > 0 {
		reservedByOther, err := s.reservations.IsPortReservedByOther(requestedPort, clientKey)
		if err == nil && reservedByOther {
			log.Printf("[server] Requested port %d is reserved by another client", requestedPort)
			requestedPort = 0
		}
	}

	// Check if client has a valid reservation
	if requestedPort > 0 && clientKey != "" {
		reservedPort, exists, err := s.reservations.GetReservedPort(clientKey)
		if err == nil && exists && reservedPort == requestedPort {
			// Check if port is in active use
			if !s.isPortInActiveUse(requestedPort) {
				// Find and remove from availablePorts
				for i, p := range s.availablePorts {
					if p == requestedPort {
						s.availablePorts = append(s.availablePorts[:i], s.availablePorts[i+1:]...)
						break
					}
				}
				s.usedPorts[requestedPort] = true
				_ = s.reservations.DeleteReservation(clientKey) // consume it

				log.Printf("[server] Assigned reserved port %d to client %s", requestedPort, clientKey)
				return requestedPort
			}
		}
	}

	// No valid reservation, get next available port
	if len(s.availablePorts) == 0 {
		log.Printf("[server] ⚠️  Port pool exhausted!")
		return publicPortStart // Fallback
	}

	port := s.availablePorts[0]
	s.availablePorts = s.availablePorts[1:]
	s.usedPorts[port] = true

	return port
}

func (s *server) releasePort(port int) {
	s.portMu.Lock()
	defer s.portMu.Unlock()

	if !s.usedPorts[port] {
		return
	}

	delete(s.usedPorts, port)
	s.availablePorts = append(s.availablePorts, port)
	sort.Ints(s.availablePorts)
}

func (s *server) getClient(clientID string) *clientSession {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return s.clients[clientID]
}

func (session *clientSession) Close() {
	session.closeOnce.Do(func() {
		close(session.done)
		if session.conn != nil {
			session.conn.Close()
		}

		session.mu.Lock()
		if session.publicListener != nil {
			session.publicListener.Close()
		}
		if session.publicUDPConn != nil {
			session.publicUDPConn.Close()
		}
		session.mu.Unlock()

		// Forget this client's UDP control address.
		if session.server != nil && session.key != "" {
			session.server.udpMu.Lock()
			delete(session.server.udpClientAddrs, session.key)
			session.server.udpMu.Unlock()
		}
	})
}

func (s *udpServerSession) Close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		if s.timer != nil {
			s.timer.Stop()
		}
		if s.conn != nil {
			s.conn.Close()
		}
	})
}

func decodeUDPField(packet []byte, offset int) (string, int, bool) {
	if offset+2 > len(packet) {
		return "", offset, false
	}
	l := int(binary.BigEndian.Uint16(packet[offset : offset+2]))
	offset += 2
	if l < 0 || offset+l > len(packet) {
		return "", offset, false
	}
	return string(packet[offset : offset+l]), offset + l, true
}

// resolvePublicHost returns the public host/IP advertised to clients for
// TCP/UDP tunnels. It prefers the configured value (PUBLIC_HOST), then falls
// back to auto-detecting the outbound IP, and finally to 127.0.0.1.
func resolvePublicHost(configured string) string {
	if configured != "" {
		return configured
	}
	if ip := detectOutboundIP(); ip != "" {
		log.Printf("[server] PUBLIC_HOST not set, auto-detected public host: %s", ip)
		return ip
	}
	log.Printf("[server] PUBLIC_HOST not set and auto-detection failed, falling back to 127.0.0.1")
	return "127.0.0.1"
}

// detectOutboundIP determines the local address used for outbound traffic
// without sending any packets (UDP "connect" only sets the route).
func detectOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return addr.IP.String()
	}
	return ""
}

func buildUDPMessage(msgType byte, key, id string, payload []byte) []byte {
	keyLen := len(key)
	idLen := len(id)
	total := 1 + 2 + keyLen
	if msgType != udpMsgHandshake {
		total += 2 + idLen
	}
	total += len(payload)
	buf := make([]byte, total)
	buf[0] = msgType
	binary.BigEndian.PutUint16(buf[1:], uint16(keyLen))
	copy(buf[3:], key)
	offset := 3 + keyLen
	if msgType != udpMsgHandshake {
		binary.BigEndian.PutUint16(buf[offset:], uint16(idLen))
		offset += 2
		copy(buf[offset:], id)
		offset += idLen
	}
	copy(buf[offset:], payload)
	return buf
}

func (s *server) Shutdown(db *database.Database) {
	log.Println("[server] Bắt đầu tắt server an toàn...")

	// 1. Close main tunnel listener to stop accepting new tunnels
	if s.listener != nil {
		log.Println("[server] Đang đóng TCP listener...")
		s.listener.Close()
	}
	if s.udpServer != nil {
		log.Println("[server] Đang đóng UDP listener...")
		s.udpServer.Close()
	}
	if s.httpServer != nil {
		log.Println("[server] Đang đóng HTTP Dashboard server...")
		s.httpServer.Close()
	}

	// 2. Disconnect and persist active clients
	s.clientsMu.Lock()
	log.Printf("[server] Đang ngắt kết nối và bảo lưu trạng thái %d client active...", len(s.clients))
	for _, client := range s.clients {
		if client.publicPort > 0 && client.key != "" {
			_ = s.reservations.ReservePort(client.key, client.publicPort, 5*time.Minute)
		}
		if client.protocol == "http" && client.subdomain != "" && client.key != "" {
			_ = s.reservations.ReserveSubdomain(client.key, client.subdomain, 5*time.Minute)
		}
		client.Close()
	}
	s.clientsMu.Unlock()

	// 3. Flush database WAL cache
	if db != nil {
		log.Println("[database] Thực hiện SQLite checkpoint...")
		if _, err := db.GetDB().Exec("PRAGMA wal_checkpoint(PASSIVE);"); err != nil {
			log.Printf("[database] Lỗi checkpoint WAL: %v", err)
		}
	}

	log.Println("[server] ✅ Tắt server hoàn tất. Tạm biệt!")
	os.Exit(0)
}

type MaskedWriter struct {
	underlying io.Writer
}

func NewMaskedWriter(underlying io.Writer) *MaskedWriter {
	return &MaskedWriter{underlying: underlying}
}

func (mw *MaskedWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	str = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9-_=\.]+`).ReplaceAllString(str, "Bearer [REDACTED]")
	str = regexp.MustCompile(`(?i)token[=:\s"]+[A-Za-z0-9-_=\.]+`).ReplaceAllString(str, "token=[REDACTED]")
	str = regexp.MustCompile(`(?i)password[=:\s"]+[a-zA-Z0-9-_]+`).ReplaceAllString(str, "password=[REDACTED]")
	str = regexp.MustCompile(`(?i)cookie[=:\s"]+[a-zA-Z0-9-_]+`).ReplaceAllString(str, "cookie=[REDACTED]")
	str = regexp.MustCompile(`(?i)api_key[=:\s"]+[a-zA-Z0-9-_]+`).ReplaceAllString(str, "api_key=[REDACTED]")

	_, err = mw.underlying.Write([]byte(str))
	return len(p), err
}
