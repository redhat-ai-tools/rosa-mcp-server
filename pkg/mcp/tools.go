package mcp

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tiwillia/rosa-mcp-go/pkg/ocm"
)

// initTools returns the ROSA HCP tools
func (s *Server) initTools() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("whoami",
			mcp.WithDescription("Get the authenticated account"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.handleWhoami},

		{Tool: mcp.NewTool("get_clusters",
			mcp.WithDescription("Retrieves the list of clusters"),
			mcp.WithString("state", mcp.Description("Filter clusters by state (e.g., ready, installing, error)"), mcp.Required()),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.handleGetClusters},

		{Tool: mcp.NewTool("get_cluster",
			mcp.WithDescription("Retrieves the details of the cluster"),
			mcp.WithString("cluster_id", mcp.Description("Unique cluster identifier"), mcp.Required()),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.handleGetCluster},

		{Tool: mcp.NewTool("create_rosa_hcp_cluster",
			mcp.WithDescription("Provision a new ROSA HCP cluster with required configuration"),
			mcp.WithString("cluster_name", mcp.Description("Name for the cluster"), mcp.Required()),
			mcp.WithString("aws_account_id", mcp.Description("AWS account ID"), mcp.Required()),
			mcp.WithString("billing_account_id", mcp.Description("AWS billing account ID"), mcp.Required()),
			mcp.WithString("role_arn", mcp.Description("IAM installer role ARN"), mcp.Required()),
			mcp.WithString("operator_role_prefix", mcp.Description("Operator role prefix"), mcp.Required()),
			mcp.WithString("oidc_config_id", mcp.Description("OIDC configuration ID"), mcp.Required()),
			mcp.WithString("support_role_arn", mcp.Description("IAM support role ARN"), mcp.Required()),
			mcp.WithString("worker_role_arn", mcp.Description("IAM worker role ARN"), mcp.Required()),
			mcp.WithString("rosa_creator_arn", mcp.Description("ROSA creator ARN"), mcp.Required()),
			mcp.WithArray("subnet_ids", mcp.Description("Array of subnet IDs"), mcp.Required()),
			mcp.WithArray("availability_zones", mcp.Description("Array of availability zones for the subnets"), mcp.Required()),
			mcp.WithString("region", mcp.Description("AWS region"), mcp.DefaultString("us-east-1")),
			mcp.WithBoolean("multi_arch_enabled", mcp.Description("Enable multi-architecture support"), mcp.DefaultBool(false)),
			mcp.WithReadOnlyHintAnnotation(false),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.handleCreateROSAHCPCluster},
	}
}

// registerTools registers all MCP tools with the server
func (s *Server) registerTools() {
	// Get profile and tools
	profile := GetDefaultProfile()
	tools := profile.GetTools(s)

	// Register tools using SetTools
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

// handleCreateROSAHCPCluster handles the create_rosa_hcp_cluster tool
func (s *Server) handleCreateROSAHCPCluster(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := ctr.GetArguments()

	// Extract required parameters with safe type assertions
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

	supportRoleArn, ok := args["support_role_arn"].(string)
	if !ok || supportRoleArn == "" {
		return NewTextResult("", errors.New("missing required argument: support_role_arn")), nil
	}

	workerRoleArn, ok := args["worker_role_arn"].(string)
	if !ok || workerRoleArn == "" {
		return NewTextResult("", errors.New("missing required argument: worker_role_arn")), nil
	}

	rosaCreatorArn, ok := args["rosa_creator_arn"].(string)
	if !ok || rosaCreatorArn == "" {
		return NewTextResult("", errors.New("missing required argument: rosa_creator_arn")), nil
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

	// Handle availability_zones array parameter
	availabilityZones := make([]string, 0)
	if availabilityZonesArg, ok := args["availability_zones"].([]interface{}); ok {
		for _, az := range availabilityZonesArg {
			if azStr, ok := az.(string); ok {
				availabilityZones = append(availabilityZones, azStr)
			}
		}
	}
	if len(availabilityZones) == 0 {
		return NewTextResult("", errors.New("missing required argument: availability_zones (must be non-empty array)")), nil
	}

	// Handle region parameter with default using mcp.ParseString
	region := mcp.ParseString(ctr, "region", "us-east-1")

	// Handle optional boolean parameters with defaults
	multiArchEnabled := mcp.ParseBoolean(ctr, "multi_arch_enabled", false)

	s.logToolCall("create_rosa_hcp_cluster", map[string]interface{}{
		"cluster_name":         clusterName,
		"aws_account_id":       awsAccountID,
		"billing_account_id":   billingAccountID,
		"role_arn":             roleArn,
		"operator_role_prefix": operatorRolePrefix,
		"oidc_config_id":       oidcConfigID,
		"support_role_arn":     supportRoleArn,
		"worker_role_arn":      workerRoleArn,
		"rosa_creator_arn":     rosaCreatorArn,
		"subnet_ids":           subnetIDs,
		"availability_zones":   availabilityZones,
		"region":               region,
		"multi_arch_enabled":   multiArchEnabled,
	})

	// Get authenticated OCM client
	client, err := s.getAuthenticatedOCMClient(ctx)
	if err != nil {
		return NewTextResult("", errors.New("authentication failed: "+err.Error())), nil
	}
	defer client.Close()

	// Call OCM client with ROSA HCP parameters (no validation, pass directly to OCM API)
	cluster, err := client.CreateROSAHCPCluster(
		clusterName, awsAccountID, billingAccountID, roleArn,
		operatorRolePrefix, oidcConfigID, supportRoleArn, workerRoleArn, rosaCreatorArn,
		subnetIDs, availabilityZones, region,
		multiArchEnabled,
	)
	if err != nil {
		// Expose OCM API errors directly without modification
		if ocmErr, ok := err.(*ocm.OCMError); ok {
			return NewTextResult("", errors.New("OCM API Error ["+ocmErr.Code+"]: "+ocmErr.Reason)), nil
		}
		return NewTextResult("", errors.New("cluster creation failed: "+err.Error())), nil
	}

	// Format response using MCP layer formatter
	formattedResponse := formatClusterCreateResponse(cluster)
	return NewTextResult(formattedResponse, nil), nil
}

// NewTextResult creates a new MCP CallToolResult
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
