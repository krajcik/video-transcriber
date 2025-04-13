GOOSE=go run github.com/pressly/goose/v3/cmd/goose

.PHONY: build test lint generate migrate-up migrate-down migrate-status

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

generate:
	go generate ./...

migrate-up:
	$(GOOSE) -dir ./migrations sqlite3 ./data.db up

migrate-down:
	$(GOOSE) -dir ./migrations sqlite3 ./data.db down

migrate-status:
	$(GOOSE) -dir ./migrations sqlite3 ./data.db status
