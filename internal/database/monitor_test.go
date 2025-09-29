package database

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestStartConnectionMonitoring(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	d := &DB{DB: sqlDB}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start monitoring with a short interval
	d.StartConnectionMonitoring(ctx, 50*time.Millisecond)

	// Wait for context to be done
	<-ctx.Done()
}

func TestGetDetailedStats(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer sqlDB.Close()

	d := &DB{DB: sqlDB}

	stats := d.GetDetailedStats()
	if stats == nil {
		t.Fatal("expected non-nil detailed stats")
	}

	expectedFields := []string{
		"open_connections",
		"in_use", 
		"idle",
		"wait_count",
		"wait_duration",
		"max_idle_closed",
		"max_idle_time_closed", 
		"max_lifetime_closed",
	}

	for _, field := range expectedFields {
		if _, ok := stats[field]; !ok {
			t.Errorf("expected field %s in detailed stats", field)
		}
	}
}