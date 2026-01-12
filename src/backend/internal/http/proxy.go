package http

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	subdomainLength = 6
	readTimeout     = 30 * time.Second
	writeTimeout    = 30 * time.Second
)

// ClientSession represents a connected client with its subdomain
type ClientSession interface {
	GetID() string
	GetKey() string
	SendHTTPRequest(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPProxyServer handles HTTPS requests and routes them to appropriate tunnels
type HTTPProxyServer struct {
	mu         sync.RWMutex
	subdomains map[string]ClientSession // subdomain -> client session
	httpServer *http.Server
	certFile   string
	keyFile    string
	baseDomain string // Configurable base domain
	httpPort   int    // Configurable HTTP port
	landingDir string // Landing page directory
	binDir     string // Bin directory for downloads
}

// HTTPRequest represents an HTTP request to be sent through the tunnel
type HTTPRequest struct {
	ID      string
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// HTTPResponse represents an HTTP response from the tunnel
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// NewHTTPProxyServer creates a new HTTP proxy server
func NewHTTPProxyServer(certFile, keyFile, baseDomain string, httpPort int) *HTTPProxyServer {
	// Determine landing page directory
	landingDir := "../frontend/landing"
	if _, err := os.Stat("frontend/landing"); err == nil {
		landingDir = "frontend/landing"
	}

	// Determine bin directory
	binDir := "bin"
	if _, err := os.Stat("bin"); os.IsNotExist(err) {
		binDir = "."
	}

	return &HTTPProxyServer{
		subdomains: make(map[string]ClientSession),
		certFile:   certFile,
		keyFile:    keyFile,
		baseDomain: baseDomain,
		httpPort:   httpPort,
		landingDir: landingDir,
		binDir:     binDir,
	}
}

// GenerateSubdomain creates a random subdomain string
func GenerateSubdomain() (string, error) {
	bytes := make([]byte, subdomainLength/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex and take only alphanumeric characters
	subdomain := hex.EncodeToString(bytes)[:subdomainLength]
	return subdomain, nil
}

// RegisterClient registers a client with a subdomain
func (p *HTTPProxyServer) RegisterClient(subdomain string, client ClientSession) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.subdomains[subdomain]; exists {
		return fmt.Errorf("subdomain %s already registered", subdomain)
	}

	p.subdomains[subdomain] = client
	log.Printf("[http] Registered subdomain: %s.%s for client %s", subdomain, p.baseDomain, client.GetID())
	return nil
}

// UnregisterClient removes a client's subdomain
func (p *HTTPProxyServer) UnregisterClient(subdomain string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.subdomains[subdomain]; exists {
		delete(p.subdomains, subdomain)
		log.Printf("[http] Unregistered subdomain: %s.%s for client %s", subdomain, p.baseDomain, client.GetID())
	}
}

// Start starts the HTTPS proxy server on configured port
func (p *HTTPProxyServer) Start() error {
	handler := http.HandlerFunc(p.handleRequest)

	p.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", p.httpPort),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	log.Printf("[http] Starting HTTPS proxy server on port %d", p.httpPort)
	log.Printf("[http] Base domain: *.%s", p.baseDomain)

	return p.httpServer.ListenAndServeTLS(p.certFile, p.keyFile)
}

// Stop gracefully stops the server
func (p *HTTPProxyServer) Stop() error {
	if p.httpServer != nil {
		return p.httpServer.Close()
	}
	return nil
}

// handleRequest handles incoming HTTP requests and routes them to the appropriate client
func (p *HTTPProxyServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Check if this is the main domain (landing page)
	if p.isMainDomain(r.Host) {
		p.serveLandingPage(w, r)
		return
	}

	// Parse subdomain from Host header
	subdomain := p.extractSubdomain(r.Host)
	if subdomain == "" {
		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
		return
	}

	// Look up client session
	p.mu.RLock()
	client, exists := p.subdomains[subdomain]
	p.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Tunnel not found for subdomain: %s", subdomain), http.StatusBadGateway)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[http] Failed to read request body: %v", err)
		http.Error(w, "Failed to read request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Convert headers to map
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Create HTTP request for tunnel
	tunnelReq := &HTTPRequest{
		ID:      generateRequestID(),
		Method:  r.Method,
		Path:    r.URL.String(),
		Headers: headers,
		Body:    body,
	}

	// Send request through tunnel
	response, err := client.SendHTTPRequest(tunnelReq)
	if err != nil {
		log.Printf("[http] Failed to send request through tunnel: %v", err)
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}

	// Write response headers
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// Write status code
	w.WriteHeader(response.StatusCode)

	// Write response body
	if _, err := w.Write(response.Body); err != nil {
		log.Printf("[http] Failed to write response: %v", err)
	}

	//log.Printf("[http] %s %s %s -> %d (%s)", r.Method, subdomain, r.URL.Path, response.StatusCode, client.GetID())
}

// isMainDomain checks if the host is the main domain (without subdomain)
func (p *HTTPProxyServer) isMainDomain(host string) bool {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx > 0 {
		host = host[:idx]
	}

	// Check if it's the main domain or www subdomain
	return host == p.baseDomain || host == "www."+p.baseDomain
}

// serveLandingPage serves the landing page for the main domain
func (p *HTTPProxyServer) serveLandingPage(w http.ResponseWriter, r *http.Request) {
	// Handle downloads
	if strings.HasPrefix(r.URL.Path, "/downloads/") {
		filename := strings.TrimPrefix(r.URL.Path, "/downloads/")
		filePath := filepath.Join(p.binDir, filename)

		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}
		http.NotFound(w, r)
		return
	}

	// Serve landing page static files
	switch r.URL.Path {
	case "/", "/index.html":
		http.ServeFile(w, r, filepath.Join(p.landingDir, "index.html"))
	case "/style.css":
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, filepath.Join(p.landingDir, "style.css"))
	case "/script.js":
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, filepath.Join(p.landingDir, "script.js"))
	default:
		// Try to serve from landing directory
		filePath := filepath.Join(p.landingDir, r.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
		} else {
			http.NotFound(w, r)
		}
	}
}

// extractSubdomain extracts the subdomain from the Host header
func (p *HTTPProxyServer) extractSubdomain(host string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx > 0 {
		host = host[:idx]
	}

	// Check if it matches our domain
	suffix := "." + p.baseDomain
	if !strings.HasSuffix(host, suffix) {
		return ""
	}

	// Extract subdomain
	subdomain := strings.TrimSuffix(host, suffix)

	// Validate subdomain (only alphanumeric and hyphens)
	for _, ch := range subdomain {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return ""
		}
	}

	return subdomain
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ReverseProxy creates a reverse proxy that forwards requests to a local address
// This is an alternative simpler implementation for direct HTTP proxying
func (p *HTTPProxyServer) CreateReverseProxy(targetURL string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   targetURL,
	})

	return proxy
}

// GetSubdomainForClient returns the subdomain for a given client, or empty if not found
func (p *HTTPProxyServer) GetSubdomainForClient(clientID string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for subdomain, client := range p.subdomains {
		if client.GetID() == clientID {
			return subdomain
		}
	}
	return ""
}

// GetActiveSubdomains returns a list of all active subdomains
func (p *HTTPProxyServer) GetActiveSubdomains() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	subdomains := make([]string, 0, len(p.subdomains))
	for subdomain := range p.subdomains {
		subdomains = append(subdomains, subdomain)
	}
	return subdomains
}

// ValidateSubdomain checks if a subdomain is valid
func ValidateSubdomain(subdomain string) bool {
	if len(subdomain) == 0 || len(subdomain) > 63 {
		return false
	}

	for _, ch := range subdomain {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}

	// Cannot start or end with hyphen
	return subdomain[0] != '-' && subdomain[len(subdomain)-1] != '-'
}

// GetFullDomain returns the full domain for a subdomain
func (p *HTTPProxyServer) GetFullDomain(subdomain string) string {
	return fmt.Sprintf("%s.%s", subdomain, p.baseDomain)
}

// GetBaseDomain returns the base domain
func (p *HTTPProxyServer) GetBaseDomain() string {
	return p.baseDomain
}
