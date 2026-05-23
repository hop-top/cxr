.PHONY: build test lint check clean fmt vet

build:
	go build ./...

test:
	go test -race -count=1 ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: vet
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "gofmt needed:"; gofmt -l .; exit 1; \
	fi

check: lint test

clean:
	go clean
