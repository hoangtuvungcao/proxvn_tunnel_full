package database

import (
	"log"
	"time"
)

// cleanupOldConnections tự động xóa connections cũ mỗi 1 giờ
func (d *Database) cleanupOldConnections() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// Xóa connections cũ hơn 24 giờ
		query := `DELETE FROM connections WHERE disconnected_at IS NOT NULL AND disconnected_at < datetime('now', '-24 hours')`
		result, err := d.db.Exec(query)
		if err != nil {
			log.Printf("[database] Error cleaning old connections: %v", err)
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Printf("[database] Cleaned up %d old connection records", rows)
		}
	}
}

// OptimizeDatabase chạy VACUUM và ANALYZE để tối ưu
func (d *Database) OptimizeDatabase() error {
	log.Println("[database] Running optimization...")

	// ANALYZE để update query planner statistics
	if _, err := d.db.Exec("ANALYZE"); err != nil {
		return err
	}

	// VACUUM để compact database (chạy ít thôi vì blocking)
	if _, err := d.db.Exec("VACUUM"); err != nil {
		return err
	}

	log.Println("[database] Optimization complete")
	return nil
}

// GetDatabaseStats trả về thống kê database
func (d *Database) GetDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Database size
	var pageCount, pageSize int64
	err := d.db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return nil, err
	}
	err = d.db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return nil, err
	}
	stats["size_bytes"] = pageCount * pageSize
	stats["size_mb"] = float64(pageCount*pageSize) / 1024 / 1024

	// WAL mode check
	var walMode string
	err = d.db.QueryRow("PRAGMA journal_mode").Scan(&walMode)
	if err != nil {
		return nil, err
	}
	stats["journal_mode"] = walMode

	// Connection stats
	dbStats := d.db.Stats()
	stats["open_connections"] = dbStats.OpenConnections
	stats["in_use"] = dbStats.InUse
	stats["idle"] = dbStats.Idle
	stats["wait_count"] = dbStats.WaitCount
	stats["wait_duration_ms"] = dbStats.WaitDuration.Milliseconds()

	return stats, nil
}
