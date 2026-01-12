package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"proxvn/backend/internal/auth"
	"proxvn/backend/internal/database"
	"proxvn/backend/internal/models"
)

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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Simple in-memory rate limiter
type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

type visitor struct {
	lastSeen time.Time
	tokens   float64
}

var limiter = &rateLimiter{
	visitors: make(map[string]*visitor),
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			lastSeen: time.Now(),
			tokens:   10, // Burst size
		}
		return true
	}

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now

	// Replenish 1 token per second
	v.tokens += elapsed
	if v.tokens > 10 {
		v.tokens = 10
	}

	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Error:   "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Internal server error",
		})
	})
}
