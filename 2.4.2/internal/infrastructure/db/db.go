package db

import (
	"database/sql"
	"log"
)

func SelectDB(path string) (db *sql.DB) {
	if path != "" {
		dbS, err := NewDataBaseSqlite(path)
		if err != nil {
			log.Fatal(err)
			return nil
		}

		err = dbS.Migrate()
		if err != nil {
			log.Fatal(err)
			return nil
		}

		db = dbS.DB
	} else {
		dbP, err := NewDataBasePostgres()
		if err != nil {
			log.Fatal(err)
			return nil
		}

		err = dbP.Migrate()
		if err != nil {
			log.Fatal(err)
			return nil
		}

		db = dbP.DB
	}

	return db
}
