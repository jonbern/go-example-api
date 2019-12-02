package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type conf struct {
	port string
	db   confDB
	jwt  confJWT
}

type confDB struct {
	host string
	port string
	user string
	pass string
	name string
}

type confJWT struct {
	secret string
}

func newConfig() conf {
	godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/jonbern/go-example-api/.env"))

	return conf{
		port: getEnvOrDefault("PORT", "8080"),
		db: confDB{
			host: getEnvOrDefault("DB_HOST", "127.0.0.1"),
			port: getEnvOrDefault("DB_PORT", "3306"),
			user: os.Getenv("DB_USER"),
			pass: os.Getenv("DB_PASS"),
			name: getEnvOrDefault("DB_NAME", "invoices"),
		},
		jwt: confJWT{
			secret: os.Getenv("JWT_SECRET"),
		},
	}
}

func (c *conf) getDSN(dbName string) string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true",
		c.db.user, c.db.pass, c.db.host, c.db.port, dbName)
}

func getEnvOrDefault(envName string, defaultValue string) string {
	value := os.Getenv(envName)

	if value == "" {
		return defaultValue
	}
	return value
}
