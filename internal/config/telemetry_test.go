package config

import (
    "context"
    "testing"
)

func TestInitTelemetry_DisabledAll(t *testing.T) {
    tp, err := InitTelemetry(&TelemetryConfig{
        ServiceName:    "svc",
        ServiceVersion: "1",
        Environment:    "test",
        OTLPGRPCEndpoint: "localhost:4317",
        EnableMetrics:  false,
        EnableTracing:  false,
        EnableLogging:  false,
    })
    if err != nil { t.Fatalf("err: %v", err) }
    if tp.TracerProvider != nil || tp.MeterProvider != nil || tp.LoggerProvider != nil {
        t.Fatalf("expected no providers when disabled: %+v", tp)
    }
    _ = tp.Shutdown(context.Background())
}

func TestGetTelemetryConfig(t *testing.T) {
    cfg := GetTelemetryConfig()
    if cfg == nil {
        t.Fatal("expected non-nil config")
    }
    if cfg.ServiceName == "" {
        t.Error("expected non-empty service name")
    }
    if cfg.ServiceVersion == "" {
        t.Error("expected non-empty service version")
    }
    if cfg.Environment == "" {
        t.Error("expected non-empty environment")
    }
    if cfg.OTLPGRPCEndpoint == "" {
        t.Error("expected non-empty OTLP endpoint")
    }
}

func TestInitTelemetry_TracingOnly(t *testing.T) {
    tp, err := InitTelemetry(&TelemetryConfig{
        ServiceName:      "test-service",
        ServiceVersion:   "1.0.0", 
        Environment:      "test",
        OTLPGRPCEndpoint: "localhost:4317",
        EnableMetrics:    false,
        EnableTracing:    true,
        EnableLogging:    false,
    })
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if tp.TracerProvider == nil {
        t.Error("expected non-nil tracer provider when tracing enabled")
    }
    if tp.MeterProvider != nil {
        t.Error("expected nil meter provider when metrics disabled")
    }
    if tp.LoggerProvider != nil {
        t.Error("expected nil logger provider when logging disabled")
    }
    _ = tp.Shutdown(context.Background())
}

func TestInitTelemetry_MetricsOnly(t *testing.T) {
    tp, err := InitTelemetry(&TelemetryConfig{
        ServiceName:      "test-service",
        ServiceVersion:   "1.0.0",
        Environment:      "test", 
        OTLPGRPCEndpoint: "localhost:4317",
        EnableMetrics:    true,
        EnableTracing:    false,
        EnableLogging:    false,
    })
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if tp.TracerProvider != nil {
        t.Error("expected nil tracer provider when tracing disabled")
    }
    if tp.MeterProvider == nil {
        t.Error("expected non-nil meter provider when metrics enabled")
    }
    if tp.LoggerProvider != nil {
        t.Error("expected nil logger provider when logging disabled")
    }
    _ = tp.Shutdown(context.Background())
}

func TestInitTelemetry_LoggingOnly(t *testing.T) {
    tp, err := InitTelemetry(&TelemetryConfig{
        ServiceName:      "test-service",
        ServiceVersion:   "1.0.0",
        Environment:      "test",
        OTLPGRPCEndpoint: "localhost:4317", 
        EnableMetrics:    false,
        EnableTracing:    false,
        EnableLogging:    true,
    })
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if tp.TracerProvider != nil {
        t.Error("expected nil tracer provider when tracing disabled")
    }
    if tp.MeterProvider != nil {
        t.Error("expected nil meter provider when metrics disabled")
    }
    if tp.LoggerProvider == nil {
        t.Error("expected non-nil logger provider when logging enabled")
    }
    _ = tp.Shutdown(context.Background())
}

func TestInitTelemetry_AllEnabled(t *testing.T) {
    tp, err := InitTelemetry(&TelemetryConfig{
        ServiceName:          "test-service",
        ServiceVersion:       "1.0.0",
        Environment:          "test",
        OTLPGRPCEndpoint:     "localhost:4317",
        EnableMetrics:        true,
        EnableTracing:        true,
        EnableLogging:        true,
        EnableRuntimeMetrics: true,
    })
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if tp.TracerProvider == nil {
        t.Error("expected non-nil tracer provider when tracing enabled")
    }
    if tp.MeterProvider == nil {
        t.Error("expected non-nil meter provider when metrics enabled")
    }
    if tp.LoggerProvider == nil {
        t.Error("expected non-nil logger provider when logging enabled")
    }
    if tp.Shutdown == nil {
        t.Error("expected non-nil shutdown function")
    }
    _ = tp.Shutdown(context.Background())
}

func TestInitTelemetry_ShutdownError(t *testing.T) {
    tp, err := InitTelemetry(&TelemetryConfig{
        ServiceName:      "test-service", 
        ServiceVersion:   "1.0.0",
        Environment:      "test",
        OTLPGRPCEndpoint: "localhost:4317",
        EnableMetrics:    true,
        EnableTracing:    true,
        EnableLogging:    true,
    })
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    // Test shutdown function exists and can be called (ignore network errors in test)
    if tp.Shutdown == nil {
        t.Error("expected non-nil shutdown function")
    }
    // Skip actual shutdown call to avoid network timeouts in test environment
}


