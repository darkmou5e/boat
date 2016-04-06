package boat

import (
	"database/sql"
	"errors"
	"fmt"
)

// Tenant is data struct for doc in master table.
type Tenant struct {
	Name   string
	Active bool
	Aux    interface{} // For any useful auxilary data.
}

const MASTER = -1

func Use(tenantId int, tx *sql.Tx) {
	var schemaName string
	if tenantId == MASTER {
		schemaName = "master"
	} else {
		schemaName = fmt.Sprintf("tenant_%d", tenantId)
	}

	_, err := tx.Exec("SET search_path = " + schemaName)
	if err != nil {
		panic(fmt.Errorf("Can't switch default schema to '%s': %s", schemaName, err))
	}
}

// It must return error or panics as all other?..
func Bootstrap(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	EnsureSchema("master", tx)
	Use(MASTER, tx)
	EnsureCollection("tenants", tx)
	EnsureGINIndex("tenants", tx)

	return tx.Commit()
}

func EnsureTenant(tenant *Tenant, tenantInit func(tx *sql.Tx), tx *sql.Tx) {
	if tenant.Name == "master" {
		panic(errors.New("Tenant name 'master' is not allowed."))
	}

	Use(MASTER, tx)
	tenantId := Insert(tenant, "tenants", tx)
	tenantName := fmt.Sprintf("tenant_%d", tenantId)
	EnsureSchema(tenantName, tx)
	Use(tenantId, tx)
	tenantInit(tx)
}

func DropTenant(tenantId int, tx *sql.Tx) {
	Use(MASTER, tx)
	Delete(tenantId, "tenants", tx)

	_, err := tx.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS tenant_%d CASCADE", tenantId))
	if err != nil {
		panic(fmt.Errorf("Can't drop tenant with id '%d': %s", tenantId, err))
	}
}

func FindTenantByName(tenantName string, tenant *Tenant, tx *sql.Tx) (tenantId int, found bool) {
	Use(MASTER, tx)
	rows := Select("tenants", tx, `WHERE doc->>'Name' = $1 LIMIT 1`, tenantName)

	empty := true
	for rows.Next() {
		empty = false
		rows.Scan(&tenantId, tenant)
	}
	rows.Close()

	if empty {
		return 0, false
	}
	return tenantId, true
}
