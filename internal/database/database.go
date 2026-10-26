package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, errors.New("initialize database pool")
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

func Check(ctx context.Context, db *sql.DB) error {
	var version int
	if err := db.QueryRowContext(ctx, `SELECT version FROM foundation_metadata WHERE singleton = true`).Scan(&version); err != nil {
		return errors.New("database or foundation schema unavailable")
	}
	if version != 1 {
		return errors.New("unexpected foundation schema version")
	}
	return nil
}
