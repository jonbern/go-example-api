package tutils

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql" //go-lint-ignore
)

// InvoicesClaims defines the JWT claims available in the application
type InvoicesClaims struct {
	GetInvoices   bool `json:"getInvoices,omitempty"`
	GetInvoice    bool `json:"getInvoice,omitempty"`
	CreateInvoice bool `json:"createInvoice,omitempty"`
}

// GenerateToken generates a JWT token using the provided JWT secret and InvoicesClaims
func GenerateToken(jwtSecret string, claims InvoicesClaims) string {
	hmacSampleSecret := []byte(jwtSecret)

	type Claims struct {
		InvoicesClaims
		jwt.StandardClaims
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		InvoicesClaims{
			GetInvoices:   true,
			GetInvoice:    true,
			CreateInvoice: true,
		},
		jwt.StandardClaims{
			ExpiresAt: getExpiry(),
		},
	})

	tokenString, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		log.Fatal(err)
	}

	return tokenString
}

func getExpiry() int64 {
	timestamp := time.Now()
	duration, err := time.ParseDuration("24h")
	if err != nil {
		log.Fatal(err)
	}

	return timestamp.Add(duration).Unix()
}

// CreateTestDB creates a test database with a random generated name and returns a fn to drop the database
func CreateTestDB(user, pass string) (string, func() error, error) {
	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("test_db_%v", rand.Intn(1000))

	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@/", user, pass))
	nilFn := func() error { return err }
	if err != nil {
		return "", nilFn, err
	}
	if err != nil {
		return "", nilFn, err
	}
	err = db.Ping()
	if err != nil {
		return "", nilFn, err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %v", dbName))
	if err != nil {
		return "", nilFn, err
	}

	dropDbFn := func() error {
		defer db.Close()
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE %v", dbName))
		return err
	}

	return dbName, dropDbFn, nil
}
