package tunnel

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
)

const Version = "7.5"

// Message is the control-plane payload exchanged between tunnel peers.
type Message struct {
	Type          string `json:"type"`
	Key           string `json:"key,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	RemotePort    int    `json:"remote_port,omitempty"`
	RequestedPort int    `json:"requested_port,omitempty"` // Port client wants to reuse on reconnect
	Target        string `json:"target,omitempty"`
	ID            string `json:"id,omitempty"`
	Error         string `json:"error,omitempty"`
	Version       string `json:"version,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	RemoteAddr    string `json:"remote_addr,omitempty"`
	Payload       string `json:"payload,omitempty"`

	// HTTP tunneling fields
	Subdomain  string            `json:"subdomain,omitempty"`
	Method     string            `json:"method,omitempty"`
	Path       string            `json:"path,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body,omitempty"`
	StatusCode int               `json:"status_code,omitempty"`

	// Security
	UDPSecret  string `json:"udp_secret,omitempty"`  // Base64 encoded AES key
	BaseDomain string `json:"base_domain,omitempty"` // Base domain for HTTP (e.g. vutrungocrong.fun)
}

// NewEncoder returns a JSON encoder with HTML escaping disabled.
func NewEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

// NewDecoder wraps the reader in a JSON decoder.
func NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

// GenerateID returns a random 16-byte hex string suitable for request IDs.
func GenerateID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
