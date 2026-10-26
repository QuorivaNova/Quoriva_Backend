package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quoriva.local/api/internal/config"
	"quoriva.local/api/internal/database"
	"quoriva.local/api/internal/handlers"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: handlers.NewHandler(func(ctx context.Context) error {
			return database.Check(ctx, db)
		}, logger, time.Second),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	// ListenAndServe blocks; keep run free to wait for a shutdown signal too.
	go func() {
		serveErr <- server.ListenAndServe()
	}()
	logger.Info("API listening", "address", cfg.HTTPAddr)
	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return errors.New("graceful shutdown timed out")
		}
		return nil
	}
}
