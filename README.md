# Go Example API template

`Go Example API` is a _batteries included_ template API for creating `RESTful` APIs in `Go`.


It includes many of the features necessary for most `RESTful` API such as:
- Basic routing (using the `github.com/gorilla/mux` router).
- `JWT` token authorization (and a `CLI` for generating tokens during development).
- Database back-end using the `database/sql` package for storing and retrieving data.
- Middleware for tracking requests using correlation IDs.
- Various middleware for logging, setting content-type, CORS headers etc.
- Database migrations for defining the initial database schema, and enabling future schema changes to be checked-in to source control, and applied as necessary.
- `E2E` (End-2-End) tests for black-box and acceptance testing.
- Convenience scripts for starting the API in a dev setting, and for running tests.
- Go Modules for managing third party dependencies
- Basic `Dockerfile` for streamlined deployments

## Requirements
- `Go` version 1.11 or later (due to [Go Modules](https://github.com/golang/go/wiki/Modules))
- `MySQL`/`MariaDB`

NOTE: This template uses `MySQL`/`MariaDB` as a database back-end, but this can be changed by modifying the implementation in `model.go` to use another database that meets your needs and preferences.

## Getting started

### Set the required environment variables:
Configuration of the API is done using environment variables, and during development it is convenient to use a local `.env` file placed in the project's root folder.

The following environment variables are required and must be configured:
- `DB_USER`: Login to be used when connecting to the database.
- `DB_PASSWORD`: Password of the login to be used when connecting to the database
- `JWT_SECRET`: Secret to be used when validating the signature of incoming JWT tokens

The environment variables below are optional, and have default values defined:

- `PORT`: The port number to listen to incoming HTTP requests. Default: 8080.
- `DB_HOST`: Hostname of database server. Default: 127.0.0.1.
- `DB_PORT`: Port number of database server. Default: 3306.
- `DB_NAME`: Name of the database to use. Default: invoices.


### Initialize an empty database:
Ensure there is an empty database on the database server with the name of the `DB_NAME` value (Default: invoices).

When the API starts, a migration script is run to ensure the configured database has the expected schema defined. However, the API will not create the database and it must already exist before starting the API.

### Start the API:
The API can be built and started using the `./start.sh` script.

This script is mainly for dev purposes, as it uses `go run` and also generates and logs outs a valid JWT token that can be used to interact with the API.

For production it is recommended to use `go build` and ship a fat binary, or alternatively generate a docker image ([Building Docker Images for Static Go Binaries](https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07)) which enables a streamlined deployment workflow in your environment.

### Running tests:
To run tests use the `./tests.sh` script which will run all tests in the project.

Out of the box, there are only `E2E` tests defined. These tests creates  transient test databases for each scenario tested. This enables the `E2E` tests to run in isolation without affecting each other. `E2E` tests
allows us to verify the expectations of the API as a client of the API with a high level of confidence.

For a simple project with limited business logic, having only `E2E` tests can be sufficient. However, if the API has extensive logic it is strongly recommended to create a unit test suite which thoroughly validates the expectations of the code.
