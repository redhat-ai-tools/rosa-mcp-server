# Milestone 2: OCM API Integration

Generated: 2025-08-06T13:21:00Z
Review status: Started
Status: Not Started
Current Phase: 1

## Overview
Implement OCM SDK integration with authentication handling, API client wrapper, and response formatting functions. This milestone establishes the core business logic for interacting with the OCM API using the official SDK, following the authentication patterns from the reference implementations.

## Guidance

### rosa-mcp-go contribution requirements
- Use OCM SDK for all API interactions (no direct HTTP calls)
- Support both stdio and SSE transport authentication patterns
- Implement simple string template response formatting
- Pass parameters directly to OCM API without validation
- Direct API error propagation without retry logic

### rosa-mcp-go testing / validation strategy
- Minimal testing - just ensure client initialization works
- Test authentication token extraction from headers/environment
- Run: `go test ./pkg/ocm/...` to validate core functionality

### Rules
- **CRITICAL** Ensure the plan is updated to reflect step completion before starting the next step

## Phases

### Phase 1: OCM Client and Authentication

Review status: Approved

Implement OCM SDK client wrapper with OAuth token handling for both transport modes.

**OCM SDK Integration**:
- [ ] (1) Implement pkg/ocm/client.go with OCM SDK client wrapper
```go
type Client struct {
    connection *sdk.Connection
    baseURL    string
}
func NewClient(baseURL string) *Client
func (c *Client) WithToken(token string) *Client
```
- [ ] (2) Implement pkg/ocm/auth.go with token extraction logic:
  - SSE transport: Extract from `X-OCM-OFFLINE-TOKEN` header
  - Stdio transport: Read from `OCM_OFFLINE_TOKEN` environment variable
  - OAuth refresh token → access token flow using OCM SDK
- [ ] (3) Add context-based authentication handling following reference patterns
- [ ] (4) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/ocm/...`
- [ ] (5) Make a commit including only the changed files.
- [ ] (6) Update the implementation plan to reflect phase completion.

### Phase 2: OCM API Error Handling

Review status: Approved

Implement proper OCM API error handling without response formatting (formatting will be done in MCP layer).

**Error Handling Implementation**:
- [ ] (7) Implement OCM API error handling in pkg/ocm/client.go:
  - Extract error code and reason from OCM API responses
  - Create error types that preserve OCM error details
  - Ensure errors bubble up with original OCM error information
- [ ] (8) Update client wrapper to return raw OCM SDK types without formatting
- [ ] (9) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/ocm/...`
- [ ] (10) Make a commit including only the changed files.
- [ ] (11) Update the implementation plan to reflect phase completion.

### Phase 3: API Wrapper Functions

Review status: Approved

Create wrapper functions for the 4 core OCM API operations required by the tools.

**API Operation Wrappers**:
- [ ] (12) Implement GetCurrentAccount() returning raw *accountsmgmt.Account
- [ ] (13) Implement GetClusters(state string) returning raw []*clustersmgmt.Cluster with state filtering
- [ ] (14) Implement GetCluster(clusterID string) returning raw *clustersmgmt.Cluster
- [ ] (15) Implement CreateROSAHCPCluster() returning raw *clustersmgmt.Cluster:
```go
func CreateROSAHCPCluster(
    clusterName, awsAccountID, billingAccountID, roleArn, 
    operatorRolePrefix, oidcConfigID string, 
    subnetIDs []string, region string
) (*clustersmgmt.Cluster, error)
```
- [ ] (16) Use cluster creation payload structure from initial_requirements.md example
- [ ] (17) All functions return raw OCM SDK types without any formatting
- [ ] (18) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/ocm/...`
- [ ] (19) Make a commit including only the changed files.
- [ ] (20) Update the implementation plan to reflect phase completion.

## TODOs
In this section, list TODOs required to be followed up on before moving onto the next milestone
- [ ] Verify OCM SDK authentication flow works with offline tokens
- [ ] Confirm cluster creation payload matches OCM API expectations
- [ ] Ensure raw OCM SDK types are properly returned without formatting
