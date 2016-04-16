package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"

	"github.com/bezrukovspb/boat"
	"github.com/bezrukovspb/mux"
	"github.com/gorilla/context"
)

var db *sql.DB

func init() {
	var err error
	dbURL := os.Getenv("BOAT_TEST_DB_URL")
	db, err = boat.Open(dbURL)
	boat.Check(err)
}

type message struct {
	Title string
	Text  string
}

func initTenant(tx *sql.Tx) {
	boat.EnsureCollection("messages", tx)
	boat.EnsureGINIndex("messages", tx)
}

func masterHome(w http.ResponseWriter, r *http.Request) {
	tx := boat.GetTx(r)

	type data struct {
		Tenants []boat.Tenant
	}

	d := data{Tenants: []boat.Tenant{}}

	rows := boat.Select("tenants", tx, "")
	defer rows.Close()
	for rows.Next() {
		var id int
		var t boat.Tenant

		rows.Scan(&id, &t)
		d.Tenants = append(d.Tenants, t)
	}

	const tpl = `<!DOCTYPE html>
							<html>
								<head>
									<meta charset="UTF-8">
									<title>Boat</title>
								</head>
								<body>
								  <ul>{{range .Tenants}}<li><a href="http://{{ .Name }}.boat.localhost:8000">{{ .Name }}</a></li>{{end}}</ul>
									<form action="add" method="POST">
									New tenant <input type="text" name="tenant">
									<input type="submit">
									</form>
								</body>
							</html>`

	var tmpl = template.Must(template.New("home").Parse(tpl))
	err := tmpl.Execute(w, d)
	boat.Check(err)
}

func appHome(w http.ResponseWriter, r *http.Request) {
	tx := boat.GetTx(r)

	type data struct {
		Messages []message
	}

	d := data{Messages: []message{}}

	rows := boat.Select("messages", tx, "")
	defer rows.Close()
	for rows.Next() {
		var id int
		var m message

		rows.Scan(&id, &m)
		d.Messages = append(d.Messages, m)
	}

	const tpl = `<!DOCTYPE html>
							<html>
								<head>
									<meta charset="UTF-8">
									<title>Boat</title>
								</head>
								<body>
								  <ul>{{range .Messages}}<li><b>{{ .Title }}</b>
									<br>
									{{ .Text }}</li>{{end}}</ul>
									<form action="add" method="POST">
									title <input type="text" name="messageTitle">
									<br>
									message<textarea name="text"></textarea>
									<input type="submit">
									</form>
								</body>
							</html>`

	tmpl := template.Must(template.New("homeApp").Parse(tpl))
	err := tmpl.Execute(w, d)
	boat.Check(err)
}

func addTenant(w http.ResponseWriter, r *http.Request) {
	tx := boat.GetTx(r)
	err := r.ParseForm()
	boat.Check(err)

	tenant := boat.Tenant{Name: r.PostForm.Get("tenant"), Active: true}
	boat.EnsureTenant(&tenant, initTenant, tx)
	boat.Check(err)

	http.Redirect(w, r, "/", 307)
}

func addMessage(w http.ResponseWriter, r *http.Request) {
	tx := boat.GetTx(r)
	err := r.ParseForm()
	boat.Check(err)

	msg := message{Title: r.Form.Get("messageTitle"), Text: r.Form.Get("text")}
	boat.Insert(msg, "messages", tx)
	http.Redirect(w, r, "/", 307)
}

func main() {
	err := boat.Bootstrap(db)
	boat.Check(err)

	setDB := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "db", db)
			h.ServeHTTP(w, r)
		})
	}

	rootRouter := mux.NewRouter().Use(setDB, boat.AtomicRequest)

	masterRouter := rootRouter.Host("boat.localhost").Subrouter().Use(boat.UseMaster)
	masterRouter.Path("/").HandlerFunc(masterHome)
	masterRouter.Path("/add").Methods("POST").HandlerFunc(addTenant)

	appRouter := rootRouter.Host("{subdomain:[a-z]+?}.boat.localhost").Subrouter().Use(boat.UseTenantBySubdomain)
	appRouter.Path("/").HandlerFunc(appHome)
	appRouter.Path("/add").Methods("POST").HandlerFunc(addMessage)

	http.ListenAndServe("127.0.0.1:8000", rootRouter)
}
