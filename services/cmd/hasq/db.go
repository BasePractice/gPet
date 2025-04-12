package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
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
    id         BIGINT    NOT NULL PRIMARY KEY,
    key        VARCHAR   NOT NULL REFERENCES keys(hash),
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
	CreateToken(title string, data []byte) (*Token, error)
	SearchToken(id *uuid.UUID, hash *string) (*Token, error)
	CreateKey(user uuid.UUID, token uuid.UUID, passphrase string) (*Key, error)
	LoadChain(token *Token) (services.Chain, error)
	Owner(user uuid.UUID, token uuid.UUID) error
	Validate(token uuid.UUID) (*ValidateResult, error)
}

type Token struct {
	Id        uuid.UUID `sql:"id"`
	Title     string    `sql:"title"`
	Hash      string    `sql:"hash"`
	Data      []byte    `sql:"data"`
	UpdatedAt time.Time `sql:"updated_at"`
}

type Key struct {
	Id     uuid.UUID `sql:"id"`
	Hash   string    `sql:"hash"`
	Num    uint64    `sql:"num"`
	UserId uuid.UUID `sql:"user_id"`
	Token  string    `sql:"token"`
}

type ValidateResult struct {
	Successful bool
	OwnerId    uuid.UUID
	LastNum    uint64
}

func (t Token) String() string {
	return t.Id.String() + ":" + t.Hash + ":\"" + t.Title + "\""
}

type ds struct {
	db *sql.DB
}

func (d *ds) Validate(token uuid.UUID) (*ValidateResult, error) {
	t, err := d.SearchToken(&token, nil)
	if err != nil {
		return nil, errors.New("token not found")
	}
	c, err := d.LoadChain(t)
	if err != nil {
		return nil, err
	}
	var userId uuid.UUID
	var ln uint64
	var v = c.Validate()
	if v {
		lastNum, k := c.GetOwner()
		if k != nil {
			err = d.db.QueryRow("SELECT user_id FROM keys WHERE hash = $1", k.String()).Scan(&userId)
			if err != nil {
				return nil, err
			}
			ln = lastNum
		}
	}
	return &ValidateResult{
		Successful: v,
		OwnerId:    userId,
		LastNum:    ln,
	}, nil
}

func (d *ds) Owner(user uuid.UUID, token uuid.UUID) error {
	t, err := d.SearchToken(&token, nil)
	if err != nil {
		return errors.New("token not found")
	}
	c, err := d.LoadChain(t)
	if err != nil {
		return err
	}
	k, err := d.loadLastKey(user, token)
	if err != nil {
		return err
	}
	lk := services.LoadKey(k.Hash)
	lastNum, key := c.GetOwner()
	if lastNum > 0 {
		if k.Hash == key.String() {
			return errors.New("token owned by this user")
		} else if k.Num != lastNum+1 {
			return errors.New("last user key does not match")
		}
	}
	owned := c.Owned(lk)
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	tb := tableName(token)
	_, err = tx.Exec("INSERT INTO "+tb+"(id, key) VALUES ($1, $2)", k.Num, k.Hash)
	if err != nil {
		return err
	}
	if owned.Gen != nil {
		_, err = tx.Exec("UPDATE "+tb+" SET generator = $1 WHERE id = $2", *owned.Gen, owned.GenId)
		if err != nil {
			return err
		}
	}
	if owned.Own != nil {
		_, err = tx.Exec("UPDATE "+tb+" SET owner = $1 WHERE id = $2", *owned.Own, owned.OwnId)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (d *ds) loadLastKey(user uuid.UUID, token uuid.UUID) (*Key, error) {
	var id uuid.UUID
	var hash string
	var num uint64

	slog.Debug("Last key searching", slog.String("token", token.String()), slog.String("user", user.String()))
	err := d.db.QueryRow("SELECT id, hash, num FROM keys WHERE user_id = $1 AND token_id = $2 ORDER BY num DESC LIMIT 1", user, token).
		Scan(&id, &hash, &num)
	if err != nil {
		return nil, err
	}
	return &Key{
		Id:     id,
		Hash:   hash,
		Num:    num,
		UserId: user,
	}, nil
}

func (d *ds) LoadChain(token *Token) (services.Chain, error) {
	var c uint64
	tb := tableName(token.Id)
	query := fmt.Sprintf("SELECT COUNT(id) FROM %s", tb)
	err := d.db.QueryRow(query).Scan(&c)
	if err != nil {
		return nil, err
	}
	query = fmt.Sprintf("SELECT id, key, generator, owner FROM %s ORDER BY id LIMIT 4", tb)
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	var ch = services.CreateEmptyChain(token.Hash, c)
	for rows.Next() {
		var id uint64
		var key string
		var generator *string
		var owner *string
		if err = rows.Scan(&id, &key, &generator, &owner); err != nil {
			break
		}
		_ = ch.Push(id, key, generator, owner)
	}
	val := ch.Validate()
	if val {
		return ch, nil
	}
	return nil, errors.New("chain damaged")
}

func (d *ds) CreateKey(user uuid.UUID, token uuid.UUID, passphrase string) (*Key, error) {
	t, err := d.SearchToken(&token, nil)
	if err != nil {
		return nil, errors.New("token not found")
	}
	c, err := d.LoadChain(t)
	if err != nil {
		return nil, err
	}
	n, k := c.GetOwner()
	if k == nil {
		n = 1
	} else {
		n = n + 1
	}
	n, k = c.KeyOn(n, passphrase)
	rows := d.db.QueryRow("INSERT INTO keys(hash, num, token_id, user_id) VALUES($1, $2, $3, $4) RETURNING id",
		k.String(), n, token, user)
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	var keyId uuid.UUID
	if err = rows.Scan(&keyId); err != nil {
		return nil, err
	}
	slog.Debug("Key created",
		slog.String("key_id", keyId.String()),
		slog.String("key_hash", k.String()),
		slog.String("token_id", token.String()),
		slog.String("user_id", user.String()))
	return &Key{
		Id:     keyId,
		Hash:   k.String(),
		Num:    n,
		UserId: user,
	}, nil
}

func textOrUndefined(id interface{}) string {
	switch id.(type) {
	case *string:
		s := id.(*string)
		if s != nil {
			return *s
		}
	case *uuid.UUID:
		u := id.(*uuid.UUID)
		if u != nil {
			return (*u).String()
		}
	}
	return "undefined"
}

func (d *ds) SearchToken(id *uuid.UUID, hash *string) (*Token, error) {
	var rows *sql.Row
	if hash != nil {
		rows = d.db.QueryRow("SELECT id, title, hash, data FROM tokens WHERE hash = $1", *hash)
	} else if id != nil {
		rows = d.db.QueryRow("SELECT id, title, hash, data FROM tokens WHERE id = $1", id.String())
	} else {
		return nil, errors.New("no token found")
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	var token Token
	if err := rows.Scan(&token.Id, &token.Title, &token.Hash, &token.Data); err != nil {
		slog.Debug("Token not found",
			slog.String("search_id", textOrUndefined(id)),
			slog.String("search_hash", textOrUndefined(hash)))
		return nil, err
	}
	slog.Debug("Token searched", slog.String("token", token.String()))
	return &token, nil
}

func (d *ds) CreateToken(title string, data []byte) (*Token, error) {
	token := services.CreateToken(data)
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	row := tx.QueryRow(
		"INSERT INTO tokens(title, hash, data) VALUES ($1, $2, $3) RETURNING id", title, token.String(), data)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var tokenId uuid.UUID
	if err = row.Scan(&tokenId); err != nil {
		return nil, err
	}
	tb := tableName(tokenId)
	_, err = tx.Exec(fmt.Sprintf(sqlCreateTokenTable, tb))
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	slog.Debug("Token created",
		slog.String("token_id", tokenId.String()), slog.String("hash", token.String()))
	return &Token{
		Id:    tokenId,
		Title: title,
		Hash:  token.String(),
		Data:  data,
	}, err
}

func tableName(tokenId uuid.UUID) string {
	tb := "token_" + strings.Replace(tokenId.String(), "-", "", -1)
	tb = strings.ToLower(tb)
	return tb
}

func NewDatabaseToken() DatabaseToken {
	db, err := services.NewDatabase(migrations)
	if err != nil {
		return nil
	}
	return &ds{db}
}
