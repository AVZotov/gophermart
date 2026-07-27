package main

import (
	"fmt"
	"log"

	"github.com/AVZotov/gophermart/internal/config"
	"github.com/AVZotov/gophermart/internal/logger"
)

func main() {
	appLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer appLogger.Sync()

	if err := run(appLogger); err != nil {
		appLogger.Error("fatal error", "err", err)
	}
}

func run(l logger.Logger) error {
	//configs
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return nil
}
