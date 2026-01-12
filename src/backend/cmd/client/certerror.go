package main

import (
	"strings"
)

// isCertError checks if an error is related to certificate verification
func isCertError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "certificate") ||
		strings.Contains(errStr, "x509") ||
		strings.Contains(errStr, "tls")
}
