package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// OtelHook is a Logrus hook that sends logs to OpenTelemetry
type OtelHook struct {
	logger log.Logger
}

// NewOtelHook creates a new OpenTelemetry hook for Logrus
func NewOtelHook(loggerProvider *sdklog.LoggerProvider) *OtelHook {
	if loggerProvider == nil {
		return &OtelHook{logger: nil}
	}
	return &OtelHook{
		logger: loggerProvider.Logger("otel-example-api"),
	}
}

// Levels returns the log levels this hook should fire for
func (hook *OtelHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Fire is called when a log entry is made
func (hook *OtelHook) Fire(entry *logrus.Entry) error {
	if hook.logger == nil {
		return nil // silently skip if no logger provider
	}
	
	// Convert logrus level to OpenTelemetry severity
	severity := hook.convertLevel(entry.Level)

	// Create log record using the correct API
	record := log.Record{}
	record.SetTimestamp(entry.Time)
	record.SetSeverity(severity)
	record.SetSeverityText(entry.Level.String())
	record.SetBody(log.StringValue(entry.Message))
	record.SetObservedTimestamp(time.Now())

	// Add attributes from logrus fields
	attrs := make([]log.KeyValue, 0, len(entry.Data)+2)

	// Add trace context if available
	if traceID, ok := entry.Data["trace_id"]; ok {
		if traceIDStr, ok := traceID.(string); ok {
			attrs = append(attrs, log.String("trace_id", traceIDStr))
		}
	}
	if spanID, ok := entry.Data["span_id"]; ok {
		if spanIDStr, ok := spanID.(string); ok {
			attrs = append(attrs, log.String("span_id", spanIDStr))
		}
	}

	// Add other fields as attributes
	for key, value := range entry.Data {
		if key == "trace_id" || key == "span_id" {
			continue // Already handled above
		}
		attrs = append(attrs, log.String(key, toString(value)))
	}

	// Add standard attributes
	attrs = append(attrs,
		log.String("logger", "logrus"),
		log.String("level", entry.Level.String()),
	)

	record.AddAttributes(attrs...)

	// Create context with trace information if available
	ctx := context.Background()
	if traceID, ok := entry.Data["trace_id"]; ok {
		if traceIDStr, ok := traceID.(string); ok {
			if spanID, ok := entry.Data["span_id"]; ok {
				if spanIDStr, ok := spanID.(string); ok {
					// Parse trace and span IDs
					if traceIDBytes, err := trace.TraceIDFromHex(traceIDStr); err == nil {
						if spanIDBytes, err := trace.SpanIDFromHex(spanIDStr); err == nil {
							// Create a span context for the log record
							spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
								TraceID: traceIDBytes,
								SpanID:  spanIDBytes,
							})
							ctx = trace.ContextWithSpanContext(ctx, spanCtx)
						}
					}
				}
			}
		}
	}

	// Emit the log record
	hook.logger.Emit(ctx, record)

	return nil
}

// convertLevel converts logrus level to OpenTelemetry severity
func (hook *OtelHook) convertLevel(level logrus.Level) log.Severity {
	switch level {
	case logrus.PanicLevel, logrus.FatalLevel:
		return log.SeverityFatal
	case logrus.ErrorLevel:
		return log.SeverityError
	case logrus.WarnLevel:
		return log.SeverityWarn
	case logrus.InfoLevel:
		return log.SeverityInfo
	case logrus.DebugLevel:
		return log.SeverityDebug
	default:
		return log.SeverityInfo
	}
}

// toString converts any value to string
func toString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// AddOtelHook adds the OpenTelemetry hook to a Logrus logger
func AddOtelHook(logger *logrus.Logger, loggerProvider *sdklog.LoggerProvider) {
	hook := NewOtelHook(loggerProvider)
	logger.AddHook(hook)
}
