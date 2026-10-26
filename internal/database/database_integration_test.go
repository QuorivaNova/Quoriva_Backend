package database_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	migratedb "github.com/golang-migrate/migrate/v4/database"

	"quoriva.local/api/internal/database"
	"quoriva.local/api/migrations"
)

// This test creates/drops only a uniquely named schema in the supplied test DB.
func TestPostgresMigrationAndReadiness(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("REQUIRE_DATABASE_TESTS") == "1" {
			t.Fatal("TEST_DATABASE_URL is required")
		}
		t.Skip("set TEST_DATABASE_URL to run real PostgreSQL integration checks")
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	schema := "qv_test_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if _, err := db.ExecContext(cleanup, `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Errorf("cleanup test schema: %v", err)
		}
	}()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	scoped, err := database.Open(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer scoped.Close()
	if err := database.Check(ctx, scoped); err == nil {
		t.Fatal("unmigrated database reported ready")
	}
	if err := migrations.Up(ctx, scoped); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Up(ctx, scoped); err != nil {
		t.Fatalf("migration rerun: %v", err)
	}
	if err := database.Check(ctx, scoped); err != nil {
		t.Fatalf("migrated database not ready: %v", err)
	}
	t.Run("migration lock wait is bounded", func(t *testing.T) {
		testMigrationLockTimeout(t, ctx, scoped, schema)
	})
	if _, err := scoped.ExecContext(ctx, `DELETE FROM foundation_metadata`); err != nil {
		t.Fatal(err)
	}
	if err := database.Check(ctx, scoped); err == nil {
		t.Fatal("missing foundation record reported ready")
	}
}

func testMigrationLockTimeout(t *testing.T, ctx context.Context, db *sql.DB, schema string) {
	t.Helper()
	blocker, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Close()
	var name string
	if err := blocker.QueryRowContext(ctx, `SELECT current_database()`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	key, err := migratedb.GenerateAdvisoryLockId(name, schema, "schema_migrations")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := blocker.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, key); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := blocker.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, key); err != nil {
			t.Errorf("release blocking test lock: %v", err)
		}
	}()
	result := make(chan error, 1)
	go func() { result <- migrations.Up(ctx, db) }()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("migration succeeded while another session held its lock")
		}
	case <-time.After(8 * time.Second):
		t.Fatal("migration initialization ignored bounded lock wait")
	}
}
