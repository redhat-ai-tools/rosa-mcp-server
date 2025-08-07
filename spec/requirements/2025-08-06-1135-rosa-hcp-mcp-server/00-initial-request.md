# Initial Request

**Timestamp:** 2025-08-06 11:35

**Request:** Build an MCP server for ROSA HCP (Red Hat OpenShift on AWS using Hosted Control Planes) to enable AI assistants to integrate with Red Hat Managed OpenShift services.

## Full Requirements from initial_requirements.md

Build an MCP server for ROSA HCP (Red Hat OpenShift on AWS using Hosted Control Planes) to enable AI assistants to integrate with Red Hat Managed OpenShift services. The server should provide a minimal MVP implementation focusing on essential cluster operations and documentation.

### Solution Overview
A golang-based MCP server following the business logic established in the python-based OCM MCP server https://github.com/redhat-ai-tools/ocm-mcp/ . The structure and best practices of the solution should follow the excellent examples set by the golang-based OpenShift MCP server https://github.com/openshift/openshift-mcp-server, using the same https://github.com/mark3labs/mcp-go framework.

### Core Requirements
- 4 core tools: List Clusters, Get Cluster Details, Create ROSA HCP Cluster, Get Authenticated Account
- 1 resource: ROSA HCP Prerequisites Documentation
- OAuth token-based authentication using OCM offline tokens
- Support both stdio and SSE transport modes
- Integration with OCM API endpoints at https://api.openshift.com