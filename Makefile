.PHONY: build test run clean

build:
	go build -o bin/mcp-dbx ./cmd/mcp-dbx

run: build
	./bin/mcp-dbx --config ./examples/mcp-dbx.yaml.example

test:
	go test ./... -v

clean:
	rm -rf bin/
