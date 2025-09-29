package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"example/otel/internal/logging"
	"example/otel/internal/middleware"
	"example/otel/internal/models"
	"example/otel/internal/repository"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userRepo *repository.UserRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GetUsers handles GET /api/users
func (h *UserHandler) GetUsers(c *gin.Context) {
	// Create custom span for this operation
	span := trace.SpanFromContext(c.Request.Context())
	span.SetAttributes(
		attribute.String("handler", "GetUsers"),
		attribute.String("operation", "list_users"),
	)

	// Log the request
	logging.WithGinContext(c).Info("Getting users list")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Add pagination info to trace and logs
	span.SetAttributes(
		attribute.Int("pagination.page", page),
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	middleware.AddSpanEvent(c, "pagination_parsed",
		attribute.Int("page", page),
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
	)

	// Get users from repository
	users, err := h.userRepo.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		logging.LogError(c.Request.Context(), err, "Failed to retrieve users from database", map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"offset": offset,
		})
		middleware.RecordError(c, err, "Failed to retrieve users from database")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve users",
		})
		return
	}

	middleware.AddSpanEvent(c, "users_retrieved", attribute.Int("count", len(users)))

	// Get total count for pagination
	total, err := h.userRepo.Count(c.Request.Context())
	if err != nil {
		logging.LogError(c.Request.Context(), err, "Failed to count users in database", nil)
		middleware.RecordError(c, err, "Failed to count users in database")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to count users",
		})
		return
	}

	middleware.AddSpanEvent(c, "total_count_retrieved", attribute.Int("total", total))

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Add result metrics to trace
	span.SetAttributes(
		attribute.Int("result.users_count", len(users)),
		attribute.Int("result.total_count", total),
		attribute.Int("result.total_pages", totalPages),
	)

	// Log successful response
	logging.WithGinContext(c).WithFields(map[string]interface{}{
		"users_count": len(users),
		"total_count": total,
		"page":        page,
		"limit":       limit,
	}).Info("Successfully retrieved users")

	response := models.PaginatedResponse{
		Success: true,
		Data:    userResponses,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve user",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    user.ToResponse(),
	})
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if email already exists
	existingUser, _ := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Success: false,
			Error:   "Email already exists",
		})
		return
	}

	user, err := h.userRepo.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Success: true,
		Message: "User created successfully",
		Data:    user.ToResponse(),
	})
}

// UpdateUser handles PUT /api/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if email already exists (if email is being updated)
	if req.Email != nil {
		existingUser, _ := h.userRepo.GetByEmail(c.Request.Context(), *req.Email)
		if existingUser != nil && existingUser.ID != id {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Success: false,
				Error:   "Email already exists",
			})
			return
		}
	}

	user, err := h.userRepo.Update(c.Request.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user.ToResponse(),
	})
}

// DeleteUser handles DELETE /api/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
		return
	}

	err = h.userRepo.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success: false,
				Error:   "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
