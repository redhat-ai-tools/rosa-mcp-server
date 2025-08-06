# Milestone 1: Foundation Setup

Generated: 2025-08-06T13:21:00Z
Review status: Started
Status: Not Started
Current Phase: 1

## Overview
Establish the foundational structure for the ROSA HCP MCP server including Go module initialization, directory structure following the requirements specification, dependency management, and basic configuration framework. This milestone creates the skeleton that all subsequent implementation will build upon.

## Guidance

### rosa-mcp-go contribution requirements
- Initialize all new code (do not fork from reference implementations)
- Use OpenShift MCP server as **REFERENCE MATERIAL ONLY**
- Follow file structure outlined in requirements specification
- Use OCM SDK as specified in requirements
- Keep implementation simple with minimal validation
- Simple string template formatting for responses

### rosa-mcp-go testing / validation strategy
- Minimal testing - just ensure the server starts and tools are registered
- Rely on ocm-sdk-go SDK testing for OCM API integration confidence
- Run: `go test ./...` to validate basic functionality

### Rules
- **CRITICAL** Ensure the plan is updated to reflect step completion before starting the next step

## Phases

### Phase 1: Go Module and Project Structure

Review status: Approved

Initialize the Go module and create the directory structure specified in requirements.

**Go Module Initialization**:
- [ ] (1) Initialize Go module with `go mod init github.com/redhat-ai-tools/rosa-mcp-go`
- [ ] (2) Create directory structure following requirements spec:
```
cmd/rosa-mcp-server/main.go
pkg/config/config.go
pkg/ocm/client.go
pkg/ocm/auth.go
pkg/mcp/server.go
pkg/mcp/tools.go
pkg/mcp/profiles.go
pkg/mcp/formatters.go
pkg/version/version.go
```
- [ ] (3) Validate the changes via typical project testing procedures.
        Run: `go mod tidy && go build ./...`
- [ ] (4) Make a commit including only the changed files.
- [ ] (5) Update the implementation plan to reflect phase completion.

### Phase 2: Core Dependencies

Review status: Approved

Add the critical dependencies specified in requirements and establish basic configuration.

**Dependency Management**:
- [ ] (6) Add core dependencies to go.mod:
```go
github.com/mark3labs/mcp-go v0.37.0
github.com/openshift-online/ocm-sdk-go
github.com/BurntSushi/toml
github.com/spf13/cobra
```
- [ ] (7) Create basic configuration struct in pkg/config/config.go supporting:
  - OCM API base URL (default: https://api.openshift.com)
  - Transport mode selection (stdio/SSE)
  - Optional SSE port and base URL
- [ ] (8) Validate the changes via typical project testing procedures.
        Run: `go mod tidy && go build ./...`
- [ ] (9) Make a commit including only the changed files.
- [ ] (10) Update the implementation plan to reflect phase completion.

### Phase 3: Basic CLI Entry Point

Review status: Approved

Create the main.go entry point with cobra CLI framework and basic transport configuration.

**CLI Framework Setup**:
- [ ] (11) Implement cmd/rosa-mcp-server/main.go with cobra CLI
- [ ] (12) Add command-line flags for:
  - `--transport` (stdio/sse)
  - `--ocm-base-url` (default: https://api.openshift.com)
  - `--port` (for SSE transport)
  - `--config` (TOML config file path)
- [ ] (13) Create basic version command in pkg/version/version.go
- [ ] (14) Validate the changes via typical project testing procedures.
        Run: `go run cmd/rosa-mcp-server/main.go --help`
- [ ] (15) Make a commit including only the changed files.
- [ ] (16) Update the implementation plan to reflect phase completion.

## TODOs
In this section, list TODOs required to be followed up on before moving onto the next milestone
- [ ] Verify mcp-go v0.37.0+ compatibility
- [ ] Confirm OCM SDK latest version compatibility
- [ ] Test basic CLI help and version commands work