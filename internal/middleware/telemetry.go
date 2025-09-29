package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryMiddleware provides OpenTelemetry instrumentation
type TelemetryMiddleware struct {
	tracer          trace.Tracer
	meter           metric.Meter
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
	activeRequests  metric.Int64UpDownCounter
}

// NewTelemetryMiddleware creates a new telemetry middleware
func NewTelemetryMiddleware(serviceName string) *TelemetryMiddleware {
	tracer := otel.Tracer(serviceName)
	meter := otel.Meter(serviceName)

	// Create metrics instruments
	requestCounter, _ := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)

	requestDuration, _ := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)

	requestSize, _ := meter.Int64Histogram(
		"http_request_size_bytes",
		metric.WithDescription("HTTP request size in bytes"),
		metric.WithUnit("bytes"),
	)

	responseSize, _ := meter.Int64Histogram(
		"http_response_size_bytes",
		metric.WithDescription("HTTP response size in bytes"),
		metric.WithUnit("bytes"),
	)

	activeRequests, _ := meter.Int64UpDownCounter(
		"http_active_requests",
		metric.WithDescription("Number of active HTTP requests"),
	)

	return &TelemetryMiddleware{
		tracer:          tracer,
		meter:           meter,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		requestSize:     requestSize,
		responseSize:    responseSize,
		activeRequests:  activeRequests,
	}
}

// GinMiddleware returns Gin middleware for OpenTelemetry tracing
func (tm *TelemetryMiddleware) GinMiddleware() gin.HandlerFunc {
	return otelgin.Middleware("otel-example-api")
}

// MetricsMiddleware returns Gin middleware for custom metrics collection
func (tm *TelemetryMiddleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Common attributes for metrics
		commonAttrs := []attribute.KeyValue{
			attribute.String("method", c.Request.Method),
			attribute.String("route", c.FullPath()),
		}

		// Increment active requests counter
		tm.activeRequests.Add(c.Request.Context(), 1, metric.WithAttributes(commonAttrs...))
		defer tm.activeRequests.Add(c.Request.Context(), -1, metric.WithAttributes(commonAttrs...))

		// Record request size
		if c.Request.ContentLength > 0 {
			tm.requestSize.Record(c.Request.Context(), c.Request.ContentLength,
				metric.WithAttributes(commonAttrs...))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get response size from header or estimate from body
		responseSize := int64(0)
		if sizeHeader := c.Writer.Header().Get("Content-Length"); sizeHeader != "" {
			if size, err := strconv.ParseInt(sizeHeader, 10, 64); err == nil {
				responseSize = size
			}
		} else {
			// Estimate response size from writer
			responseSize = int64(c.Writer.Size())
		}

		// Final attributes including status
		finalAttrs := append(commonAttrs,
			attribute.String("status_code", strconv.Itoa(c.Writer.Status())),
			attribute.String("status_class", getStatusClass(c.Writer.Status())),
		)

		// Record metrics
		tm.requestCounter.Add(c.Request.Context(), 1, metric.WithAttributes(finalAttrs...))
		tm.requestDuration.Record(c.Request.Context(), duration, metric.WithAttributes(finalAttrs...))

		if responseSize > 0 {
			tm.responseSize.Record(c.Request.Context(), responseSize, metric.WithAttributes(finalAttrs...))
		}

		// Add custom span attributes
		if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
			span.SetAttributes(
				attribute.String("http.route", c.FullPath()),
				attribute.String("http.method", c.Request.Method),
				attribute.Int("http.status_code", c.Writer.Status()),
				attribute.String("http.status_class", getStatusClass(c.Writer.Status())),
				attribute.Int64("http.request.size", c.Request.ContentLength),
				attribute.Int64("http.response.size", responseSize),
				attribute.String("user.agent", c.Request.UserAgent()),
				attribute.String("client.ip", c.ClientIP()),
				attribute.Float64("http.duration", duration),
			)

			// Add error information if present
			if len(c.Errors) > 0 {
				span.SetAttributes(
					attribute.String("error.message", c.Errors.String()),
					attribute.Bool("error", true),
				)
				span.SetStatus(codes.Error, c.Errors.String())
			} else if c.Writer.Status() >= 400 {
				span.SetAttributes(attribute.Bool("error", true))
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		}
	}
}

// getStatusClass returns the HTTP status class (2xx, 3xx, 4xx, 5xx)
func getStatusClass(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "1xx"
	}
}

// CustomSpan creates a custom span for manual instrumentation
func (tm *TelemetryMiddleware) CustomSpan(c *gin.Context, spanName string, attrs ...attribute.KeyValue) (trace.Span, func()) {
	ctx, span := tm.tracer.Start(c.Request.Context(), spanName)
	c.Request = c.Request.WithContext(ctx)

	// Add custom attributes
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	return span, func() {
		span.End()
	}
}

// AddSpanAttribute adds an attribute to the current span
func AddSpanAttribute(c *gin.Context, key string, value interface{}) {
	if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		default:
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(c *gin.Context, name string, attrs ...attribute.KeyValue) {
	if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordError records an error in the current span
func RecordError(c *gin.Context, err error, description string) {
	if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("error.description", description),
			attribute.Bool("error", true),
		)
		span.SetStatus(codes.Error, description)
	}
}

// RecordMetric records a custom metric
func (tm *TelemetryMiddleware) RecordMetric(c *gin.Context, name string, value int64, attrs ...attribute.KeyValue) {
	// Create a counter for custom metrics if it doesn't exist
	counter, _ := tm.meter.Int64Counter(
		name,
		metric.WithDescription("Custom metric: "+name),
	)
	counter.Add(c.Request.Context(), value, metric.WithAttributes(attrs...))
}
