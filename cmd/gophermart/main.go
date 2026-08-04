package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AVZotov/gophermart/internal/config"
	"github.com/AVZotov/gophermart/internal/handler"
	"github.com/AVZotov/gophermart/internal/logger"
	"github.com/AVZotov/gophermart/internal/service"
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

	secret := []byte(cfg.JWTSecret)
	svc := service.NewService(store, secret)
	h := handler.NewHandler(svc, l)
	router := handler.NewRouter(h, l, secret)

	srv := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		l.Info("starting server", "address", cfg.RunAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("run server: %w", err)
			return
		}
		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		l.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	l.Info("server stopped gracefully")
	return nil
}
