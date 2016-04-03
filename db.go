package boat

import (
	"database/sql"

	_ "github.com/lib/pq" // The driver for Postgres.
)

var DB *sql.DB

func Use(schemaName string, tx *sql.Tx) (*sql.Tx, error) {

	_, err := tx.Exec("SET search_path = " + schemaName)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func Open(connectionParams string) (*sql.DB, error) {
	DB, err := sql.Open("postgres", connectionParams) // Boat uses Postgres only.

	if err != nil {
		return nil, err
	}

	return DB, err
}
