package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"arquivolivre.com.br/otel/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUserStore struct {
	users      []models.User
	nextID     int
	failOnCall map[string]bool
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{
		users:      []models.User{},
		nextID:     1,
		failOnCall: map[string]bool{},
	}
}

func (m *mockUserStore) GetAll(_ context.Context, limit, offset int) ([]models.User, error) {
	if m.failOnCall["GetAll"] {
		return nil, fmt.Errorf("mock error")
	}
	end := offset + limit
	if end > len(m.users) {
		end = len(m.users)
	}
	if offset > len(m.users) {
		offset = len(m.users)
	}
	return m.users[offset:end], nil
}

func (m *mockUserStore) GetByID(_ context.Context, id int) (*models.User, error) {
	if m.failOnCall["GetByID"] {
		return nil, fmt.Errorf("mock error")
	}
	for i := range m.users {
		if m.users[i].ID == id {
			u := m.users[i]
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserStore) Create(_ context.Context, req models.CreateUserRequest) (*models.User, error) {
	u := models.User{ID: m.nextID, Name: req.Name, Email: req.Email, Bio: req.Bio}
	m.nextID++
	m.users = append(m.users, u)
	return &u, nil
}

func (m *mockUserStore) Update(_ context.Context, id int, req models.UpdateUserRequest) (*models.User, error) {
	if m.failOnCall["Update"] {
		return nil, fmt.Errorf("mock error")
	}
	for i := range m.users {
		if m.users[i].ID == id {
			if req.Name != nil {
				m.users[i].Name = *req.Name
			}
			if req.Email != nil {
				m.users[i].Email = *req.Email
			}
			if req.Bio != nil {
				m.users[i].Bio = *req.Bio
			}
			u := m.users[i]
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockUserStore) Delete(_ context.Context, id int) error {
	if m.failOnCall["Delete"] {
		return fmt.Errorf("mock error")
	}
	for i := range m.users {
		if m.users[i].ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func (m *mockUserStore) Count(_ context.Context) (int, error) {
	if m.failOnCall["Count"] {
		return 0, fmt.Errorf("mock error")
	}
	return len(m.users), nil
}
func (m *mockUserStore) GetByEmail(_ context.Context, email string) (*models.User, error) {
	for i := range m.users {
		if m.users[i].Email == email {
			u := m.users[i]
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func setupRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	users := api.Group("/users")
	users.GET("", handler.GetUsers)
	users.POST("", handler.CreateUser)
	users.GET(":id", handler.GetUser)
	users.PUT(":id", handler.UpdateUser)
	users.DELETE(":id", handler.DeleteUser)
	return r
}

func TestCreateAndGetUser(t *testing.T) {
	store := newMockUserStore()
	handler := NewUserHandler(store)
	r := setupRouter(handler)

	body := models.CreateUserRequest{Name: "Alice", Email: "alice@example.com", Bio: "bio"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/users?page=1&limit=10", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestGetUserNotFound(t *testing.T) {
	store := newMockUserStore()
	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateAndDeleteUser(t *testing.T) {
	store := newMockUserStore()
	_, _ = store.Create(context.TODO(), models.CreateUserRequest{Name: "Bob", Email: "bob@example.com"})

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	newName := "Bobby"
	upd := models.UpdateUserRequest{Name: &newName}
	b, _ := json.Marshal(upd)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestCreateUserInvalidPayload(t *testing.T) {
	store := newMockUserStore()
	handler := NewUserHandler(store)
	r := setupRouter(handler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateUserConflict(t *testing.T) {
	store := newMockUserStore()
	_, _ = store.Create(context.Background(), models.CreateUserRequest{Name: "X", Email: "x@example.com"})

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	body := models.CreateUserRequest{Name: "Y", Email: "x@example.com"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateUserEmailConflict(t *testing.T) {
	store := newMockUserStore()
	_, _ = store.Create(context.Background(), models.CreateUserRequest{Name: "A", Email: "a@example.com"}) // id=1
	_, _ = store.Create(context.Background(), models.CreateUserRequest{Name: "B", Email: "b@example.com"}) // id=2

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	newEmail := "a@example.com" // conflicts with id=1
	upd := models.UpdateUserRequest{Email: &newEmail}
	b, _ := json.Marshal(upd)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/users/2", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetUserInvalidID(t *testing.T) {
	store := newMockUserStore()
	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserSuccess(t *testing.T) {
	store := newMockUserStore()
	_, _ = store.Create(context.Background(), models.CreateUserRequest{Name: "Alice", Email: "alice@example.com"})

	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateUserInvalidID(t *testing.T) {
	store := newMockUserStore()
	handler := NewUserHandler(store)
	r := setupRouter(handler)

	upd := models.UpdateUserRequest{Name: func() *string { s := "New Name"; return &s }()}
	b, _ := json.Marshal(upd)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/users/invalid", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserInvalidID(t *testing.T) {
	store := newMockUserStore()
	handler := NewUserHandler(store)
	r := setupRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/users/invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
