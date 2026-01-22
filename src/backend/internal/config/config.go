package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config chứa tất cả cấu hình hệ thống
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Auth         AuthConfig
	Monitoring   MonitoringConfig
	Performance  PerformanceConfig
	RateLimit    RateLimitConfig
	FileServer   FileServerConfig
	Backup       BackupConfig
	Notification NotificationConfig
	TLS          TLSConfig
	CORS         CORSConfig
}

// ServerConfig cấu hình server
type ServerConfig struct {
	Host            string
	Port            int
	PublicPortStart int
	PublicPortEnd   int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	HTTPDomain      string // Base domain for HTTP tunneling
	HTTPPort        int    // Port for HTTP proxy
}

// DatabaseConfig cấu hình database
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AuthConfig cấu hình authentication
type AuthConfig struct {
	JWTSecret       string
	TokenExpiry     time.Duration
	AdminUsername   string
	AdminPassword   string
	GenerateAPIKeys bool
}

// MonitoringConfig cấu hình monitoring
type MonitoringConfig struct {
	Enabled   bool
	Port      int
	Path      string
	DebugMode bool
	LogLevel  string
}

// PerformanceConfig cấu hình hiệu suất
type PerformanceConfig struct {
	MaxConnections       int
	BufferSize           int
	EnableCompression    bool
	CompressionLevel     int
	EnableHTTP2          bool
	EnableCache          bool
	CacheSizeMB          int
	CacheTTL             time.Duration
	WorkerPoolSize       int
	TCPKeepAlive         time.Duration
	TCPKeepAliveInterval time.Duration
	TCPNoDelay           bool
	MaxIdleConns         int
	MaxIdleConnsPerHost  int
	IdleConnTimeout      time.Duration
}

// RateLimitConfig cấu hình rate limiting
type RateLimitConfig struct {
	RPS                  int
	Burst                int
	EnableDDoSProtection bool
}

// FileServerConfig cấu hình file server
type FileServerConfig struct {
	Enabled            bool
	Port               int
	MaxUploadSizeMB    int
	UserStorageQuotaMB int
	BandwidthLimitMBps int
	MaxPreviewSizeMB   int
	EnableVersioning   bool
	MaxFileVersions    int
	WebDAVEnabled      bool
	WebDAVPath         string
}

// BackupConfig cấu hình backup
type BackupConfig struct {
	AutoBackup     bool
	BackupInterval time.Duration
	BackupDir      string
	RetentionDays  int
}

// NotificationConfig cấu hình notifications
type NotificationConfig struct {
	SMTPEnabled  bool
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	WebhookEnabled bool
	WebhookURL     string

	TelegramEnabled  bool
	TelegramBotToken string
	TelegramChatID   string
}

// TLSConfig cấu hình TLS/SSL
type TLSConfig struct {
	MinVersion string
	AutoTLS    bool
	CertFile   string
	KeyFile    string
}

// CORSConfig cấu hình CORS
type CORSConfig struct {
	Enabled        bool
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

// Load đọc cấu hình từ file .env và biến môi trường
func Load() (*Config, error) {
	// Thử load file .env nếu có
	if err := godotenv.Load(); err != nil {
		log.Println("[config] Không tìm thấy file .env, sử dụng biến môi trường hoặc giá trị mặc định")
	} else {
		log.Println("[config] Đã load cấu hình từ file .env")
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvInt("SERVER_PORT", 8882),
			PublicPortStart: getEnvInt("PUBLIC_PORT_START", 10000),
			PublicPortEnd:   getEnvInt("PUBLIC_PORT_END", 20000),
			ReadTimeout:     getEnvDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
			HTTPDomain:      getEnv("HTTP_DOMAIN", ""),
			HTTPPort:        getEnvInt("HTTP_PORT", 443),
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
			JWTSecret:       getEnv("JWT_SECRET", "proxvn-secret-key-change-this"),
			TokenExpiry:     getEnvDuration("TOKEN_EXPIRY", 24*time.Hour),
			AdminUsername:   getEnv("ADMIN_USERNAME", "admin"),
			AdminPassword:   getEnv("ADMIN_PASSWORD", "admin123"),
			GenerateAPIKeys: getEnvBool("GENERATE_API_KEYS", true),
		},
		Monitoring: MonitoringConfig{
			Enabled:   getEnvBool("MONITORING_ENABLED", true),
			Port:      getEnvInt("MONITORING_PORT", 9090),
			Path:      getEnv("MONITORING_PATH", "/metrics"),
			DebugMode: getEnvBool("DEBUG_MODE", false),
			LogLevel:  getEnv("LOG_LEVEL", "info"),
		},
		Performance: PerformanceConfig{
			MaxConnections:       getEnvInt("MAX_CONNECTIONS", 10000),
			BufferSize:           getEnvInt("BUFFER_SIZE", 32768),
			EnableCompression:    getEnvBool("ENABLE_COMPRESSION", true),
			CompressionLevel:     getEnvInt("COMPRESSION_LEVEL", 6),
			EnableHTTP2:          getEnvBool("ENABLE_HTTP2", true),
			EnableCache:          getEnvBool("ENABLE_CACHE", true),
			CacheSizeMB:          getEnvInt("CACHE_SIZE_MB", 256),
			CacheTTL:             getEnvDuration("CACHE_TTL", 3600*time.Second),
			WorkerPoolSize:       getEnvInt("WORKER_POOL_SIZE", 100),
			TCPKeepAlive:         getEnvDuration("TCP_KEEPALIVE", 30*time.Second),
			TCPKeepAliveInterval: getEnvDuration("TCP_KEEPALIVE_INTERVAL", 15*time.Second),
			TCPNoDelay:           getEnvBool("TCP_NODELAY", true),
			MaxIdleConns:         getEnvInt("MAX_IDLE_CONNS", 100),
			MaxIdleConnsPerHost:  getEnvInt("MAX_IDLE_CONNS_PER_HOST", 10),
			IdleConnTimeout:      getEnvDuration("IDLE_CONN_TIMEOUT", 90*time.Second),
		},
		RateLimit: RateLimitConfig{
			RPS:                  getEnvInt("RATE_LIMIT_RPS", 10),
			Burst:                getEnvInt("RATE_LIMIT_BURST", 20),
			EnableDDoSProtection: getEnvBool("ENABLE_DDOS_PROTECTION", true),
		},
		FileServer: FileServerConfig{
			Enabled:            getEnvBool("FILE_SERVER_ENABLED", true),
			Port:               getEnvInt("FILE_SERVER_PORT", 8080),
			MaxUploadSizeMB:    getEnvInt("MAX_UPLOAD_SIZE", 1000),
			UserStorageQuotaMB: getEnvInt("USER_STORAGE_QUOTA", 10000),
			BandwidthLimitMBps: getEnvInt("BANDWIDTH_LIMIT", 0),
			MaxPreviewSizeMB:   getEnvInt("MAX_PREVIEW_SIZE", 10),
			EnableVersioning:   getEnvBool("ENABLE_FILE_VERSIONING", true),
			MaxFileVersions:    getEnvInt("MAX_FILE_VERSIONS", 5),
			WebDAVEnabled:      getEnvBool("WEBDAV_ENABLED", true),
			WebDAVPath:         getEnv("WEBDAV_PATH", "/webdav"),
		},
		Backup: BackupConfig{
			AutoBackup:     getEnvBool("AUTO_BACKUP", true),
			BackupInterval: getEnvDuration("BACKUP_INTERVAL", 24*time.Hour),
			BackupDir:      getEnv("BACKUP_DIR", "./backups"),
			RetentionDays:  getEnvInt("BACKUP_RETENTION_DAYS", 7),
		},
		Notification: NotificationConfig{
			SMTPEnabled:      getEnvBool("SMTP_ENABLED", false),
			SMTPHost:         getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:         getEnvInt("SMTP_PORT", 587),
			SMTPUser:         getEnv("SMTP_USER", ""),
			SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
			SMTPFrom:         getEnv("SMTP_FROM", "noreply@proxvn.com"),
			WebhookEnabled:   getEnvBool("WEBHOOK_ENABLED", false),
			WebhookURL:       getEnv("WEBHOOK_URL", ""),
			TelegramEnabled:  getEnvBool("TELEGRAM_ENABLED", false),
			TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
			TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
		},
		TLS: TLSConfig{
			MinVersion: getEnv("TLS_MIN_VERSION", "1.3"),
			AutoTLS:    getEnvBool("AUTO_TLS", true),
			CertFile:   getEnv("TLS_CERT_FILE", "./server.crt"),
			KeyFile:    getEnv("TLS_KEY_FILE", "./server.key"),
		},
		CORS: CORSConfig{
			Enabled:        getEnvBool("CORS_ENABLED", true),
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Authorization,X-API-Key"),
		},
	}

	return cfg, nil
}

// GetDatabaseDSN trả về database connection string
func (c *Config) GetDatabaseDSN() string {
	// For SQLite3, return path from DB_PATH env or default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./proxvn.db"
	}
	return dbPath
}

// Helper functions
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
