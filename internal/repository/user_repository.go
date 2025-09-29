package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"example/otel/internal/database"
	"example/otel/internal/models"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UserRepository handles user data operations
type UserRepository struct {
	db     *database.DB
	tracer trace.Tracer
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{
		db:     db,
		tracer: otel.Tracer("user-repository"),
	}
}

// GetAll retrieves all users with pagination
func (r *UserRepository) GetAll(ctx context.Context, limit, offset int) ([]models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetAll")
	defer span.End()

	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)

	query := `
		SELECT id, name, email, bio, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`

	// Record query metrics
	start := time.Now()
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	duration := time.Since(start)

	// Record database metrics
	r.db.RecordQueryMetrics(ctx, "SELECT", "users", duration, err)

	if err != nil {
		span.SetAttributes(attribute.Bool("db.query.success", false))
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Bio,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			span.SetAttributes(attribute.Bool("db.query.success", false))
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		span.SetAttributes(attribute.Bool("db.query.success", false))
		return nil, fmt.Errorf("error iterating over users: %w", err)
	}

	// Add result count to span
	span.SetAttributes(
		attribute.Int("result.count", len(users)),
		attribute.Bool("db.query.success", true),
	)

	return users, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user.id", id),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)

	query := `
		SELECT id, name, email, bio, created_at, updated_at 
		FROM users 
		WHERE id = ?
	`

	// Record query metrics
	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, id)
	duration := time.Since(start)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Record database metrics
	r.db.RecordQueryMetrics(ctx, "SELECT", "users", duration, err)

	if err != nil {
		if err == sql.ErrNoRows {
			span.SetAttributes(
				attribute.Bool("user.found", false),
				attribute.Bool("db.query.success", true),
			)
			return nil, fmt.Errorf("user not found")
		}
		span.SetAttributes(attribute.Bool("db.query.success", false))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	span.SetAttributes(
		attribute.Bool("user.found", true),
		attribute.Bool("db.query.success", true),
	)
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.name", req.Name),
		attribute.String("user.email", req.Email),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "users"),
	)

	query := `
		INSERT INTO users (name, email, bio) 
		VALUES (?, ?, ?)
	`

	// Record query metrics
	start := time.Now()
	result, err := r.db.ExecContext(ctx, query, req.Name, req.Email, req.Bio)
	duration := time.Since(start)

	// Record database metrics
	r.db.RecordQueryMetrics(ctx, "INSERT", "users", duration, err)

	if err != nil {
		span.SetAttributes(attribute.Bool("db.query.success", false))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		span.SetAttributes(attribute.Bool("db.query.success", false))
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	span.SetAttributes(
		attribute.Int64("user.id", id),
		attribute.Bool("db.query.success", true),
	)
	return r.GetByID(ctx, int(id))
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, id int, req models.UpdateUserRequest) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Update")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user.id", id),
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.table", "users"),
	)

	// First check if user exists
	existingUser, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
		span.SetAttributes(attribute.String("user.name", *req.Name))
	}
	if req.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *req.Email)
		span.SetAttributes(attribute.String("user.email", *req.Email))
	}
	if req.Bio != nil {
		setParts = append(setParts, "bio = ?")
		args = append(args, *req.Bio)
		span.SetAttributes(attribute.String("user.bio", *req.Bio))
	}

	if len(setParts) == 0 {
		span.SetAttributes(attribute.Bool("user.no_changes", true))
		return existingUser, nil // No changes
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	// Rebuild query properly
	query := "UPDATE users SET "
	for i, part := range setParts {
		if i > 0 {
			query += ", "
		}
		query += part
	}
	query += " WHERE id = ?"

	_, err = r.db.ExecContext(ctx, query, args...)
	r.db.RecordQueryMetrics(ctx, query, err)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user.id", id),
		attribute.String("db.operation", "DELETE"),
		attribute.String("db.table", "users"),
	)

	// First check if user exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	query := "DELETE FROM users WHERE id = ?"
	start := time.Now()
	_, err = r.db.ExecContext(ctx, query, id)
	r.db.RecordQueryMetrics(ctx, query, start, err)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	span.SetAttributes(attribute.Bool("user.deleted", true))
	return nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Count")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)

	query := "SELECT COUNT(*) FROM users"

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	r.db.RecordQueryMetrics(ctx, query, err)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	span.SetAttributes(attribute.Int("result.count", count))
	return count, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.email", email),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)

	query := `
		SELECT id, name, email, bio, created_at, updated_at 
		FROM users 
		WHERE email = ?
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// Record database query metrics
	r.db.RecordQueryMetrics(ctx, query, err)
	if err != nil {
		if err == sql.ErrNoRows {
			span.SetAttributes(attribute.Bool("user.found", false))
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	span.SetAttributes(attribute.Bool("user.found", true))
	return &user, nil
}
