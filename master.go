package boat

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Tenant is data struct for doc in master table.
type Tenant struct {
	Name      string
	Subdomain string
	Active    bool
}

func Bootstrap(db *sql.DB) error {
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

	err = EnsureSchema("master", tx)

	if err != nil {
		return err
	}

	tx, err = Use("master", tx)

	if err != nil {
		return err
	}

	err = EnsureCollection("tenants", tx)
	if err != nil {
		return err
	}

	err = EnsureGINIndex("tenants", tx)
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
