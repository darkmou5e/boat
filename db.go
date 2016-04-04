package boat

import (
	"database/sql"

	_ "github.com/lib/pq" // The driver for Postgres.
)

var DB *sql.DB

// connectionURL like "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
// Boat uses Postgres only.
func Open(connectionURL string) (*sql.DB, error) {
	DB, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, err
	}
	return DB, err
}
