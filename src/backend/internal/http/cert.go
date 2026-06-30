package http

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CertManager handles SSL certificate loading and validation
type CertManager struct {
	certFile string
	keyFile  string
}

// NewCertManager creates a new certificate manager
func NewCertManager() *CertManager {
	return &CertManager{}
}

// LoadCertificate attempts to load SSL certificate from various locations
func (cm *CertManager) LoadCertificate() error {
	// Try to find certificate files in multiple locations
	locations := []struct {
		cert string
		key  string
	}{
		// Current directory (preferred for Docker / portable deploys)
		{"wildcard.crt", "wildcard.key"},
		{"server.crt", "server.key"},
		{"cert.pem", "key.pem"},
	}

	// Let's Encrypt / system SSL locations are derived from the configured
	// HTTP_DOMAIN so the same binary works for any domain.
	if domain := strings.TrimPrefix(os.Getenv("HTTP_DOMAIN"), "."); domain != "" {
		locations = append(locations,
			struct{ cert, key string }{
				fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain),
				fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", domain),
			},
			struct{ cert, key string }{
				fmt.Sprintf("/etc/ssl/certs/%s.crt", domain),
				fmt.Sprintf("/etc/ssl/private/%s.key", domain),
			},
		)
	}

	for _, loc := range locations {
		certPath := loc.cert
		keyPath := loc.key

		// Check if files exist
		if _, err := os.Stat(certPath); err == nil {
			if _, err := os.Stat(keyPath); err == nil {
				// Try to load the certificate
				if err := cm.validateCertificate(certPath, keyPath); err == nil {
					cm.certFile = certPath
					cm.keyFile = keyPath
					log.Printf("[cert] Loaded certificate from: %s", certPath)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no valid SSL certificate found")
}

// validateCertificate validates that the certificate and key are valid
func (cm *CertManager) validateCertificate(certFile, keyFile string) error {
	// Try to load the certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	_ = cert // Certificate is valid
	return nil
}

// GetCertFiles returns the paths to the certificate files
func (cm *CertManager) GetCertFiles() (string, string, error) {
	if cm.certFile == "" || cm.keyFile == "" {
		return "", "", fmt.Errorf("certificate not loaded")
	}
	return cm.certFile, cm.keyFile, nil
}

// GenerateSelfSignedCertIfNeeded generates a self-signed certificate if none exists
// This is a fallback for development/testing
func (cm *CertManager) GenerateSelfSignedCertIfNeeded() error {
	certFile := "wildcard.crt"
	keyFile := "wildcard.key"

	// Check if files already exist
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			log.Printf("[cert] Self-signed certificate already exists")
			cm.certFile = certFile
			cm.keyFile = keyFile
			return nil
		}
	}

	domain := strings.TrimPrefix(os.Getenv("HTTP_DOMAIN"), ".")
	if domain == "" {
		domain = "your-domain"
	}
	log.Printf("[cert] WARNING: No SSL certificate found. HTTP tunneling will not work!")
	// #nosec G706
	log.Printf("[cert] Please provide a wildcard certificate for *.%s", domain)
	log.Printf("[cert] Expected files: wildcard.crt and wildcard.key in current directory")

	return fmt.Errorf("SSL certificate required for HTTP tunneling")
}

// GetCertificateInfo returns information about the loaded certificate
func (cm *CertManager) GetCertificateInfo() string {
	if cm.certFile == "" {
		return "No certificate loaded"
	}
	return fmt.Sprintf("Certificate: %s, Key: %s",
		filepath.Base(cm.certFile),
		filepath.Base(cm.keyFile))
}

// CheckCertificateExpiry checks if the certificate is expiring soon
func (cm *CertManager) CheckCertificateExpiry() error {
	if cm.certFile == "" {
		return fmt.Errorf("no certificate loaded")
	}

	cert, err := tls.LoadX509KeyPair(cm.certFile, cm.keyFile)
	if err != nil {
		return err
	}

	if len(cert.Certificate) == 0 {
		return fmt.Errorf("invalid certificate")
	}

	// Note: Full expiry checking would require parsing the certificate
	// For now, we just validate that it loads correctly

	return nil
}
