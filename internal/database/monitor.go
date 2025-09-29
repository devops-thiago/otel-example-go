package database

import (
	"context"
	"log"
	"time"
)

// StartConnectionMonitoring starts periodic monitoring of database connection metrics
func (db *DB) StartConnectionMonitoring(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Database connection monitoring stopped")
				return
			case <-ticker.C:
				// Record connection pool metrics
				db.RecordConnectionMetrics(ctx)

				// Log connection stats for debugging
				stats := db.GetConnectionStats()
				log.Printf("DB Stats - Open: %d, InUse: %d, Idle: %d, WaitCount: %d, WaitDuration: %v",
					stats.OpenConnections,
					stats.InUse,
					stats.Idle,
					stats.WaitCount,
					stats.WaitDuration,
				)
			}
		}
	}()
}

// GetDetailedStats returns detailed database statistics
func (db *DB) GetDetailedStats() map[string]interface{} {
	stats := db.GetConnectionStats()

	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
