# Expert Detail Questions

## Q6: Should the create_rosa_hcp_cluster tool require ALL ROSA HCP parameters (AWS account, role ARNs, OIDC config, subnet IDs) to be provided by the user?
**Default if unknown:** No (provide sensible defaults for MVP, require only cluster name and optionally region)

## Q7: Should we implement the same profile system as the OpenShift MCP server (e.g., "rosa-basic", "rosa-full") to control which tools are available?
**Default if unknown:** No (expose all 4 tools by default for MVP simplicity)

## Q8: Should the X-OCM-OFFLINE-TOKEN header support follow the exact pattern from ocm_mcp_server.py with fallback to environment variable for stdio transport?
**Default if unknown:** Yes (maintains compatibility with existing AI assistant integrations)

## Q9: Should cluster creation return the raw OCM API response JSON or a formatted human-readable string like the other tools?
**Default if unknown:** Formatted string (consistent with other tool responses and human readability requirements)

## Q10: Should the server validate that provided AWS IAM role ARNs exist before making the cluster creation API call?
**Default if unknown:** No (let OCM API handle validation and return meaningful errors directly to the user)