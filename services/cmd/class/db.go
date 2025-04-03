package main

import (
	"database/sql"
	"embed"
	"errors"
	"log"

	"pet/services"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrations embed.FS

type DatabaseClass interface {
}

type ds struct {
	db *sql.DB
}

func migrationScheme(db *sql.DB) {
	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatal(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
		return
	}
	instance, err := migrate.NewWithInstance("iofs", d, "tm", driver)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = instance.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
		return
	}
}

func NewDatabaseClass() DatabaseClass {
	db, err := sql.Open("postgres", services.PostgresUrl)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	migrationScheme(db)
	return &ds{db}
}
