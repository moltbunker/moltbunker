.PHONY: build daemon cli test clean install

build: daemon cli

daemon:
	@echo "Building daemon..."
	@go build -o bin/moltbunker-daemon ./cmd/daemon

cli:
	@echo "Building CLI..."
	@go build -o bin/moltbunker ./cmd/cli

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/

install: build
	@echo "Installing binaries..."
	@cp bin/moltbunker-daemon /usr/local/bin/
	@cp bin/moltbunker /usr/local/bin/
