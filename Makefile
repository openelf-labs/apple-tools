VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build install test clean

build:
	CGO_ENABLED=0 GOOS=darwin go build -ldflags "-X main.version=$(VERSION)" -o apple-tools ./cmd/apple-tools

install: build
	mkdir -p ~/.openelf/bin
	cp apple-tools ~/.openelf/bin/apple-tools
	@echo "Installed to ~/.openelf/bin/apple-tools"

test:
	go test ./...

clean:
	rm -f apple-tools
