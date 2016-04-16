package boat

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/bezrukovspb/mux"
	"github.com/gorilla/context"
)

func AtomicRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := context.Get(r, "db").(*sql.DB)
		if db == nil {
			log.Fatal("db must be in request context")
		}
		tx, err := db.Begin()
		Check(err)
		defer tx.Rollback()
		context.Set(r, "tx", tx)

		h.ServeHTTP(w, r)

		tx.Commit()
	})
}

func UseTenantBySubdomain(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := GetTx(r)
		if tx == nil {
			log.Fatal("Tx in the context is nil.")
		}

		params := mux.Vars(r)
		subdomain, ok := params["subdomain"]
		if !ok {
			log.Fatal("subdomain must be in request vars.")
		}

		var tenant Tenant
		tenantId, found := FindTenantByName(subdomain, &tenant, tx)
		if !found {
			http.NotFound(w, r)
			return
		}
		context.Set(r, "tenantId", tenantId)
		context.Set(r, "tenant", tenant)
		Use(tenantId, tx)

		h.ServeHTTP(w, r)
	})
}

func UseMaster(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := GetTx(r)
		if tx == nil {
			log.Fatal("Tx in the context is nil.")
		}
		Use(MASTER, tx)

		h.ServeHTTP(w, r)
	})
}
