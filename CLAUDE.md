# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server for ROSA HCP (Red Hat OpenShift Service on AWS using Hosted Control Planes) written in Go. It enables AI assistants to integrate with Red Hat Managed OpenShift services through 4 core tools: `whoami`, `get_clusters`, `get_cluster`, and `create_rosa_hcp_cluster`.

## Build and Development Commands

```bash
# Build the server binary
make build
# or: go build -o rosa-mcp-server ./cmd/rosa-mcp-server

# Build and run with stdio transport (for local testing)
make run

# Clean build artifacts
make clean

# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for specific package
go test ./pkg/ocm -v

# Test the built server
./rosa-mcp-server --help
./rosa-mcp-server version
```

## Architecture Overview

The codebase follows the Model Context Protocol (MCP) server pattern with clear separation of concerns:

### Core Components

**MCP Layer (`pkg/mcp/`):**
- `server.go` - Main MCP server implementation, handles transport (stdio/SSE) and authentication
- `tools.go` - Implements all 4 ROSA HCP tools with parameter validation and OCM client interaction
- `formatters.go` - Human-readable response formatters (not JSON) for AI assistant consumption
- `profiles.go` - Tool profile system for selective tool exposure (currently default profile only)

**OCM Integration (`pkg/ocm/`):**
- `client.go` - OCM SDK wrapper with authenticated connections and structured error handling
- `auth.go` - Transport-agnostic token extraction (SSE headers vs stdio environment variables)

**Configuration (`pkg/config/`):**
- `config.go` - TOML configuration file support with CLI flag overrides

### Authentication Flow

The server supports dual transport authentication:

1. **Stdio Transport**: Uses `OCM_OFFLINE_TOKEN` environment variable
2. **SSE Transport**: Extracts token from `X-OCM-OFFLINE-TOKEN` HTTP header

Token extraction is handled by `ocm.ExtractTokenFromContext()` which determines the appropriate method based on transport mode. The OCM SDK handles OAuth refresh flow automatically.

### Error Handling Pattern

All OCM API errors are preserved and exposed directly to users without modification through the `OCMError` type, which extracts structured error details (code, reason, operation ID) from the OCM SDK.

### Tool Implementation Pattern

Each tool follows this pattern:
1. Extract and validate parameters from MCP request
2. Get authenticated OCM client via `getAuthenticatedOCMClient()`
3. Call OCM client method
4. Format response using dedicated formatter
5. Return `NewTextResult()` with formatted text or error

### Response Formatting

All responses are human-readable formatted strings (not JSON) designed for AI assistant consumption. Each tool has a dedicated formatter in `formatters.go` that structures the output consistently.

## Development Notes

- The project uses `github.com/mark3labs/mcp-go` v0.37.0+ as the MCP framework
- OCM integration via `github.com/openshift-online/ocm-sdk-go`
- Configuration supports CLI flags, environment variables, and TOML files
- Only the `pkg/ocm` package currently has test coverage focusing on authentication logic
- The server binary is built to `rosa-mcp-server` in the project root

## OCM Authentication Setup

For development and testing, you need an OCM offline token:

1. Visit [console.redhat.com/openshift/token](https://console.redhat.com/openshift/token)
2. Copy the offline token
3. For stdio: `export OCM_OFFLINE_TOKEN="your-token"`
4. For SSE: Include `X-OCM-OFFLINE-TOKEN` header in requests

The server supports concurrent users via header-based authentication in SSE mode.