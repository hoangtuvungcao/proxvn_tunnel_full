package fileserver

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizePath chuẩn hóa đường dẫn cho cross-platform
// Windows: C:\Users\Admin → C:\Users\Admin (giữ nguyên)
// Linux: /home/user → /home/user (giữ nguyên)
// Xử lý: ~, .., ., \\, /
func NormalizePath(path string) string {
	// Expand ~ to home directory
	path = ExpandPath(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	// Clean the path (remove .., ., etc.)
	return filepath.Clean(absPath)
}

// ValidatePath đảm bảo path không thoát khỏi thư mục root
// Ngăn chặn directory traversal attacks như: ../../../etc/passwd
func ValidatePath(root, requested string) (string, error) {
	// Normalize both paths
	cleanRoot := filepath.Clean(root)
	cleanRequested := filepath.Clean(requested)

	// If requested is relative, join with root
	var fullPath string
	if !filepath.IsAbs(cleanRequested) {
		fullPath = filepath.Join(cleanRoot, cleanRequested)
	} else {
		fullPath = cleanRequested
	}

	// Get absolute paths
	absRoot, err := filepath.Abs(cleanRoot)
	if err != nil {
		return "", fmt.Errorf("invalid root path: %v", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid requested path: %v", err)
	}

	// Ensure the requested path is within root
	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("cannot determine relative path: %v", err)
	}

	// Check if path tries to escape (starts with ..)
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path escapes root directory: %s", requested)
	}

	return absPath, nil
}

// ExpandPath mở rộng ~ và biến môi trường
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	return os.ExpandEnv(path)
}

// IsWindowsPath kiểm tra xem có phải đường dẫn Windows không
// Ví dụ: C:\, D:\, E:\Users\...
func IsWindowsPath(path string) bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Check for drive letter pattern: C:\, D:\, etc.
	if len(path) >= 2 && path[1] == ':' {
		drive := path[0]
		return (drive >= 'A' && drive <= 'Z') || (drive >= 'a' && drive <= 'z')
	}

	// Check for UNC path: \\server\share
	return strings.HasPrefix(path, "\\\\")
}

// PathExists kiểm tra xem đường dẫn có tồn tại không
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory kiểm tra xem đường dẫn có phải là thư mục không
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDirectory tạo thư mục nếu chưa tồn tại
func EnsureDirectory(path string) error {
	if PathExists(path) {
		if !IsDirectory(path) {
			return fmt.Errorf("path exists but is not a directory: %s", path)
		}
		return nil
	}

	return os.MkdirAll(path, 0755)
}
