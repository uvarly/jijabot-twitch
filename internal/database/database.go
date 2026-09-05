package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Open(ctx context.Context, path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_texttotime=1", path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxIdleConns(1)
	db.SetMaxIdleConns(1)

	if err := Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	sub, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to build migration filesystem: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, sub)

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
