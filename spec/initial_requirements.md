# Initial Requirements

# ROSA HCP MCP Server Initial Requirements

## Problem Statement

Build an MCP server for ROSA HCP (Red Hat OpenShift on AWS using Hosted Control Planes) to enable AI assistants to integrate with Red Hat Managed OpenShift services. The server should provide a minimal MVP implementation focusing on essential cluster operations and documentation.

## Solution Overview

A golang-based MCP server following the business logic established in the python-based OCM MCP server https://github.com/redhat-ai-tools/ocm-mcp/ . The structure and best practices of the solution should follow the excellent examples set by the golang-based OpenShift MCP server https://github.com/openshift/openshift-mcp-server, using the same https://github.com/mark3labs/mcp-go framework.

Design must be kept simple to allow for quick implementation. Testing should be limited to unit tests.

## Functional Requirements

### Core Tools (4 required)
1. **List Clusters** - Retrieve list of ROSA clusters with state filtering
2. **Get Cluster Details** - Retrieve detailed information for a specific cluster
3. **Create ROSA HCP Cluster** - Provision new ROSA HCP clusters with minimal configuration
4. **Get Authenticated Account** - Verify authentication and return account information

### Resources (1 required)
1. **ROSA HCP Prerequisites Documentation** - Static documentation of cluster creation requirements

### Authentication
- OAuth token-based authentication using OCM offline tokens
- Token source: https://console.redhat.com/openshift/token
- Follow OCM MCP server authentication pattern exactly

### Transport Support
- Support both stdio and SSE transport modes
- SSE transport must support the X-OCM-OFFLINE-TOKEN header to allow authentication to be provided to the server
- Command-line configuration for transport selection

### API Integration
- **Base URL:** https://api.openshift.com (OCM API endpoints) - must be configurable
- **Authentication:** Bearer token via OAuth refresh flow
- **Endpoints:** Reuse existing OCM API patterns from python reference implementation

### Tool Implementations

#### 1. `whoami()`
- **Description:** "Get the authenticated account"
- **Parameters:** None
- **Returns:** Formatted account information

#### 2. `get_clusters(state: str)`
- **Description:** "Retrieves the list of clusters"
- **Parameters:**
  - `state`: Filter clusters by state (required)
- **Returns:** Formatted cluster list

#### 3. `get_cluster(cluster_id: str)`
- **Description:** "Retrieves the details of the cluster"
- **Parameters:**
  - `cluster_id`: Unique cluster identifier (required)
- **Returns:** Formatted cluster details

#### 4. `create_rosa_hcp_cluster(...)`
- **Description:** "Provision a new ROSA HCP cluster with bare minimum configuration"
- **Parameters:** Minimal required parameters for cluster creation
- **Returns:** Cluster creation status and details

### Implementation Patterns

**Response Formatting:**
- Implement dedicated formatter functions for human-readable output
- Convert JSON API responses to readable strings

**Error Handling:**
- Expose OCM API error details in MCP tool call responses without retry logic
- Enable MCP server consumers to resolve errors
- No local error recovery or retry mechanisms

## Examples and Details

Golang-based OpenShift MCP Server: https://github.com/openshift/openshift-mcp-server
Golang-based Kubernetes MCP Server: https://github.com/redhat-ai-tools/ocm-mcp
Python-based OCM MCP Server: https://github.com/redhat-ai-tools/ocm-mcp

Minimal ROSA HCP Cluster creation documentation: https://cloud.redhat.com/learning/learn:getting-started-red-hat-openshift-service-aws-rosa/resource/resources:creating-rosa-hcp-clusters-using-default-options
Basic ROSA HCP Cluster Creation Workshop: https://www.rosaworkshop.io/rosa/18-deploy_hcp/

### OpenAPI Specs

**OCM Clusters_mgmt API**: https://api.openshift.com/api/clusters_mgmt/v1/openapi
**OCM Accounts_mgmt API Spec**: https://api.openshift.com/api/accounts_mgmt/v1/openapi

### Example POST /api/clusters_mgmt/v1/clusters Request

```json
{
  "aws": {
    "account_id": "****",
    "billing_account_id": "****",
    "sts": {
      "auto_mode": true,
      "oidc_config": {
        "id": "2kg4slloso10aa8q0jdscjoaeb97bevq"
      },
      "operator_role_prefix": "cs-ci-qkhmb-4krd",
      "role_arn": "arn:aws:iam::****:role/CSE2ETests-HCP-ROSA-Installer-Role"
    },
    "subnet_ids": [
      "subnet-020d9293d85d6c842",
      "subnet-001dbc2264d3432c2"
    ]
  },
  "billing_model": "marketplace-aws",
  "ccs": {
    "enabled": true,
    "kind": "CCS"
  },
  "hypershift": {
    "enabled": true
  },
  "kind": "Cluster",
  "name": "example-hcp",
  "product": {
    "id": "rosa",
    "kind": "Product"
  },
  "region": {
    "id": "us-east-1",
    "kind": "CloudRegion"
  }
}
```

### Example OCM API Error Response

```json
{
  "kind":"Error",
  "id":"400",
  "href":"/api/clusters_mgmt/v1/errors/400",
  "code":"CLUSTERS-MGMT-400",
  "reason":"Attribute 'aws.sts.role_arn' is not a valid ARN",
  "operation_id":"2ac99562-cd61-4626-afcb-acff13eddb00",
  "timestamp":"2025-08-06T15:24:16.640485899Z"
}
```

All errors will contain these fields, with the `reason` and `code` fields being the most important to expose in the MCP server response when an error occurs in OCM API client calls.
