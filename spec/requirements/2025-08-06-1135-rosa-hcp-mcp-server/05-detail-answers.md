# Expert Detail Answers

## Q6: Should the create_rosa_hcp_cluster tool require ALL ROSA HCP parameters (AWS account, role ARNs, OIDC config, subnet IDs) to be provided by the user?
**Answer:** Yes - require everything listed (AWS account, role ARNs, OIDC config, subnet IDs). These are user-specific and must be set up beforehand. Region can have a sensible default.

## Q7: Should we implement the same profile system as the OpenShift MCP server (e.g., "rosa-basic", "rosa-full") to control which tools are available?
**Answer:** No - expose all 4 tools by default for MVP simplicity. Consider as future option.

## Q8: Should the X-OCM-OFFLINE-TOKEN header support follow the exact pattern from ocm_mcp_server.py with fallback to environment variable for stdio transport?
**Answer:** Yes - X-OCM-OFFLINE-TOKEN header for SSE transport, with fallback to OCM_OFFLINE_TOKEN environment variable for stdio transport.

## Q9: Should cluster creation return the raw OCM API response JSON or a formatted human-readable string like the other tools?
**Answer:** Formatted string

## Q10: Should the server validate that provided AWS IAM role ARNs exist before making the cluster creation API call?
**Answer:** No - let OCM API handle validation. Note: this may change after the MVP.