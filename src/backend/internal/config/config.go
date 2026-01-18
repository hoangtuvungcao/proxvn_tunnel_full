package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Monitoring MonitoringConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	PublicPortStart int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	HTTPDomain   string // Base domain for HTTP tunneling (e.g., "vutrungocrong.fun")
	HTTPPort     int    // Port for HTTP proxy (default 443)
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret     string
	TokenExpiry   time.Duration
	AdminUsername string
	AdminPassword string
}

type MonitoringConfig struct {
	Enabled bool
	Port    int
	Path    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			Port:           getEnvInt("SERVER_PORT", 8881),
			PublicPortStart: getEnvInt("PUBLIC_PORT_START", 10000),
			ReadTimeout:    getEnvDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout:   getEnvDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:    getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
			HTTPDomain:     getEnv("HTTP_DOMAIN", ""), // Leave empty for IP:port fallback
			HTTPPort:       getEnvInt("HTTP_PORT", 443),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "proxvn"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "proxvn_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "proxvn-secret-key"),
			TokenExpiry:   getEnvDuration("TOKEN_EXPIRY", 24*time.Hour),
			AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		},
		Monitoring: MonitoringConfig{
			Enabled: getEnvBool("MONITORING_ENABLED", true),
			Port:    getEnvInt("MONITORING_PORT", 9090),
			Path:    getEnv("MONITORING_PATH", "/metrics"),
		},
	}

	return cfg, nil
}

func (c *Config) GetDatabaseDSN() string {
	// For SQLite3, return path from DB_PATH env or default
	// Leave empty to disable database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./proxvn.db" // Default SQLite file
	}
	return dbPath
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
