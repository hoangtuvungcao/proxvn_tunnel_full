package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"log"
)

// verifyCertFingerprint is a custom certificate verification callback
// that checks if the server's certificate matches the expected fingerprint
func verifyCertFingerprint(rawCerts [][]byte, expectedFingerprint string) error {
	if expectedFingerprint == "" {
		// No pinning required
		return nil
	}

	if len(rawCerts) == 0 {
		return fmt.Errorf("no certificates presented by server")
	}

	// Parse the leaf certificate
	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Calculate SHA256 fingerprint
	fingerprint := sha256.Sum256(cert.Raw)
	actualFingerprint := hex.EncodeToString(fingerprint[:])

	// Compare with expected
	if actualFingerprint != expectedFingerprint {
		return fmt.Errorf("certificate fingerprint mismatch: expected %s, got %s",
			expectedFingerprint, actualFingerprint)
	}

	return nil
}

// buildTLSConfig creates a TLS config with optional certificate pinning
func (c *client) buildTLSConfig() *tls.Config {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// Skip certificate verification by default (works with self-signed certs)
		InsecureSkipVerify: true,
	}

	// If certificate pinning is requested, enforce strict validation
	if c.certFingerprint != "" {
		config.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return verifyCertFingerprint(rawCerts, c.certFingerprint)
		}
		log.Printf("[client] 🔐 Certificate pinning enabled")
	}

	return config
}
