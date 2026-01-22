package fileserver

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/webdav"
)

// WebDAVServer quản lý WebDAV file server
type WebDAVServer struct {
	handler    *webdav.Handler
	root       string
	username   string
	password   string
	perms      Permissions
	sessionMgr *SessionManager
}

// NewWebDAVServer tạo WebDAV server mới
func NewWebDAVServer(root, prefix, username, password string, perms Permissions) (*WebDAVServer, error) {
	// Validate root path
	root = NormalizePath(root)
	if !PathExists(root) {
		return nil, fmt.Errorf("đường dẫn không tồn tại: %s", root)
	}

	if !IsDirectory(root) {
		return nil, fmt.Errorf("đường dẫn không phải là thư mục: %s", root)
	}

	// Create WebDAV handler
	handler := &webdav.Handler{
		Prefix:     prefix,
		FileSystem: webdav.Dir(root),
		LockSystem: webdav.NewMemLS(),
		// Logger disabled to prevent spam for normal "file not found" errors
		// (e.g. .hidden, autocomplete paths checked by file managers)
		Logger: func(r *http.Request, err error) {
			// Only log critical errors, not "file not found" which is normal
			// if err != nil && !os.IsNotExist(err) {
			if err != nil && r.Method != "PROPFIND" { // PROPFIND often "fails" naturally on hidden files
				log.Printf("[WebDAV] %s %s - Error: %v", r.Method, r.URL.Path, err)
			}
		},
	}

	// Create session manager with username and password
	sessionMgr := NewSessionManager(username, password)

	return &WebDAVServer{
		handler:    handler,
		root:       root,
		username:   username,
		password:   password,
		perms:      perms,
		sessionMgr: sessionMgr,
	}, nil
}

// ServeHTTP handles WebDAV requests with authentication and permissions
func (s *WebDAVServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply middlewares in order:
	// 1. Authentication
	// 2. Permissions
	// 3. WebDAV handler

	authMiddleware := PasswordAuthMiddleware(s.sessionMgr)
	permMiddleware := PermissionMiddleware(s.perms)

	handler := authMiddleware(permMiddleware(s.handler))
	handler.ServeHTTP(w, r)
}

// GetRoot returns the root directory being shared
func (s *WebDAVServer) GetRoot() string {
	return s.root
}

// GetPermissions returns the permissions
func (s *WebDAVServer) GetPermissions() Permissions {
	return s.perms
}

// GetSessionManager returns the session manager
func (s *WebDAVServer) GetSessionManager() *SessionManager {
	return s.sessionMgr
}
