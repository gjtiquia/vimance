package main

import (
	"database/sql"
	"embed"
	"testing"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed db/migrations/*.sql
var testMigrations embed.FS

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	_, err = database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	goose.SetBaseFS(testMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := goose.Up(database, "db/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return database
}

func now() int64 {
	return time.Now().Unix()
}

func TestUserCRUD(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	queries := db.New(database)

	user, err := queries.CreateUser(t.Context(), db.CreateUserParams{
		Username:  "testuser",
		CreatedAt: now(),
		UpdatedAt: now(),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}

	fetched, err := queries.GetUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if fetched.Username != user.Username {
		t.Errorf("expected username '%s', got '%s'", user.Username, fetched.Username)
	}

	err = queries.SoftDeleteUser(t.Context(), db.SoftDeleteUserParams{
		DeletedAt: sql.NullInt64{Int64: now(), Valid: true},
		DeletedBy: sql.NullInt64{Int64: user.ID, Valid: true},
		ID:        user.ID,
	})
	if err != nil {
		t.Fatalf("failed to soft delete user: %v", err)
	}

	activeUsers, err := queries.ListActiveUsers(t.Context())
	if err != nil {
		t.Fatalf("failed to list active users: %v", err)
	}
	for _, u := range activeUsers {
		if u.ID == user.ID {
			t.Error("soft deleted user should not appear in active_users view")
		}
	}

	_, err = queries.RestoreUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("failed to restore user: %v", err)
	}

	activeUsers, err = queries.ListActiveUsers(t.Context())
	if err != nil {
		t.Fatalf("failed to list active users: %v", err)
	}
	found := false
	for _, u := range activeUsers {
		if u.ID == user.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("restored user should appear in active_users view")
	}
}

func TestRecordDateConstraint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	queries := db.New(database)

	user, _ := queries.CreateUser(t.Context(), db.CreateUserParams{
		Username:  "testuser",
		CreatedAt: now(),
		UpdatedAt: now(),
	})

	currency, _ := queries.CreateCurrency(t.Context(), db.CreateCurrencyParams{
		Code:      "USD",
		CreatedAt: now(),
		UpdatedAt: now(),
	})

	validDates := []string{"2026-01-15", "2026-12-31", "2000-06-01"}
	for _, date := range validDates {
		_, err := queries.CreateRecord(t.Context(), db.CreateRecordParams{
			Date:         date,
			AmountCents:  1000,
			CurrencyID:   currency.ID,
			Notes:        "test",
			CreatedAt:    now(),
			CreatedBy:    user.ID,
			UpdatedAt:    now(),
			UpdatedBy:    user.ID,
		})
		if err != nil {
			t.Errorf("valid date '%s' should be accepted: %v", date, err)
		}
	}

	invalidDates := []string{"2026-13-01", "2026-01-32", "2026-00-15", "invalid"}
	for _, date := range invalidDates {
		_, err := queries.CreateRecord(t.Context(), db.CreateRecordParams{
			Date:         date,
			AmountCents:  1000,
			CurrencyID:   currency.ID,
			Notes:        "test",
			CreatedAt:    now(),
			CreatedBy:    user.ID,
			UpdatedAt:    now(),
			UpdatedBy:    user.ID,
		})
		if err == nil {
			t.Errorf("invalid date '%s' should be rejected", date)
		}
	}
}

func TestForeignKeyConstraints(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	queries := db.New(database)

	_, err := queries.CreateRecord(t.Context(), db.CreateRecordParams{
		Date:         "2026-01-15",
		AmountCents:  1000,
		CurrencyID:   999,
		Notes:        "test",
		CreatedAt:    now(),
		CreatedBy:    999,
		UpdatedAt:    now(),
		UpdatedBy:    999,
	})
	if err == nil {
		t.Error("expected error when creating record with non-existent currency_id")
	}

	_, err = queries.CreateTag(t.Context(), db.CreateTagParams{
		Name:        "test",
		Description: "",
		Notes:       "",
		CreatedAt:   now(),
		CreatedBy:   999,
		UpdatedAt:   now(),
		UpdatedBy:   999,
	})
	if err == nil {
		t.Error("expected error when creating tag with non-existent created_by user")
	}
}

func TestCascadeDelete(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	queries := db.New(database)

	user, _ := queries.CreateUser(t.Context(), db.CreateUserParams{
		Username:  "testuser",
		CreatedAt: now(),
		UpdatedAt: now(),
	})

	currency, _ := queries.CreateCurrency(t.Context(), db.CreateCurrencyParams{
		Code:      "USD",
		CreatedAt: now(),
		UpdatedAt: now(),
	})

	record, _ := queries.CreateRecord(t.Context(), db.CreateRecordParams{
		Date:         "2026-01-15",
		AmountCents:  1000,
		CurrencyID:   currency.ID,
		Notes:        "test",
		CreatedAt:    now(),
		CreatedBy:    user.ID,
		UpdatedAt:    now(),
		UpdatedBy:    user.ID,
	})

	tag, _ := queries.CreateTag(t.Context(), db.CreateTagParams{
		Name:        "food",
		Description: "food expenses",
		Notes:       "",
		CreatedAt:   now(),
		CreatedBy:   user.ID,
		UpdatedAt:   now(),
		UpdatedBy:   user.ID,
	})

	err := queries.AddRecordTag(t.Context(), db.AddRecordTagParams{
		RecordID:  record.ID,
		TagID:     tag.ID,
		CreatedAt: now(),
		CreatedBy: user.ID,
		UpdatedAt: now(),
		UpdatedBy: user.ID,
	})
	if err != nil {
		t.Fatalf("failed to add record tag: %v", err)
	}

	recordTags, _ := queries.GetRecordTags(t.Context(), record.ID)
	if len(recordTags) != 1 {
		t.Errorf("expected 1 record tag, got %d", len(recordTags))
	}

	err = queries.HardDeleteTag(t.Context(), tag.ID)
	if err != nil {
		t.Fatalf("failed to delete tag: %v", err)
	}

	recordTags, _ = queries.GetRecordTags(t.Context(), record.ID)
	if len(recordTags) != 0 {
		t.Errorf("expected 0 record tags after cascade delete, got %d", len(recordTags))
	}
}
