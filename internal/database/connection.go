package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"example/otel/internal/config"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// DB holds the database connection with metrics
type DB struct {
	*sql.DB
	meter               metric.Meter
	queryDuration       metric.Float64Histogram
	queryCount          metric.Int64Counter
	queryErrors         metric.Int64Counter
	connectionCount     metric.Int64UpDownCounter
	connectionErrors    metric.Int64Counter
	healthCheckDuration metric.Float64Histogram
}

// NewConnection creates a new database connection with OpenTelemetry instrumentation
func NewConnection(cfg *config.Config) (*DB, error) {
	// Open connection using the instrumented mysql driver with XSAM/otelsql
	db, err := otelsql.Open("mysql", cfg.Database.DSN,
		otelsql.WithAttributes(
			semconv.DBSystemMySQL,
			semconv.DBName(cfg.Database.Name),
			semconv.DBConnectionString(cfg.Database.DSN),
		),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitConnQuery:        false,
			OmitRows:             false,
			OmitConnectorConnect: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Register database stats for metrics collection
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
		semconv.DBName(cfg.Database.Name),
	))
	if err != nil {
		log.Printf("Warning: Failed to register database stats metrics: %v", err)
	}

	// Create meter for custom metrics
	meter := otel.Meter("database")

	// Create custom metrics
	queryDuration, err := meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create query duration metric: %v", err)
	}

	queryCount, err := meter.Int64Counter(
		"db.query.count",
		metric.WithDescription("Total number of database queries"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create query count metric: %v", err)
	}

	queryErrors, err := meter.Int64Counter(
		"db.query.errors",
		metric.WithDescription("Total number of database query errors"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create query errors metric: %v", err)
	}

	connectionCount, err := meter.Int64UpDownCounter(
		"db.connections.active",
		metric.WithDescription("Number of active database connections"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create connection count metric: %v", err)
	}

	connectionErrors, err := meter.Int64Counter(
		"db.connection.errors",
		metric.WithDescription("Total number of database connection errors"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create connection errors metric: %v", err)
	}

	healthCheckDuration, err := meter.Float64Histogram(
		"db.health_check.duration",
		metric.WithDescription("Database health check duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		log.Printf("Warning: Failed to create health check duration metric: %v", err)
	}

	log.Println("Successfully connected to database with comprehensive OpenTelemetry instrumentation")

	return &DB{
		DB:                  db,
		meter:               meter,
		queryDuration:       queryDuration,
		queryCount:          queryCount,
		queryErrors:         queryErrors,
		connectionCount:     connectionCount,
		connectionErrors:    connectionErrors,
		healthCheckDuration: healthCheckDuration,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health checks if the database connection is healthy
func (db *DB) Health() error {
	start := time.Now()
	err := db.Ping()
	duration := time.Since(start).Seconds()

	// Record health check duration
	if db.healthCheckDuration != nil {
		db.healthCheckDuration.Record(context.Background(), duration, metric.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.Bool("db.health.status", err == nil),
		))
	}

	// Record connection errors
	if err != nil && db.connectionErrors != nil {
		db.connectionErrors.Add(context.Background(), 1, metric.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("error.type", "health_check_failed"),
		))
	}

	return err
}

// RecordQueryMetrics records metrics for database queries
func (db *DB) RecordQueryMetrics(ctx context.Context, operation, table string, duration time.Duration, err error) {
	attrs := []attribute.KeyValue{
		semconv.DBSystemMySQL,
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	// Record query duration
	if db.queryDuration != nil {
		db.queryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	// Record query count
	if db.queryCount != nil {
		db.queryCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	// Record query errors
	if err != nil && db.queryErrors != nil {
		errorAttrs := append(attrs, attribute.String("error.type", "query_failed"))
		db.queryErrors.Add(ctx, 1, metric.WithAttributes(errorAttrs...))
	}
}

// RecordConnectionMetrics records connection pool metrics
func (db *DB) RecordConnectionMetrics(ctx context.Context) {
	stats := db.Stats()

	// Record active connections
	if db.connectionCount != nil {
		db.connectionCount.Add(ctx, int64(stats.OpenConnections), metric.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("connection.type", "active"),
		))
		db.connectionCount.Add(ctx, -int64(stats.Idle), metric.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("connection.type", "idle"),
		))
	}
}

// GetConnectionStats returns current connection pool statistics
func (db *DB) GetConnectionStats() sql.DBStats {
	return db.Stats()
}
