package migrations

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var files embed.FS

// Up uses the migration library's PostgreSQL advisory lock and version tracking.
// A dedicated connection keeps advisory lock/unlock on the same DB session.
// Its session settings and any surviving advisory locks are discarded afterward.
func Up(ctx context.Context, db *sql.DB) error {
	source, err := iofs.New(files, ".")
	if err != nil {
		return err
	}
	defer source.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Discard this physical session instead of returning modified settings or
		// a surviving advisory lock to the application pool after an error.
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		_ = conn.Close()
	}()
	// Driver initialization acquires a lock before migrate.LockTimeout applies.
	// Server-side limits also cover the library's background-context queries.
	if _, err := conn.ExecContext(ctx, `SET lock_timeout = '2s'; SET statement_timeout = '10s'`); err != nil {
		return err
	}
	migrationDriver, err := postgres.WithConnection(ctx, conn, &postgres.Config{StatementTimeout: 10 * time.Second})
	if err != nil {
		return err
	}
	runner, err := migrate.NewWithInstance("iofs", source, "postgres", migrationDriver)
	if err != nil {
		return err
	}
	runner.LockTimeout = 5 * time.Second
	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
