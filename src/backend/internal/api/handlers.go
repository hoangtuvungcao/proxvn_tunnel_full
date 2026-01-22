package api

import (
	"log"
	"net/http"
	"time"

	"proxvn/backend/internal/auth"
	"proxvn/backend/internal/database"
	"proxvn/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var startTime = time.Now()

type Handler struct {
	db          *database.Database
	authService *auth.AuthService
}

func NewHandler(db *database.Database, authService *auth.AuthService) *Handler {
	return &Handler{
		db:          db,
		authService: authService,
	}
}

// ============================================
// AUTH HANDLERS
// ============================================

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"token":    token,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to process password",
		})
		return
	}

	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     models.UserRoleUser,
		APIKey:   uuid.New().String(),
	}

	if err := h.db.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Username or email already exists",
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"api_key":  user.APIKey,
		},
	})
}

func (h *Handler) GetProfile(c *gin.Context) {
	username := c.GetString("username")

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"api_key":  user.APIKey,
		},
	})
}

// ============================================
// ADMIN - USER MANAGEMENT
// ============================================

func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to fetch users",
		})
		return
	}

	// Don't send passwords
	sanitized := make([]map[string]interface{}, len(users))
	for i, user := range users {
		sanitized[i] = map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"api_key":    user.APIKey,
			"created_at": user.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    sanitized,
	})
}

func (h *Handler) CreateUserByAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to process password",
		})
		return
	}

	role := req.Role
	if role == "" {
		role = models.UserRoleUser
	}

	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
		APIKey:   uuid.New().String(),
	}

	if err := h.db.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"api_key":  user.APIKey,
		},
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Don't allow deleting yourself
	currentUserID := c.GetString("user_id")
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Cannot delete your own account",
		})
		return
	}

	if err := h.db.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "User deleted successfully",
	})
}

// ============================================
// ADMIN - TUNNEL MANAGEMENT
// ============================================

func (h *Handler) GetAllTunnels(c *gin.Context) {
	tunnels, err := h.db.GetAllTunnels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to fetch tunnels",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tunnels,
	})
}

func (h *Handler) DeleteTunnelByAdmin(c *gin.Context) {
	tunnelID := c.Param("id")

	if err := h.db.DeleteTunnelByAdmin(tunnelID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to delete tunnel",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "Tunnel deleted successfully",
	})
}

// ============================================
// ADMIN - SYSTEM STATS
// ============================================

func (h *Handler) GetSystemStats(c *gin.Context) {
	dbStats, err := h.db.GetDatabaseStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get stats",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    dbStats,
	})
}

// ============================================
// TUNNEL HANDLERS (User)
// ============================================

func (h *Handler) GetTunnels(c *gin.Context) {
	userID := c.GetString("user_id")

	tunnels, err := h.db.GetTunnelsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to fetch tunnels",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tunnels,
	})
}

func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.db.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to fetch metrics",
		})
		return
	}

	uptime := time.Since(startTime)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"active_tunnels":    metrics.ActiveTunnels,
			"total_connections": metrics.TotalConnections,
			"total_bytes_up":    metrics.TotalBytesUp,
			"total_bytes_down":  metrics.TotalBytesDown,
			"active_users":      metrics.ActiveUsers,
			"uptime_seconds":    uptime.Seconds(),
			"uptime_formatted":  uptime.String(),
		},
	})
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status": "healthy",
			"uptime": time.Since(startTime).String(),
		},
	})
}

// ============================================
// WEBSOCKET
// ============================================

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Send updates every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics, err := h.db.GetMetrics()
		if err != nil {
			continue
		}

		if err := conn.WriteJSON(metrics); err != nil {
			break
		}
	}
}
