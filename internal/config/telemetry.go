package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// TelemetryConfig holds telemetry configuration
type TelemetryConfig struct {
	ServiceName          string
	ServiceVersion       string
	Environment          string
	OTLPGRPCEndpoint     string
	EnableMetrics        bool
	EnableTracing        bool
	EnableLogging        bool
	EnableRuntimeMetrics bool
}

// TelemetryProvider holds the telemetry providers
type TelemetryProvider struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	Shutdown       func(context.Context) error
}

// InitTelemetry initializes OpenTelemetry with tracing and metrics
func InitTelemetry(cfg *TelemetryConfig) (*TelemetryProvider, error) {
	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error
	var tracerProvider *sdktrace.TracerProvider
	var meterProvider *sdkmetric.MeterProvider
	var loggerProvider *sdklog.LoggerProvider

	// Initialize tracing if enabled
	if cfg.EnableTracing {
		tp, shutdown, err := initTracing(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
		tracerProvider = tp
		shutdownFuncs = append(shutdownFuncs, shutdown)

		// Set global tracer provider
		otel.SetTracerProvider(tracerProvider)

		// Set global propagator
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	// Initialize metrics if enabled
	if cfg.EnableMetrics {
		mp, shutdown, err := initMetrics(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
		meterProvider = mp
		shutdownFuncs = append(shutdownFuncs, shutdown)

		// Set global meter provider
		otel.SetMeterProvider(meterProvider)
	}

	// Initialize logging if enabled
	if cfg.EnableLogging {
		lp, shutdown, err := initLogging(ctx, res, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize logging: %w", err)
		}
		loggerProvider = lp
		shutdownFuncs = append(shutdownFuncs, shutdown)

		// Set global logger provider
		// Note: otel.SetLoggerProvider doesn't exist in the current API
		// The logger provider is used directly by the bridge
	}

	// Combined shutdown function
	shutdown := func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("telemetry shutdown errors: %v", errs)
		}
		return nil
	}

	return &TelemetryProvider{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
		LoggerProvider: loggerProvider,
		Shutdown:       shutdown,
	}, nil
}

// initTracing initializes tracing with OTLP gRPC exporter
func initTracing(ctx context.Context, res *resource.Resource, cfg *TelemetryConfig) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	otlpExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPGRPCEndpoint),
		otlptracegrpc.WithInsecure(), // Use WithTLSClientConfig for secure connections
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP gRPC trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(otlpExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	log.Println("OTLP gRPC trace exporter initialized for Grafana Tempo via Alloy")
	return tracerProvider, tracerProvider.Shutdown, nil
}

// initMetrics initializes metrics with OTLP gRPC exporter
func initMetrics(ctx context.Context, res *resource.Resource, cfg *TelemetryConfig) (*sdkmetric.MeterProvider, func(context.Context) error, error) {
	otlpExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTLPGRPCEndpoint),
		otlpmetricgrpc.WithInsecure(), // Use WithTLSClientConfig for secure connections
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP gRPC metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(otlpExporter, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)

	// Start runtime metrics collection if enabled
	if cfg.EnableRuntimeMetrics {
		err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(15 * time.Second))
		if err != nil {
			log.Printf("Warning: Failed to start runtime metrics collection: %v", err)
		} else {
			log.Println("Go runtime metrics collection started")
		}
	}

	log.Println("OTLP gRPC metric exporter initialized for Grafana Mimir via Alloy")
	return meterProvider, meterProvider.Shutdown, nil
}

// initLogging initializes logging with OTLP gRPC exporter
func initLogging(ctx context.Context, res *resource.Resource, cfg *TelemetryConfig) (*sdklog.LoggerProvider, func(context.Context) error, error) {
	otlpExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.OTLPGRPCEndpoint),
		otlploggrpc.WithInsecure(), // Use WithTLSClientConfig for secure connections
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP gRPC log exporter: %w", err)
	}

	// Create batch processor
	processor := sdklog.NewBatchProcessor(otlpExporter)

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)

	log.Println("OTLP gRPC log exporter initialized for Grafana Loki via Alloy")
	return loggerProvider, loggerProvider.Shutdown, nil
}

// GetTelemetryConfig creates telemetry configuration from environment
func GetTelemetryConfig() *TelemetryConfig {
	return &TelemetryConfig{
		ServiceName:          getEnv("OTEL_SERVICE_NAME", "otel-example-api"),
		ServiceVersion:       getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:          getEnv("OTEL_ENVIRONMENT", getEnv("APP_ENV", "development")),
		OTLPGRPCEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		EnableMetrics:        getEnv("OTEL_ENABLE_METRICS", "true") == "true",
		EnableTracing:        getEnv("OTEL_ENABLE_TRACING", "true") == "true",
		EnableLogging:        getEnv("OTEL_ENABLE_LOGGING", "true") == "true",
		EnableRuntimeMetrics: getEnv("OTEL_ENABLE_RUNTIME_METRICS", "true") == "true",
	}
}
