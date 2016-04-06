package boat

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// All CRUD functions are here. If you need something else - use SQL.

func Insert(doc interface{}, collection string, tx *sql.Tx) (docId int) {
	d, err := json.Marshal(doc)
	if err != nil {
		panic(fmt.Errorf("Can't marshal to json the doc '%+v': %s", doc, err))
	}

	queryText := fmt.Sprintf(`INSERT INTO %s (doc) VALUES ($1) RETURNING id`, collection)
	err = tx.QueryRow(queryText, d).Scan(&docId)
	if err != nil {
		panic(fmt.Errorf("Can't insert the doc '%s' to the collection '%s': %s", string(d), collection, err))
	}
	return docId
}

func Delete(docId int, collection string, tx *sql.Tx) {
	queryText := fmt.Sprintf("DELETE FROM %s WHERE id = $1", collection)
	_, err := tx.Exec(queryText, docId)
	if err != nil {
		panic(fmt.Errorf("Can't delete the doc with id '%d' from the collection '%s': %s", docId, collection, err))
	}
}

func Update(doc interface{}, docId int, collection string, tx *sql.Tx) {
	d, err := json.Marshal(doc)
	if err != nil {
		panic(fmt.Errorf("Can't marshal to json the doc '%+v': %s", doc, err))
	}

	queryText := fmt.Sprintf(`UPDATE %s SET doc = $1 WHERE id = $2`, collection)
	_, err = tx.Exec(queryText, d, docId)
	if err != nil {
		panic(fmt.Errorf("Can't update the doc with id '%d' in the collection '%s' with new doc '%s': %s", docId, collection, string(d), err))
	}
}

func Find(docId int, collection string, doc interface{}, tx *sql.Tx) (found bool) {
	var d []byte
	queryText := fmt.Sprintf(`SELECT doc FROM %s WHERE id = $1`, collection)
	err := tx.QueryRow(queryText, docId).Scan(&d)

	switch err {
	case sql.ErrNoRows:
		return false
	case nil:
	default:
		panic(fmt.Errorf("Can't select the doc with id '%d' in the collection '%s': %s", docId, collection, err))
	}

	err = json.Unmarshal(d, &doc)
	if err != nil {
		panic(fmt.Errorf("Can't unmarshal to struct '%+v' the json '%s': %s", doc, string(d), err))
	}

	return true
}

// Make our Rows with a blackjack
type Rows struct {
	rows *sql.Rows
}

func (r *Rows) Next() bool {
	return r.rows.Next()
}

func (r *Rows) Scan(docId *int, doc interface{}) {
	var d []byte

	err := r.rows.Scan(docId, &d)
	if err != nil {
		panic(fmt.Errorf("Can't scan docId and/or doc to struct '%+v': %s", d, err))
	}

	err = json.Unmarshal(d, doc)
	if err != nil {
		panic(fmt.Errorf("Can't unmarshal to struct '%+v' the json '%s': %s", doc, string(d), err))
	}
}

func (r *Rows) Close() {
	r.rows.Close()
}

func Select(collection string, tx *sql.Tx, where string, params ...interface{}) (rows Rows) {
	queryText := fmt.Sprintf(`SELECT id, doc FROM %s %s`, collection, where)
	var err error
	rows.rows, err = tx.Query(queryText, params...)
	if err != nil {
		panic(fmt.Errorf("Can't select the docs from the collection '%s' with the query '%s' and params %v: %s", collection, queryText, params, err))
	}
	return rows
}
