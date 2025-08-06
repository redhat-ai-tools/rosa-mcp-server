package mcp

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/redhat-ai-tools/rosa-mcp-go/pkg/ocm"
)

// initTools returns the ROSA HCP tools following OpenShift MCP server patterns
func (s *Server) initTools() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("whoami",
			mcp.WithDescription("Get the authenticated account"),
		), Handler: s.handleWhoami},
		
		{Tool: mcp.NewTool("get_clusters",
			mcp.WithDescription("Retrieves the list of clusters"),
			mcp.WithString("state", mcp.Description("Filter clusters by state (e.g., ready, installing, error)"), mcp.Required()),
		), Handler: s.handleGetClusters},
		
		{Tool: mcp.NewTool("get_cluster",
			mcp.WithDescription("Retrieves the details of the cluster"),
			mcp.WithString("cluster_id", mcp.Description("Unique cluster identifier"), mcp.Required()),
		), Handler: s.handleGetCluster},
		
		// create_rosa_hcp_cluster tool will be implemented in Phase 4
	}
}

// registerTools registers all MCP tools with the server following OpenShift MCP patterns
func (s *Server) registerTools() {
	// Get profile and tools
	profile := GetDefaultProfile()
	tools := profile.GetTools(s)
	
	// Register tools using SetTools like OpenShift MCP server
	s.mcpServer.SetTools(tools...)
}

// handleWhoami handles the whoami tool
func (s *Server) handleWhoami(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logToolCall("whoami", convertParamsToMap())

	// Get authenticated OCM client
	client, err := s.getAuthenticatedOCMClient(ctx)
	if err != nil {
		return NewTextResult("", errors.New("authentication failed: "+err.Error())), nil
	}
	defer client.Close()

	// Call OCM client to get current account
	account, err := client.GetCurrentAccount()
	if err != nil {
		// Handle OCM API errors with code and reason exposure
		if ocmErr, ok := err.(*ocm.OCMError); ok {
			return NewTextResult("", errors.New("OCM API Error ["+ocmErr.Code+"]: "+ocmErr.Reason)), nil
		}
		return NewTextResult("", errors.New("failed to get account: "+err.Error())), nil
	}

	// Format response using MCP layer formatter
	formattedResponse := formatAccountResponse(account)
	return NewTextResult(formattedResponse, nil), nil
}

// handleGetClusters handles the get_clusters tool
func (s *Server) handleGetClusters(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := ctr.GetArguments()

	state, ok := args["state"].(string)
	if !ok || state == "" {
		return NewTextResult("", errors.New("missing required argument: state")), nil
	}

	s.logToolCall("get_clusters", map[string]interface{}{"state": state})

	// Get authenticated OCM client
	client, err := s.getAuthenticatedOCMClient(ctx)
	if err != nil {
		return NewTextResult("", errors.New("authentication failed: "+err.Error())), nil
	}
	defer client.Close()

	// Call OCM client to get clusters with state filter
	clusters, err := client.GetClusters(state)
	if err != nil {
		// Handle OCM API errors with code and reason exposure
		if ocmErr, ok := err.(*ocm.OCMError); ok {
			return NewTextResult("", errors.New("OCM API Error ["+ocmErr.Code+"]: "+ocmErr.Reason)), nil
		}
		return NewTextResult("", errors.New("failed to get clusters: "+err.Error())), nil
	}

	// Format response using MCP layer formatter
	formattedResponse := formatClustersResponse(clusters)
	return NewTextResult(formattedResponse, nil), nil
}

// handleGetCluster handles the get_cluster tool
func (s *Server) handleGetCluster(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := ctr.GetArguments()

	clusterID, ok := args["cluster_id"].(string)
	if !ok || clusterID == "" {
		return NewTextResult("", errors.New("missing required argument: cluster_id")), nil
	}

	s.logToolCall("get_cluster", map[string]interface{}{"cluster_id": clusterID})

	// Get authenticated OCM client
	client, err := s.getAuthenticatedOCMClient(ctx)
	if err != nil {
		return NewTextResult("", errors.New("authentication failed: "+err.Error())), nil
	}
	defer client.Close()

	// Call OCM client to get cluster details
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		// Handle OCM API errors with code and reason exposure
		if ocmErr, ok := err.(*ocm.OCMError); ok {
			return NewTextResult("", errors.New("OCM API Error ["+ocmErr.Code+"]: "+ocmErr.Reason)), nil
		}
		return NewTextResult("", errors.New("failed to get cluster: "+err.Error())), nil
	}

	// Format response using MCP layer formatter
	formattedResponse := formatClusterResponse(cluster)
	return NewTextResult(formattedResponse, nil), nil
}

// NewTextResult creates a new MCP CallToolResult following OpenShift MCP server patterns
func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}