package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"proxvn/backend/internal/auth"
	"proxvn/backend/internal/database"
	"proxvn/backend/internal/models"
)

var startTime = time.Now()

type Handler struct {
	db         *database.Database
	authService *auth.AuthService
}

func NewHandler(db *database.Database, authService *auth.AuthService) *Handler {
	return &Handler{
		db:         db,
		authService: authService,
	}
}

// Auth handlers
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
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

	if !h.authService.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	user.Password = "" // Don't send password to client
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user":  user,
			"token": token,
		},
	})
}

func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Validate input
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Password must be at least 6 characters",
		})
		return
	}
	if len(req.Username) < 3 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Username must be at least 3 characters",
		})
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
		return
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     models.UserRoleUser,
		APIKey:   h.authService.GenerateAPIKey(),
	}

	if err := h.db.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create user",
		})
		return
	}

	user.Password = "" // Don't send password to client
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    user,
		Message: "User created successfully",
	})
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	user, err := h.db.GetUserByUsername(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	user.Password = "" // Don't send password to client
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// Tunnel handlers
func (h *Handler) GetTunnels(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	tunnels, err := h.db.GetTunnelsByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get tunnels",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tunnels,
	})
}

func (h *Handler) CreateTunnel(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req models.CreateTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Generate unique public port
	publicPort := h.generatePublicPort()
	
	// Generate auth token
	authToken := h.authService.GenerateAPIKey()

	tunnel := &models.Tunnel{
		UserID:     uuid.MustParse(userID.(string)),
		Name:       req.Name,
		Protocol:   req.Protocol,
		LocalHost:  req.LocalHost,
		LocalPort:  req.LocalPort,
		PublicPort: publicPort,
		Status:     models.TunnelStatusInactive,
		AuthToken:  authToken,
	}

	if err := h.db.CreateTunnel(tunnel); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to create tunnel",
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    tunnel,
		Message: "Tunnel created successfully",
	})
}

func (h *Handler) GetTunnel(c *gin.Context) {
	tunnelID := c.Param("id")
	
	tunnel, err := h.db.GetTunnelByID(tunnelID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Tunnel not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tunnel,
	})
}

func (h *Handler) UpdateTunnel(c *gin.Context) {
	tunnelID := c.Param("id")
	
	var req models.UpdateTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.db.UpdateTunnel(tunnelID, userID.(string), req.Name, req.LocalHost, req.LocalPort); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Tunnel updated successfully",
	})
}

func (h *Handler) DeleteTunnel(c *gin.Context) {
	tunnelID := c.Param("id")
	userID, _ := c.Get("user_id")
	
	if err := h.db.DeleteTunnel(tunnelID, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Tunnel deleted successfully",
	})
}

// Metrics handlers
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.db.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get metrics",
		})
		return
	}

	// Calculate uptime (using process start time)
	uptime := time.Since(startTime)
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60
	metrics.Uptime = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    metrics,
	})
}

func (h *Handler) GetTunnelStats(c *gin.Context) {
	tunnelID := c.Param("id")
	
	stats, err := h.db.GetTunnelStats(tunnelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get tunnel stats",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// Admin handlers
func (h *Handler) GetAllUsers(c *gin.Context) {
	users, err := h.db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get users",
		})
		return
	}
	
	// Remove passwords from response
	for _, user := range users {
		user.Password = ""
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    users,
	})
}

func (h *Handler) GetAllTunnels(c *gin.Context) {
	tunnels, err := h.db.GetAllTunnels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get tunnels",
		})
		return
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tunnels,
	})
}

// Helper functions
func (h *Handler) generatePublicPort() int {
	// Generate unique port in range 10000-65535
	for i := 0; i < 100; i++ {
		port := 10000 + (int(time.Now().UnixNano()/1000) % 55535)
		if h.db.IsPortAvailable(port) {
			return port
		}
		time.Sleep(time.Millisecond)
	}
	// Fallback: random port
	return 10000 + int(time.Now().Unix()%55535)
}

// WebSocket handler for real-time updates
func (h *Handler) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // In production, validate origin properly
		},
	}
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	
	// Send initial tunnel status
	userID, exists := c.Get("user_id")
	if !exists {
		return
	}
	
	// Send updates every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			tunnels, err := h.db.GetTunnelsByUserID(userID.(string))
			if err != nil {
				continue
			}
			
			msg := models.WebSocketMessage{
				Type:      models.WSMessageTypeTunnelUpdate,
				Data:      tunnels,
				Timestamp: time.Now(),
			}
			
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}
}
