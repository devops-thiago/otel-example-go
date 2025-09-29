TEST_PKGS := ./...

.PHONY: test cover coverhtml lint

test:
	go test $(TEST_PKGS) -count=1

cover:
	go test $(TEST_PKGS) -coverpkg=./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -n 1

coverhtml: cover
	go tool cover -html=coverage.out -o coverage.html


