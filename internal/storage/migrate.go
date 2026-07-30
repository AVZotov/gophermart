package storage

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/AVZotov/gophermart"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	migrationsFS, err := fs.Sub(gophermart.EmbedMigrations, "migrations")
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationsFS)
	if err != nil {
		return err
	}
	defer provider.Close()

	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	return nil
}
