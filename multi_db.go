package boat

import (
	"database/sql"
	"encoding/json"
	"errors"
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

func EnsureTenant(tenant *Tenant, tenantInit func(tx *sql.Tx) error, db *sql.DB) error {
	if tenant.Subdomain == "master" {
		return errors.New(`Tenant name "master" is not allowed.`)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	err = EnsureSchema(tenant.Subdomain, tx)
	if err != nil {
		return err
	}

	tx, err = Use("master", tx)
	if err != nil {
		return err
	}

	t, err := json.Marshal(tenant)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO tenants (doc) VALUES ($1)`, t)
	if err != nil {
		fmt.Println(string(t))
		return err
	}

	tx, err = Use(tenant.Subdomain, tx)
	if err != nil {
		return err
	}

	err = tenantInit(tx)
	if err != nil {
		return err
	}

	return nil
}
