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
		// For SSE transport, the token should be provided via X-OCM-OFFLINE-TOKEN header
		// In a real implementation, this would extract from the HTTP request context
		// For MVP, we'll check environment variable as fallback
		token := os.Getenv(StdioTokenEnv)
		if token == "" {
			return "", fmt.Errorf("SSE transport requires %s header or %s environment variable", SSETokenHeader, StdioTokenEnv)
		}
		return token, nil
	default:
		return "", fmt.Errorf("unsupported transport mode: %s", transport)
	}
}