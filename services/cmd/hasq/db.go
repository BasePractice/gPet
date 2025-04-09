package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"pet/services"

	_ "github.com/lib/pq"
)

const (
	sqlCreateTokenTable = `
CREATE TABLE %s
(
    id         SERIAL    NOT NULL PRIMARY KEY,
    key        VARCHAR   NOT NULL,
    generator  VARCHAR   DEFAULT NULL,
    owner      VARCHAR   DEFAULT NULL,
    updated_at TIMESTAMP NOT NULL                                                    DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL                                                    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (key, generator, owner)
)`
)

//go:embed migrations/*.sql
var migrations embed.FS

type DatabaseToken interface {
	CreateToken(title string, data []byte) (error, *Token)
	SearchToken(id *uuid.UUID, hash *string) (error, *Token)
}

type Token struct {
	Id        uuid.UUID `sql:"id"`
	Title     string    `sql:"title"`
	Hash      string    `sql:"hash"`
	Data      []byte    `sql:"data"`
	UpdatedAt time.Time `sql:"updated_at"`
}

func (t Token) String() string {
	return fmt.Sprintf("Token{id: %s, hash: %s, title: %s}", t.Id, t.Hash, t.Title)
}

type ds struct {
	db *sql.DB
}

func (d *ds) SearchToken(id *uuid.UUID, hash *string) (error, *Token) {
	var rows *sql.Row
	if hash != nil {
		rows = d.db.QueryRow("SELECT id, title, hash, data FROM hasq.token WHERE hash = $1", *hash)
	} else if id != nil {
		rows = d.db.QueryRow("SELECT id, title, hash, data FROM hasq.token WHERE id = $1", *id)
	} else {
		return errors.New("no token found"), nil
	}
	if rows.Err() != nil {
		return rows.Err(), nil
	}
	var token Token
	if err := rows.Scan(&token.Id, &token.Title, &token.Hash, &token.Data); err != nil {
		return err, nil
	}
	return nil, &token
}

func (d *ds) CreateToken(title string, data []byte) (error, *Token) {
	token := services.CreateToken(data)
	tx, err := d.db.Begin()
	if err != nil {
		return err, nil
	}
	row := tx.QueryRow(
		"INSERT INTO hasq.token(title, hash, data) VALUES ($1, $2, $3) RETURNING id", title, token.String(), data)
	if row.Err() != nil {
		return row.Err(), nil
	}
	var tokenId uuid.UUID
	if err := row.Scan(&tokenId); err != nil {
		return err, nil
	}
	tableName := "token_" + strings.Replace(tokenId.String(), "-", "", -1)
	tableName = strings.ToLower(tableName)
	_, err = tx.Exec(fmt.Sprintf(sqlCreateTokenTable, tableName))
	if err != nil {
		return err, nil
	}
	err = tx.Commit()
	if err != nil {
		return err, nil
	}
	return nil, &Token{
		Id:    tokenId,
		Title: title,
		Hash:  token.String(),
		Data:  data,
	}
}

func NewDatabaseToken() DatabaseToken {
	db, err := services.NewDatabase(migrations)
	if err != nil {
		return nil
	}
	return &ds{db}
}
