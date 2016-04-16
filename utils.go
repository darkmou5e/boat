package boat

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/context"
)

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetTx(r *http.Request) *sql.Tx {
	return context.Get(r, "tx").(*sql.Tx)
}
