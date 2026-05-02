.PHONY: build run test test-fast test-ci lint clean help lint

BINARY_NAME=avito-shop
BINARY_PATH=bin/$(BINARY_NAME)

build:
	go build -v -o $(BINARY_PATH) ./cmd/

run: build
	$(BINARY_PATH)

test-fast:
	go test -v -short ./...

test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-ci:
	go test -v ./...

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

lint:
	golangci-lint run