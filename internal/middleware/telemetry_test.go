package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.Use(tm.MetricsMiddleware())
	r.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestGetStatusClass(t *testing.T) {
	cases := map[int]string{199: "1xx", 200: "2xx", 301: "3xx", 404: "4xx", 500: "5xx"}
	for code, cls := range cases {
		if got := getStatusClass(code); got != cls {
			t.Fatalf("%d => %s, got %s", code, cls, got)
		}
	}
}

func TestGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.Use(tm.GinMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		span, endSpan := tm.CustomSpan(c, "test-span", attribute.String("test", "value"))
		assert.NotNil(t, span)
		endSpan()
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanAttribute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		AddSpanAttribute(c, "string_attr", "value")
		AddSpanAttribute(c, "int_attr", 42)
		AddSpanAttribute(c, "int64_attr", int64(64))
		AddSpanAttribute(c, "float64_attr", 3.14)
		AddSpanAttribute(c, "bool_attr", true)
		AddSpanAttribute(c, "other_attr", []string{"test"})
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		AddSpanEvent(c, "test-event", attribute.String("event", "data"))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecordError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		RecordError(c, errors.New("test error"), "test description")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecordMetric(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		tm.RecordMetric(c, "test_metric", 1, attribute.String("test", "value"))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricsMiddlewareWithContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.Use(tm.MetricsMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("test data"))
	req.Header.Set("Content-Length", "9")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricsMiddlewareWithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.Use(tm.MetricsMiddleware())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(errors.New("test error"))
		c.String(http.StatusInternalServerError, "error")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMetricsMiddlewareWithResponseSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := NewTelemetryMiddleware("test-service")
	r := gin.New()
	r.Use(tm.MetricsMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.Header("Content-Length", "5")
		c.String(http.StatusOK, "hello")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanAttributeErrorPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		AddSpanAttribute(c, "key", "value")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanEventErrorPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		AddSpanEvent(c, "test-event")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecordErrorPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		// Test when span is not recording
		RecordError(c, errors.New("test error"), "test description")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanEvent_WithRecordingSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Use telemetry middleware to create a recording span
	middleware := NewTelemetryMiddleware("test-service")
	r.Use(middleware.GinMiddleware())

	r.GET("/test", func(c *gin.Context) {
		// Test AddSpanEvent with a recording span
		AddSpanEvent(c, "test-event", attribute.String("key", "value"))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanEvent_NoRecordingSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		// Test when span is not recording (no telemetry middleware)
		AddSpanEvent(c, "test-event", attribute.String("key", "value"))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddSpanAttribute_WithRecordingSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	middleware := NewTelemetryMiddleware("test-service")
	r.Use(middleware.GinMiddleware())

	r.GET("/test", func(c *gin.Context) {
		// Test AddSpanAttribute with different types
		AddSpanAttribute(c, "string_attr", "test_value")
		AddSpanAttribute(c, "int_attr", 42)
		AddSpanAttribute(c, "int64_attr", int64(100))
		AddSpanAttribute(c, "float64_attr", 3.14)
		AddSpanAttribute(c, "bool_attr", true)
		AddSpanAttribute(c, "other_attr", []string{"array", "value"}) // default case
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecordError_WithRecordingSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	middleware := NewTelemetryMiddleware("test-service")
	r.Use(middleware.GinMiddleware())

	r.GET("/test", func(c *gin.Context) {
		// Test RecordError with a recording span
		RecordError(c, errors.New("test error"), "test error description")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
