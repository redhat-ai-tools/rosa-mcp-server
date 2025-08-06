# Milestone 3: MCP Server Implementation

Generated: 2025-08-06T13:21:00Z
Review status: Started
Status: Not Started
Current Phase: 1

## Overview
Implement the Model Context Protocol server using mcp-go framework, register the 4 required tools, implement the static resource, and establish the profile system. This milestone creates the MCP interface that AI assistants will interact with.

## Guidance

### rosa-mcp-go contribution requirements
- Use mcp-go framework v0.37.0+ for all MCP functionality
- Follow OpenShift MCP server architectural patterns (reference only)
- Register all 4 tools with proper descriptions and parameter schemas
- Implement simple profile system (all tools enabled for MVP)
- Add static ROSA HCP prerequisites documentation resource

### rosa-mcp-go testing / validation strategy
- Test that MCP server starts successfully
- Verify all tools are registered and discoverable
- Test static resource is accessible
- Run: `go test ./pkg/mcp/...`

### Rules
- **CRITICAL** Ensure the plan is updated to reflect step completion before starting the next step

## Phases

### Phase 1: MCP Server Framework

Review status: Approved

Implement the core MCP server using mcp-go framework with transport support.

**MCP Server Implementation**:
- [ ] (1) Implement pkg/mcp/server.go with mcp-go framework:
```go
type Server struct {
    mcpServer *mcp.Server
    ocmClient *ocm.Client
    config    *config.Configuration
}
func NewServer(cfg *config.Configuration) *Server
func (s *Server) Start() error
```
- [ ] (2) Add support for both stdio and SSE transports based on configuration
- [ ] (3) Implement context-based OCM client initialization with authentication
- [ ] (4) Add logging middleware for tool calls following OpenShift MCP patterns:
```go
// Add structured logging for tool execution
func (s *Server) logToolCall(toolName string, params map[string]interface{}) {
    log.Printf("Tool called: %s with params: %v", toolName, params)
}

// Wrap tool handlers with logging middleware
func (s *Server) wrapToolHandler(toolName string, handler func(...) mcp.ToolResult) func(...) mcp.ToolResult {
    return func(params ...interface{}) mcp.ToolResult {
        s.logToolCall(toolName, convertParamsToMap(params))
        result := handler(params...)
        if result.IsError {
            log.Printf("Tool %s failed: %s", toolName, result.Content)
        } else {
            log.Printf("Tool %s completed successfully", toolName)
        }
        return result
    }
}
```
- [ ] (5) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/mcp/...`
- [ ] (6) Make a commit including only the changed files.
- [ ] (7) Update the implementation plan to reflect phase completion.

### Phase 2: Response Formatters

Review status: Approved

Implement simple string template formatters for OCM API responses in the MCP layer.

**Formatter Implementation**:
- [ ] (8) Implement pkg/mcp/formatters.go with functions:
```go
func formatAccountResponse(account *accountsmgmt.Account) string
func formatClustersResponse(clusters []*clustersmgmt.Cluster) string
func formatClusterResponse(cluster *clustersmgmt.Cluster) string
func formatClusterCreateResponse(cluster *clustersmgmt.Cluster) string
```
- [ ] (9) Use simple string templates showing key information:
  - Account: username, organization, email
  - Clusters: name, ID, state, API URL, console URL, version
  - Cluster details: comprehensive cluster information
  - Create response: cluster name, ID, state, creation status
- [ ] (10) Handle OCM API error responses with code and reason exposure
- [ ] (11) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/mcp/...`
- [ ] (12) Make a commit including only the changed files.
- [ ] (13) Update the implementation plan to reflect phase completion.

### Phase 3: Basic Tool Implementations

Review status: Approved

Implement the 3 basic MCP tools (whoami, get_clusters, get_cluster) with proper parameter schemas and error handling.

**Tool Registration and Implementation**:
- [ ] (14) Implement pkg/mcp/tools.go with tool registration:
```go
func (s *Server) registerTools()
func (s *Server) handleWhoami() mcp.ToolResult
func (s *Server) handleGetClusters(state string) mcp.ToolResult  
func (s *Server) handleGetCluster(clusterID string) mcp.ToolResult
```
- [ ] (15) Register whoami tool:
  - Description: "Get the authenticated account"
  - Parameters: None
  - Returns: Formatted account information using formatAccountResponse()
- [ ] (16) Register get_clusters tool:
  - Description: "Retrieves the list of clusters"
  - Parameters: state (required string)
  - Returns: Formatted cluster list using formatClustersResponse()
- [ ] (17) Register get_cluster tool with detailed implementation:
  - Description: "Retrieves the details of the cluster"  
  - Parameters: cluster_id (required string)
  - Returns: Formatted cluster details using formatClusterResponse()
```go
func (s *Server) handleGetCluster(clusterID string) mcp.ToolResult {
    // Extract OCM client from context with authentication
    client, err := s.getAuthenticatedOCMClient()
    if err != nil {
        return mcp.NewTextResult(fmt.Sprintf("Authentication failed: %s", err.Error()))
    }
    
    // Call OCM client to get cluster details
    cluster, err := client.GetCluster(clusterID)
    if err != nil {
        // Handle OCM API errors with code and reason exposure
        if ocmErr, ok := err.(*ocm.OCMError); ok {
            return mcp.NewTextResult(fmt.Sprintf("OCM API Error [%s]: %s", ocmErr.Code, ocmErr.Reason))
        }
        return mcp.NewTextResult(fmt.Sprintf("Failed to get cluster: %s", err.Error()))
    }
    
    // Format response using MCP layer formatter
    formattedResponse := formatClusterResponse(cluster)
    return mcp.NewTextResult(formattedResponse)
}
```
- [ ] (18) Implement error handling using NewTextResult() pattern from OpenShift MCP
- [ ] (19) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/mcp/...`
- [ ] (20) Make a commit including only the changed files.
- [ ] (21) Update the implementation plan to reflect phase completion.

### Phase 4: ROSA HCP Cluster Creation Tool

Review status: Approved

Implement the complex create_rosa_hcp_cluster tool with all required parameters and ROSA HCP payload structure.

**ROSA HCP Cluster Creation Implementation**:
- [ ] (22) Register create_rosa_hcp_cluster tool with complete parameter schema:
```go
// Tool registration with all required parameters
func (s *Server) registerCreateROSAHCPClusterTool() {
    s.mcpServer.RegisterTool("create_rosa_hcp_cluster", mcp.ToolDefinition{
        Description: "Provision a new ROSA HCP cluster with required configuration",
        Parameters: map[string]mcp.ParameterSchema{
            "cluster_name":         {Type: "string", Required: true, Description: "Name for the cluster"},
            "aws_account_id":       {Type: "string", Required: true, Description: "AWS account ID"},
            "billing_account_id":   {Type: "string", Required: true, Description: "AWS billing account ID"},
            "role_arn":            {Type: "string", Required: true, Description: "IAM installer role ARN"},
            "operator_role_prefix": {Type: "string", Required: true, Description: "Operator role prefix"},
            "oidc_config_id":      {Type: "string", Required: true, Description: "OIDC configuration ID"},
            "subnet_ids":          {Type: "array", Items: "string", Required: true, Description: "Array of subnet IDs"},
            "region":              {Type: "string", Required: false, Default: "us-east-1", Description: "AWS region"},
        },
        Handler: s.handleCreateROSAHCPCluster,
    })
}
```
- [ ] (23) Implement handleCreateROSAHCPCluster following OpenShift MCP patterns:
```go
func (s *Server) handleCreateROSAHCPCluster(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Extract required parameters with safe type assertions (following OpenShift MCP pattern)
    args := ctr.GetArguments()
    
    clusterName, ok := args["cluster_name"].(string)
    if !ok || clusterName == "" {
        return NewTextResult("", errors.New("missing required argument: cluster_name")), nil
    }
    
    awsAccountID, ok := args["aws_account_id"].(string)
    if !ok || awsAccountID == "" {
        return NewTextResult("", errors.New("missing required argument: aws_account_id")), nil
    }
    
    billingAccountID, ok := args["billing_account_id"].(string)
    if !ok || billingAccountID == "" {
        return NewTextResult("", errors.New("missing required argument: billing_account_id")), nil
    }
    
    roleArn, ok := args["role_arn"].(string)
    if !ok || roleArn == "" {
        return NewTextResult("", errors.New("missing required argument: role_arn")), nil
    }
    
    operatorRolePrefix, ok := args["operator_role_prefix"].(string)
    if !ok || operatorRolePrefix == "" {
        return NewTextResult("", errors.New("missing required argument: operator_role_prefix")), nil
    }
    
    oidcConfigID, ok := args["oidc_config_id"].(string)
    if !ok || oidcConfigID == "" {
        return NewTextResult("", errors.New("missing required argument: oidc_config_id")), nil
    }
    
    // Handle subnet_ids array parameter
    subnetIDs := make([]string, 0)
    if subnetIDsArg, ok := args["subnet_ids"].([]interface{}); ok {
        for _, subnetID := range subnetIDsArg {
            if subnetIDStr, ok := subnetID.(string); ok {
                subnetIDs = append(subnetIDs, subnetIDStr)
            }
        }
    }
    if len(subnetIDs) == 0 {
        return NewTextResult("", errors.New("missing required argument: subnet_ids (must be non-empty array)")), nil
    }
    
    // Handle optional region parameter with default
    region := "us-east-1"
    if regionArg, ok := args["region"].(string); ok && regionArg != "" {
        region = regionArg
    }
    
    // Get authenticated OCM client
    client, err := s.getAuthenticatedOCMClient(ctx)
    if err != nil {
        return NewTextResult("", fmt.Errorf("authentication failed: %s", err.Error())), nil
    }
    
    // Call OCM client with ROSA HCP parameters (no validation, pass directly to OCM API)
    cluster, err := client.CreateROSAHCPCluster(
        clusterName, awsAccountID, billingAccountID, roleArn,
        operatorRolePrefix, oidcConfigID, subnetIDs, region,
    )
    if err != nil {
        // Expose OCM API errors directly without modification
        if ocmErr, ok := err.(*ocm.OCMError); ok {
            return NewTextResult("", fmt.Errorf("OCM API Error [%s]: %s (Operation ID: %s)", 
                ocmErr.Code, ocmErr.Reason, ocmErr.OperationID)), nil
        }
        return NewTextResult("", fmt.Errorf("cluster creation failed: %s", err.Error())), nil
    }
    
    // Format response using MCP layer formatter
    formattedResponse := formatClusterCreateResponse(cluster)
    return NewTextResult(formattedResponse, nil), nil
}
```
- [ ] (24) Update other tool handlers to follow the same safe pattern:
```go
// Update handleGetCluster to use OpenShift MCP pattern
func (s *Server) handleGetCluster(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    args := ctr.GetArguments()
    
    clusterID, ok := args["cluster_id"].(string)
    if !ok || clusterID == "" {
        return NewTextResult("", errors.New("missing required argument: cluster_id")), nil
    }
    
    // Get authenticated OCM client
    client, err := s.getAuthenticatedOCMClient(ctx)
    if err != nil {
        return NewTextResult("", fmt.Errorf("authentication failed: %s", err.Error())), nil
    }
    
    // Call OCM client and handle response...
    // (rest of implementation as previously planned)
}
```
- [ ] (25) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/mcp/...`
- [ ] (26) Make a commit including only the changed files.
- [ ] (27) Update the implementation plan to reflect phase completion.

### Phase 5: Resources and Profiles

Review status: Approved

Implement the static resource system and basic profile management.

**Resource and Profile Implementation**:
- [ ] (28) Implement pkg/mcp/profiles.go with simple profile system and explanatory comment:
```go
/*
Package profiles implements a basic profile system inspired by the OpenShift MCP server.

The profiles concept allows us to selectively expose different sets of OCM tools
to avoid overloading an LLM's context window with too many tool options. For the MVP,
we implement only a basic "all tools enabled" profile, but this foundation will
allow us to support more complex OCM operations in the future without overwhelming
the AI assistant with excessive tool choices.

Future profile examples could include:
- rosa-classic: Legacy classic ROSA clusters
- rosa-full: All cluster lifecycle operations for all ROSA
- rosa-admin: Administrative, billing, and SRE operations
- aro-hcp: All ARO HCP cluster lifecycle operations
*/
type Profile struct {
    Name  string
    Tools []string
}
func GetDefaultProfile() Profile // Returns all tools enabled
```
- [ ] (29) Create static ROSA HCP Prerequisites Documentation resource:
  - Type: Static resource
  - Content: Documentation of cluster creation requirements
  - Include: IAM role setup, OIDC configuration, networking prerequisites
  - Reference: https://cloud.redhat.com/learning/learn:getting-started-red-hat-openshift-service-aws-rosa/resource/resources:creating-rosa-hcp-clusters-using-default-options
- [ ] (30) Register resource with MCP framework for client access
- [ ] (31) Validate the changes via typical project testing procedures.
        Run: `go test ./pkg/mcp/...`
- [ ] (32) Make a commit including only the changed files.
- [ ] (33) Update the implementation plan to reflect phase completion.

## TODOs
In this section, list TODOs required to be followed up on before moving onto the next milestone
- [ ] Verify all 4 tools are discoverable by MCP clients
- [ ] Test static resource content is accessible and helpful
- [ ] Confirm tool parameter schemas match requirements exactly
