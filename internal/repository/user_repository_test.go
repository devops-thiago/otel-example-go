package repository

import (
    "context"
    "database/sql"
    "fmt"
    "regexp"
    "testing"
    "time"

    "example/otel/internal/database"
    "example/otel/internal/models"

    sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newTestDB(t *testing.T) (*database.DB, sqlmock.Sqlmock, func()) {
    t.Helper()
    sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
    if err != nil { t.Fatalf("sqlmock new: %v", err) }
    db := &database.DB{DB: sqlDB}
    cleanup := func() { _ = sqlDB.Close() }
    return db, mock, cleanup
}

func TestGetByID_NotFound(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE id = ?`)).WithArgs(99).WillReturnRows(sqlmock.NewRows([]string{"id","name","email","bio","created_at","updated_at"}))

    u, err := repo.GetByID(context.Background(), 99)
    if err == nil || u != nil { t.Fatalf("expected not found, got %v, %v", u, err) }
}

func TestCreate_Success(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (name, email, bio) 
        VALUES (?, ?, ?)`)).WithArgs("Alice","alice@example.com","bio").WillReturnResult(sqlmock.NewResult(1,1))

    now := time.Now()
    rows := sqlmock.NewRows([]string{"id","name","email","bio","created_at","updated_at"}).AddRow(1,"Alice","alice@example.com","bio", now, now)
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE id = ?`)).WithArgs(1).WillReturnRows(rows)

    u, err := repo.Create(context.Background(), models.CreateUserRequest{Name:"Alice", Email:"alice@example.com", Bio:"bio"})
    if err != nil { t.Fatalf("create err: %v", err) }
    if u == nil || u.ID != 1 { t.Fatalf("unexpected user: %+v", u) }
}

func TestGetAll_Pagination(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    now := time.Now()
    rows := sqlmock.NewRows([]string{"id","name","email","bio","created_at","updated_at"}).
        AddRow(1,"A","a@x","", now, now).
        AddRow(2,"B","b@x","", now, now)
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        ORDER BY created_at DESC 
        LIMIT ? OFFSET ?`)).WithArgs(2,0).WillReturnRows(rows)

    users, err := repo.GetAll(context.Background(), 2, 0)
    if err != nil || len(users) != 2 { t.Fatalf("unexpected: %v %d", err, len(users)) }
}

func TestCount_Success(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
    c, err := repo.Count(context.Background())
    if err != nil || c != 5 { t.Fatalf("unexpected: %v %d", err, c) }
}

func TestDelete_Success(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    now := time.Now()
    sel := sqlmock.NewRows([]string{"id","name","email","bio","created_at","updated_at"}).AddRow(3,"C","c@x","", now, now)
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE id = ?`)).WithArgs(3).WillReturnRows(sel)
    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = ?`)).WithArgs(3).WillReturnResult(sqlmock.NewResult(0,1))

    if err := repo.Delete(context.Background(), 3); err != nil { t.Fatalf("delete err: %v", err) }
}

func TestUpdate_SetsFields(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    now := time.Now()
    // initial select existing user
    sel := sqlmock.NewRows([]string{"id","name","email","bio","created_at","updated_at"}).AddRow(5,"Old","old@x","bio", now, now)
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE id = ?`)).WithArgs(5).WillReturnRows(sel)

    // expect update
    mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET name = ?, email = ?, updated_at = NOW() WHERE id = ?`)).
        WithArgs("New","new@x",5).WillReturnResult(sqlmock.NewResult(0,1))

    // select after update
    sel2 := sqlmock.NewRows([]string{"id","name","email","bio","created_at","updated_at"}).AddRow(5,"New","new@x","bio", now, now)
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE id = ?`)).WithArgs(5).WillReturnRows(sel2)

    newName := "New"
    newEmail := "new@x"
    u, err := repo.Update(context.Background(), 5, models.UpdateUserRequest{Name: &newName, Email: &newEmail})
    if err != nil { t.Fatalf("update err: %v", err) }
    if u == nil || u.Name != "New" { t.Fatalf("unexpected user: %+v", u) }
}

func TestGetByEmail_Found(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    now := time.Now()
    rows := sqlmock.NewRows([]string{"id", "name", "email", "bio", "created_at", "updated_at"}).
        AddRow(1, "John Doe", "john@example.com", "Bio", now, now)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE email = ?`)).
        WithArgs("john@example.com").
        WillReturnRows(rows)

    user, err := repo.GetByEmail(context.Background(), "john@example.com")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if user == nil {
        t.Fatal("expected user, got nil")
    }
    if user.Email != "john@example.com" {
        t.Errorf("expected email john@example.com, got: %s", user.Email)
    }
}

func TestGetByEmail_NotFound(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE email = ?`)).
        WithArgs("notfound@example.com").
        WillReturnError(sql.ErrNoRows)

    user, err := repo.GetByEmail(context.Background(), "notfound@example.com")
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if user != nil {
        t.Errorf("expected nil user, got: %+v", user)
    }
}

func TestGetAll_DatabaseError(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        LIMIT ? OFFSET ?`)).
        WithArgs(10, 0).
        WillReturnError(fmt.Errorf("database error"))

    users, err := repo.GetAll(context.Background(), 10, 0)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if users != nil {
        t.Errorf("expected nil users, got: %+v", users)
    }
}

func TestGetByID_DatabaseError(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, bio, created_at, updated_at 
        FROM users 
        WHERE id = ?`)).
        WithArgs(1).
        WillReturnError(fmt.Errorf("database error"))

    user, err := repo.GetByID(context.Background(), 1)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if user != nil {
        t.Errorf("expected nil user, got: %+v", user)
    }
}

func TestCreate_DatabaseError(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (name, email, bio, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())`)).
        WithArgs("John", "john@example.com", "Bio").
        WillReturnError(fmt.Errorf("database error"))

    user, err := repo.Create(context.Background(), models.CreateUserRequest{
        Name:  "John",
        Email: "john@example.com",
        Bio:   "Bio",
    })
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if user != nil {
        t.Errorf("expected nil user, got: %+v", user)
    }
}

func TestDelete_DatabaseError(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = ?`)).
        WithArgs(1).
        WillReturnError(fmt.Errorf("database error"))

    err := repo.Delete(context.Background(), 1)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}

func TestCount_DatabaseError(t *testing.T) {
    db, mock, cleanup := newTestDB(t)
    defer cleanup()
    repo := NewUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).
        WillReturnError(fmt.Errorf("database error"))

    count, err := repo.Count(context.Background())
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if count != 0 {
        t.Errorf("expected 0 count, got: %d", count)
    }
}


