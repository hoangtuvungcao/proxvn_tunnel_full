package compression

import (
	"compress/gzip"
	"io"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"
)

// Compression levels
const (
	DefaultCompressionLevel = 6
	BestSpeed               = 1
	BestCompression         = 9
)

// gzipWriterPool để tái sử dụng gzip writers
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, DefaultCompressionLevel)
		return w
	},
}

// CompressedWriter wrapper cho response writer
type CompressedWriter struct {
	gin.ResponseWriter
	writer    io.Writer
	compress  bool
	algorithm string
}

func (w *CompressedWriter) Write(data []byte) (int, error) {
	if w.compress {
		return w.writer.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

// Middleware trả về Gin middleware cho compression
func Middleware(config ...CompressionConfig) gin.HandlerFunc {
	var cfg CompressionConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultConfig()
	}

	return func(c *gin.Context) {
		// Skip nếu không enable
		if !cfg.Enable {
			c.Next()
			return
		}

		// Skip compression cho các loại file đã nén
		if shouldSkipCompression(c.Request.Header.Get("Content-Type")) {
			c.Next()
			return
		}

		// Kiểm tra client có hỗ trợ compression không
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		if acceptEncoding == "" {
			c.Next()
			return
		}

		// Chọn thuật toán compression
		var algorithm string
		var writer io.Writer
		var closer io.WriteCloser

		if strings.Contains(acceptEncoding, "zstd") && cfg.EnableZstd {
			algorithm = "zstd"
			encoder, err := zstd.NewWriter(c.Writer, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(cfg.ZstdLevel)))
			if err == nil {
				writer = encoder
				closer = encoder
			}
		} else if strings.Contains(acceptEncoding, "br") && cfg.EnableBrotli {
			algorithm = "br"
			// Brotli support có thể thêm sau nếu cần
		} else if strings.Contains(acceptEncoding, "gzip") {
			algorithm = "gzip"
			gzw := gzipWriterPool.Get().(*gzip.Writer)
			gzw.Reset(c.Writer)
			writer = gzw
			closer = gzw
			defer func() {
				gzw.Close()
				gzipWriterPool.Put(gzw)
			}()
		}

		if writer != nil {
			c.Header("Content-Encoding", algorithm)
			c.Header("Vary", "Accept-Encoding")
			c.Header("X-Compressed", "true")

			cw := &CompressedWriter{
				ResponseWriter: c.Writer,
				writer:         writer,
				compress:       true,
				algorithm:      algorithm,
			}
			c.Writer = cw

			defer func() {
				if closer != nil {
					closer.Close()
				}
			}()
		}

		c.Next()
	}
}

// CompressionConfig cấu hình compression
type CompressionConfig struct {
	Enable       bool
	GzipLevel    int
	ZstdLevel    int
	EnableBrotli bool
	EnableZstd   bool
	SkipTypes    []string
}

// DefaultConfig trả về cấu hình mặc định
func DefaultConfig() CompressionConfig {
	return CompressionConfig{
		Enable:       true,
		GzipLevel:    DefaultCompressionLevel,
		ZstdLevel:    3,
		EnableBrotli: false,
		EnableZstd:   true,
		SkipTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"video/",
			"audio/",
			"application/zip",
			"application/gzip",
			"application/x-gzip",
		},
	}
}

// shouldSkipCompression kiểm tra xem có nên skip compression không
func shouldSkipCompression(contentType string) bool {
	skipTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/zip",
		"application/gzip",
		"application/x-gzip",
		"video/",
		"audio/",
	}

	for _, skip := range skipTypes {
		if strings.Contains(contentType, skip) {
			return true
		}
	}
	return false
}

// CompressResponse helper để compress response
func CompressResponse(data []byte, level int) ([]byte, error) {
	var buf strings.Builder
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}
