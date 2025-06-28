package database

import (
	"database/sql"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

func InitDB(p string) (*Queries, error) {
	migDrv, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, err
	}
	dbOpen, err := sql.Open("mysql", p)
	if err != nil {
		return nil, err
	}
	dbDrv, err := mysql.WithInstance(dbOpen, &mysql.Config{})
	if err != nil {
		return nil, err
	}
	mig, err := migrate.NewWithInstance("iofs", migDrv, "mysql", dbDrv)
	if err != nil {
		return nil, err
	}
	err = mig.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}
	return New(dbOpen), nil
}
