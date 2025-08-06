# Milestone 4: Integration and Final Testing

Generated: 2025-08-06T13:21:00Z
Review status: Started
Status: Not Started
Current Phase: 1

## Overview
Complete the integration by connecting all components, implement basic testing, validate both transport modes, and ensure the server meets all acceptance criteria. This milestone delivers the complete MVP implementation ready for use by AI assistants.

## Guidance

### rosa-mcp-go contribution requirements
- Wire all components together in main.go
- Support both stdio and SSE transports with authentication
- Minimal testing focused on server startup and tool registration
- Validate against all acceptance criteria from requirements
- No parameter validation - pass directly to OCM API

### rosa-mcp-go testing / validation strategy
- Test server starts successfully with both transports
- Verify tools are registered and accessible
- Test authentication handling for both transport modes
- Run: `go test ./...` for complete test suite
- Manual validation with MCP client if possible

### Rules
- **CRITICAL** Ensure the plan is updated to reflect step completion before starting the next step

## Phases

### Phase 1: Component Integration

Review status: Approved

Wire together all components and complete the main.go implementation.

**Full Integration**:
- [ ] (1) Complete cmd/rosa-mcp-server/main.go integration:
```go
func main() {
    // Parse CLI flags and config
    // Initialize configuration
    // Create MCP server with OCM client
    // Start appropriate transport (stdio/SSE)
}
```
- [ ] (2) Add TOML configuration file support using BurntSushi/toml
- [ ] (3) Implement proper shutdown handling and cleanup
- [ ] (4) Add version information and build metadata
- [ ] (5) Validate the changes via typical project testing procedures.
        Run: `go run cmd/rosa-mcp-server/main.go --help`
- [ ] (6) Make a commit including only the changed files.
- [ ] (7) Update the implementation plan to reflect phase completion.

### Phase 2: Transport Validation

Review status: Approved

Test both stdio and SSE transport modes with authentication handling following OpenShift MCP testing patterns.

**Transport Testing** (following pkg/mcp/mcp_test.go patterns):
- [ ] (8) Create pkg/mcp/server_test.go following OpenShift MCP patterns:
```go
func TestOCMHeaders(t *testing.T) {
    // Mock OCM server similar to OpenShift's mock Kubernetes server
    mockOCMServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        // Capture headers and verify Bearer token from OCM offline token
        // Return mock OCM API responses for current_account, clusters, etc.
    }))
    defer mockOCMServer.Close()
    
    // Test that X-OCM-OFFLINE-TOKEN header is properly converted to Bearer token
    // Similar to TestSseHeaders in OpenShift MCP server
}

func TestStdioTokenHandling(t *testing.T) {
    // Test OCM_OFFLINE_TOKEN environment variable handling
    // Verify token extraction and Bearer token generation
}
```
- [ ] (9) Test OCM offline token to Bearer token conversion:
  - Mock OCM token endpoint responses
  - Verify X-OCM-OFFLINE-TOKEN header handling (SSE transport)
  - Verify OCM_OFFLINE_TOKEN environment variable handling (stdio transport)
  - Test authentication error scenarios
- [ ] (10) Test tool registration and basic functionality:
  - Verify all 4 tools are registered correctly
  - Test basic tool execution with mock OCM responses
  - Follow testCase pattern from OpenShift MCP's common_test.go
- [ ] (11) Manual validation tests:
  - Run `go run cmd/rosa-mcp-server/main.go --transport=stdio` locally
  - Run `go run cmd/rosa-mcp-server/main.go --transport=sse --port=8080` locally
  - Test with basic MCP client calls or curl for SSE mode
- [ ] (12) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/mcp/...` and manual tests with both transport modes
- [ ] (13) Make a commit including only the changed files.
- [ ] (14) Update the implementation plan to reflect phase completion.

### Phase 3: Acceptance Criteria Validation

Review status: Approved

Verify all acceptance criteria from requirements specification are met.

**Final Validation**:
- [ ] (15) Verify Transport Support: Both stdio and SSE work with token authentication ✓
- [ ] (16) Verify Tool Functionality: All 4 tools return properly formatted responses ✓
- [ ] (17) Verify Error Handling: OCM API errors are exposed without modification ✓
- [ ] (18) Verify Multi-Region: Cluster creation supports configurable AWS regions ✓
- [ ] (19) Verify Authentication: Multiple concurrent users supported via header tokens ✓
- [ ] (20) Verify Documentation: Prerequisites resource available to clients ✓
- [ ] (21) Verify Configuration: Configurable via CLI, environment, and config files ✓
- [ ] (22) Create basic README.md with:
  - Installation instructions
  - Configuration examples
  - Usage examples for both transports
  - Authentication setup
- [ ] (23) Validate the changes via typical project testing procedures.
        Run: `go test ./...` and manual verification
- [ ] (24) Make a commit including only the changed files.
- [ ] (25) Update the implementation plan to reflect phase completion.

### Phase 4: Documentation and Cleanup

Review status: Approved

Final documentation and code cleanup for MVP release.

**Final Documentation**:
- [ ] (26) Add comprehensive code comments where needed
- [ ] (27) Update static ROSA HCP prerequisites resource with accurate information from reference documentation:
  - Reference: https://cloud.redhat.com/learning/learn:getting-started-red-hat-openshift-service-aws-rosa/resource/resources:creating-rosa-hcp-clusters-using-default-options
  - Reference: https://www.rosaworkshop.io/rosa/18-deploy_hcp/
  - Include: IAM role setup, OIDC configuration, networking prerequisites
  - Include: AWS account requirements and subnet configuration
  - Include: Example ARN formats and role naming conventions
- [ ] (28) Create example configuration files (TOML)
- [ ] (29) Validate the changes via typical project testing procedures.
        Run: `go test ./...` and `go build ./...`
- [ ] (30) Make a commit including only the changed files.
- [ ] (31) Update the implementation plan to reflect phase completion.

## TODOs
In this section, list TODOs required to be followed up on before moving onto the next milestone
- [ ] Test with real OCM offline token if available
- [ ] Verify tool responses are useful for AI assistants
- [ ] Confirm cluster creation parameters work with OCM API
- [ ] Document any known limitations or future enhancements needed