.PHONY: build run clean

# Build the rosa-mcp-server binary
build:
	go build -o rosa-mcp-server ./cmd/rosa-mcp-server

# Build and run the server with stdio transport
run: build
	./rosa-mcp-server --transport=stdio

# Clean build artifacts
clean:
	rm -f rosa-mcp-server
