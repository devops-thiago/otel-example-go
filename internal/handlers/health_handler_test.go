package handlers

import (
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

type mockHealthDB struct{ healthy bool }

func (m *mockHealthDB) Health() error {
    if m.healthy {
        return nil
    }
    return errors.New("db down")
}

func TestHealthCheck(t *testing.T) {
    gin.SetMode(gin.TestMode)
    h := &HealthHandler{db: &mockDBWrapper{&mockHealthDB{healthy: true}}}
    r := gin.New()
    r.GET("/health", h.HealthCheck)

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheck_Unhealthy(t *testing.T) {
    gin.SetMode(gin.TestMode)
    h := &HealthHandler{db: &mockDBWrapper{&mockHealthDB{healthy: false}}}
    r := gin.New()
    r.GET("/health", h.HealthCheck)

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// mockDBWrapper adapts mockHealthDB to the subset of methods used by handler
type mockDBWrapper struct{ m *mockHealthDB }

func (w *mockDBWrapper) Health() error { return w.m.Health() }

func TestNewHealthHandler(t *testing.T) {
    db := &mockDBWrapper{&mockHealthDB{healthy: true}}
    handler := NewHealthHandler(db)
    if handler == nil {
        t.Fatal("expected non-nil health handler")
    }
    if handler.db != db {
        t.Error("expected handler to store provided db")
    }
}

func TestReadinessCheck_Ready(t *testing.T) {
    gin.SetMode(gin.TestMode)
    h := &HealthHandler{db: &mockDBWrapper{&mockHealthDB{healthy: true}}}
    r := gin.New()
    r.GET("/ready", h.ReadinessCheck)

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/ready", nil)
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadinessCheck_NotReady(t *testing.T) {
    gin.SetMode(gin.TestMode)
    h := &HealthHandler{db: &mockDBWrapper{&mockHealthDB{healthy: false}}}
    r := gin.New()
    r.GET("/ready", h.ReadinessCheck)

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/ready", nil)
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}


