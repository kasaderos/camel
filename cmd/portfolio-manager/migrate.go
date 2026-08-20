package main

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"
	"gorm.io/gorm"
)

func migrateUp(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	db, err := do.Invoke[*gorm.DB](injector)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func migrateDrop(ctx context.Context, c *cli.Command) error {
	injector, err := provide()
	if err != nil {
		return err
	}
	defer terminate(injector)

	db, err := do.Invoke[*gorm.DB](injector)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
