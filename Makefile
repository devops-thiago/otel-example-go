TEST_PKGS := ./...

.PHONY: test cover coverhtml lint fmt fmt-check vet trim-whitespace

test:
	go test $(TEST_PKGS) -count=1

cover:
	go test $(TEST_PKGS) -coverpkg=./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -n 1

coverhtml: cover
	go tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

fmt:
	@echo "Running gofmt..."
	gofmt -s -w .

fmt-check:
	@echo "Checking formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "The following files need formatting:"; \
		gofmt -s -l .; \
		exit 1; \
	fi

vet:
	@echo "Running go vet..."
	go vet ./...

trim-whitespace:
	@echo "Removing trailing whitespaces..."
	@find . -type f \( -name "*.go" -o -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" -o -name "*.sh" -o -name "Makefile" -o -name "Dockerfile" \) ! -path "./vendor/*" ! -path "./.git/*" -exec sed -i 's/[[:space:]]*$$//' {} +
	@echo "Trailing whitespaces removed"


