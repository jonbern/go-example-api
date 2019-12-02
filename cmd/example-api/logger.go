package main

import (
	"fmt"
	"log"
	"net/http"
)

type requestLogger struct{}

func (l *requestLogger) getID(r *http.Request) string {
	return r.Header.Get("X-Correlation-ID")
}

func (l *requestLogger) info(r *http.Request, msg string) {
	log.Println(fmt.Sprintf("ID=%q: %q", l.getID(r), msg))
}

func (l *requestLogger) warn(r *http.Request, msg string) {
	log.Println(fmt.Sprintf("WARN: ID=%q: %q", l.getID(r), msg))
}

func (l *requestLogger) error(r *http.Request, err error) {
	log.Println(fmt.Sprintf("ERROR: ID=%q: %q", l.getID(r), err))
}

func (l *requestLogger) panic(r *http.Request, err error) {
	log.Panic(fmt.Sprintf("ERROR: ID=%q: %q", l.getID(r), err))
}

func (l *requestLogger) fatal(r *http.Request, err error) {
	log.Fatal(fmt.Sprintf("ERROR: ID=%q: %q", l.getID(r), err))
}
