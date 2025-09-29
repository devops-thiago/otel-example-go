package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"arquivolivre.com.br/otel/internal/config"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type DatabaseConnector interface {
	Open(driverName, dataSourceName string, options ...otelsql.Option) (*sql.DB, error)
	RegisterDBStatsMetrics(db *sql.DB, options ...otelsql.Option) error
}

type MeterProvider interface {
	Meter(name string, options ...metric.MeterOption) metric.Meter
}

type MetricsFactory interface {
	CreateMetrics(meter metric.Meter) (*DBMetrics, error)
}

type DBMetrics struct {
	QueryDuration       metric.Float64Histogram
	QueryCount          metric.Int64Counter
	QueryErrors         metric.Int64Counter
	ConnectionCount     metric.Int64UpDownCounter
	ConnectionErrors    metric.Int64Counter
	HealthCheckDuration metric.Float64Histogram
}

type ConnectionConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

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

type OtelDatabaseConnector struct{}

func (o *OtelDatabaseConnector) Open(driverName, dataSourceName string, options ...otelsql.Option) (*sql.DB, error) {
	return otelsql.Open(driverName, dataSourceName, options...)
}

func (o *OtelDatabaseConnector) RegisterDBStatsMetrics(db *sql.DB, options ...otelsql.Option) error {
	return otelsql.RegisterDBStatsMetrics(db, options...)
}

type OtelMeterProvider struct{}

func (o *OtelMeterProvider) Meter(name string, options ...metric.MeterOption) metric.Meter {
	return otel.Meter(name, options...)
}

type DefaultMetricsFactory struct{}

func (f *DefaultMetricsFactory) CreateMetrics(meter metric.Meter) (*DBMetrics, error) {
	queryDuration, err := meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create query duration metric: %w", err)
	}

	queryCount, err := meter.Int64Counter(
		"db.query.count",
		metric.WithDescription("Total number of database queries"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create query count metric: %w", err)
	}

	queryErrors, err := meter.Int64Counter(
		"db.query.errors",
		metric.WithDescription("Total number of database query errors"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create query errors metric: %w", err)
	}

	connectionCount, err := meter.Int64UpDownCounter(
		"db.connections.active",
		metric.WithDescription("Number of active database connections"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection count metric: %w", err)
	}

	connectionErrors, err := meter.Int64Counter(
		"db.connection.errors",
		metric.WithDescription("Total number of database connection errors"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection errors metric: %w", err)
	}

	healthCheckDuration, err := meter.Float64Histogram(
		"db.health_check.duration",
		metric.WithDescription("Database health check duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check duration metric: %w", err)
	}

	return &DBMetrics{
		QueryDuration:       queryDuration,
		QueryCount:          queryCount,
		QueryErrors:         queryErrors,
		ConnectionCount:     connectionCount,
		ConnectionErrors:    connectionErrors,
		HealthCheckDuration: healthCheckDuration,
	}, nil
}

func NewConnection(cfg *config.Config) (*DB, error) {
	return NewConnectionWithDeps(
		cfg,
		&OtelDatabaseConnector{},
		&OtelMeterProvider{},
		&DefaultMetricsFactory{},
		DefaultConnectionConfig(),
	)
}

func NewConnectionWithDeps(
	cfg *config.Config,
	connector DatabaseConnector,
	meterProvider MeterProvider,
	metricsFactory MetricsFactory,
	connCfg ConnectionConfig,
) (*DB, error) {
	db, err := connector.Open("mysql", cfg.Database.DSN,
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

	err = configureConnectionPool(db, connCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Register database stats for metrics collection
	err = connector.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
		semconv.DBName(cfg.Database.Name),
	))
	if err != nil {
		log.Printf("Warning: Failed to register database stats metrics: %v", err)
	}

	// Create meter and metrics
	dbInstance, err := createDBWithMetrics(db, meterProvider, metricsFactory)
	if err != nil {
		return nil, fmt.Errorf("failed to create database with metrics: %w", err)
	}

	log.Println("Successfully connected to database with comprehensive OpenTelemetry instrumentation")
	return dbInstance, nil
}

// configureConnectionPool configures the database connection pool
func configureConnectionPool(db *sql.DB, config ConnectionConfig) error {
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	return nil
}

// createDBWithMetrics creates a DB instance with OpenTelemetry metrics
func createDBWithMetrics(db *sql.DB, meterProvider MeterProvider, metricsFactory MetricsFactory) (*DB, error) {
	// Create meter for custom metrics
	meter := meterProvider.Meter("database")

	// Create custom metrics
	metrics, err := metricsFactory.CreateMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics: %w", err)
	}

	return &DB{
		DB:                  db,
		meter:               meter,
		queryDuration:       metrics.QueryDuration,
		queryCount:          metrics.QueryCount,
		queryErrors:         metrics.QueryErrors,
		connectionCount:     metrics.ConnectionCount,
		connectionErrors:    metrics.ConnectionErrors,
		healthCheckDuration: metrics.HealthCheckDuration,
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
