package service_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/gjtiquia/vimance/internal/db"
	"github.com/gjtiquia/vimance/internal/service"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

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

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "db", "migrations")
	if err := goose.Up(database, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return database
}

func setupTestService(t *testing.T) *service.Service {
	t.Helper()

	database := setupTestDB(t)
	return service.New(database)
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
		CreatedAt:    0,
		CreatedBy:    999,
		UpdatedAt:    0,
		UpdatedBy:    999,
	})
	if err == nil {
		t.Error("expected error when creating record with non-existent currency_id")
	}

	_, err = queries.CreateTag(t.Context(), db.CreateTagParams{
		Name:        "test",
		Description: "",
		Notes:       "",
		CreatedAt:   0,
		CreatedBy:   999,
		UpdatedAt:   0,
		UpdatedBy:   999,
	})
	if err == nil {
		t.Error("expected error when creating tag with non-existent created_by user")
	}
}

func TestCascadeDelete(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	record, _ := s.CreateRecordWithTags(t.Context(), "2026-01-15", 1000, currency.ID, "", user.ID, []int64{tag.ID})

	tags, _ := s.GetRecordTags(t.Context(), record.ID)
	if len(tags) != 1 {
		t.Errorf("expected 1 record tag, got %d", len(tags))
	}

	err := s.HardDeleteTag(t.Context(), tag.ID)
	if err != nil {
		t.Fatalf("failed to delete tag: %v", err)
	}

	tags, _ = s.GetRecordTags(t.Context(), record.ID)
	if len(tags) != 0 {
		t.Errorf("expected 0 record tags after cascade delete, got %d", len(tags))
	}
}

func TestRecordDateConstraint(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")

	validDates := []string{"2026-01-15", "2026-12-31", "2000-06-01"}
	for _, date := range validDates {
		_, err := s.CreateRecord(t.Context(), date, 1000, currency.ID, "", user.ID)
		if err != nil {
			t.Errorf("valid date '%s' should be accepted: %v", date, err)
		}
	}

	invalidDates := []string{"2026-13-01", "2026-01-32", "2026-00-15", "invalid"}
	for _, date := range invalidDates {
		_, err := s.CreateRecord(t.Context(), date, 1000, currency.ID, "", user.ID)
		if err == nil {
			t.Errorf("invalid date '%s' should be rejected", date)
		}
	}
}

func TestTransactionRollback(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")
	currency, _ := s.CreateCurrency(t.Context(), "USD")
	tag, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)

	recordsBefore, _ := s.ListRecords(t.Context())

	_, err := s.CreateRecordWithTags(t.Context(), "invalid-date", 1000, currency.ID, "", user.ID, []int64{tag.ID})
	if err == nil {
		t.Error("expected error when creating record with invalid date")
	}

	recordsAfter, _ := s.ListRecords(t.Context())
	if len(recordsAfter) != len(recordsBefore) {
		t.Error("transaction should have rolled back, no record should be created")
	}
}
