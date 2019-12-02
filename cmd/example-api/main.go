package main

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var logger requestLogger = requestLogger{}
var model invoicesModel

var config conf = newConfig()

const schemaVersion = 1

func main() {
	config := newConfig()
	router := newAPI(config.db.name)
	go heartBeat()
	log.Println(fmt.Sprintf("Listening to request on port=%v", config.port))
	log.Fatal(http.ListenAndServe(":"+config.port, router))
}

func heartBeat() {
	var m runtime.MemStats
	for {
		runtime.ReadMemStats(&m)
		log.Println(fmt.Sprintf("Heart beat: TotalAlloc=%v, Go routines=%v", m.TotalAlloc, runtime.NumGoroutine()))
		time.Sleep(time.Minute)
	}
}

func newAPI(dbName string) *mux.Router {
	dsn := config.getDSN(dbName)

	m, err := migrate.New(
		fmt.Sprintf("file://%v", os.ExpandEnv("$GOPATH/src/github.com/jonbern/go-example-api/cmd/example-api/migrations")),
		fmt.Sprintf("mysql://%v", dsn))
	if err != nil {
		log.Panic(err)
	}
	m.Steps(schemaVersion)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Panic(err.Error())
	}

	model = newInvoicesModel(db)

	router := mux.NewRouter().StrictSlash(true)
	router.Use(ensureCorrelationID)
	router.Use(logRequest)
	router.Use(checkAuthorization)
	router.Use(setContentType)
	router.Use(setAccessControlAllowOrigin)

	router.Methods(http.MethodOptions).
		Path("/invoices").
		HandlerFunc(optionsResponse("GET,POST,OPTIONS"))
	router.Methods(http.MethodGet).
		Path("/invoices").
		HandlerFunc(checkPermission(getInvoices, "getInvoices"))
	router.Methods(http.MethodPost).
		Path("/invoices").
		HandlerFunc(checkPermission(createInvoice, "createInvoice"))

	router.Methods(http.MethodOptions).
		Path("/invoices/{id}").
		HandlerFunc(optionsResponse("GET,OPTIONS"))
	router.Methods(http.MethodGet).
		Path("/invoices/{id}").
		HandlerFunc(checkPermission(getInvoice, "getInvoice"))

	router.PathPrefix("/").HandlerFunc(notFoundHandler)
	return router
}
