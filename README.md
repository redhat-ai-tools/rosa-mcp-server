# ROSA HCP MCP Server

A Model Context Protocol (MCP) server for ROSA HCP (Red Hat OpenShift Service on AWS using Hosted Control Planes) that enables AI assistants to integrate with Red Hat Managed OpenShift services.

## Features

- **4 Core Tools**: `whoami`, `get_clusters`, `get_cluster`, `create_rosa_hcp_cluster`
- **Dual Transport Support**: stdio and Server-Sent Events (SSE)
- **OCM API Integration**: Direct integration with OpenShift Cluster Manager
- **Multi-Region Support**: Configurable AWS regions (default: us-east-1)
- **Enterprise Authentication**: OCM offline tokens with concurrent user support

## Installation

### Prerequisites

- Go 1.21 or later
- OCM offline token from [console.redhat.com/openshift/token](https://console.redhat.com/openshift/token)

### Build from Source

```bash
git clone https://github.com/redhat-ai-tools/rosa-mcp-go
cd rosa-mcp-go
go build -o rosa-mcp-server cmd/rosa-mcp-server/main.go
```

## Configuration

### Command Line Flags

```bash
rosa-mcp-server [flags]

Flags:
  --config string         path to configuration file
  --ocm-base-url string   OCM API base URL (default "https://api.openshift.com")
  --port int              port for SSE transport (default 8080)
  --sse-base-url string   SSE base URL for public endpoints
  --transport string      transport mode (stdio/sse) (default "stdio")
```

### Environment Variables

- `OCM_OFFLINE_TOKEN`: Your OCM offline token for authentication

### TOML Configuration File

```toml
ocm_base_url = "https://api.openshift.com"
transport = "stdio"
port = 8080
sse_base_url = "https://example.com:8080"
```

## Usage Examples

### Stdio Transport (Local)

```bash
# Set your OCM token
export OCM_OFFLINE_TOKEN="your-ocm-token-here"

# Start server
./rosa-mcp-server --transport=stdio
```

### SSE Transport (Remote)

```bash
# Start SSE server
./rosa-mcp-server --transport=sse --port=8080

# Server will be available at http://localhost:8080/sse
# Send X-OCM-OFFLINE-TOKEN header with requests
```

## Authentication Setup

### 1. Get OCM Offline Token

1. Visit [console.redhat.com/openshift/token](https://console.redhat.com/openshift/token)
2. Log in with your Red Hat account
3. Copy the offline token

### 2. Configure Authentication

**For stdio transport:**
```bash
export OCM_OFFLINE_TOKEN="your-token-here"
```

**For SSE transport:**
```bash
# Include header in HTTP requests
X-OCM-OFFLINE-TOKEN: your-token-here
```

## Available Tools

### 1. whoami
Get information about the authenticated account.
```json
{
  "name": "whoami",
  "description": "Get the authenticated account"
}
```

### 2. get_clusters
Retrieve a list of clusters filtered by state.
```json
{
  "name": "get_clusters",
  "parameters": {
    "state": {
      "type": "string",
      "description": "Filter clusters by state (e.g., ready, installing, error)",
      "required": true
    }
  }
}
```

### 3. get_cluster
Get detailed information about a specific cluster.
```json
{
  "name": "get_cluster",
  "parameters": {
    "cluster_id": {
      "type": "string", 
      "description": "Unique cluster identifier",
      "required": true
    }
  }
}
```

### 4. create_rosa_hcp_cluster
Provision a new ROSA HCP cluster with required AWS configuration.
```json
{
  "name": "create_rosa_hcp_cluster",
  "parameters": {
    "cluster_name": {"type": "string", "required": true},
    "aws_account_id": {"type": "string", "required": true},
    "billing_account_id": {"type": "string", "required": true},
    "role_arn": {"type": "string", "required": true},
    "operator_role_prefix": {"type": "string", "required": true},
    "oidc_config_id": {"type": "string", "required": true},
    "subnet_ids": {"type": "array", "required": true},
    "region": {"type": "string", "default": "us-east-1"}
  }
}
```

## ROSA HCP Prerequisites

Before creating clusters, ensure you have:

- **AWS Account**: Account ID and billing account ID
- **IAM Roles**: Installer, support, worker, and control plane roles configured
- **OIDC Configuration**: OIDC config ID for secure authentication
- **Networking**: At least 2 subnet IDs in different availability zones
- **Operator Roles**: Role prefix for cluster operators

### Example Cluster Creation

```bash
# All required parameters for ROSA HCP cluster
{
  "cluster_name": "my-rosa-hcp",
  "aws_account_id": "123456789012",
  "billing_account_id": "123456789012", 
  "role_arn": "arn:aws:iam::123456789012:role/ManagedOpenShift-Installer-Role",
  "operator_role_prefix": "my-cluster-operators",
  "oidc_config_id": "2kg4slloso10aa8q0jdscjoaeb97bevq",
  "subnet_ids": ["subnet-12345", "subnet-67890"],
  "region": "us-east-1"
}
```

## Development

### Project Structure

```
├── cmd/rosa-mcp-server/     # Main entry point
├── pkg/
│   ├── config/              # Configuration management
│   ├── mcp/                 # MCP server implementation
│   ├── ocm/                 # OCM API client wrapper
│   └── version/             # Version information
├── go.mod                   # Go module definition
└── README.md               # This file
```

### Testing

```bash
# Run all tests
go test ./...

# Build project
go build ./...

# Test server startup
./rosa-mcp-server --help
./rosa-mcp-server version
```

## Integration with AI Assistants

### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "rosa-hcp": {
      "command": "/path/to/rosa-mcp-server",
      "args": ["--transport=stdio"],
      "env": {
        "OCM_OFFLINE_TOKEN": "your-token-here"
      }
    }
  }
}
```

### SSE Integration

For remote integrations, use the SSE endpoint:

```bash
# Server endpoint
POST http://localhost:8080/sse

# Required header
X-OCM-OFFLINE-TOKEN: your-token-here
```

## Error Handling

The server exposes OCM API errors directly without modification:

```
OCM API Error [CLUSTERS-MGMT-400]: Invalid cluster configuration
```

Common error scenarios:
- **Authentication**: Invalid or expired OCM token
- **Permissions**: Insufficient permissions for cluster operations
- **AWS Resources**: Missing or misconfigured AWS prerequisites
- **Validation**: Invalid cluster parameters

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0.

## Support

For issues and questions:
- GitHub Issues: [redhat-ai-tools/rosa-mcp-go/issues](https://github.com/redhat-ai-tools/rosa-mcp-go/issues)
- Red Hat Support: Contact Red Hat support for ROSA-related issues
- OCM Documentation: [console.redhat.com](https://console.redhat.com)