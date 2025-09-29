package handlers

import (
	"testing"

	"arquivolivre.com.br/otel/internal/database"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	d := &database.DB{DB: sqlDB}

	router := SetupRoutes(d)
	if router == nil {
		t.Fatal("expected non-nil router")
	}

	// Check that the routes are registered by checking the routes info
	routes := router.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}

	// Check for specific expected routes
	expectedRoutes := map[string]bool{
		"GET /health":           false,
		"GET /ready":            false,
		"GET /metrics":          false,
		"GET /api/":             false,
		"GET /api/users":        false,
		"POST /api/users":       false,
		"GET /api/users/:id":    false,
		"PUT /api/users/:id":    false,
		"DELETE /api/users/:id": false,
	}

	for _, route := range routes {
		routeKey := route.Method + " " + route.Path
		if _, exists := expectedRoutes[routeKey]; exists {
			expectedRoutes[routeKey] = true
		}
	}

	for route, found := range expectedRoutes {
		if !found {
			t.Errorf("expected route %s to be registered", route)
		}
	}
}
