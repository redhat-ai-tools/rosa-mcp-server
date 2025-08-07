# ROSA HCP MCP Server - Comprehensive Requirements Specification

## Problem Statement
Build a golang-based MCP server for ROSA HCP (Red Hat OpenShift on AWS using Hosted Control Planes) that enables AI assistants to integrate with Red Hat Managed OpenShift services. The server provides essential cluster operations through the Model Context Protocol.

## Solution Overview
A minimal MVP implementation following the architectural patterns of the OpenShift MCP server (github.com/openshift/openshift-mcp-server) and business logic of the OCM MCP server (github.com/redhat-ai-tools/ocm-mcp), using the mcp-go framework (github.com/mark3labs/mcp-go).

## Functional Requirements

### Core Tools (4 Required)

#### 1. `whoami()`
- **Description:** "Get the authenticated account"
- **Parameters:** None
- **Returns:** Formatted account information (username, organization, etc.)
- **Implementation:** GET `/api/accounts_mgmt/v1/current_account`

#### 2. `get_clusters(state: string)`
- **Description:** "Retrieves the list of clusters"
- **Parameters:**
  - `state` (required): Filter clusters by state (e.g., "ready", "installing", "error")
- **Returns:** Formatted cluster list with name, ID, API URL, console URL, state, version
- **Implementation:** GET `/api/clusters_mgmt/v1/clusters` with state filtering

#### 3. `get_cluster(cluster_id: string)`
- **Description:** "Retrieves the details of the cluster"
- **Parameters:**
  - `cluster_id` (required): Unique cluster identifier
- **Returns:** Formatted detailed cluster information
- **Implementation:** GET `/api/clusters_mgmt/v1/clusters/{cluster_id}`

#### 4. `create_rosa_hcp_cluster(...)`
- **Description:** "Provision a new ROSA HCP cluster with required configuration"
- **Parameters (all required except region):**
  - `cluster_name` (required): Name for the cluster
  - `aws_account_id` (required): AWS account ID
  - `billing_account_id` (required): AWS billing account ID
  - `role_arn` (required): IAM installer role ARN
  - `operator_role_prefix` (required): Operator role prefix
  - `oidc_config_id` (required): OIDC configuration ID
  - `subnet_ids` (required): Array of subnet IDs
  - `region` (optional): AWS region (default: "us-east-1")
- **Returns:** Formatted cluster creation response
- **Implementation:** POST `/api/clusters_mgmt/v1/clusters` with ROSA HCP payload

### Resources (1 Required)

#### 1. ROSA HCP Prerequisites Documentation
- **Type:** Static resource
- **Content:** Documentation of cluster creation requirements
- **Includes:** IAM role setup, OIDC configuration, networking prerequisites

## Technical Requirements

### Architecture
- **Framework:** `github.com/mark3labs/mcp-go` v0.37.0+
- **Structure:** Follow OpenShift MCP server patterns
  - `Server` struct with OCM client manager
  - `Configuration` struct for settings
  - Profile system (expose all tools for MVP)
  - Context-based authentication handling

### Authentication
- **Token Source:** OCM offline tokens from https://console.redhat.com/openshift/token
- **Flow:** OAuth refresh token → access token
- **Transport Handling:**
  - **SSE Transport:** `X-OCM-OFFLINE-TOKEN` header
  - **Stdio Transport:** `OCM_OFFLINE_TOKEN` environment variable
- **Multiple Users:** Support concurrent tokens via header-based authentication

### Transport Support
- **Stdio:** Standard input/output for local integrations
- **SSE:** Server-Sent Events with HTTP headers for remote access
- **Configuration:** Command-line flags and TOML config files

### API Integration
- **SDK:** Use `github.com/openshift-online/ocm-sdk-go` for OCM API client
- **Base URL:** https://api.openshift.com (configurable)
- **Authentication:** Bearer token via OAuth refresh flow (handled by OCM SDK)
- **Endpoints:** 
  - OCM Clusters_mgmt API: `/api/clusters_mgmt/v1/`
  - OCM Accounts_mgmt API: `/api/accounts_mgmt/v1/`
- **Error Handling:** Direct API error propagation without retry logic

### Response Formatting
- **Output:** Human-readable formatted strings (not JSON)
- **Patterns:** Dedicated formatter functions like `formatClustersResponse()`
- **Consistency:** All tools return formatted text for AI assistant consumption

### Configuration
- **Methods:** Command-line flags, environment variables, TOML config files
- **Required Settings:**
  - OCM API base URL (default: https://api.openshift.com)
  - Transport mode selection
  - Optional: Port for SSE/HTTP transport
  - Optional: SSE base URL for public endpoints

## Implementation Hints

### File Structure (following OpenShift MCP patterns)
```
cmd/rosa-mcp-server/main.go          # Entry point
pkg/
  config/config.go                   # Configuration management
  ocm/
    client.go                        # OCM SDK client wrapper
    auth.go                          # OAuth token handling (via OCM SDK)
    formatter.go                     # Response formatting
  mcp/
    server.go                        # MCP server implementation
    tools.go                         # Tool implementations
    profiles.go                      # Profile definitions
  version/version.go                 # Version information
```

### Key Patterns to Follow
1. **Server Initialization:** Similar to `mcp.NewServer()` in OpenShift MCP
2. **Context Handling:** Extract tokens from headers/environment in context functions
3. **Tool Registration:** Register all 4 tools with the mcp-go framework
4. **Error Wrapping:** Use `NewTextResult()` pattern for consistent error responses
5. **Middleware:** Logging middleware for tool calls

### Critical Dependencies
- `github.com/mark3labs/mcp-go` - MCP framework
- `github.com/openshift-online/ocm-sdk-go` - OCM API client SDK
- `github.com/BurntSushi/toml` - Configuration files
- `github.com/spf13/cobra` - CLI interface

## Acceptance Criteria
1. **Transport Support:** Both stdio and SSE work with token authentication
2. **Tool Functionality:** All 4 tools return properly formatted responses
3. **Error Handling:** OCM API errors are exposed to users without modification
4. **Multi-Region:** Cluster creation supports configurable AWS regions
5. **Authentication:** Multiple concurrent users supported via header tokens
6. **Documentation:** Prerequisites resource available to clients
7. **Configuration:** Configurable via CLI, environment, and config files

## Assumptions
- Users have pre-configured AWS IAM roles and OIDC configurations
- OCM API handles all parameter validation for cluster creation
- No caching required - real-time API data preferred
- MVP scope excludes cluster lifecycle management beyond creation
- Unit testing sufficient for MVP (integration tests future enhancement)
- Profile system will be simple (all tools enabled) with future configurability

## Future Considerations (Do not implement in MVP)
- Tool selection profiles (rosa-basic, rosa-full)
- Pre-validation of AWS resources before API calls
- Cluster lifecycle operations (scaling, upgrading, deletion)
- Caching for performance optimization
- Enhanced error recovery and retry logic
