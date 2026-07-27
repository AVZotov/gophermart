package main

import (
	"context"
	"fmt"
	"log"
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
	defer appLogger.Sync()

	if err := run(appLogger); err != nil {
		appLogger.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func run(l logger.Logger) error {
	//configs
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	//db
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = storage.New(ctx, cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	//migrations
	if err := storage.RunMigrations(ctx, cfg.DatabaseURI); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
