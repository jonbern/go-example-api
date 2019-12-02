package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"regexp"
)

type claimsContextKey string

var ctxKeyClaims claimsContextKey = claimsContextKey("claims")

func init() {
	if config.jwt.secret == "" {
		log.Fatal("JWT_SECRET env variable not defined")
	}
}

func checkAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := regexp.MustCompile("(?i)(Bearer\\s)").ReplaceAllString(authorizationHeader, "")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.jwt.secret), nil
		})

		if token == nil {
			http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), ctxKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			logger.info(r, err.Error())
			http.Error(w, "Invalid or expired JWT token", http.StatusUnauthorized)
		}
	})
}

func checkPermission(handlerFunc func(w http.ResponseWriter, r *http.Request), permission string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if claims := r.Context().Value(ctxKeyClaims).(jwt.MapClaims); claims != nil {
			v := claims[permission]
			if v == nil || v.(bool) != true {
				http.Error(w, "Operation not permitted", http.StatusForbidden)
				return
			}
		} else {
			logger.panic(r, errors.New("claims not found in context"))
		}
		handlerFunc(w, r)
	}
}
