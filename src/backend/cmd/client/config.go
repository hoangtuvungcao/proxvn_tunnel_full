package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ClientConfig holds default values that can be set from a config file so the
// client can be pointed at a different server / domain WITHOUT recompiling.
// Precedence at runtime:  command-line flags  >  config file  >  built-in defaults.
type ClientConfig struct {
	Server   string `json:"server"`    // tunnel server address (host:port)
	Host     string `json:"host"`      // local host to forward
	Port     int    `json:"port"`      // local port to forward
	Proto    string `json:"proto"`     // tcp | udp | http
	UI       bool   `json:"ui"`        // enable TUI
	CertPin  string `json:"cert_pin"`  // optional server cert SHA256 pin
	Insecure bool   `json:"insecure"`  // skip TLS verify (testing only)
}

// defaultClientConfig returns the built-in defaults (used when no config file).
func defaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Server: defaultServerAddr,
		Host:   defaultLocalHost,
		Port:   defaultLocalPort,
		Proto:  "tcp",
		UI:     true,
	}
}

// configSearchPaths returns the ordered list of locations to look for a config
// file. An explicit path (from --config or PROXVN_CONFIG) wins; otherwise the
// client looks next to the binary, in the working directory and in the user's
// home directory.
func configSearchPaths(explicit string) []string {
	var paths []string
	if explicit != "" {
		paths = append(paths, explicit)
		return paths
	}
	if env := strings.TrimSpace(os.Getenv("PROXVN_CONFIG")); env != "" {
		paths = append(paths, env)
	}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), "proxvn.json"))
	}
	paths = append(paths, "proxvn.json", "config.json")
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".proxvn", "config.json"))
	}
	return paths
}

// preParseConfigFlag scans os.Args for --config / -config (=value or next arg)
// before the main flag set is defined, so the file can seed flag defaults.
func preParseConfigFlag(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "--config="):
			return strings.TrimPrefix(a, "--config=")
		case strings.HasPrefix(a, "-config="):
			return strings.TrimPrefix(a, "-config=")
		case a == "--config" || a == "-config":
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

// loadClientConfig builds the effective default config: built-in defaults
// overlaid with the first config file found (if any).
func loadClientConfig(explicit string) *ClientConfig {
	cfg := defaultClientConfig()
	for _, p := range configSearchPaths(explicit) {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			log.Printf("[client] config file %s không hợp lệ: %v (bỏ qua)", p, err)
			continue
		}
		log.Printf("[client] đã nạp cấu hình từ: %s", p)
		break
	}
	// Guard against empty / invalid values from the file.
	if strings.TrimSpace(cfg.Server) == "" {
		cfg.Server = defaultServerAddr
	}
	if strings.TrimSpace(cfg.Host) == "" {
		cfg.Host = defaultLocalHost
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		cfg.Port = defaultLocalPort
	}
	if cfg.Proto == "" {
		cfg.Proto = "tcp"
	}
	return cfg
}
