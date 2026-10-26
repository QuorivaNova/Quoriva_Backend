package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"quoriva.local/api/internal/config"
	"quoriva.local/api/internal/database"
	"quoriva.local/api/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if !run(logger) {
		os.Exit(1)
	}
}

func run(logger *slog.Logger) bool {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		return false
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("database pool initialization failed")
		return false
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		logger.Error("database unavailable for migration")
		return false
	}
	if err := migrations.Up(ctx, db); err != nil {
		// Avoid publishing driver errors that can contain connection information.
		logger.Error("migration failed; inspect database/schema state before retrying")
		return false
	}
	logger.Info("migrations applied or already current")
	return true
}
