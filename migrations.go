package boat

import (
	"database/sql"
	"fmt"
)

// EnsureSchema checks existance of a schema and creates it, if it is
// not exist.
func EnsureSchema(name string, tx *sql.Tx) error {
	queryText := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, name)

	_, err := tx.Exec(queryText)
	if err != nil {
		return err
	}

	return nil
}

// EnsureCollection checks existance of a collection and creates it, if it is
// not exist.
func EnsureCollection(name string, tx *sql.Tx) error {
	queryText := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
                              id serial primary key,
                              doc jsonb
                              )`, name)

	_, err := tx.Exec(queryText)
	if err != nil {
		return err
	}

	return nil
}

// EnsureGINIndex checks existance of a GIN index on doc fileld for table tableName
// and creates it, if it is not exist.
func EnsureGINIndex(tableName string, tx *sql.Tx) error {
	queryText := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s_doc_gin on %s USING GIN (doc)`, tableName, tableName)

	_, err := tx.Exec(queryText)
	if err != nil {
		return err
	}

	return nil
}
