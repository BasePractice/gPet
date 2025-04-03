package main

import (
	"database/sql"
	"embed"
	"errors"
	"log"
	"time"

	"pet/services"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrations embed.FS

type DatabaseClass interface {
	Classes(nameFilter *string, status *string, version *uint32) ([]Class, error)
	Class(name string) (*Class, error)
	Elements(c Class, version *uint32, status *string) ([]Element, error)
}

type Class struct {
	Id        int64     `sql:"id"`
	Name      string    `sql:"name"`
	Title     string    `sql:"title"`
	TableName string    `sql:"table_name"`
	Current   uint32    `sql:"current"`
	Status    string    `sql:"status"`
	UpdatedAt time.Time `sql:"updated_at"`
}
type Element struct {
	Key     string `sql:"key"`
	Value   string `sql:"value"`
	Version uint32 `sql:"version"`
	Status  string `sql:"status"`
}

type ds struct {
	db *sql.DB
}

func (d *ds) Elements(c Class, version *uint32, status *string) ([]Element, error) {
	query := "SELECT key, value, version, status FROM class.$1 WHERE 1 = 1"
	args := make([]interface{}, 1)
	args[0] = c.TableName
	if status != nil {
		query += " AND status  = $2"
		args = append(args, *status)
	}
	if version != nil {
		query += " AND version = $3"
		args = append(args, *version)
	}
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var elements []Element
	for rows.Next() {
		var element Element
		err := rows.Scan(&element.Key, &element.Value, &element.Version, &element.Status)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	return elements, nil
}

func (d *ds) Class(name string) (*Class, error) {
	rows, err := d.db.Query("SELECT id, name, title, table_name, current, status, updated_at  FROM class.classes WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var class Class
	if rows.Next() {
		err := rows.Scan(&class.Id, &class.Name, &class.Title, &class.TableName, &class.Current,
			&class.Status, &class.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return &class, nil
	}
	return nil, nil
}

func (d *ds) Classes(nameFilter *string, status *string, version *uint32) ([]Class, error) {
	query := "SELECT id, name, title, table_name, current, status, updated_at FROM class.classes WHERE 1 = 1"
	args := make([]interface{}, 0)
	if nameFilter != nil {
		query += " AND name LIKE '%$1%'"
		args = append(args, *nameFilter)
	}
	if status != nil {
		query += " AND status  = $2"
		args = append(args, *status)
	}
	if version != nil {
		query += " AND version = $3"
		args = append(args, *version)
	}

	var rows *sql.Rows
	var err error
	if len(args) == 0 {
		rows, err = d.db.Query(query)
	} else {
		rows, err = d.db.Query(query, args)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Class, 0)
	for rows.Next() {
		var class Class
		err := rows.Scan(&class.Id, &class.Name, &class.Title, &class.TableName, &class.Current,
			&class.Status, &class.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, class)
	}
	return result, nil
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
