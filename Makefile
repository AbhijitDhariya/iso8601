.PHONY: test lint lint-fix fmt vet

# Run tests
test:
	go test -v -cover ./...

# Run linter
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Run linter and auto-fix issues
lint-fix:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

# Format code
fmt:
	go fmt ./...
	gofmt -w .

# Run go vet
vet:
	go vet ./...

# Run all checks
check: fmt vet lint test

