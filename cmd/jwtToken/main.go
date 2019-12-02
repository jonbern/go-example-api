package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("SECRET env variable not defined")
	}

	hmacSampleSecret := []byte(secret)

	type invoicesClaims struct {
		GetInvoices   bool `json:"getInvoices,omitempty"`
		GetInvoice    bool `json:"getInvoice,omitempty"`
		CreateInvoice bool `json:"createInvoice,omitempty"`
	}

	type Claims struct {
		invoicesClaims
		jwt.StandardClaims
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		invoicesClaims{
			GetInvoices: true,
			GetInvoice: true,
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

	fmt.Println(tokenString)
}

func getExpiry() int64 {
	timestamp := time.Now()
	duration, err := time.ParseDuration("24h")
	if err != nil {
		log.Fatal(err)
	}

	return timestamp.Add(duration).Unix()
}