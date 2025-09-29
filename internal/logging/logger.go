package logging

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// Logger wraps logrus with OpenTelemetry integration
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new structured logger with OpenTelemetry integration
func NewLogger() *Logger {
	logger := logrus.New()

	// Set JSON formatter for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Set log level from environment
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return &Logger{Logger: logger}
}

// WithTraceContext adds trace context to log entries
func (l *Logger) WithTraceContext(ctx context.Context) *logrus.Entry {
	entry := l.WithFields(logrus.Fields{})

	// Extract trace information from context
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		spanContext := span.SpanContext()
		entry = entry.WithFields(logrus.Fields{
			"trace_id": spanContext.TraceID().String(),
			"span_id":  spanContext.SpanID().String(),
		})
	}

	return entry
}

// WithGinContext adds Gin context information to log entries
func (l *Logger) WithGinContext(c *gin.Context) *logrus.Entry {
	entry := l.WithTraceContext(c.Request.Context())

	// Add request information
	entry = entry.WithFields(logrus.Fields{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"user_agent": c.Request.UserAgent(),
		"client_ip":  c.ClientIP(),
		"request_id": c.GetString("request_id"),
	})

	return entry
}

// Middleware returns a Gin middleware for request logging
func (l *Logger) Middleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		entry := l.WithFields(logrus.Fields{
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency.String(),
			"client_ip":   param.ClientIP,
			"user_agent":  param.Request.UserAgent(),
		})

		// Add trace context if available
		if span := trace.SpanFromContext(param.Request.Context()); span.SpanContext().IsValid() {
			spanContext := span.SpanContext()
			entry = entry.WithFields(logrus.Fields{
				"trace_id": spanContext.TraceID().String(),
				"span_id":  spanContext.SpanID().String(),
			})
		}

		// Log based on status code
		if param.StatusCode >= 500 {
			entry.Error("HTTP request completed with server error")
		} else if param.StatusCode >= 400 {
			entry.Warn("HTTP request completed with client error")
		} else {
			entry.Info("HTTP request completed successfully")
		}

		return "" // Return empty string since we're using structured logging
	})
}

// LogError logs an error with trace context
func (l *Logger) LogError(ctx context.Context, err error, message string, fields map[string]interface{}) {
	entry := l.WithTraceContext(ctx).WithError(err)

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Error(message)
}

// LogInfo logs info with trace context
func (l *Logger) LogInfo(ctx context.Context, message string, fields map[string]interface{}) {
	entry := l.WithTraceContext(ctx)

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Info(message)
}

// LogWarn logs warning with trace context
func (l *Logger) LogWarn(ctx context.Context, message string, fields map[string]interface{}) {
	entry := l.WithTraceContext(ctx)

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Warn(message)
}

// LogDebug logs debug with trace context
func (l *Logger) LogDebug(ctx context.Context, message string, fields map[string]interface{}) {
	entry := l.WithTraceContext(ctx)

	if fields != nil {
		entry = entry.WithFields(fields)
	}

	entry.Debug(message)
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger() {
	globalLogger = NewLogger()
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		InitGlobalLogger()
	}
	return globalLogger
}

// Helper functions for global logger access
func WithTraceContext(ctx context.Context) *logrus.Entry {
	return GetLogger().WithTraceContext(ctx)
}

func WithGinContext(c *gin.Context) *logrus.Entry {
	return GetLogger().WithGinContext(c)
}

func LogError(ctx context.Context, err error, message string, fields map[string]interface{}) {
	GetLogger().LogError(ctx, err, message, fields)
}

func LogInfo(ctx context.Context, message string, fields map[string]interface{}) {
	GetLogger().LogInfo(ctx, message, fields)
}

func LogWarn(ctx context.Context, message string, fields map[string]interface{}) {
	GetLogger().LogWarn(ctx, message, fields)
}

func LogDebug(ctx context.Context, message string, fields map[string]interface{}) {
	GetLogger().LogDebug(ctx, message, fields)
}

// SetupOtelHook sets up the OpenTelemetry hook for the global logger
func SetupOtelHook(loggerProvider *sdklog.LoggerProvider) {
	if globalLogger != nil {
		AddOtelHook(globalLogger.Logger, loggerProvider)
	}
}
