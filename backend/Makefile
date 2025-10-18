APP_NAME=zamanbek-backend
GOFILES=$(shell find . -name '*.go' -not -path "./vendor/*")

.PHONY: run test lint fmt vet tidy

run:
	@echo "→ starting $(APP_NAME) on :$${PORT:-8080}"
	@PORT=$${PORT:-8080} go run .

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

tidy:
	go mod tidy
