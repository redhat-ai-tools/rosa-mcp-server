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
			mcp.WithDescription(`Provision a new ROSA HCP cluster with basic configuration.

Use the workflow from the get_rosa_hcp_prerequisites_guide tool or prompt to guide a user through completing the necessary pre-requisite steps and collecting the required configuration values.`),
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

		{Tool: mcp.NewTool("get_rosa_hcp_prerequisites_guide",
			mcp.WithDescription(`Get the complete workflow prompt for ROSA HCP cluster installation prerequisites and setup.

Use this workflow to guide a user through the complete setup process for creating a ROSA HCP (Red Hat OpenShift Service on AWS with Hosted Control Planes) cluster. using the create_rosa_hcp_cluster tool.`),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.handleGetROSAHCPPrerequisitesGuide},

		{Tool: mcp.NewTool("setup_htpasswd_identity_provider",
			mcp.WithDescription(`Setup an HTPasswd identity provider for a ROSA HCP cluster.

HTPasswd is a common identity provider for development and testing environments. This tool allows creating users with username/password authentication.`),
			mcp.WithString("cluster_id", mcp.Description("Target cluster identifier"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Identity provider name"), mcp.DefaultString("htpasswd")),
			mcp.WithString("mapping_method", mcp.Description("User mapping method - options: add, claim, generate, lookup"), mcp.DefaultString("claim")),
			mcp.WithArray("users", mcp.Description("List of username:password pairs [\"user1:password1\", \"user2:password2\"]"), mcp.Required()),
			mcp.WithBoolean("overwrite_existing", mcp.Description("Whether to overwrite if IDP with same name exists"), mcp.DefaultBool(false)),
			mcp.WithReadOnlyHintAnnotation(false),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
		), Handler: s.handleSetupHTPasswdIdentityProvider},
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
	if errorResult := handleOCMError(err, "failed to get account"); errorResult != nil {
		return errorResult, nil
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
	if errorResult := handleOCMError(err, "failed to get clusters"); errorResult != nil {
		return errorResult, nil
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
	if errorResult := handleOCMError(err, "failed to get cluster"); errorResult != nil {
		return errorResult, nil
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
	if errorResult := handleOCMError(err, "cluster creation"); errorResult != nil {
		return errorResult, nil
	}

	// Format response using MCP layer formatter
	formattedResponse := formatClusterCreateResponse(cluster)
	return NewTextResult(formattedResponse, nil), nil
}

// handleGetROSAHCPPrerequisitesGuide handles the get_rosa_hcp_prerequisites_guide tool
func (s *Server) handleGetROSAHCPPrerequisitesGuide(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logToolCall("get_rosa_hcp_prerequisites_guide", convertParamsToMap())

	// Return the embedded prerequisites guide content directly
	return NewTextResult(prereqsGuide, nil), nil
}

// handleSetupHTPasswdIdentityProvider handles the setup_htpasswd_identity_provider tool
func (s *Server) handleSetupHTPasswdIdentityProvider(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := ctr.GetArguments()
	s.logToolCall("setup_htpasswd_identity_provider", params)

	// Extract required parameters
	clusterID, ok := params["cluster_id"].(string)
	if !ok || clusterID == "" {
		return NewTextResult("", errors.New("cluster_id parameter is required")), nil
	}

	// Extract optional parameters with defaults
	name := "htpasswd"
	if n, ok := params["name"].(string); ok && n != "" {
		name = n
	}

	mappingMethod := "claim"
	if mm, ok := params["mapping_method"].(string); ok && mm != "" {
		mappingMethod = mm
	}

	overwriteExisting := false
	if ow, ok := params["overwrite_existing"].(bool); ok {
		overwriteExisting = ow
	}

	// Get authenticated OCM client
	client, err := s.getAuthenticatedOCMClient(ctx)
	if err != nil {
		return NewTextResult("", errors.New("authentication failed: "+err.Error())), nil
	}
	defer client.Close()

	// Setup HTPasswd identity provider using OCM client
	idp, err := client.SetupHTPasswdIdentityProvider(clusterID, name, mappingMethod, params, overwriteExisting)
	if errorResult := handleOCMError(err, "failed to setup HTPasswd identity provider"); errorResult != nil {
		return errorResult, nil
	}

	// Get cluster details for response formatting
	cluster, err := client.GetCluster(clusterID)
	if errorResult := handleOCMError(err, "failed to get cluster details"); errorResult != nil {
		return errorResult, nil
	}

	// Count the number of users created
	userCount := 0
	if idp.Htpasswd() != nil && idp.Htpasswd().Users() != nil {
		userCount = len(idp.Htpasswd().Users().Slice())
	}

	// Format response using MCP layer formatter
	formattedResponse := FormatHTPasswdIdentityProviderResult(idp, cluster, userCount)
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

// handleOCMError processes OCM API errors with enhanced token expiration detection
// Returns an appropriate MCP CallToolResult for the error, or nil if no error
func handleOCMError(err error, operation string) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	// Handle OCM API errors with enhanced token expiration detection
	if ocmErr, ok := err.(*ocm.OCMError); ok {
		// Return specific error for token expiration
		if ocm.IsAccessTokenExpiredError(ocmErr) {
			return mcp.NewToolResultError("AUTHENTICATION_FAILED: " + ocmErr.Error())
		}
		return NewTextResult("", errors.New("OCM API Error ["+ocmErr.Code+"]: "+ocmErr.Reason))
	}
	return NewTextResult("", errors.New(operation+" failed: "+err.Error()))
}
