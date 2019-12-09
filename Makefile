MAIN_CMD=example-api

.PHONY: start
start:
	go run ./cmd/$(MAIN_CMD)

build:
	go build ./cmd/$(MAIN_CMD)

.PHONY: test
test:
	go test ./cmd/example-api/...

.PHONY: token
token:
	echo "Generating JWT token":
	go run ./cmd/jwtToken

.PHONY: clean
clean:
	rm $(MAIN_CMD)