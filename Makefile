.PHONY: build run clean

# Build the rosa-mcp-go binary
build:
	go build -o rosa-mcp-go ./cmd/rosa-mcp-server

# Build and run the server with stdio transport
run: build
	./rosa-mcp-go --transport=stdio

# Clean build artifacts
clean:
	rm -f rosa-mcp-go