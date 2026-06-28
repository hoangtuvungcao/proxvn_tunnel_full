package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"proxvn/backend/internal/auth"
	"proxvn/backend/internal/database"
	"proxvn/backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// AuthMiddleware xác thực JWT token
func AuthMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Authorization required",
			})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// APIKeyMiddleware xác thực API key
func APIKeyMiddleware(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "API key required",
			})
			c.Abort()
			return
		}

		user, err := db.GetUserByAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID.String())
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Set("api_key", apiKey)
		c.Next()
	}
}

// AdminMiddleware kiểm tra quyền admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORSMiddleware với cấu hình CORS bảo mật
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		if origin != "" {
			if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://localhost") {
				allowed = true
			} else {
				httpDomain := os.Getenv("HTTP_DOMAIN")
				if httpDomain != "" {
					cleanDomain := strings.TrimPrefix(httpDomain, ".")
					if strings.HasSuffix(origin, cleanDomain) {
						allowed = true
					}
				}
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Credentials", "false")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key, X-Requested-With, X-Timestamp")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware thêm security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HSTS - Force HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS Protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// Advanced rate limiter với token bucket algorithm
type advancedRateLimiter struct {
	visitors map[string]*visitorInfo
	mu       sync.RWMutex
	cleanup  *time.Ticker
}

type visitorInfo struct {
	limiter    *rate.Limiter
	lastSeen   time.Time
	attempts   int // Failed login attempts
	blocked    bool
	blockUntil time.Time
}

var advLimiter *advancedRateLimiter

func init() {
	advLimiter = &advancedRateLimiter{
		visitors: make(map[string]*visitorInfo),
		cleanup:  time.NewTicker(5 * time.Minute),
	}

	// Cleanup old visitors
	go func() {
		for range advLimiter.cleanup.C {
			advLimiter.cleanupOldVisitors()
		}
	}()
}

func (rl *advancedRateLimiter) cleanupOldVisitors() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-30 * time.Minute)
	for ip, info := range rl.visitors {
		if info.lastSeen.Before(cutoff) && !info.blocked {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *advancedRateLimiter) getVisitor(ip string, rps int, burst int) *visitorInfo {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(rps), burst)
		v = &visitorInfo{
			limiter:  limiter,
			lastSeen: time.Now(),
			attempts: 0,
			blocked:  false,
		}
		rl.visitors[ip] = v
	}

	v.lastSeen = time.Now()
	return v
}

func (rl *advancedRateLimiter) blockIP(ip string, duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v, exists := rl.visitors[ip]; exists {
		v.blocked = true
		v.blockUntil = time.Now().Add(duration)
	} else {
		limiter := rate.NewLimiter(rate.Limit(1), 1)
		rl.visitors[ip] = &visitorInfo{
			limiter:    limiter,
			lastSeen:   time.Now(),
			blocked:    true,
			blockUntil: time.Now().Add(duration),
		}
	}
}

func (rl *advancedRateLimiter) isBlocked(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if v, exists := rl.visitors[ip]; exists {
		if v.blocked {
			if time.Now().Before(v.blockUntil) {
				return true
			}
			// Unblock
			v.blocked = false
			v.attempts = 0
		}
	}
	return false
}

// RateLimitMiddleware với token bucket algorithm
func RateLimitMiddleware(rps, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Check if IP is blocked
		if advLimiter.isBlocked(ip) {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Error:   "IP temporarily blocked due to excessive requests",
			})
			c.Abort()
			return
		}

		visitor := advLimiter.getVisitor(ip, rps, burst)

		if !visitor.limiter.Allow() {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Error:   "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BruteForceProtectionMiddleware chống brute force attack
func BruteForceProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		c.Next()

		// Check if login failed
		if c.Writer.Status() == http.StatusUnauthorized {
			advLimiter.mu.Lock()
			if v, exists := advLimiter.visitors[ip]; exists {
				v.attempts++

				// Block after 5 failed attempts
				if v.attempts >= 5 {
					v.blocked = true
					v.blockUntil = time.Now().Add(15 * time.Minute)
				}
			}
			advLimiter.mu.Unlock()
		} else if c.Writer.Status() == http.StatusOK {
			// Reset attempts on successful login
			advLimiter.mu.Lock()
			if v, exists := advLimiter.visitors[ip]; exists {
				v.attempts = 0
			}
			advLimiter.mu.Unlock()
		}
	}
}

// RequestSizeLimitMiddleware giới hạn kích thước request
func RequestSizeLimitMiddleware(maxSizeMB int) gin.HandlerFunc {
	maxSize := int64(maxSizeMB * 1024 * 1024)
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// LoggingMiddleware với format tùy chỉnh
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.ErrorMessage,
		)
	})
}

// RecoveryMiddleware với custom error handling
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Internal server error",
		})
	})
}

// TimestampValidationMiddleware ngăn chặn replay attacks bằng cách kiểm tra clock skew
func TimestampValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			tsHeader := c.GetHeader("X-Timestamp")
			if tsHeader != "" {
				ts, err := time.Parse(time.RFC3339, tsHeader)
				if err == nil {
					diff := time.Since(ts)
					if diff < 0 {
						diff = -diff
					}
					if diff > 5*time.Minute {
						c.JSON(http.StatusBadRequest, models.APIResponse{
							Success: false,
							Error:   "Request timestamp skew too large (> 5 minutes)",
						})
						c.Abort()
						return
					}
				} else {
					c.JSON(http.StatusBadRequest, models.APIResponse{
						Success: false,
						Error:   "Invalid X-Timestamp header format (RFC3339 required)",
					})
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// GzipAndCacheMiddleware nén gzip và thiết lập Cache-Control cho tài nguyên tĩnh
func GzipAndCacheMiddleware() gin.HandlerFunc {
	var gzipPool = sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
			return w
		},
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		isAsset := strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".png") ||
			strings.HasSuffix(path, ".jpg") ||
			strings.HasSuffix(path, ".svg") ||
			strings.HasSuffix(path, ".woff") ||
			strings.HasSuffix(path, ".woff2")

		if isAsset {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") ||
			c.GetHeader("Content-Encoding") != "" {
			c.Next()
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(c.Writer)
		defer func() {
			gz.Close()
			gzipPool.Put(gz)
		}()

		c.Writer = &gzipWriter{ResponseWriter: c.Writer, writer: gz}
		c.Next()
	}
}
