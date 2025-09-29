package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/otel/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestGetUsers_WithErrorFromStore(t *testing.T) {
	store := newMockUserStore()
	store.failOnCall["GetAll"] = true

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUsers_WithCountError(t *testing.T) {
	store := newMockUserStore()
	store.failOnCall["Count"] = true

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUser_StoreError(t *testing.T) {
	store := newMockUserStore()
	store.failOnCall["GetByID"] = true

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateUser_StoreError(t *testing.T) {
	store := newMockUserStore()
	store.users = []models.User{{ID: 1, Name: "Test", Email: "test@example.com"}}
	store.failOnCall["Update"] = true

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	upd := models.UpdateUserRequest{Name: func() *string { s := "New Name"; return &s }()}
	b, _ := json.Marshal(upd)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteUser_StoreError(t *testing.T) {
	store := newMockUserStore()
	store.users = []models.User{{ID: 1, Name: "Test", Email: "test@example.com"}}
	store.failOnCall["Delete"] = true

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}