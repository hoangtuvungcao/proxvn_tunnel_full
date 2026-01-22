package fileserver

import (
	"fmt"
	"net/http"
	"strings"
)

// Permissions định nghĩa quyền hạn cho file sharing
type Permissions string

const (
	PermRead      Permissions = "r"   // Chỉ đọc: browse, download
	PermReadWrite Permissions = "rw"  // Đọc-ghi: browse, download, upload, mkdir, delete
	PermFull      Permissions = "rwx" // Đầy đủ: all operations
)

// ParsePermissions chuyển đổi string sang Permissions
func ParsePermissions(s string) Permissions {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "r", "read":
		return PermRead
	case "rw", "readwrite", "read-write":
		return PermReadWrite
	case "rwx", "full":
		return PermFull
	default:
		return PermReadWrite // Default to read-write
	}
}

// HasRead kiểm tra quyền đọc
func (p Permissions) HasRead() bool {
	return p == PermRead || p == PermReadWrite || p == PermFull
}

// HasWrite kiểm tra quyền ghi
func (p Permissions) HasWrite() bool {
	return p == PermReadWrite || p == PermFull
}

// HasExecute kiểm tra quyền execute (future use)
func (p Permissions) HasExecute() bool {
	return p == PermFull
}

// Description trả về mô tả quyền bằng tiếng Việt
func (p Permissions) Description() string {
	switch p {
	case PermRead:
		return "Chỉ đọc"
	case PermReadWrite:
		return "Đọc và Ghi"
	case PermFull:
		return "Quyền đầy đủ"
	default:
		return "Không xác định"
	}
}

// String returns string representation
func (p Permissions) String() string {
	return string(p)
}

// PermissionMiddleware kiểm tra quyền trước khi thực hiện operation
func PermissionMiddleware(perms Permissions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := r.Method

			// Log request for debugging
			// log.Printf("[FileServer] %s %s (Permissions: %s)", method, r.URL.Path, perms)

			// Kiểm tra quyền dựa trên HTTP method
			switch method {
			case "GET", "HEAD", "OPTIONS", "PROPFIND":
				// Các method này yêu cầu quyền đọc (r)
				if !perms.HasRead() {
					http.Error(w, "Forbidden: Read permission required", http.StatusForbidden)
					return
				}

			case "PUT", "POST", "DELETE", "MKCOL", "COPY", "MOVE", "PROPPATCH", "PATCH":
				// Các method này yêu cầu quyền ghi (rw hoặc rwx)
				if !perms.HasWrite() {
					http.Error(w, "Forbidden: Write permission required", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckMethodPermission kiểm tra quyền cho một method cụ thể
func CheckMethodPermission(method string, perms Permissions) error {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS", "PROPFIND":
		if !perms.HasRead() {
			return fmt.Errorf("read permission required")
		}
	case "PUT", "POST", "DELETE", "MKCOL", "COPY", "MOVE", "PROPPATCH", "PATCH":
		if !perms.HasWrite() {
			return fmt.Errorf("write permission required")
		}
	}
	return nil
}
