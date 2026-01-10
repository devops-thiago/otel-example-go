package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"arquivolivre.com.br/otel/internal/config"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestDBHealth_Closed(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	d := &DB{DB: sqlDB}
	_ = sqlDB.Close()
	if err := d.Health(); err == nil {
		t.Fatal("expected error when DB is closed")
	}
}

func TestGetConnectionStats_Zero(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	d := &DB{DB: sqlDB}
	_ = d.GetConnectionStats()
}

func TestDBClose(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	d := &DB{DB: sqlDB}
	mock.ExpectClose()
	err = d.Close()
	if err != nil {
		t.Errorf("expected no error closing DB, got: %v", err)
	}
}

func TestRecordQueryMetrics_Success(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	d := &DB{DB: sqlDB}
	d.RecordQueryMetrics(context.Background(), "SELECT", "users", 100*1000000, nil)
}

func TestRecordQueryMetrics_WithMetrics(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	d := &DB{DB: sqlDB}

	d.RecordQueryMetrics(context.Background(), "SELECT", "users", 100*1000000, nil)
	d.RecordQueryMetrics(context.Background(), "INSERT", "users", 50*1000000, fmt.Errorf("constraint error"))
}

func TestRecordQueryMetrics_Error(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	d := &DB{DB: sqlDB}
	d.RecordQueryMetrics(context.Background(), "SELECT", "users", 100*1000000, fmt.Errorf("query error"))
}

func TestRecordConnectionMetrics(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	d := &DB{DB: sqlDB}
	d.RecordConnectionMetrics(context.Background())
}

func TestRecordConnectionMetrics_WithMetrics(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	d := &DB{DB: sqlDB}
	d.RecordConnectionMetrics(context.Background())
}

func TestDBHealth_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectPing()

	d := &DB{DB: sqlDB}
	if err := d.Health(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestNewConnection_InvalidDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "invalid-dsn-format",
			Name: "test",
		},
	}

	_, err := NewConnection(cfg)
	if err == nil {
		t.Error("expected error with invalid DSN, got nil")
	}
}

func TestNewConnection_PingFails(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "user:password@tcp(nonexistent-host:3306)/dbname",
			Name: "test",
		},
	}

	_, err := NewConnection(cfg)
	if err == nil {
		t.Error("expected error with unreachable host, got nil")
	}
}

func TestNewConnection_EmptyDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "",
			Name: "test",
		},
	}

	_, err := NewConnection(cfg)
	if err == nil {
		t.Error("expected error with empty DSN, got nil")
	}
}

func TestNewConnection_MalformedDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "malformed:dsn:format",
			Name: "test",
		},
	}

	_, err := NewConnection(cfg)
	if err == nil {
		t.Error("expected error with malformed DSN, got nil")
	}
}

type mockDatabaseConnector struct {
	openError          error
	registerStatsError error
}

func (m *mockDatabaseConnector) Open(driverName, dataSourceName string, options ...otelsql.Option) (*sql.DB, error) {
	if m.openError != nil {
		return nil, m.openError
	}

	// Use sqlmock for testing
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	return sqlDB, nil
}

func (m *mockDatabaseConnector) RegisterDBStatsMetrics(db *sql.DB, options ...otelsql.Option) error {
	return m.registerStatsError
}

// NoopMeterProvider implements MeterProvider using noop implementation
type NoopMeterProvider struct{}

func (n *NoopMeterProvider) Meter(name string, options ...metric.MeterOption) metric.Meter {
	return noop.NewMeterProvider().Meter(name, options...)
}

// Tests for refactored functions using noop OpenTelemetry

func TestNewConnectionWithDeps_Success(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "test:dsn",
			Name: "testdb",
		},
	}

	connector := &mockDatabaseConnector{}
	meterProvider := &NoopMeterProvider{}
	metricsFactory := &DefaultMetricsFactory{}
	connCfg := DefaultConnectionConfig()

	db, err := NewConnectionWithDeps(cfg, connector, meterProvider, metricsFactory, connCfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if db == nil {
		t.Error("expected non-nil DB, got nil")
	}

	if db != nil {
		// Test that the DB has all the expected metrics
		if db.meter == nil {
			t.Error("expected non-nil meter")
		}
		if db.queryDuration == nil {
			t.Error("expected non-nil queryDuration metric")
		}
		if db.queryCount == nil {
			t.Error("expected non-nil queryCount metric")
		}
		if db.queryErrors == nil {
			t.Error("expected non-nil queryErrors metric")
		}
		if db.connectionCount == nil {
			t.Error("expected non-nil connectionCount metric")
		}
		if db.connectionErrors == nil {
			t.Error("expected non-nil connectionErrors metric")
		}
		if db.healthCheckDuration == nil {
			t.Error("expected non-nil healthCheckDuration metric")
		}

		// Clean up
		_ = db.Close()
	}
}

func TestNewConnectionWithDeps_ConnectorOpenError(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "test:dsn",
			Name: "testdb",
		},
	}

	connector := &mockDatabaseConnector{openError: fmt.Errorf("connection failed")}
	meterProvider := &NoopMeterProvider{}
	metricsFactory := &DefaultMetricsFactory{}
	connCfg := DefaultConnectionConfig()

	_, err := NewConnectionWithDeps(cfg, connector, meterProvider, metricsFactory, connCfg)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "failed to open database connection: connection failed" {
		t.Errorf("expected specific error message, got: %v", err)
	}
}

func TestNewConnectionWithDeps_RegisterStatsError(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:  "test:dsn",
			Name: "testdb",
		},
	}

	connector := &mockDatabaseConnector{registerStatsError: fmt.Errorf("stats registration failed")}
	meterProvider := &NoopMeterProvider{}
	metricsFactory := &DefaultMetricsFactory{}
	connCfg := DefaultConnectionConfig()

	// This should succeed but log a warning
	db, err := NewConnectionWithDeps(cfg, connector, meterProvider, metricsFactory, connCfg)
	if err != nil {
		t.Errorf("expected no error despite stats registration failure, got: %v", err)
	}
	if db == nil {
		t.Error("expected non-nil DB despite stats registration failure")
	}

	// Clean up
	if db != nil {
		_ = db.Close()
	}
}

func TestConfigureConnectionPool(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	config := ConnectionConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    3,
		ConnMaxLifetime: time.Minute,
	}

	err = configureConnectionPool(sqlDB, config)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify configuration was applied
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("expected MaxOpenConnections=10, got: %d", stats.MaxOpenConnections)
	}
}

func TestCreateDBWithMetrics_Success(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	meterProvider := &NoopMeterProvider{}
	metricsFactory := &DefaultMetricsFactory{}

	db, err := createDBWithMetrics(sqlDB, meterProvider, metricsFactory)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if db == nil {
		t.Error("expected non-nil DB, got nil")
		return
	}

	// Verify all metrics are properly created
	if db.meter == nil {
		t.Error("expected non-nil meter")
	}
	if db.queryDuration == nil {
		t.Error("expected non-nil queryDuration metric")
	}
	if db.queryCount == nil {
		t.Error("expected non-nil queryCount metric")
	}
	if db.queryErrors == nil {
		t.Error("expected non-nil queryErrors metric")
	}
	if db.connectionCount == nil {
		t.Error("expected non-nil connectionCount metric")
	}
	if db.connectionErrors == nil {
		t.Error("expected non-nil connectionErrors metric")
	}
	if db.healthCheckDuration == nil {
		t.Error("expected non-nil healthCheckDuration metric")
	}
}

func TestDefaultConnectionConfig(t *testing.T) {
	config := DefaultConnectionConfig()
	if config.MaxOpenConns != 25 {
		t.Errorf("expected MaxOpenConns=25, got: %d", config.MaxOpenConns)
	}
	if config.MaxIdleConns != 5 {
		t.Errorf("expected MaxIdleConns=5, got: %d", config.MaxIdleConns)
	}
	if config.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("expected ConnMaxLifetime=5m, got: %v", config.ConnMaxLifetime)
	}
}

func TestDefaultMetricsFactory_CreateMetrics_Success(t *testing.T) {
	factory := &DefaultMetricsFactory{}
	meterProvider := &NoopMeterProvider{}
	meter := meterProvider.Meter("test")

	metrics, err := factory.CreateMetrics(meter)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if metrics == nil {
		t.Error("expected non-nil metrics, got nil")
		return
	}

	// Verify all metrics are created
	if metrics.QueryDuration == nil {
		t.Error("expected non-nil QueryDuration")
	}
	if metrics.QueryCount == nil {
		t.Error("expected non-nil QueryCount")
	}
	if metrics.QueryErrors == nil {
		t.Error("expected non-nil QueryErrors")
	}
	if metrics.ConnectionCount == nil {
		t.Error("expected non-nil ConnectionCount")
	}
	if metrics.ConnectionErrors == nil {
		t.Error("expected non-nil ConnectionErrors")
	}
	if metrics.HealthCheckDuration == nil {
		t.Error("expected non-nil HealthCheckDuration")
	}
}

// Test the real implementations
func TestOtelDatabaseConnector_Open(t *testing.T) {
	connector := &OtelDatabaseConnector{}

	// Test with invalid DSN - should return error
	_, err := connector.Open("mysql", "invalid-dsn")
	if err == nil {
		t.Error("expected error with invalid DSN")
	}
}

func TestOtelDatabaseConnector_RegisterDBStatsMetrics(t *testing.T) {
	connector := &OtelDatabaseConnector{}

	// Use sqlmock to create a test DB
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	// RegisterDBStatsMetrics should work with a valid DB connection
	err = connector.RegisterDBStatsMetrics(sqlDB)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestOtelMeterProvider_Methods(t *testing.T) {
	provider := &OtelMeterProvider{}

	// Just test that the method exists
	meter := provider.Meter("test")
	if meter == nil {
		t.Error("expected non-nil meter")
	}
}

func TestNoopMeterProvider_Methods(t *testing.T) {
	provider := &NoopMeterProvider{}

	// Test that the noop provider works
	meter := provider.Meter("test")
	if meter == nil {
		t.Error("expected non-nil meter from noop provider")
	}

	// Test that we can create metrics without errors
	histogram, err := meter.Float64Histogram("test.histogram")
	if err != nil {
		t.Errorf("expected no error creating histogram, got: %v", err)
	}
	if histogram == nil {
		t.Error("expected non-nil histogram")
	}

	counter, err := meter.Int64Counter("test.counter")
	if err != nil {
		t.Errorf("expected no error creating counter, got: %v", err)
	}
	if counter == nil {
		t.Error("expected non-nil counter")
	}
}
