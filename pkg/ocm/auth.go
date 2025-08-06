package ocm

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

const (
	// SSE transport header for OCM offline token
	// NOTE: http.Header keys are stored in canonical format, hence the different casing required here.
	// The provided X-OCM-OFFLINE-TOKEN header key will be translated to X-Ocm-Offline-Token
	SSETokenHeader = "X-Ocm-Offline-Token"

	// Stdio transport environment variable for OCM offline token
	StdioTokenEnv = "OCM_OFFLINE_TOKEN"
)

// ExtractTokenFromSSE extracts OCM offline token from X-OCM-OFFLINE-TOKEN header
func ExtractTokenFromSSE(headers map[string]string) (string, error) {
	// Try exact header name first
	token, exists := headers[SSETokenHeader]
	if exists && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("missing or empty %s header", SSETokenHeader)
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
		// For SSE transport, extract token from HTTP headers in the context
		if headers := extractHeadersFromContext(ctx); headers != nil {
			if token, err := ExtractTokenFromSSE(headers); err == nil {
				return token, nil
			}
		}

		// Fallback to environment variable for MVP compatibility
		token := os.Getenv(StdioTokenEnv)
		if token == "" {
			return "", fmt.Errorf("SSE transport requires %s header or %s environment variable", SSETokenHeader, StdioTokenEnv)
		}
		return token, nil
	default:
		return "", fmt.Errorf("unsupported transport mode: %s", transport)
	}
}

// extractHeadersFromContext extracts HTTP headers from the context
// This function looks for headers stored in the context by the mcp-go SSE server
func extractHeadersFromContext(ctx context.Context) map[string]string {
	// Check for headers stored in context by mcp-go framework
	if headers, ok := ctx.Value("headers").(map[string]string); ok {
		return headers
	}

	// Check for HTTP request in context
	if req, ok := ctx.Value("http.request").(*http.Request); ok {
		headers := make(map[string]string)
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		return headers
	}

	// Check for request context pattern used by some HTTP frameworks
	if req, ok := ctx.Value("request").(*http.Request); ok {
		headers := make(map[string]string)
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		return headers
	}

	return nil
}
