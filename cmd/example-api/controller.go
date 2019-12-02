package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func optionsResponse(methods string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", methods)
		w.WriteHeader(http.StatusNoContent)
	}
}

func getInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	errCh := make(chan error)
	ch := make(chan []invoice)
	timeout := time.After(50 * time.Millisecond)

	go func() {
		invoices, err := model.getAll(ctx)
		if err != nil {
			errCh <- err
		} else {
			ch <- invoices
		}
	}()

	select {
	case invoices := <-ch:
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "    ")

		if err := encoder.Encode(invoices); err != nil {
			logger.panic(r, err)
		}
	case <-timeout:
		http.Error(w, "Request timed out", http.StatusGatewayTimeout)
	case err := <-errCh:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		logger.error(r, err)
	}
}

func getInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not convert id=%d to integer", id), http.StatusUnprocessableEntity)
		logger.error(r, err)
		return
	}

	ctx := context.TODO()
	invoice, err := model.getByID(ctx, id)
	if err != nil {
		switch err.(type) {
		case NotFoundError:
			http.Error(w, err.Error(), http.StatusNotFound)
			logger.error(r, err)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.error(r, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(invoice); err != nil {
		logger.panic(r, err)
	}

}

func createInvoice(w http.ResponseWriter, r *http.Request) {
	var i invoice
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000))
	if err != nil {
		logger.panic(r, err)
	}
	if err := r.Body.Close(); err != nil {
		logger.panic(r, err)
	}
	if err := json.Unmarshal(body, &i); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			logger.panic(r, err)
		}
		return
	}

	ctx := context.TODO()
	result, err := model.create(ctx, i)
	if err != nil {
		logger.error(r, err)
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.panic(r, err)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.Method + " " + r.URL.RequestURI()
	logger.info(r, msg+" "+strconv.Itoa(http.StatusNotFound))
	http.Error(w, msg, http.StatusNotFound)
}
