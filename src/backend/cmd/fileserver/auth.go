package fileserver

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Session quản lý phiên đăng nhập
type Session struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionManager quản lý các session
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	username string
	password string
}

// NewSessionManager tạo session manager mới
func NewSessionManager(username, password string) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		username: username,
		password: password,
	}

	// Cleanup expired sessions every 1 hour
	go sm.cleanupLoop()

	return sm
}

// cleanupLoop xóa session hết hạn
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for token, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, token)
			}
		}
		sm.mu.Unlock()
	}
}

// CreateSession tạo session mới
func (sm *SessionManager) CreateSession() (string, error) {
	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	session := &Session{
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
	}

	sm.mu.Lock()
	sm.sessions[token] = session
	sm.mu.Unlock()

	return token, nil
}

// ValidateSession kiểm tra session có hợp lệ không
func (sm *SessionManager) ValidateSession(token string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[token]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return false
	}

	return true
}

// DeleteSession xóa session
func (sm *SessionManager) DeleteSession(token string) {
	sm.mu.Lock()
	delete(sm.sessions, token)
	sm.mu.Unlock()
}

// ValidatePassword kiểm tra password (constant-time comparison)
func (sm *SessionManager) ValidatePassword(password string) bool {
	return subtle.ConstantTimeCompare([]byte(sm.password), []byte(password)) == 1
}

// ValidateCredentials kiểm tra cả username và password (constant-time comparison)
func (sm *SessionManager) ValidateCredentials(username, password string) bool {
	usernameMatch := subtle.ConstantTimeCompare([]byte(sm.username), []byte(username)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(sm.password), []byte(password)) == 1
	return usernameMatch && passwordMatch
}

// PasswordAuthMiddleware middleware xác thực password
func PasswordAuthMiddleware(sessionMgr *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// [DEBUG] Log request
			// log.Printf("[AUTH] Checking %s %s (Remote: %s)", r.Method, r.URL.Path, r.RemoteAddr)

			// Check session cookie first
			if cookie, err := r.Cookie("proxvn_session"); err == nil {
				if sessionMgr.ValidateSession(cookie.Value) {
					// log.Printf("[AUTH] ✅ Cookie valid")
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check Authorization header
			if auth := r.Header.Get("Authorization"); auth != "" {
				if validateAuthHeader(auth, sessionMgr) {
					// log.Printf("[AUTH] ✅ Header valid")
					next.ServeHTTP(w, r)
					return
				}
			}

			// Allow OPTIONS requests to pass through - required for some WebDAV clients
			// to discover server capabilities (DAV header) before authenticating.
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Check HTTP Basic Auth
			username, password, ok := r.BasicAuth()
			if ok {
				// Validate both username and password
				if sessionMgr.ValidateCredentials(username, password) {
					// log.Printf("[AUTH] ✅ Credentials accepted for user '%s'", username)
					next.ServeHTTP(w, r)
					return
				} else {
					log.Printf("[AUTH] ❌ Invalid credentials for user '%s'", username)
				}
			}

			// No valid authentication found
			w.Header().Set("WWW-Authenticate", `Basic realm="ProxVN File Share"`)
			// Add DAV header to 401 response to help clients identify this as a WebDAV share
			w.Header().Set("Dav", "1, 2")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// validateAuthHeader validates Authorization header
func validateAuthHeader(auth string, sessionMgr *SessionManager) bool {
	// Support "Basic" and "Bearer" tokens
	if strings.HasPrefix(auth, "Basic ") {
		// Decode Basic auth
		encoded := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return false
		}

		// Format is "username:password"
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) == 2 {
			// Validate both username and password
			return sessionMgr.ValidateCredentials(parts[0], parts[1])
		}
	} else if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		return sessionMgr.ValidateSession(token)
	}

	return false
}

// SetSessionCookie sets session cookie
func SetSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	// Detect if on HTTPS
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "proxvn_session",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie clears session cookie
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	// Detect if on HTTPS
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "proxvn_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// HandleLogin xử lý đăng nhập
func HandleLogin(sessionMgr *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Try Form Value first (backend preference)
		username := r.FormValue("username")
		password := r.FormValue("password")

		// If empty, try JSON body (for cached clients sending JSON)
		if password == "" {
			var jsonData struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&jsonData); err == nil {
				username = jsonData.Username
				password = jsonData.Password
			}
		}

		// Validate both username and password
		if !sessionMgr.ValidateCredentials(username, password) {
			log.Printf("[DEBUG] ❌ Invalid credentials for user '%s'", username)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("[DEBUG] ✅ Credentials matched for user '%s'", username)

		// Create session
		token, err := sessionMgr.CreateSession()
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set cookie
		SetSessionCookie(w, r, token)

		// Return success
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "token": "%s"}`, token)
	}
}

// HandleLogout xử lý đăng xuất
func HandleLogout(sessionMgr *SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		if cookie, err := r.Cookie("proxvn_session"); err == nil {
			sessionMgr.DeleteSession(cookie.Value)
		}

		// Clear cookie
		ClearSessionCookie(w, r)

		// Return success
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true}`)
	}
}
