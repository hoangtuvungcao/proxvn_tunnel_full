package main

import (
	"time"

	"golang.org/x/time/rate"
)

// getRateLimiter gets or creates a rate limiter for a client IP
func (s *server) getRateLimiter(clientIP string) *rateLimiter {
	s.rateLimitersMu.Lock()
	defer s.rateLimitersMu.Unlock()

	limiter, exists := s.rateLimiters[clientIP]
	if !exists {
		limiter = &rateLimiter{
			// Registration: 5 attempts per minute, burst of 10
			registrations: rate.NewLimiter(rate.Every(12*time.Second), 10),
			// HTTP Requests: 100 requests per second, burst of 200
			httpRequests: rate.NewLimiter(100, 200),
			// UDP Sessions: 50 new sessions per minute, burst of 100
			udpSessions: rate.NewLimiter(rate.Every(1200*time.Millisecond), 100),
			lastSeen:    time.Now(),
		}
		s.rateLimiters[clientIP] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter
}

// checkRegistrationRateLimit returns true if allowed, false if rate limited
func (s *server) checkRegistrationRateLimit(clientIP string) bool {
	limiter := s.getRateLimiter(clientIP)
	return limiter.registrations.Allow()
}

// checkHTTPRequestRateLimit returns true if allowed, false if rate limited
func (s *server) checkHTTPRequestRateLimit(clientIP string) bool {
	limiter := s.getRateLimiter(clientIP)
	return limiter.httpRequests.Allow()
}

// checkUDPSessionRateLimit returns true if allowed, false if rate limited  
func (s *server) checkUDPSessionRateLimit(clientIP string) bool {
	limiter := s.getRateLimiter(clientIP)
	return limiter.udpSessions.Allow()
}

// cleanupRateLimiters removes old rate limiters every 10 minutes
func (s *server) cleanupRateLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.rateLimitersMu.Lock()
		now := time.Now()
		for ip, limiter := range s.rateLimiters {
			// Remove limiters not seen in last hour
			if now.Sub(limiter.lastSeen) > time.Hour {
				delete(s.rateLimiters, ip)
			}
		}
		s.rateLimitersMu.Unlock()
	}
}
