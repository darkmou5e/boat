package boat

import (
	"database/sql"
	"os"
	"testing"
)

func connectDB(t *testing.T) *sql.DB {
	url := os.Getenv("BOAT_TEST_DB_URL")
	db, err := Open(url)
	if err != nil {
		t.Fatalf("Can't connect to database '%s': %s", url, err)
	}
	return db
}

type message struct {
	Title string
	Text  string
}

func initTenant(tx *sql.Tx) {
	EnsureCollection("messages", tx)
	EnsureGINIndex("messages", tx)
}

func removeExistingTenants(t *testing.T, tx *sql.Tx) {
	Use(MASTER, tx)

	rows := Select("tenants", tx, "")
	var id int
	var tenant Tenant
	var tenantsIds []int
	for rows.Next() {
		rows.Scan(&id, &tenant)
		tenantsIds = append(tenantsIds, id)
	}
	rows.Close()

	for _, tid := range tenantsIds {
		DropTenant(tid, tx)
	}
}

func testCRUD(tenantId int, t *testing.T, tx *sql.Tx) {
	Use(tenantId, tx)

	// Insert and Find
	msg := message{Title: "First", Text: "This is the first message for test"}
	msgId := Insert(msg, "messages", tx)

	var msgFound message
	found := Find(msgId, "messages", &msgFound, tx)

	if !found {
		t.Fatalf("Doc with id '%d' not found in collection after insert, but it must be there", msgId)
	}

	if msg != msgFound {
		t.Fatalf("One doc inserted and gotten from a collection must be the same Inserted = %s; Gotten = %s", msg, msgFound)
	}

	msg.Title = "Updated"
	Update(msg, msgId, "messages", tx)

	found = Find(msgId, "messages", &msgFound, tx)
	if !found {
		t.Fatalf("Doc with id '%d' not found in collection after update, but it must be there", msgId)
	}

	if msg != msgFound {
		t.Fatalf("One doc updated and gotten from a collection must be the same Updated = %s; Gotten = %s", msg, msgFound)
	}
}

func createTestTenants(t *testing.T, tx *sql.Tx) (tenantId int) {
	Use(MASTER, tx)

	tenant := Tenant{Name: "test", Active: true}
	EnsureTenant(&tenant, initTenant, tx)
	tenantId, found := FindTenantByName("test", &tenant, tx)
	if !found {
		t.Fatalf("Problem with finding of tenantId for test tenant")
	}
	return tenantId
}

func TestBoat(t *testing.T) {
	db := connectDB(t)

	err := Bootstrap(db)
	if err != nil {
		t.Fatalf("Can't to bootstrap Boat: %s", err)
	}

	// Drop already exiting tenants
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Can't start a transaction: %s", err)
	}
	defer tx.Rollback() // Rollback if the transaction is still not commited.

	removeExistingTenants(t, tx)
	tenantId := createTestTenants(t, tx)
	testCRUD(tenantId, t, tx)

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Can't commit transaction: %s", err)
	}
}
