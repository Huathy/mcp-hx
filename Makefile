.PHONY: build test run clean

build:
	go build -o bin/mcp-dbx ./cmd/mcp-dbx

run: build
	./bin/mcp-dbx --config ./.kilo/mcp_x.yaml

test:
	go test ./... -v

clean:
	rm -rf bin/
