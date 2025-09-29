package database

import (
    "context"
    "testing"
    "time"
    sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestRecordQueryMetrics_NoPanic(t *testing.T) {
    d := &DB{}
    d.RecordQueryMetrics(context.Background(), "SELECT", "users", 10*time.Millisecond, nil)
    d.RecordQueryMetrics(context.Background(), "SELECT", "users", 10*time.Millisecond, assertErr{})
}

type assertErr struct{}
func (assertErr) Error() string { return "err" }

func TestRecordConnectionMetrics_NoPanic(t *testing.T) {
    // Use real sqlmock DB to avoid nil deref inside Stats
    sqlDB, _, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock new: %v", err) }
    defer sqlDB.Close()
    d := &DB{DB: sqlDB}
    d.RecordConnectionMetrics(context.Background())
}

