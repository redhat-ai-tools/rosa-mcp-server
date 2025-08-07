# Context Findings

## Current Codebase State
- **Status:** Fresh project - no existing Go code
- **Available files:**
  - `initial_requirements.md` - Comprehensive requirements document
  - `example_cluster_create_request.json` - Example OCM API request for ROSA HCP cluster creation

## Reference Implementation Analysis

### Python OCM MCP Server (ocm_mcp_server.py)
**Key Patterns:**
- Uses FastMCP framework with `@mcp.tool()` decorators
- Token handling: Environment variable for stdio, HTTP header `X-OCM-Offline-Token` for SSE
- OAuth refresh flow: POST to SSO with `refresh_token` grant type
- Response formatting: Dedicated `format_clusters_response()` functions
- Error handling: Direct `response.raise_for_status()` without retry logic
- Transport detection: `MCP_TRANSPORT` environment variable

### Go OpenShift MCP Server Patterns  
**Architecture:**
- Uses `github.com/mark3labs/mcp-go` framework version 0.37.0
- Server struct with configuration and kubernetes manager
- Profile-based tool registration system
- Middleware support for authentication and logging
- Both stdio and SSE transport support with context functions

**Authentication:**
- OAuth header support: `Authorization` and custom fallback headers
- Context-based token passing
- Middleware for scoped authorization

**Transport Support:**
- `ServeStdio()` for stdio transport
- `ServeSse()` for SSE with base URL and HTTP server options
- Context functions for header extraction

## Technical Architecture Plan
- **Framework:** `github.com/mark3labs/mcp-go`
- **Structure:** Follow OpenShift MCP server patterns (Server struct, Configuration, Profile)
- **Authentication:** OCM offline token → OAuth access token flow (like Python reference)
- **Transport:** Both stdio and SSE with `X-OCM-OFFLINE-TOKEN` header support
- **API Client:** HTTP client for OCM API calls
- **Formatting:** Dedicated formatter functions for human-readable output

## Implementation Patterns Required
- Server struct with OCM client manager
- Configuration struct for settings
- Profile system for tool registration  
- Context-based token handling for SSE transport
- Direct API error propagation without retry logic
- Formatter functions returning strings (not JSON)