package db

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type DataBaseSqlite struct {
	DB *sql.DB
}

func NewDataBaseSqlite(path string) (*DataBaseSqlite, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DataBaseSqlite{
		DB: db,
	}, nil
}

func (d *DataBaseSqlite) Migrate() error {
	driver, err := sqlite3.WithInstance(d.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/infrastructure/db/migrations",
		//"ql",
		"sqlite3",
		driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
