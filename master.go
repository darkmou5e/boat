package boat

import "database/sql"

// Tenant is data struct for doc in master table.
type Tenant struct {
	Name      string
	Subdomain string
	Active    bool
}

// Bootstrap ...
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
