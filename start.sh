#!/bin/bash

# Generates a JWT token using configured JWT_SECRET for convenience
# when developing. Avoid using this helper script in production, as we may
# not want to print a valid JWT token to console.

echo "JWT token for dev purposes:"
go run ./cmd/jwtToken

printf "\nStarting api:\n"
go run ./cmd/example-api