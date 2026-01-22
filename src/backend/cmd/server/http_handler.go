package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"proxvn/backend/internal/config"
	httpproxy "proxvn/backend/internal/http"
	"proxvn/backend/internal/tunnel"
)

// Implement clientSession's interface methods for HTTP proxy

// GetID returns the client ID
func (session *clientSession) GetID() string {
	return session.clientID
}

// GetKey returns the client key
func (session *clientSession) GetKey() string {
	return session.key
}

// SendHTTPRequest sends an HTTP request through the tunnel and waits for response
func (session *clientSession) SendHTTPRequest(req *httpproxy.HTTPRequest) (*httpproxy.HTTPResponse, error) {
	// Create response channel
	respCh := make(chan *httpproxy.HTTPResponse, 1)

	// Register response channel in server's map
	session.server.httpReqMu.Lock()
	session.server.httpRequests[req.ID] = respCh
	session.server.httpReqMu.Unlock()

	// Clean up on return
	defer func() {
		session.server.httpReqMu.Lock()
		delete(session.server.httpRequests, req.ID)
		session.server.httpReqMu.Unlock()
	}()

	// Send HTTP request message through tunnel
	msg := tunnel.Message{
		Type:    "http_request",
		ID:      req.ID,
		Method:  req.Method,
		Path:    req.Path,
		Headers: req.Headers,
		Body:    req.Body,
	}

	if err := session.enc.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	// Wait for response (with timeout)
	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("HTTP request timeout")
	case <-session.done:
		return nil, fmt.Errorf("client disconnected")
	}
}

// initHTTPProxy initializes the HTTP proxy server
func (s *server) initHTTPProxy(cfg *config.Config) {
	// Check if HTTP domain is configured
	httpDomain := strings.TrimSpace(cfg.Server.HTTPDomain)
	if httpDomain == "" {
		log.Printf("[http] HTTP_DOMAIN not configured - HTTP tunneling disabled")
		log.Printf("[http] HTTP clients will fallback to IP:port mode")
		log.Printf("[http] To enable: set HTTP_DOMAIN environment variable (e.g., HTTP_DOMAIN=vutrungocrong.fun)")
		return
	}

	// Try to load SSL certificate
	certManager := httpproxy.NewCertManager()

	if err := certManager.LoadCertificate(); err != nil {
		log.Printf("[http] Failed to load SSL certificate: %v", err)
		log.Printf("[http] HTTP tunneling disabled - wildcard SSL cert required for *.%s", httpDomain)
		log.Printf("[http] HTTP clients will fallback to IP:port mode")
		return
	}

	certFile, keyFile, err := certManager.GetCertFiles()
	if err != nil {
		log.Printf("[http] Certificate error: %v", err)
		return
	}

	// Create HTTP proxy server with configured domain
	httpPort := cfg.Server.HTTPPort
	s.httpProxy = httpproxy.NewHTTPProxyServer(certFile, keyFile, httpDomain, httpPort)

	// Set dashboard target to internal listener port
	s.httpProxy.SetDashboardTarget(s.listenPort)

	log.Printf("[http] %s", certManager.GetCertificateInfo())
	log.Printf("[http] HTTP Domain: *.%s", httpDomain)

	// Start proxy server
	// Use blocking check to ensure port is available
	// Since Start() is blocking, we should ideally check Listen first, or rely on Start returning error fast.
	// But p.Start() calls ListenAndServeTLS, which blocks.

	go func() {
		if err := s.httpProxy.Start(); err != nil {
			log.Fatalf("[http] Failed to start HTTPS proxy on port %d: %v (Is the port already in use by Apache/Nginx?)", cfg.Server.HTTPPort, err)
		}
	}()

	// Give it a moment to fail startup if port is busy
	time.Sleep(100 * time.Millisecond)
}

// registerHTTPClient registers a client for HTTP tunneling
func (s *server) registerHTTPClient(session *clientSession) error {
	if s.httpProxy == nil {
		return fmt.Errorf("HTTP proxy not available")
	}

	// Generate unique subdomain
	subdomain, err := httpproxy.GenerateSubdomain()
	if err != nil {
		return fmt.Errorf("failed to generate subdomain: %w", err)
	}

	// Register with HTTP proxy
	if err := s.httpProxy.RegisterClient(subdomain, session); err != nil {
		return err
	}

	session.subdomain = subdomain
	return nil
}

// unregisterHTTPClient removes a client from HTTP tunneling
func (s *server) unregisterHTTPClient(session *clientSession) {
	if s.httpProxy != nil && session.subdomain != "" {
		s.httpProxy.UnregisterClient(session.subdomain)
	}
}

// handleHTTPResponse handles HTTP response from client
func (s *server) handleHTTPResponse(msg tunnel.Message) {
	s.httpReqMu.Lock()
	respCh, exists := s.httpRequests[msg.ID]
	if exists {
		delete(s.httpRequests, msg.ID)
	}
	s.httpReqMu.Unlock()

	if !exists {
		log.Printf("[http] Received response for unknown request ID: %s", msg.ID)
		return
	}

	response := &httpproxy.HTTPResponse{
		StatusCode: msg.StatusCode,
		Headers:    msg.Headers,
		Body:       msg.Body,
	}

	select {
	case respCh <- response:
	default:
	}
}
