package fileserver

import (
	_ "embed"
	"net/http"
)

//go:embed static/login.html
var loginHTML []byte

//go:embed static/browse.html
var browseHTML []byte

// UIHandler quản lý các trang web UI
type UIHandler struct {
	sessionMgr *SessionManager
}

// NewUIHandler tạo UI handler mới
func NewUIHandler(sessionMgr *SessionManager) *UIHandler {
	return &UIHandler{
		sessionMgr: sessionMgr,
	}
}

// ServeHTTP handles UI requests
func (h *UIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch path {
	case "/", "/login":
		h.serveLogin(w, r)
	case "/browse":
		h.serveBrowse(w, r)
	case "/api/login":
		HandleLogin(h.sessionMgr)(w, r)
	case "/api/logout":
		HandleLogout(h.sessionMgr)(w, r)
	default:
		http.NotFound(w, r)
	}
}

// serveLogin serves the login page
func (h *UIHandler) serveLogin(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	if cookie, err := r.Cookie("proxvn_session"); err == nil {
		if h.sessionMgr.ValidateSession(cookie.Value) {
			http.Redirect(w, r, "/browse", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(loginHTML)
}

// serveBrowse serves the file browser page (requires authentication)
func (h *UIHandler) serveBrowse(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	authenticated := false

	if cookie, err := r.Cookie("proxvn_session"); err == nil {
		if h.sessionMgr.ValidateSession(cookie.Value) {
			authenticated = true
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(browseHTML)
}
