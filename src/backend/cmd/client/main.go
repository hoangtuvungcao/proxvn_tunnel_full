package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"

	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/term"

	"proxvn/backend/internal/tunnel"
)

const (
	defaultServerAddr  = "103.77.246.196:8882"
	defaultLocalHost   = "localhost"
	defaultLocalPort   = 80
	heartbeatInterval  = 5 * time.Second // Faster detection
	backendIdleTimeout = 5 * time.Second
	backendIdleRetries = 3
	udpControlInterval = 2 * time.Second
	udpControlTimeout  = 6 * time.Second
)

const debugUDP = false

const (
	udpMsgHandshake byte = 1
	udpMsgData      byte = 2
	udpMsgClose     byte = 3
	udpMsgPing      byte = 4
	udpMsgPong      byte = 5
)

type client struct {
	serverAddr         string
	localAddr          string
	key                string
	clientID           string
	remotePort         int
	publicHost         string
	protocol           string
	subdomain          string // Subdomain assigned by server for HTTP mode
	baseDomain         string // Base domain assigned by server for HTTP mode
	certFingerprint    string // Optional: Server certificate fingerprint for pinning
	insecureSkipVerify bool   // Skip TLS certificate verification
	uiEnabled          bool
	generation         int64
	state              *tunnel.StateMachine
	ctrlQueue          *tunnel.ControlMessageQueue
	rttMu              sync.Mutex
	rtt                time.Duration

	// Control connection
	control        net.Conn
	enc            *jsonWriter
	dec            *jsonReader
	closeOnce      sync.Once
	done           chan struct{}
	trafficQuit    chan struct{}
	statusCh       chan trafficStats
	bytesUp        uint64
	bytesDown      uint64
	pingCh         chan time.Duration
	pingSent       int64
	pingMs         int64
	exitFlag       uint32
	activeSessions int64
	totalSessions  uint64

	udpMu       sync.Mutex
	udpSessions map[string]*udpClientSession
	udpConn     *net.UDPConn
	udpReady    bool

	udpCtrlMu        sync.Mutex
	udpPingTicker    *time.Ticker
	udpPingStop      chan struct{}
	udpLastPing      time.Time
	udpLastPong      time.Time
	udpControlWarned bool
	udpCtrlStatus    string

	dataMu           sync.Mutex
	lastServerData   time.Time
	lastBackendData  time.Time
	totalUDPSessions uint64
	udpSecret        []byte // Key for UDP encryption
}

type trafficStats struct {
	upRate    string
	downRate  string
	totalUp   string
	totalDown string
}

type udpClientSession struct {
	id         string
	conn       *net.UDPConn
	remoteAddr string
	closeOnce  sync.Once
	closed     chan struct{}
	timer      *time.Timer
	idleCount  int
}

func (s *udpClientSession) Close() {
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
	// Custom usage message with examples
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
╔════════════════════════════════════════════════════════════════════════════╗
║                 ProxVN v%s - Client                                   ║
║            Tunnel Localhost ra Internet - Miễn Phí 100%%                   ║
╚════════════════════════════════════════════════════════════════════════════╝

🌟 TÍNH NĂNG:
  • HTTP Tunnel:  Nhận subdomain HTTPS tự động (https://abc.domain.com)
  • TCP Tunnel:   Public bất kỳ service TCP nào (Web, SSH, RDP, Database...)
  • UDP Tunnel:   Cho game server (Minecraft PE, CS:GO, Palworld...)
  • File Sharing: Chia sẻ file/folder như Windows Network Share
  • TLS Security: Mã hóa end-to-end cho tất cả kết nối
  • Auto Reconnect: Tự động kết nối lại khi mất mạng

📖 CÚ PHÁP:
  proxvn [OPTIONS] [LOCAL_PORT]

⚙️  CÁC THAM SỐ:
`, tunnel.Version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
💡 VÍ DỤ SỬ DỤNG:

▶ HTTP Tunnel - Nhận Subdomain HTTPS:
  proxvn --proto http 80              # Share website port 80
  proxvn --proto http 3000            # Share Node.js/React app
  proxvn --proto http 443             # Tunnel local HTTPS site
  → Kết quả: https://abc123.bacsycay.click

▶ TCP Tunnel - Nhận IP:Port:
  proxvn 80                           # Public web server
  proxvn 3389                         # Remote Desktop (RDP)
  proxvn 22                           # SSH server
  → Kết quả: 103.77.246.196:10000

▶ UDP Tunnel - Game Server:
  proxvn --proto udp 19132            # Minecraft Bedrock Edition
  proxvn --proto udp 25565            # Minecraft Java (UDP mode)
  proxvn --proto udp 7777             # Palworld server
  → Kết quả: 103.77.246.196:10000

▶ File Sharing - Chia Sẻ File/Folder:
  proxvn --file /home/user/Documents --pass matkhau123
  proxvn --file "C:\\Projects" --pass abc123         # Windows
  proxvn --file ~/Downloads --pass secret --permissions r  # Read-only
  → Kết quả: Mount như ổ đĩa mạng (Z:\\) hoặc truy cập qua web

▶ Kết nối tới VPS riêng:
  proxvn --server YOUR_VPS_IP:8882 --proto http 80

🔗 THÔNG TIN:
  • Website:        https://bacsycay.click
  • Documentation:  https://github.com/hoangtuvungcao/proxvn_tunnel
  • Issues:         https://github.com/hoangtuvungcao/proxvn_tunnel/issues

© 2026 ProxVN - Developed by TrongDev
Licensed under FREE TO USE - NON-COMMERCIAL ONLY

`)
	}

	// Load config file (if any) so its values become the flag defaults.
	// Precedence: command-line flags > config file > built-in defaults.
	cfgFile := loadClientConfig(preParseConfigFlag(os.Args[1:]))

	_ = flag.String("config", "", "Đường dẫn file cấu hình JSON (mặc định: proxvn.json cạnh binary / ~/.proxvn/config.json)")
	serverAddr := flag.String("server", cfgFile.Server, "Địa chỉ tunnel server (host:port)")
	hostFlag := flag.String("host", cfgFile.Host, "Host nội bộ cần tunnel (mặc định: localhost)")
	portFlag := flag.Int("port", cfgFile.Port, "Port nội bộ (bị ghi đè nếu truyền trực tiếp)")
	id := flag.String("id", "", "Client ID (optional)")
	proto := flag.String("proto", cfgFile.Proto, "Protocol: tcp, udp, or http")
	UI := flag.Bool("ui", cfgFile.UI, "Enable TUI (disable with --ui=false)")
	certPin := flag.String("cert-pin", cfgFile.CertPin, "Optional: Server certificate SHA256 fingerprint for pinning (hex format)")
	insecure := flag.Bool("insecure", cfgFile.Insecure, "Skip TLS certificate verification (for testing with localhost)")

	// File sharing flags
	fileFlag := flag.String("file", "", "Đường dẫn file/folder cần chia sẻ (vd: /home/user/docs, C:\\\\Users\\\\Admin\\\\Documents)")
	userFlag := flag.String("user", "proxvn", "Username để truy cập file share (mặc định: proxvn)")
	passFlag := flag.String("pass", "", "Mật khẩu để truy cập file share (bắt buộc khi dùng --file)")
	permsFlag := flag.String("permissions", "rw", "Quyền hạn: r (chỉ đọc), rw (đọc-ghi), rwx (đầy đủ)")

	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags)

	// Check if file sharing mode
	if *fileFlag != "" {
		if *passFlag == "" {
			log.Fatal("❌ Lỗi: --pass bắt buộc khi dùng --file")
		}
		// Trim spaces to prevent auth errors
		username := strings.TrimSpace(*userFlag)
		password := strings.TrimSpace(*passFlag)
		perms := strings.TrimSpace(*permsFlag)

		if err := runFileShareMode(*fileFlag, username, password, perms, *serverAddr, *insecure); err != nil {
			log.Fatalf("❌ File sharing lỗi: %v", err)
		}
		return
	}

	clientID := strings.TrimSpace(*id)
	if clientID == "" {
		host, _ := os.Hostname()
		clientID = fmt.Sprintf("client-%s", host)
	}

	localHost := strings.TrimSpace(*hostFlag)
	if localHost == "" {
		localHost = defaultLocalHost
	}
	localPort := *portFlag

	args := normalizedArgs(flag.Args())
	switch len(args) {
	case 0:
		// use flag defaults
	case 1:
		if p, err := strconv.Atoi(args[0]); err == nil && p > 0 && p <= 65535 {
			localPort = p
		} else {
			log.Fatalf("[client] port không hợp lệ: %q", args[0])
		}
	default:
		if strings.TrimSpace(args[0]) != "" {
			localHost = args[0]
		}
		if p, err := strconv.Atoi(args[1]); err == nil && p > 0 && p <= 65535 {
			localPort = p
		} else {
			log.Fatalf("[client] port không hợp lệ: %q", args[1])
		}
	}

	if localPort <= 0 || localPort > 65535 {
		log.Fatalf("[client] port không hợp lệ: %d", localPort)
	}

	protocol := strings.ToLower(strings.TrimSpace(*proto))
	if protocol != "udp" && protocol != "http" {
		protocol = "tcp"
	}

	cl := &client{
		serverAddr:      *serverAddr,
		localAddr:       net.JoinHostPort(localHost, strconv.Itoa(localPort)),
		clientID:        clientID,
		protocol:        protocol,
		certFingerprint: strings.ToLower(strings.TrimSpace(*certPin)),
		uiEnabled:       *UI && term.IsTerminal(int(os.Stdout.Fd())),
		state:           tunnel.NewStateMachine(),
		ctrlQueue:       tunnel.NewControlMessageQueue(),
	}

	if err := cl.run(); err != nil {
		log.Fatalf("[client] lỗi: %v", err)
	}
}

func (c *client) run() error {
	backoff := 2 * time.Second
	maxBackoff := 30 * time.Second

	for {
		_ = c.state.TransitionTo(tunnel.StateConnecting)
		if err := c.connectControl(); err != nil {
			log.Printf("[client] kết nối control thất bại: %v", err)

			// Calculate backoff with jitter
			jitter := time.Duration(time.Now().UnixNano() % int64(500*time.Millisecond))
			sleepTime := backoff + jitter
			log.Printf("[client] retry sau %v...", sleepTime)

			_ = c.state.TransitionTo(tunnel.StateReconnecting)
			time.Sleep(sleepTime)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// Reset backoff on success
		backoff = 2 * time.Second

		if err := c.receiveLoop(); err != nil {
			log.Printf("[client] control lỗi: %v", err)
		}
		c.closeControl()
		if atomic.LoadUint32(&c.exitFlag) == 1 {
			return nil
		}

		_ = c.state.TransitionTo(tunnel.StateReconnecting)
		jitter := time.Duration(time.Now().UnixNano() % int64(500*time.Millisecond))
		time.Sleep(backoff + jitter)
	}
}

func (c *client) connectControl() error {
	// Connect with TLS (certificate verification skipped by default)
	tlsConfig := c.buildTLSConfig()
	conn, err := tls.Dial("tcp", c.serverAddr, tlsConfig)
	if err != nil {
		return err
	}

	c.closeOnce = sync.Once{}
	c.done = make(chan struct{})
	c.trafficQuit = make(chan struct{})
	c.statusCh = make(chan trafficStats, 1)
	c.pingCh = make(chan time.Duration, 1)
	c.control = conn
	c.enc = &jsonWriter{enc: tunnel.NewEncoder(conn)}
	c.dec = &jsonReader{dec: tunnel.NewDecoder(bufio.NewReader(conn))}
	c.stopUDPPing()
	c.setUDPCtrlStatus("offline")
	atomic.StoreUint64(&c.bytesUp, 0)
	atomic.StoreUint64(&c.bytesDown, 0)
	atomic.StoreInt64(&c.pingSent, 0)
	atomic.StoreInt64(&c.pingMs, -1)
	select {
	case c.pingCh <- time.Duration(-1):
	default:
	}
	success := false
	defer func() {
		if !success {
			if c.control != nil {
				c.control.Close()
				c.control = nil
			}
			if c.trafficQuit != nil {
				close(c.trafficQuit)
				c.trafficQuit = nil
			}
			c.enc = nil
			c.dec = nil
			c.udpMu.Lock()
			if c.udpConn != nil {
				c.udpConn.Close()
				c.udpConn = nil
			}
			c.udpMu.Unlock()
			c.stopUDPPing()
		}
	}()

	c.generation++
	_ = c.state.TransitionTo(tunnel.StateConnecting)
	_ = c.state.TransitionTo(tunnel.StateTLSHandshake)
	_ = c.state.TransitionTo(tunnel.StateAuthenticating)
	_ = c.state.TransitionTo(tunnel.StateRegistering)

	register := tunnel.Message{
		Type:       "register",
		Key:        c.key,
		ClientID:   c.clientID,
		Target:     c.localAddr,
		Protocol:   c.protocol,
		Generation: c.generation,
		Subdomain:  c.subdomain,
	}

	// If reconnecting and we had a port before, request the same port
	if c.remotePort > 0 {
		register.RequestedPort = c.remotePort
	}

	if err := c.enc.Encode(register); err != nil {
		return err
	}

	resp := tunnel.Message{}
	if err := c.dec.Decode(&resp); err != nil {
		return err
	}
	if resp.Type != "registered" {
		_ = c.state.TransitionTo(tunnel.StateClosed)
		return fmt.Errorf("đăng ký thất bại: %+v", resp)
	}
	_ = c.state.TransitionTo(tunnel.StateActive)
	go c.replayQueuedMessages()
	if strings.TrimSpace(resp.Key) != "" {
		c.key = strings.TrimSpace(resp.Key)
	}
	c.remotePort = resp.RemotePort
	if strings.TrimSpace(resp.Protocol) != "" {
		c.protocol = strings.ToLower(strings.TrimSpace(resp.Protocol))
	}

	// For HTTP mode, server assigns a subdomain
	if c.protocol == "http" && resp.Subdomain != "" {
		c.subdomain = resp.Subdomain
	}

	// Handle UDP Encryption Key
	if resp.UDPSecret != "" {
		secret, err := base64.StdEncoding.DecodeString(resp.UDPSecret)
		if err == nil && len(secret) == 32 {
			c.udpSecret = secret
		}
	}
	// Also store base domain if provided
	if resp.BaseDomain != "" {
		c.baseDomain = resp.BaseDomain
	}

	hostPart := c.serverAddr
	if host, _, err := net.SplitHostPort(c.serverAddr); err == nil {
		hostPart = host
	}
	c.publicHost = net.JoinHostPort(hostPart, strconv.Itoa(c.remotePort))
	c.setUDPCtrlStatus("n/a")

	// Log success based on protocol
	if c.protocol == "http" {
		domain := c.baseDomain
		if domain == "" {
			domain = "bacsycay.click"
		}
		log.Printf("[client] ✅ HTTP Tunnel Active")
		log.Printf("[client] 🌐 Public URL: https://%s.%s", c.subdomain, domain)
		log.Printf("[client] 📍 Forwarding to: %s", c.localAddr)
	} else {
		log.Printf("[client] đăng ký thành công, public port %d", c.remotePort)
	}

	if c.protocol == "udp" {
		c.setUDPCtrlStatus("offline")
		if err := c.setupUDPChannel(); err != nil {
			log.Printf("[client] thiết lập UDP control lỗi: %v", err)
		} else if debugUDP {
			log.Printf("[client] UDP control đang chờ handshake với %s", c.serverAddr)
		}
	}
	go c.heartbeatLoop()
	go c.trafficLoop()
	go c.displayLoop()
	success = true
	return nil
}

func (c *client) receiveLoop() error {
	for {
		msg := tunnel.Message{}
		if err := c.dec.Decode(&msg); err != nil {
			if isEOF(err) {
				return io.EOF
			}
			return err
		}

		switch msg.Type {
		case "proxy":
			go c.handleProxy(msg.ID)
		case "udp_open":
			c.handleUDPOpen(msg)
		case "udp_close":
			c.handleUDPClose(msg.ID)
		case "ping":
			_ = c.enc.Encode(tunnel.Message{Type: "pong"})
		case "pong":
			c.recordPingReply()
		case "http_request":
			// Handle HTTP request
			go c.handleHTTPRequest(msg)
		case "error":
			log.Printf("[client] server báo lỗi: %s", msg.Error)
		default:
			log.Printf("[client] thông điệp không hỗ trợ: %+v", msg)
		}
	}
}

func (c *client) heartbeatLoop() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			if c.enc == nil {
				return
			}
			if err := c.enc.Encode(tunnel.Message{Type: "ping"}); err != nil {
				return
			}
			atomic.StoreInt64(&c.pingSent, start.UnixNano())

			c.rttMu.Lock()
			currentRTT := c.rtt
			c.rttMu.Unlock()

			pongTimeout := 3 * currentRTT
			if pongTimeout < 3*time.Second {
				pongTimeout = 3 * time.Second
			}
			if pongTimeout > 10*time.Second {
				pongTimeout = 10 * time.Second
			}

			// Spawn check routine
			go func(sentTime int64, timeout time.Duration) {
				time.Sleep(timeout)
				sent := atomic.LoadInt64(&c.pingSent)
				if sent == sentTime && sent > 0 {
					log.Printf("[client] Ping timeout sau %v (RTT=%v), đóng control...", timeout, currentRTT)
					c.closeControl()
				}
			}(start.UnixNano(), pongTimeout)
		case <-c.done:
			return
		}
	}
}

func (c *client) trafficLoop() {
	const interval = 1 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastUp, lastDown uint64
	firstStats := trafficStats{
		upRate:    formatRate(0, interval),
		downRate:  formatRate(0, interval),
		totalUp:   formatBytes(0),
		totalDown: formatBytes(0),
	}
	select {
	case c.statusCh <- firstStats:
	default:
	}
	for {
		select {
		case <-ticker.C:
			up := atomic.LoadUint64(&c.bytesUp)
			down := atomic.LoadUint64(&c.bytesDown)
			upDelta := up - lastUp
			downDelta := down - lastDown
			lastUp = up
			lastDown = down
			stats := trafficStats{
				upRate:    formatRate(upDelta, interval),
				downRate:  formatRate(downDelta, interval),
				totalUp:   formatBytes(up),
				totalDown: formatBytes(down),
			}
			select {
			case c.statusCh <- stats:
			default:
				select {
				case <-c.statusCh:
				default:
				}
				c.statusCh <- stats
			}
		case <-c.trafficQuit:
			return
		case <-c.done:
			return
		}
	}
}

func (c *client) displayLoop() {
	if !c.uiEnabled {
		return
	}

	if c.uiEnabled {
		fmt.Print("\033[2J\033[H\033[?25l")
		defer fmt.Print("\033[?25h\033[2J\033[H")
	}

	traffic := trafficStats{
		upRate:    formatRate(0, time.Second),
		downRate:  formatRate(0, time.Second),
		totalUp:   formatBytes(0),
		totalDown: formatBytes(0),
	}
	ping := time.Duration(-1)
	hasTraffic := false

	render := func() {
		if !hasTraffic {
			return
		}
		c.renderFrame(traffic, ping)
	}

	for {
		// Force redraw every second even if no stats update
		redrawTicker := time.NewTicker(1 * time.Second)
		defer redrawTicker.Stop()

		select {
		case <-redrawTicker.C:
			render()
		case stats, ok := <-c.statusCh:
			if !ok {
				return
			}
			traffic = stats
			hasTraffic = true
			render()
		case duration, ok := <-c.pingCh:
			if !ok {
				ping = time.Duration(-1)
				continue
			}
			ping = duration
			render()
		case <-c.done:
			return
		case <-c.trafficQuit:
			return
		}
	}
}

func (c *client) handleProxy(id string) {
	if c.protocol == "udp" {
		log.Printf("[client] bỏ qua proxy TCP vì tunnel đang ở chế độ UDP")
		return
	}
	if strings.TrimSpace(id) == "" {
		return
	}

	localConn, err := net.Dial("tcp", c.localAddr)
	if err != nil {
		log.Printf("[client] không kết nối được backend %s: %v", c.localAddr, err)
		c.reportProxyError(id, err)
		return
	}

	atomic.AddInt64(&c.activeSessions, 1)
	atomic.AddUint64(&c.totalSessions, 1)

	// Connect to server with TLS
	tlsConfig := c.buildTLSConfig()
	srvConn, err := tls.Dial("tcp", c.serverAddr, tlsConfig)
	if err != nil {
		log.Printf("[client] không connect server cho proxy: %v", err)
		localConn.Close()
		c.reportProxyError(id, err)
		return
	}

	enc := tunnel.NewEncoder(srvConn)
	if err := enc.Encode(tunnel.Message{
		Type:     "proxy",
		Key:      c.key,
		ClientID: c.clientID,
		ID:       id,
	}); err != nil {
		log.Printf("[client] gửi proxy handshake lỗi: %v", err)
		localConn.Close()
		srvConn.Close()
		return
	}

	go func() {
		defer atomic.AddInt64(&c.activeSessions, -1)
		proxyCopyCount(srvConn, localConn, &c.bytesUp)
	}()
	go proxyCopyCount(localConn, srvConn, &c.bytesDown)
}

func (c *client) handleUDPOpen(msg tunnel.Message) {
	if debugUDP {
		log.Printf("[client] handleUDPOpen called id=%s proto=%q", msg.ID, msg.Protocol)
	}
	if c.protocol != "udp" {
		return
	}
	if strings.TrimSpace(msg.ID) == "" {
		return
	}
	if msg.Protocol != "" && strings.ToLower(msg.Protocol) != "udp" {
		return
	}
	backend, err := c.resolveBackendUDP()
	if err != nil {
		log.Printf("[client] resolve backend UDP lỗi: %v", err)
		c.sendUDPClose(msg.ID)
		return
	}
	conn, err := net.DialUDP("udp", nil, backend)
	if err != nil {
		log.Printf("[client] không kết nối được backend UDP %s: %v", backend, err)
		c.sendUDPClose(msg.ID)
		return
	}
	sess := &udpClientSession{
		id:         msg.ID,
		conn:       conn,
		remoteAddr: strings.TrimSpace(msg.RemoteAddr),
		closed:     make(chan struct{}),
	}
	c.udpMu.Lock()
	if c.udpSessions == nil {
		c.udpSessions = make(map[string]*udpClientSession)
	}
	if old, ok := c.udpSessions[msg.ID]; ok {
		delete(c.udpSessions, msg.ID)
		old.Close()
	}
	c.udpSessions[msg.ID] = sess
	atomic.AddUint64(&c.totalUDPSessions, 1)
	c.udpMu.Unlock()
	go c.readFromUDPLocal(sess)
}

func (c *client) handleUDPClose(id string) {
	if strings.TrimSpace(id) == "" {
		return
	}
	c.removeUDPSession(id, false)
}

func (c *client) setupUDPChannel() error {
	addr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	_ = conn.SetReadBuffer(4 * 1024 * 1024)
	_ = conn.SetWriteBuffer(4 * 1024 * 1024)
	c.udpMu.Lock()
	if c.udpConn != nil {
		c.udpConn.Close()
	}
	c.udpConn = conn
	c.udpReady = false
	c.udpMu.Unlock()
	c.stopUDPPing()
	c.setUDPCtrlStatus("handshake")
	go c.readUDPControl(conn)
	for i := 0; i < 3; i++ {
		if err := c.sendUDPHandshake(); err != nil {
			log.Printf("[client] gửi UDP handshake burst #%d lỗi: %v", i+1, err)
		} else if debugUDP {
			log.Printf("[client] gửi UDP handshake burst #%d tới %s", i+1, addr)
		}
		if i < 2 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	go c.udpHandshakeRetry()
	return nil
}

func (c *client) readUDPControl(conn *net.UDPConn) {
	defer c.stopUDPPing()
	buf := make([]byte, 65535)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("[client] đọc UDP control lỗi: %v", err)
			}
			return
		}
		if n == 0 {
			continue
		}
		packet := make([]byte, n)
		copy(packet, buf[:n])
		c.handleUDPControlPacket(packet)
	}
}

func (c *client) handleUDPControlPacket(packet []byte) {
	if len(packet) < 3 {
		return
	}
	msgType := packet[0]
	key, idx, ok := decodeUDPField(packet, 1)
	if !ok || key == "" || key != c.key {
		return
	}
	switch msgType {
	case udpMsgData:
		id, next, ok := decodeUDPField(packet, idx)
		if !ok || id == "" {
			return
		}
		payload := make([]byte, len(packet)-next)
		copy(payload, packet[next:])
		c.handleUDPDataPacket(id, payload)
	case udpMsgClose:
		id, _, ok := decodeUDPField(packet, idx)
		if !ok || id == "" {
			return
		}
		c.handleUDPClose(id)
	case udpMsgHandshake:
		c.udpMu.Lock()
		if !c.udpReady && debugUDP {
			log.Printf("[client] UDP control handshake thành công từ %s", c.serverAddr)
		}
		c.udpReady = true
		c.udpMu.Unlock()
		c.startUDPPing()
	case udpMsgPong:
		_, next, ok := decodeUDPField(packet, idx)
		if !ok {
			return
		}
		payload := make([]byte, len(packet)-next)
		copy(payload, packet[next:])
		c.handleUDPPong(payload)
	case udpMsgPing:
		_, next, ok := decodeUDPField(packet, idx)
		if !ok {
			return
		}
		payload := make([]byte, len(packet)-next)
		copy(payload, packet[next:])
		c.sendUDPPong(payload)
	default:
	}
}

func (c *client) handleUDPDataPacket(id string, payload []byte) {
	if debugUDP {
		log.Printf("[client] handleUDPDataPacket id=%s len=%d", id, len(payload))
	}
	if len(payload) == 0 {
		return
	}

	// Decrypt if secret is available
	if c.udpSecret != nil {
		decrypted, err := tunnel.DecryptUDP(c.udpSecret, payload)
		if err != nil {
			if debugUDP {
				log.Printf("[client] giải mã UDP thất bại: %v", err)
			}
			return
		}
		payload = decrypted
	}

	sess := c.getUDPSession(id)
	if sess == nil {
		return
	}
	c.markServerData()
	if _, err := sess.conn.Write(payload); err != nil {
		log.Printf("[client] ghi về backend UDP lỗi: %v", err)
		c.removeUDPSession(id, true)
		return
	}
	c.startBackendWait(id)
	if debugUDP {
		log.Printf("[client] nhận %d bytes UDP từ server cho phiên %s", len(payload), id)
	}
	atomic.AddUint64(&c.bytesDown, uint64(len(payload)))
}

func (c *client) readFromUDPLocal(sess *udpClientSession) {
	buf := make([]byte, 65535)
	for {
		n, err := sess.conn.Read(buf)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("[client] đọc UDP backend lỗi: %v", err)
			}
			break
		}
		if n == 0 {
			continue
		}
		payload := make([]byte, n)
		copy(payload, buf[:n])
		c.cancelBackendWait(sess.id)
		c.markBackendData()
		c.sendUDPData(sess.id, payload)
	}
	c.removeUDPSession(sess.id, true)
}

func (c *client) resolveBackendUDP() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", c.localAddr)
}

func (c *client) getUDPSession(id string) *udpClientSession {
	c.udpMu.Lock()
	defer c.udpMu.Unlock()
	if c.udpSessions == nil {
		return nil
	}
	return c.udpSessions[id]
}

func (c *client) handleBackendTimeout(id string) {
	sess := c.getUDPSession(id)
	remote := ""
	if sess != nil {
		remote = sess.remoteAddr
	}
	if sess != nil {
		sess.idleCount++
		if sess.idleCount < backendIdleRetries {
			if debugUDP {
				log.Printf("[client] backend phiên %s (remote %s) chưa phản hồi (%d/%d)", id, remote, sess.idleCount, backendIdleRetries)
			}
			// restart timer
			c.startBackendWait(id)
			return
		}
	}
	log.Printf("[client] backend không phản hồi cho phiên %s (remote %s) - đóng phiên", id, remote)
	if c.enc != nil {
		_ = c.sendControlMessage(tunnel.Message{Type: "udp_idle", ID: id, Protocol: "udp"})
	}
	c.removeUDPSession(id, true)
}

func (c *client) removeUDPSession(id string, notify bool) {
	c.udpMu.Lock()
	sess := c.udpSessions[id]
	if sess != nil {
		delete(c.udpSessions, id)
	}
	c.udpMu.Unlock()
	if sess == nil {
		return
	}
	sess.Close()
	if notify {
		c.sendUDPClose(id)
	}
}

func (c *client) startBackendWait(id string) {
	c.udpMu.Lock()
	defer c.udpMu.Unlock()
	if sess, ok := c.udpSessions[id]; ok {
		if sess.timer != nil {
			sess.timer.Stop()
		}
		sess.idleCount = 0
		sess.timer = time.AfterFunc(backendIdleTimeout, func() {
			c.handleBackendTimeout(id)
		})
	}
}

func (c *client) cancelBackendWait(id string) {
	c.udpMu.Lock()
	defer c.udpMu.Unlock()
	if sess, ok := c.udpSessions[id]; ok && sess.timer != nil {
		sess.timer.Stop()
		sess.timer = nil
		sess.idleCount = 0
	}
}

func (c *client) markServerData() {
	c.dataMu.Lock()
	c.lastServerData = time.Now()
	c.dataMu.Unlock()
}

func (c *client) markBackendData() {
	c.dataMu.Lock()
	c.lastBackendData = time.Now()
	c.dataMu.Unlock()
}

func (c *client) getLastServerData() time.Time {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()
	return c.lastServerData
}

func (c *client) getLastBackendData() time.Time {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()
	return c.lastBackendData
}

func (c *client) closeAllUDPSessions() {
	c.udpMu.Lock()
	sessions := make([]*udpClientSession, 0, len(c.udpSessions))
	for _, sess := range c.udpSessions {
		if sess.timer != nil {
			sess.timer.Stop()
			sess.timer = nil
		}
		sessions = append(sessions, sess)
	}
	c.udpSessions = make(map[string]*udpClientSession)
	c.udpMu.Unlock()
	for _, sess := range sessions {
		sess.Close()
	}
}

func (c *client) sendUDPData(id string, payload []byte) {
	if len(payload) == 0 {
		return
	}

	// Encrypt if secret is available
	if c.udpSecret != nil {
		encrypted, err := tunnel.EncryptUDP(c.udpSecret, payload)
		if err != nil {
			log.Printf("[client] mã hóa UDP lỗi: %v", err)
			return
		}
		payload = encrypted
	}

	if err := c.writeUDP(udpMsgData, id, payload); err != nil {
		log.Printf("[client] gửi udp_data lỗi: %v", err)
		return
	}
	atomic.AddUint64(&c.bytesUp, uint64(len(payload)))
}

func (c *client) sendUDPClose(id string) {
	if err := c.writeUDP(udpMsgClose, id, nil); err != nil {
		log.Printf("[client] gửi udp_close lỗi: %v", err)
	}
	if c.enc != nil {
		_ = c.sendControlMessage(tunnel.Message{Type: "udp_close", ID: id, Protocol: "udp"})
	}
}

func (c *client) sendUDPHandshake() error {
	return c.writeUDP(udpMsgHandshake, "", nil)
}

func (c *client) sendUDPPing(payload []byte) error {
	return c.writeUDP(udpMsgPing, "", payload)
}

func (c *client) sendUDPPong(payload []byte) {
	if err := c.writeUDP(udpMsgPong, "", payload); err != nil && debugUDP {
		log.Printf("[client] gửi udp_pong lỗi: %v", err)
	}
}

func (c *client) udpHandshakeRetry() {
	const (
		retryInterval    = 500 * time.Millisecond
		handshakeTimeout = 10 * time.Second
		maxRetries       = 20
	)

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()

	attempts := 0
	for {
		c.udpMu.Lock()
		ready := c.udpReady
		connPresent := c.udpConn != nil
		c.udpMu.Unlock()
		if ready || !connPresent {
			if attempts > 0 && ready {
				log.Printf("[client] UDP handshake thành công sau %d lần thử (%d ms)", attempts+1, attempts*int(retryInterval/time.Millisecond))
			}
			return
		}
		select {
		case <-ticker.C:
			attempts++
			if attempts > maxRetries {
				log.Printf("[client] UDP handshake thất bại sau %d lần thử", maxRetries)
				c.udpMu.Lock()
				if c.udpConn != nil {
					c.udpConn.Close()
					c.udpConn = nil
				}
				c.udpMu.Unlock()
				c.setUDPCtrlStatus("offline")
				return
			}
			if err := c.sendUDPHandshake(); err != nil {
				if debugUDP {
					log.Printf("[client] retry handshake #%d lỗi: %v", attempts, err)
				}
			} else if debugUDP {
				log.Printf("[client] retry handshake #%d/%d", attempts, maxRetries)
			}
		case <-timeout.C:
			log.Printf("[client] UDP handshake timeout sau %v", handshakeTimeout)
			c.udpMu.Lock()
			if c.udpConn != nil {
				c.udpConn.Close()
				c.udpConn = nil
			}
			c.udpMu.Unlock()
			c.setUDPCtrlStatus("offline")
			return
		case <-c.done:
			return
		}
	}
}

func (c *client) writeUDP(msgType byte, id string, payload []byte) error {
	c.udpMu.Lock()
	conn := c.udpConn
	key := c.key
	ready := c.udpReady
	c.udpMu.Unlock()
	if conn == nil {
		return errors.New("udp chưa sẵn sàng")
	}
	if !ready && msgType != udpMsgHandshake && msgType != udpMsgPing {
		if debugUDP {
			log.Printf("[client] cảnh báo: gửi UDP khi handshake chưa hoàn tất (msg=%d)", msgType)
		}
	}
	buf := buildUDPMessage(msgType, key, id, payload)
	_, err := conn.Write(buf)
	if debugUDP && err == nil && msgType != udpMsgHandshake && !ready {
		log.Printf("[client] cảnh báo: gửi UDP nhưng handshake chưa được xác nhận")
	}
	return err
}

func (c *client) startUDPPing() {
	c.udpCtrlMu.Lock()
	if c.udpPingTicker != nil {
		c.udpCtrlMu.Unlock()
		return
	}
	ticker := time.NewTicker(udpControlInterval)
	stopCh := make(chan struct{})
	c.udpPingTicker = ticker
	c.udpPingStop = stopCh
	c.udpLastPong = time.Now()
	c.udpControlWarned = false
	c.udpCtrlMu.Unlock()
	c.setUDPCtrlStatus("pinging")
	go c.udpPingLoop(ticker, stopCh)
}

func (c *client) stopUDPPing() {
	c.udpCtrlMu.Lock()
	if c.udpPingTicker != nil {
		c.udpPingTicker.Stop()
		c.udpPingTicker = nil
	}
	if c.udpPingStop != nil {
		close(c.udpPingStop)
		c.udpPingStop = nil
	}
	c.udpControlWarned = false
	c.udpCtrlMu.Unlock()
}

func (c *client) udpPingLoop(ticker *time.Ticker, stopCh chan struct{}) {
	for {
		select {
		case <-ticker.C:
			ts := time.Now()
			payload := make([]byte, 8)
			binary.BigEndian.PutUint64(payload, uint64(ts.UnixNano()))
			c.udpCtrlMu.Lock()
			c.udpLastPing = ts
			c.udpCtrlMu.Unlock()
			if err := c.sendUDPPing(payload); err != nil && debugUDP {
				log.Printf("[client] gửi udp_ping lỗi: %v", err)
			}
			c.checkUDPPingTimeout()
		case <-stopCh:
			return
		case <-c.done:
			return
		}
	}
}

func (c *client) checkUDPPingTimeout() {
	c.udpCtrlMu.Lock()
	last := c.udpLastPong
	warned := c.udpControlWarned
	if time.Since(last) > udpControlTimeout {
		if !warned {
			c.udpControlWarned = true
			c.udpCtrlMu.Unlock()
			c.setUDPCtrlStatus("timeout")
			log.Printf("[client] UDP control timeout (>%v)", udpControlTimeout)
			return
		}
		c.udpCtrlMu.Unlock()
		return
	}
	if warned {
		c.udpControlWarned = false
	}
	c.udpCtrlMu.Unlock()
}

func (c *client) handleUDPPong(payload []byte) {
	if len(payload) < 8 {
		if debugUDP {
			log.Printf("[client] udp_pong payload quá ngắn")
		}
		return
	}
	sent := int64(binary.BigEndian.Uint64(payload))
	now := time.Now()
	rtt := time.Duration(now.UnixNano()-sent) * time.Nanosecond
	c.udpCtrlMu.Lock()
	c.udpLastPong = now
	c.udpControlWarned = false
	c.udpCtrlMu.Unlock()
	c.setUDPCtrlStatus(fmt.Sprintf("ok (%d ms)", rtt.Milliseconds()))
	if debugUDP {
		log.Printf("[client] nhận udp_pong, rtt %d ms", rtt.Milliseconds())
	}
}

func (c *client) setUDPCtrlStatus(status string) {
	c.udpCtrlMu.Lock()
	c.udpCtrlStatus = status
	c.udpCtrlMu.Unlock()
}

func (c *client) getUDPCtrlStatus() string {
	if strings.ToLower(c.protocol) != "udp" {
		return "n/a"
	}
	c.udpCtrlMu.Lock()
	status := c.udpCtrlStatus
	c.udpCtrlMu.Unlock()
	if status == "" {
		return "unknown"
	}
	return status
}

func (c *client) getSessionStats() (active int, total uint64) {
	activeTCP := atomic.LoadInt64(&c.activeSessions)
	totalTCP := atomic.LoadUint64(&c.totalSessions)

	c.udpMu.Lock()
	activeUDP := int64(len(c.udpSessions))
	c.udpMu.Unlock()
	totalUDP := atomic.LoadUint64(&c.totalUDPSessions)

	return int(activeTCP + activeUDP), totalTCP + totalUDP
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

func (c *client) reportProxyError(id string, err error) {
	if c.enc == nil {
		return
	}
	_ = c.sendControlMessage(tunnel.Message{
		Type:  "proxy_error",
		ID:    id,
		Error: err.Error(),
	})
}

func (c *client) closeControl() {
	c.closeOnce.Do(func() {
		close(c.done)
	})
	c.closeAllUDPSessions()
	c.stopUDPPing()
	c.setUDPCtrlStatus("offline")
	c.udpMu.Lock()
	if c.udpConn != nil {
		c.udpConn.Close()
		c.udpConn = nil
	}
	c.udpReady = false
	c.udpMu.Unlock()
	if c.control != nil {
		c.control.Close()
	}
	c.control = nil
	c.enc = nil
	c.dec = nil
	if c.trafficQuit != nil {
		close(c.trafficQuit)
		c.trafficQuit = nil
	}
	if c.statusCh != nil {
		close(c.statusCh)
		c.statusCh = nil
	}
	if c.pingCh != nil {
		close(c.pingCh)
		c.pingCh = nil
	}
}

func normalizedArgs(input []string) []string {
	filtered := make([]string, 0, len(input))
	for _, arg := range input {
		if arg == "" {
			continue
		}
		if arg == os.Args[0] || strings.HasSuffix(arg, "/"+filepath.Base(os.Args[0])) {
			continue
		}
		if strings.Contains(arg, "/") {
			// likely a path accidentally forwarded via shell wrapper
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}

func formatRate(delta uint64, interval time.Duration) string {
	if interval <= 0 {
		return formatBytes(delta)
	}
	perSecond := float64(delta) / interval.Seconds()
	return formatBytesFloat(perSecond)
}

func formatSince(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	d := time.Since(t)
	if d < time.Millisecond {
		return "just now"
	}
	if d < time.Second {
		return fmt.Sprintf("%d ms ago", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1f s ago", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1f m ago", d.Minutes())
	}
	return fmt.Sprintf("%.1f h ago", d.Hours())
}

func formatBytes(n uint64) string {
	return formatBytesFloat(float64(n))
}

func formatBytesFloat(value float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if value < 0 {
		value = 0
	}
	unit := 0
	for unit < len(units)-1 && value >= 1024 {
		value /= 1024
		unit++
	}
	switch {
	case value >= 100:
		return fmt.Sprintf("%.0f %s", value, units[unit])
	case value >= 10:
		return fmt.Sprintf("%.1f %s", value, units[unit])
	default:
		return fmt.Sprintf("%.2f %s", value, units[unit])
	}
}

type byteCounter struct {
	counter *uint64
}

func (b *byteCounter) Write(p []byte) (int, error) {
	if len(p) > 0 && b.counter != nil {
		atomic.AddUint64(b.counter, uint64(len(p)))
	}
	return len(p), nil
}

func proxyCopyCount(dst, src net.Conn, counter *uint64) {
	defer dst.Close()
	defer src.Close()
	reader := io.TeeReader(src, &byteCounter{counter: counter})
	_, _ = io.Copy(dst, reader)
}

func (c *client) recordPingReply() {
	sent := atomic.SwapInt64(&c.pingSent, 0)
	if sent <= 0 {
		return
	}
	ms := time.Since(time.Unix(0, sent))
	c.rttMu.Lock()
	c.rtt = ms
	c.rttMu.Unlock()
	atomic.StoreInt64(&c.pingMs, ms.Milliseconds())
	if c.pingCh == nil {
		return
	}
	select {
	case c.pingCh <- ms:
	default:
		select {
		case <-c.pingCh:
		default:
		}
		select {
		case c.pingCh <- ms:
		default:
		}
	}
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func formatPingDisplay(d time.Duration) (string, string) {
	if d < 0 {
		return "N/A", "[----]"
	}
	ms := d.Milliseconds()
	var bars string
	switch {
	case ms <= 50:
		bars = "[||||]"
	case ms <= 120:
		bars = "[||| ]"
	case ms <= 250:
		bars = "[||  ]"
	case ms <= 500:
		bars = "[|   ]"
	default:
		bars = "[    ]"
	}
	return fmt.Sprintf("%d ms", ms), bars
}

func (c *client) renderFrame(stats trafficStats, ping time.Duration) {
	activeSessions, totalSessions := c.getSessionStats()

	// ANSI colors
	const (
		reset       = "\033[0m"
		bold        = "\033[1m"
		cyan        = "\033[36m"
		green       = "\033[32m"
		yellow      = "\033[33m"
		red         = "\033[31m"
		magenta     = "\033[35m"
		blue        = "\033[34m"
		brightCyan  = "\033[96m"
		brightGreen = "\033[92m"
	)

	// Status emoji and color
	statusEmoji := "🟢"
	statusColor := green
	statusText := "ACTIVE"
	if c.publicHost == "" || c.publicHost == "pending..." {
		statusEmoji = "🟡"
		statusColor = yellow
		statusText = "CONNECTING"
	}

	pingText, bars := formatPingDisplay(ping)
	pingColor := green
	// Status line special case for emoji
	// Status line special case for emoji
	statusLine := func() string {
		now := time.Now().Format("15:04:05")
		return fmt.Sprintf(bold+brightCyan+"║"+reset+"  %s Status   : %s%s%s (%s)", statusEmoji, statusColor, bold, statusText, now)
	}

	// Helper to create a row with an emoji label
	makeRow := func(emoji, label, val, color string) string {
		// "  emoji Label    : Value"
		// Align colon at specific column?
		// "  🔗 Local    : " -> 16 chars

		prefixVisible := 16
		currentPrefix := 2 + 2 + 1 + len(label) + 2
		padLabel := prefixVisible - currentPrefix
		if padLabel < 0 {
			padLabel = 0
		}

		labelStr := label + strings.Repeat(" ", padLabel)

		return fmt.Sprintf(bold+brightCyan+"║"+reset+"  %s %s : %s%s%s", emoji, labelStr, color, val, reset)
	}

	lines := []string{
		bold + brightCyan + "╔══════════════════════════════════════════════════════",
		bold + brightCyan + "║" + reset + bold + "      TrongDev | ProxVN - Tunnel Việt Nam Free",
		bold + brightCyan + "╠══════════════════════════════════════════════════════",
		statusLine(),
		makeRow("🔗", "Local", c.localAddr, cyan),
		func() string {
			displayHost := nonEmpty(c.publicHost, "pending...")
			if c.protocol == "http" && c.subdomain != "" {
				domain := c.baseDomain
				if domain == "" {
					domain = "bacsycay.click" // Fallback default
				}
				displayHost = fmt.Sprintf("https://%s.%s", c.subdomain, domain)
			}
			return makeRow("🌐", "Public", displayHost, brightGreen+bold)
		}(),
		makeRow("📡", "Protocol", strings.ToUpper(nonEmpty(c.protocol, "tcp")), magenta),
		bold + brightCyan + "╠══════════════════════════════════════════════════════",
		func() string {
			v1 := fmt.Sprintf("⬆️  %s%s/s%s", green, stats.upRate, reset)
			v2 := fmt.Sprintf("⬇️  %s%s/s%s", blue, stats.downRate, reset)
			return fmt.Sprintf(bold+brightCyan+"║"+reset+"  📊 Traffic  : %s %s", v1, v2)
		}(),
		func() string {
			return fmt.Sprintf(bold+brightCyan+"║"+reset+"  📈 Total    : %s%s%s ↑  %s%s%s ↓", cyan, stats.totalUp, reset, cyan, stats.totalDown, reset)
		}(),
		func() string {
			ac := strconv.Itoa(activeSessions)
			to := strconv.FormatUint(totalSessions, 10)
			return fmt.Sprintf(bold+brightCyan+"║"+reset+"  🔌 Sessions : active %s%s%s | total %s%s%s", yellow, ac, reset, cyan, to, reset)
		}(),
		func() string {
			return fmt.Sprintf(bold+brightCyan+"║"+reset+"  🏓 Ping     : %s%s %s%s", pingColor, pingText, bars, reset)
		}(),
		makeRow("🔐", "Key", nonEmpty(c.key, "(none)"), yellow),
		makeRow("⚙️", "Version", tunnel.Version, magenta),
		bold + brightCyan + "╚══════════════════════════════════════════════════════",
		"",
		cyan + "  Press 'q' or ESC to quit" + reset,
	}

	if c.uiEnabled {
		var builder strings.Builder

		// Move cursor to top-left (Home)
		builder.WriteString("\033[H")

		// Write all lines
		for _, line := range lines {
			builder.WriteString(line)
			builder.WriteByte('\n')
		}

		// Clear from cursor to end of screen (cleans up any partial leftovers from prev frame)
		builder.WriteString("\033[J")

		// Print everything in one go to minimize tearing/scrolling artifacts
		fmt.Print(builder.String())
	}
}

func terminalSize() (int, int) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			return width, height
		}
	}
	return 80, 24
}

func isEOF(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection")
}

func (c *client) sendControlMessage(msg tunnel.Message) error {
	if c.state.Get() != tunnel.StateActive {
		c.ctrlQueue.Push(msg)
		log.Printf("[client] Control plane offline. Message queued: %s", msg.Type)
		return nil
	}
	if c.enc == nil {
		c.ctrlQueue.Push(msg)
		return fmt.Errorf("control plane encoder not initialized")
	}
	if err := c.enc.Encode(msg); err != nil {
		c.ctrlQueue.Push(msg)
		return err
	}
	return nil
}

func (c *client) replayQueuedMessages() {
	messages := c.ctrlQueue.PopAll()
	if len(messages) == 0 {
		return
	}
	log.Printf("[client] Replaying %d queued control messages...", len(messages))
	for _, msg := range messages {
		if err := c.sendControlMessage(msg); err != nil {
			log.Printf("[client] Failed to replay message %s: %v", msg.Type, err)
			break
		}
	}
}
