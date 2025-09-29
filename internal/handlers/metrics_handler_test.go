package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"arquivolivre.com.br/otel/internal/database"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestNewMetricsHandler(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	d := &database.DB{DB: sqlDB}

	handler := NewMetricsHandler(d)
	if handler == nil {
		t.Fatal("expected non-nil metrics handler")
	}
	if handler.db != d {
		t.Error("expected handler to store provided db")
	}
}

func TestGetMetrics_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	d := &database.DB{DB: sqlDB}
	h := &MetricsHandler{db: d}
	r := gin.New()
	r.GET("/metrics", h.GetMetrics)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
}

func TestGetMetrics_UnhealthyDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	_ = sqlDB.Close() // Close to simulate unhealthy DB
	d := &database.DB{DB: sqlDB}
	h := &MetricsHandler{db: d}
	r := gin.New()
	r.GET("/metrics", h.GetMetrics)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
