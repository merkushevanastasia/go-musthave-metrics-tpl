package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed *.sql
var fs embed.FS

func RunMigrations(pool *pgxpool.Pool) error {
	sourceDriver, err := iofs.New(fs, ".")
	defer sourceDriver.Close()

	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	// Получаем стандартный *sql.DB из существующего pgxpool.Pool
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	dbDriver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create pgx driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "pgx", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
