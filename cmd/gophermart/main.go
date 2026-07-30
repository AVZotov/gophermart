package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AVZotov/gophermart/internal/config"
	"github.com/AVZotov/gophermart/internal/logger"
	"github.com/AVZotov/gophermart/internal/storage"
)

func main() {
	appLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if syncErr := appLogger.Sync(); syncErr != nil {
			log.Printf("logger sync: %v", syncErr)
		}
	}()

	if err := run(appLogger); err != nil {
		appLogger.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func run(l logger.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctxConnect, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	store, err := storage.New(ctxConnect, cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer store.Close()

	ctxMigrate, cancelMigrate := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelMigrate()
	if err := storage.RunMigrations(ctxMigrate, cfg.DatabaseURI); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	srv := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	l.Info("starting server", "address", cfg.RunAddress)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
