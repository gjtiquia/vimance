# Plan 001: Database Setup

## Overview

Set up SQLite database with migrations (goose) and type-safe Go code generation (sqlc).

## Directory Structure

```
db/
├── migrations/     # goose migration files
└── queries/        # sqlc query files
internal/
└── db/             # generated go code from sqlc
```

## Schema

### Tables

#### users
- id (INTEGER PRIMARY KEY)
- username (TEXT UNIQUE, NOT NULL)
- created_at (INTEGER, Unix timestamp)
- updated_at (INTEGER, Unix timestamp)
- deleted_at (INTEGER, nullable, Unix timestamp)
- deleted_by (INTEGER, nullable, FK → users.id)

#### currencies
- id (INTEGER PRIMARY KEY)
- code (TEXT UNIQUE, NOT NULL) -- e.g., USD, EUR, CAD
- created_at (INTEGER, Unix timestamp)
- updated_at (INTEGER, Unix timestamp)

#### tags
- id (INTEGER PRIMARY KEY)
- name (TEXT UNIQUE, NOT NULL)
- description (TEXT NOT NULL, can be empty string)
- notes (TEXT NOT NULL, can be empty string)
- created_at (INTEGER, Unix timestamp)
- created_by (INTEGER NOT NULL, FK → users.id)
- updated_at (INTEGER, Unix timestamp)
- updated_by (INTEGER NOT NULL, FK → users.id)
- deleted_at (INTEGER, nullable, Unix timestamp)
- deleted_by (INTEGER, nullable, FK → users.id)

#### records
- id (INTEGER PRIMARY KEY)
- date (TEXT NOT NULL) -- ISO 8601 format (YYYY-MM-DD), with constraint
- amount_cents (INTEGER NOT NULL) -- store in cents to avoid floating point issues
- currency_id (INTEGER NOT NULL, FK → currencies.id)
- notes (TEXT NOT NULL, can be empty string)
- created_at (INTEGER, Unix timestamp)
- created_by (INTEGER NOT NULL, FK → users.id)
- updated_at (INTEGER, Unix timestamp)
- updated_by (INTEGER NOT NULL, FK → users.id)
- deleted_at (INTEGER, nullable, Unix timestamp)
- deleted_by (INTEGER, nullable, FK → users.id)

#### records_tags
- record_id (INTEGER NOT NULL, FK → records.id ON DELETE CASCADE)
- tag_id (INTEGER NOT NULL, FK → tags.id ON DELETE CASCADE)
- created_at (INTEGER, Unix timestamp)
- created_by (INTEGER NOT NULL, FK → users.id)
- updated_at (INTEGER, Unix timestamp)
- updated_by (INTEGER NOT NULL, FK → users.id)
- PRIMARY KEY (record_id, tag_id)

#### pinned_tabs
- tab_id (INTEGER PRIMARY KEY) -- placeholder for future TUI tab tracking
- position (INTEGER NOT NULL) -- ordering of pinned tabs
- created_at (INTEGER, Unix timestamp)
- created_by (INTEGER NOT NULL, FK → users.id)
- updated_at (INTEGER, Unix timestamp)
- updated_by (INTEGER NOT NULL, FK → users.id)

### Views

#### active_users
Filters users where `deleted_at IS NULL`.

#### active_tags
Filters tags where `deleted_at IS NULL`.

#### active_records
Filters records where `deleted_at IS NULL`.

## Configuration

### PRAGMA
- `PRAGMA foreign_keys = ON;` in migration (SQLite doesn't enable FK constraints by default)

### Date Constraint
Add CHECK constraint for ISO 8601 date format on records.date:
```sql
CHECK (date LIKE '____-__-__' AND 
       substr(date, 1, 4) BETWEEN '0000' AND '9999' AND
       substr(date, 6, 2) BETWEEN '01' AND '12' AND
       substr(date, 9, 2) BETWEEN '01' AND '31')
```

## Tooling

### goose
Migration tool for managing schema versions.

Install:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

CLI Usage (for development):
```bash
# Create new migration
goose -dir db/migrations sqlite3 vimance.db create <name> sql

# Apply migrations
goose -dir db/migrations sqlite3 vimance.db up

# Rollback
goose -dir db/migrations sqlite3 vimance.db down
```

#### Auto-migrate on App Startup
Embed migrations and run `goose.Up()` on startup:

```go
import (
    "database/sql"
    "embed"
    
    "github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func runMigrations(db *sql.DB) error {
    goose.SetBaseFS(migrations)
    if err := goose.SetDialect("sqlite3"); err != nil {
        return err
    }
    return goose.Up(db, "db/migrations")
}
```

Call this in `main.go` after opening the database connection.

### sqlc
Generates type-safe Go code from SQL queries.

Install:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Usage:
```bash
# Generate code
sqlc generate
```

## sqlc.yaml Configuration

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "db/queries"
    schema: "db/migrations"
    gen:
      go:
        package: "db"
        out: "internal/db"
```

## Implementation Steps

1. [x] Install goose and sqlc
2. [x] Create directory structure (`db/migrations`, `db/queries`, `internal/db`)
3. [x] Create initial migration file with full schema
4. [x] Create `sqlc.yaml` config
5. [x] Create query files in `db/queries/` for CRUD operations
6. [x] Run sqlc to generate Go code
7. [x] Add auto-migrate logic to `main.go`

## Decisions Made

- **Single user**: App starts with single-user model, auth can be added later
- **Unix timestamps**: Store as INTEGER (Unix epoch), easier for Go
- **Currencies**: Start empty, user adds what they need
- **Soft deletes**: Keep deleted records for recovery, use views for active records
- **Default currency**: No user preference stored; frontend provides default selection
- **pinned_tabs**: Include in initial schema as stub for future TUI functionality

## Files to Create

```
db/
├── migrations/
│   └── 20260423000000_initial_schema.sql
└── queries/
    ├── users.sql
    ├── currencies.sql
    ├── tags.sql
    ├── records.sql
    └── records_tags.sql
sqlc.yaml
main.go (update - add auto-migrate on startup)
```

## Notes

- Generated code in `internal/db` will be committed to the repo (no .gitignore needed)
- No Makefile for now - keep it minimal, add later if needed
