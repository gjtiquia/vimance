package tui_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/gjtiquia/vimance/internal/service"
	"github.com/gjtiquia/vimance/internal/tui"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	_, err = database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "db", "migrations")
	if err := goose.Up(database, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return database
}

func seedTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec("INSERT INTO users (id, name) VALUES (1, 'testuser')")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, err = db.Exec("INSERT INTO currencies (id, code) VALUES (1, 'USD')")
	if err != nil {
		t.Fatalf("seed currency: %v", err)
	}
	_, err = db.Exec("INSERT INTO tags (id, name) VALUES (1, 'food')")
	if err != nil {
		t.Fatalf("seed tag: %v", err)
	}
}

func newTestModel(t *testing.T) tui.Model {
	t.Helper()
	return tui.NewModel(setupTestDB(t))
}

func newTestService(t *testing.T) *service.Service {
	t.Helper()
	db := setupTestDB(t)
	seedTestDB(t, db)
	return service.New(db)
}
