package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestNewLoggerLevelFromEnv(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "debug")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()
	l := NewLogger()
	if l.Level.String() != "debug" {
		t.Fatalf("expected debug, got %s", l.Level.String())
	}
}

func TestWithTraceAndGinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := NewLogger()
	// Trace context with background should not add IDs but should not panic
	_ = l.WithTraceContext(context.Background())

	// Build a minimal gin.Context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/x?y=1", nil)
	c.Request = req
	c.Set("request_id", "req-1")
	_ = l.WithGinContext(c)
}

func TestNewLoggerDifferentLevels(t *testing.T) {
	tests := []struct {
		envValue string
		expected string
	}{
		{"info", "info"},
		{"warn", "warning"},
		{"error", "error"},
		{"invalid", "info"}, // defaults to info
		{"", "info"},        // defaults to info
	}

	for _, test := range tests {
		_ = os.Setenv("LOG_LEVEL", test.envValue)
		l := NewLogger()
		if l.Level.String() != test.expected {
			t.Errorf("for env %s expected %s, got %s", test.envValue, test.expected, l.Level.String())
		}
		_ = os.Unsetenv("LOG_LEVEL")
	}
}

func TestLoggerMethods(t *testing.T) {
	l := NewLogger()
	ctx := context.Background()
	fields := map[string]interface{}{"key": "value"}

	// Test all logging methods
	l.LogError(ctx, nil, "test error", fields)
	l.LogInfo(ctx, "test info", fields)
	l.LogWarn(ctx, "test warn", fields)
	l.LogDebug(ctx, "test debug", nil)
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	logger := GetLogger()
	if logger == nil {
		t.Error("expected non-nil global logger")
	}

	InitGlobalLogger()
	if globalLogger == nil {
		t.Error("expected global logger to be initialized")
	}

	ctx := context.Background()
	fields := map[string]interface{}{"test": "value"}

	// Test global functions
	WithTraceContext(ctx)
	LogError(ctx, nil, "global error", fields)
	LogInfo(ctx, "global info", fields)
	LogWarn(ctx, "global warn", fields)
	LogDebug(ctx, "global debug", nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req
	WithGinContext(c)
}

func TestSetupOtelHook(t *testing.T) {
	globalLogger = NewLogger()
	SetupOtelHook(nil)

	globalLogger = nil
	SetupOtelHook(nil) // Should not panic when globalLogger is nil
}

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := NewLogger()
	r := gin.New()
	r.Use(l.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggerMiddleware_ErrorCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := NewLogger()
	r := gin.New()
	r.Use(l.Middleware())

	// Test 4xx error
	r.GET("/badreq", func(c *gin.Context) {
		c.String(http.StatusBadRequest, "bad request")
	})

	// Test 5xx error
	r.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "server error")
	})

	// Test 4xx
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/badreq", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 5xx
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/error", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWithTraceContext_ValidSpan(t *testing.T) {
	l := NewLogger()

	// Create a context with a valid span
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	entry := l.WithTraceContext(ctx)

	// Verify the entry has trace fields when span is valid
	assert.NotNil(t, entry)
	// The span context should be valid and add trace fields
	if span.SpanContext().IsValid() {
		assert.Contains(t, entry.Data, "trace_id")
		assert.Contains(t, entry.Data, "span_id")
	}
}
