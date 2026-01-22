package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CacheMiddleware creates a caching middleware for Gin
func CacheMiddleware(cache *Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip cache for authenticated requests
		if c.GetHeader("Authorization") != "" {
			c.Next()
			return
		}

		// Generate cache key from request
		cacheKey := generateCacheKey(c.Request)

		// Try to get from cache
		if cached, found := cache.Get(cacheKey); found {
			if response, ok := cached.(*CachedResponse); ok {
				// Serve from cache
				for key, values := range response.Headers {
					for _, value := range values {
						c.Header(key, value)
					}
				}
				c.Header("X-Cache", "HIT")
				c.Data(response.StatusCode, response.ContentType, response.Body)
				c.Abort()
				return
			}
		}

		// Cache miss - capture response
		c.Header("X-Cache", "MISS")

		// Use custom writer to capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
			body:           make([]byte, 0),
		}
		c.Writer = writer

		c.Next()

		// Cache the response if successful
		if writer.statusCode == http.StatusOK && len(writer.body) > 0 {
			cached := &CachedResponse{
				StatusCode:  writer.statusCode,
				Headers:     writer.Header(),
				Body:        writer.body,
				ContentType: writer.Header().Get("Content-Type"),
			}
			cache.Set(cacheKey, cached)
		}
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ContentType string
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// generateCacheKey creates a unique cache key from request
func generateCacheKey(req *http.Request) string {
	h := sha256.New()
	io.WriteString(h, req.Method)
	io.WriteString(h, req.URL.String())
	io.WriteString(h, req.Header.Get("Accept"))
	io.WriteString(h, req.Header.Get("Accept-Encoding"))

	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("http:%s", hash[:16])
}
