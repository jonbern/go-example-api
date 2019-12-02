package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jonbern/go-example-api/pkg/tutils"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func setup() (*httptest.Server, func()) {
	dbName, dropDB, err := tutils.CreateTestDB(config.db.user, config.db.pass)
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(newAPI(dbName))

	return ts, func() {
		defer ts.Close()
		err := dropDB()
		if err != nil {
			panic(err)
		}
	}
}

var endpoints = []struct {
	verb string
	path string
}{
	{"GET", "/invoices"},
	{"GET", "/invoices/1"},
	{"POST", "/invoices"},
}

func TestEndpoints_WithoutToken(t *testing.T) {
	ts, teardown := setup()
	defer teardown()

	for _, x := range endpoints {
		req, err := http.NewRequest(x.verb, ts.URL+x.path, nil)
		if err != nil {
			t.Errorf(err.Error())
		}

		client := &http.Client{}
		res, err := client.Do(req)

		if err != nil {
			t.Errorf(err.Error())
		}

		t.Run(fmt.Sprintf("%v %v: Responds with 401", x.verb, x.path), func(t *testing.T) {
			if res.StatusCode != 401 {
				t.Errorf("Should return status code %v. Returned code was: %v", 401, res.StatusCode)
			}
		})

		t.Run(fmt.Sprintf("%v %v: Returns error message content", x.verb, x.path), func(t *testing.T) {
			body, err := ioutil.ReadAll(res.Body)
			defer res.Body.Close()
			if err != nil {
				t.Errorf(err.Error())
			}

			expected := "Missing Authorization header\n"

			if string(body) != expected {
				t.Errorf("Expected error message: %q, but got %q", expected, string(body))
			}
		})

	}
}

func TestEndpoints_WithInvalidToken(t *testing.T) {
	ts, teardown := setup()
	defer teardown()

	for _, x := range endpoints {
		req, err := http.NewRequest(x.verb, ts.URL+x.path, nil)
		if err != nil {
			t.Errorf(err.Error())
		}

		req.Header.Add("Authorization", "Bearer this-is-not-the-token-you're-looking-for")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Errorf(err.Error())
		}

		t.Run(fmt.Sprintf("%v %v: Responds with 401", x.verb, x.path), func(t *testing.T) {
			if res.StatusCode != 401 {
				t.Errorf("Should return status code %v. Returned code was: %v", 401, res.StatusCode)
			}
		})

		t.Run(fmt.Sprintf("%v %v: Returns error message content", x.verb, x.path), func(t *testing.T) {
			body, err := ioutil.ReadAll(res.Body)
			defer res.Body.Close()
			if err != nil {
				t.Errorf(err.Error())
			}

			expected := "Invalid JWT token\n"

			if string(body) != expected {
				t.Errorf("Expected error message: %q, but got %q", expected, string(body))
			}
		})
	}
}

func TestEndpoints_WithCorrelationID(t *testing.T) {
	ts, teardown := setup()
	defer teardown()

	for _, x := range endpoints {
		req, err := http.NewRequest(x.verb, ts.URL+x.path, nil)
		if err != nil {
			t.Errorf(err.Error())
		}

		req.Header.Add("Authorization", "Bearer "+tutils.GenerateToken(config.jwt.secret, tutils.InvoicesClaims{GetInvoices: true}))

		correlationID := "correlation-ID-bla-bla"
		req.Header.Add("X-Correlation-ID", correlationID)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Errorf(err.Error())
		}

		value := res.Header.Get("X-Correlation-ID")

		if value == "" {
			t.Errorf("Missing or empty X-Correlation-ID header")
		}

		if value != correlationID {
			t.Errorf("It honours given correlation id (%v)", correlationID)
		}
	}
}

func TestCreateInvoice(t *testing.T) {
	ts, teardown := setup()
	defer teardown()

	customerID := 24002
	description := "Office supplies, and other fascinating items"
	dueDate, err := time.Parse(time.RFC3339, "2019-10-23T00:00:00Z")
	if err != nil {
		t.Errorf(err.Error())
	}
	amount := 1024.12

	expected := invoice{
		CustomerID:  customerID,
		Description: description,
		DueDate:     dueDate,
		Amount:      amount,
	}

	jsonPayload, err := json.Marshal(expected)
	if err != nil {
		t.Errorf(err.Error())
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/invoices", ts.URL), bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Errorf(err.Error())
	}

	req.Header.Add("Authorization", "Bearer "+tutils.GenerateToken(config.jwt.secret, tutils.InvoicesClaims{GetInvoices: true}))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf(err.Error())
	}

	t.Run("Responds with 201", func(t *testing.T) {
		if res.StatusCode != 201 {
			t.Errorf("Should return status code %v", 201)
		}
	})

	t.Run("Has Content-Type application/json", func(t *testing.T) {
		contentType := res.Header.Get("Content-Type")

		if contentType == "" {
			t.Errorf("Missing or empty Content-Type header")
		}

		if strings.Index(contentType, "application/json") != 0 {
			t.Errorf("Content-Type should be application/json")
		}
	})

	t.Run("Has CORS wild-card", func(t *testing.T) {
		value := res.Header.Get("Access-Control-Allow-Origin")

		if value == "" {
			t.Errorf("Missing or empty Access-Control-Origin header")
		}

		if value != "*" {
			t.Errorf("Access-Control-Origin should be wildcard (*)")
		}
	})

	t.Run("Sets X-Correlation-ID header", func(t *testing.T) {
		value := res.Header.Get("X-Correlation-ID")

		if value == "" {
			t.Errorf("Missing or empty X-Correlation-ID header")
		}

		re := regexp.MustCompile("([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}")

		if !re.Match([]byte(value)) {
			t.Errorf("X-Correlation-ID should be set to a valid uuid")
		}
	})

	t.Run("Returns created invoice", func(t *testing.T) {
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			t.Errorf(err.Error())
		}

		var result invoice
		if err := json.Unmarshal(body, &result); err != nil {
			t.Errorf(err.Error())
		}

		expected.ID = result.ID // We don't know the ID before it has been created
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("API should create invoice with provided values")
		}

	})
}

func TestGetInvoice(t *testing.T) {
	ts, teardown := setup()
	defer teardown()

	ctx := context.Background()

	_, err := model.create(ctx, invoice{CustomerID: 0, Description: "First invoice", DueDate: time.Now(), Amount: 123.43})
	if err != nil {
		t.Errorf(err.Error())
	}
	expected, err := model.create(ctx, invoice{CustomerID: 0, Description: "Another invoice", DueDate: time.Now(), Amount: 1})
	if err != nil {
		t.Errorf(err.Error())
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/invoices/%v", ts.URL, expected.ID), nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	req.Header.Add("Authorization", "Bearer "+tutils.GenerateToken(config.jwt.secret, tutils.InvoicesClaims{GetInvoices: true}))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf(err.Error())
	}

	t.Run("Responds with 200", func(t *testing.T) {
		if res.StatusCode != 200 {
			t.Errorf("Should return status code %v", 200)
		}
	})

	t.Run("Has Content-Type application/json", func(t *testing.T) {
		contentType := res.Header.Get("Content-Type")

		if contentType == "" {
			t.Errorf("Missing or empty Content-Type header")
		}

		if strings.Index(contentType, "application/json") != 0 {
			t.Errorf("Content-Type should be application/json")
		}
	})

	t.Run("Has CORS wild-card", func(t *testing.T) {
		value := res.Header.Get("Access-Control-Allow-Origin")

		if value == "" {
			t.Errorf("Missing or empty Access-Control-Origin header")
		}

		if value != "*" {
			t.Errorf("Access-Control-Origin should be wildcard (*)")
		}
	})

	t.Run("Sets X-Correlation-ID header", func(t *testing.T) {
		value := res.Header.Get("X-Correlation-ID")

		if value == "" {
			t.Errorf("Missing or empty X-Correlation-ID header")
		}

		re := regexp.MustCompile("([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}")

		if !re.Match([]byte(value)) {
			t.Errorf("X-Correlation-ID should be set to a valid uuid")
		}
	})

	t.Run("Returns specified invoice", func(t *testing.T) {
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			t.Errorf(err.Error())
		}

		var result invoice

		if err := json.Unmarshal(body, &result); err != nil {
			t.Errorf(err.Error())
		}

		if result.ID != expected.ID {
			t.Errorf("API should return specified invoice")
		}

	})
}

func TestGetInvoices(t *testing.T) {
	ts, teardown := setup()
	defer teardown()

	ctx := context.Background()
	model.create(ctx, invoice{CustomerID: 0, Description: "First invoice", DueDate: time.Now(), Amount: 123.43})

	req, err := http.NewRequest("GET", ts.URL+"/invoices", nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	req.Header.Add("Authorization", "Bearer "+tutils.GenerateToken(config.jwt.secret, tutils.InvoicesClaims{GetInvoices: true}))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf(err.Error())
	}

	t.Run("Responds with 200", func(t *testing.T) {
		if res.StatusCode != 200 {
			t.Errorf("Should return status code %v", 200)
		}
	})

	t.Run("Has Content-Type application/json", func(t *testing.T) {
		contentType := res.Header.Get("Content-Type")

		if contentType == "" {
			t.Errorf("Missing or empty Content-Type header")
		}

		if strings.Index(contentType, "application/json") != 0 {
			t.Errorf("Content-Type should be application/json")
		}
	})

	t.Run("Has CORS wild-card", func(t *testing.T) {
		value := res.Header.Get("Access-Control-Allow-Origin")

		if value == "" {
			t.Errorf("Missing or empty Access-Control-Origin header")
		}

		if value != "*" {
			t.Errorf("Access-Control-Origin should be wildcard (*)")
		}
	})

	t.Run("Sets X-Correlation-ID header", func(t *testing.T) {
		value := res.Header.Get("X-Correlation-ID")

		if value == "" {
			t.Errorf("Missing or empty X-Correlation-ID header")
		}

		re := regexp.MustCompile("([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}){1}")

		if !re.Match([]byte(value)) {
			t.Errorf("X-Correlation-ID should be set to a valid uuid")
		}
	})

	t.Run("Returns Invoices", func(t *testing.T) {
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			t.Errorf(err.Error())
		}

		var invoices []invoice

		if err := json.Unmarshal(body, &invoices); err != nil {
			t.Errorf(err.Error())
		}

		if len(invoices) == 0 {
			t.Error("API should return Invoices")
		}
	})
}
