package boat

import (
	"database/sql"
	"errors"
)

// Tenant is data struct for doc in master table.
type Tenant struct {
	Name   string
	Active bool
	Aux    interface{} // For any useful auxilary data.
}

const MASTER = -1

func Use(tenantID int, tx *sql.Tx) (*sql.Tx, error) {
	var schemaName string
	if tenantID == MASTER {
		schemaName = "Master"
	} else {
		schemaName = "Tenant" + string(tenantID)
	}

	_, err := tx.Exec("SET search_path = " + schemaName)
	if err != nil {
		return nil, err
	}
	return tx, nil
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

	err = EnsureSchema("Master", tx)

	if err != nil {
		return err
	}

	tx, err = Use(MASTER, tx)

	if err != nil {
		return err
	}

	err = EnsureCollection("Tenants", tx)
	if err != nil {
		return err
	}

	err = EnsureGINIndex("Tenants", tx)
	if err != nil {
		return err
	}

	return nil
}

func EnsureTenant(tenant *Tenant, tenantInit func(tx *sql.Tx) error, tx *sql.Tx) error {
	if tenant.Name == "Master" {
		return errors.New(`Tenant name "Master" is not allowed.`)
	}

	tx, err := Use(MASTER, tx)
	if err != nil {
		return err
	}

	tenantID, err := Insert(tenant, "Tenants", tx)
	if err != nil {
		return err
	}

	tenantName := "Tenant" + string(tenantID)
	err = EnsureSchema(tenantName, tx)
	if err != nil {
		return err
	}

	tx, err = Use(tenantID, tx)
	if err != nil {
		return err
	}

	err = tenantInit(tx)
	if err != nil {
		return err
	}

	return nil
}

func DropTenant(tenantID int, tx *sql.Tx) error {
	tx, err := Use(MASTER, tx)
	if err != nil {
		return err
	}

	err = Delete(tenantID, "Tenants", tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP SCHEMA Tenant" + string(tenantID))
	if err != nil {
		return err
	}
	return nil
}
