package boat

import (
	"database/sql"

	_ "github.com/lib/pq" // The driver for Postgres.
)

// DB extends sql.DB type with methods related to a multi-tenant use case.
var DB *sql.DB

// Use ...
func Use(schemaName string, tx *sql.Tx) (*sql.Tx, error) {

	_, err := tx.Exec("SET search_path = " + schemaName)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

// Open wraps sql.Open function and returns pointer on *DB instead of *sql.DB.
func Open(connectionParams string) (*sql.DB, error) {
	DB, err := sql.Open("postgres", connectionParams) // Boat uses Postgres only.

	if err != nil {
		return nil, err
	}

	return DB, err
}
