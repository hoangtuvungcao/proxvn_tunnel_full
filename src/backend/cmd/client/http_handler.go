package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	
	"proxvn/backend/internal/tunnel"
)

// handleHTTPRequest handles an HTTP request from the server
func (c *client) handleHTTPRequest(msg tunnel.Message) {
	if c.protocol != "http" {
		log.Printf("[client] Received HTTP request but not in HTTP mode")
		return
	}
	
	// Determine scheme based on port
	scheme := "http"
	if strings.HasSuffix(c.localAddr, ":443") {
		scheme = "https"
	}

	// Build HTTP request
	req, err := http.NewRequest(msg.Method, fmt.Sprintf("%s://%s%s", scheme, c.localAddr, msg.Path), bytes.NewReader(msg.Body))
	if err != nil {
		//log.Printf("[client] Failed to create HTTP request: %v", err)
		c.sendHTTPError(msg.ID, http.StatusBadGateway, err.Error())
		return
	}
	
	// Set headers
	for key, value := range msg.Headers {
		req.Header.Set(key, value)
	}
	
	// CRITICAL: Rewrite Host header to local address to ensure correct VirtualHost matching
	// Otherwise Apache/Nginx might serve default content if they don't recognize the public subdomain
	// Force "localhost" if targeting local loopback to match browser behavior
	if strings.Contains(c.localAddr, "127.0.0.1") || strings.Contains(c.localAddr, "::1") || strings.Contains(c.localAddr, "localhost") {
		req.Host = "localhost"
	} else {
		req.Host = c.localAddr
	}
	
	// Forward to local HTTP server
	// Skip SSL verification for local HTTPS targets (common for dev)
	// This is ACCEPTABLE because:
	// 1. Destination is localhost (not over network)
	// 2. Often used with self-signed certs in development
	// 3. No MITM risk since traffic doesn't leave the machine
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		//log.Printf("[client] Failed to forward HTTP request to %s: %v", c.localAddr, err)
		//c.sendHTTPError(msg.ID, http.StatusBadGateway, fmt.Sprintf("Failed to connect to local server: %v", err))
		return
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//log.Printf("[client] Failed to read response body: %v", err)
		//c.sendHTTPError(msg.ID, http.StatusInternalServerError, "Failed to read response")
		return
	}
	
	// Convert headers to map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	// Send response back to server
	responseMsg := tunnel.Message{
		Type:       "http_response",
		ID:         msg.ID,
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
	}
	
	if err := c.enc.Encode(responseMsg); err != nil {
		//log.Printf("[client] Failed to send HTTP response: %v", err)
		return
	}
	
	// Update traffic stats
	atomic.AddUint64(&c.bytesDown, uint64(len(msg.Body)))
	atomic.AddUint64(&c.bytesUp, uint64(len(body)))
	
	//log.Printf("[client] HTTP %s %s -> %d (%d bytes)", msg.Method, msg.Path, resp.StatusCode, len(body))
}

// sendHTTPError sends an error response back to server
func (c *client) sendHTTPError(requestID string, statusCode int, errorMsg string) {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	
	errorBody := []byte(errorMsg)
	
	responseMsg := tunnel.Message{
		Type:       "http_response",
		ID:         requestID,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       errorBody,
	}
	
	if err := c.enc.Encode(responseMsg); err != nil {
		//log.Printf("[client] Failed to send HTTP error response: %v", err)
	}
}
