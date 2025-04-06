package services

import (
	"database/sql"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"log/slog"
)

func MigrationScheme(db *sql.DB, migrations embed.FS) {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		slog.Error("Can't open migration resource", slog.String("err", err.Error()))
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Can't create postgres instance", slog.String("err", err.Error()))
		return
	}
	instance, err := migrate.NewWithInstance("iofs", d, "pet", driver)
	if err != nil {
		slog.Error("Can't create migration", slog.String("err", err.Error()))
		return
	}
	err = instance.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("Can't up migration", slog.String("err", err.Error()))
		return
	}
}
