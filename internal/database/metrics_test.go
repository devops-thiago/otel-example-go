package database

import (
	"context"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestRecordQueryMetrics_NoPanic(t *testing.T) {
	d := &DB{}
	d.RecordQueryMetrics(context.Background(), "SELECT", "users", 10*time.Millisecond, nil)
	d.RecordQueryMetrics(context.Background(), "SELECT", "users", 10*time.Millisecond, assertErr{})
}

type assertErr struct{}

func (assertErr) Error() string { return "err" }

func TestRecordConnectionMetrics_NoPanic(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	d := &DB{DB: sqlDB}
	d.RecordConnectionMetrics(context.Background())
}
