package ocm

import (
	"context"
	"fmt"
	"os"
)

const (
	// SSE transport header for OCM offline token
	SSETokenHeader = "X-OCM-OFFLINE-TOKEN"
	
	// Stdio transport environment variable for OCM offline token
	StdioTokenEnv = "OCM_OFFLINE_TOKEN"
)

// ExtractTokenFromSSE extracts OCM offline token from X-OCM-OFFLINE-TOKEN header
func ExtractTokenFromSSE(headers map[string]string) (string, error) {
	token, exists := headers[SSETokenHeader]
	if !exists || token == "" {
		return "", fmt.Errorf("missing or empty %s header", SSETokenHeader)
	}
	return token, nil
}

// ExtractTokenFromStdio extracts OCM offline token from OCM_OFFLINE_TOKEN environment variable
func ExtractTokenFromStdio() (string, error) {
	token := os.Getenv(StdioTokenEnv)
	if token == "" {
		return "", fmt.Errorf("missing or empty %s environment variable", StdioTokenEnv)
	}
	return token, nil
}

// ExtractTokenFromContext extracts OCM offline token from context based on transport mode
func ExtractTokenFromContext(ctx context.Context, transport string) (string, error) {
	switch transport {
	case "stdio":
		return ExtractTokenFromStdio()
	case "sse":
		// Extract headers from context (implementation depends on MCP framework)
		// For now, return error - this will be implemented in MCP layer
		return "", fmt.Errorf("SSE token extraction from context not yet implemented")
	default:
		return "", fmt.Errorf("unsupported transport mode: %s", transport)
	}
}