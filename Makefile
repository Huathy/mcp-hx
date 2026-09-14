.PHONY: build test run clean

build:
	go build -o bin/mcp-x ./cmd/mcp-x

run: build
	./bin/mcp-x --config ./.kilo/mcp-x.yaml

test:
	go test ./... -v

clean:
	rm -rf bin/
