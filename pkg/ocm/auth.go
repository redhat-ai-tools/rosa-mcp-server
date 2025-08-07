package ocm

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/golang/glog"
)

// contextKey type for context value storage
type contextKey int

const (
	// requestHeader key used by mcp-go framework to store HTTP headers in context
	requestHeader contextKey = iota
)

// RequestHeaderKey returns the context key used for storing HTTP headers
func RequestHeaderKey() contextKey {
	return requestHeader
}

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
	glog.V(3).Infof("SSE headers received: %+v", headers)
	
	// Try exact header name first
	token, exists := headers[SSETokenHeader]
	if exists && token != "" {
		glog.V(3).Infof("Found OCM token in header %s", SSETokenHeader)
		return token, nil
	}

	glog.Warningf("Missing or empty %s header in SSE request. Available headers: %+v", SSETokenHeader, headers)
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
	glog.V(2).Infof("Extracting token for transport mode: %s", transport)
	
	switch transport {
	case "stdio":
		return ExtractTokenFromStdio()
	case "sse":
		// For SSE transport, extract token from HTTP headers in the context
		headers := extractHeadersFromContext(ctx)
		if headers != nil {
			glog.V(2).Infof("Headers found in context for SSE transport")
			if token, err := ExtractTokenFromSSE(headers); err == nil {
				return token, nil
			} else {
				glog.Warningf("Failed to extract token from SSE headers: %v", err)
			}
		} else {
			glog.Warningf("No headers found in context for SSE transport")
		}

		// Fallback to environment variable for MVP compatibility
		token := os.Getenv(StdioTokenEnv)
		if token == "" {
			err := fmt.Errorf("SSE transport requires %s header or %s environment variable", SSETokenHeader, StdioTokenEnv)
			glog.Errorf("Authentication failed: %v", err)
			return "", err
		}
		glog.V(2).Infof("Using fallback environment variable for SSE transport")
		return token, nil
	default:
		err := fmt.Errorf("unsupported transport mode: %s", transport)
		glog.Errorf("Authentication failed: %v", err)
		return "", err
	}
}

// extractHeadersFromContext extracts HTTP headers from the context
// This function looks for headers stored in the context by the mcp-go SSE server
func extractHeadersFromContext(ctx context.Context) map[string]string {
	glog.V(2).Infof("Extracting headers from context")
	
	// Check for headers stored by mcp-go framework using requestHeader key
	if headerValue := ctx.Value(requestHeader); headerValue != nil {
		glog.V(2).Infof("Found requestHeader context value")
		if httpHeader, ok := headerValue.(http.Header); ok {
			headers := make(map[string]string)
			for key, values := range httpHeader {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}
			glog.V(2).Infof("Extracted %d headers from mcp-go context", len(headers))
			return headers
		} else {
			glog.V(2).Infof("requestHeader context value is not http.Header type: %T", headerValue)
		}
	} else {
		glog.V(2).Infof("No requestHeader found in context")
	}

	// Fallback: Check for headers stored as map[string]string (legacy)
	if headers, ok := ctx.Value("headers").(map[string]string); ok {
		glog.V(2).Infof("Found legacy headers map in context with %d entries", len(headers))
		return headers
	}

	// Fallback: Check for HTTP request in context
	if req, ok := ctx.Value("http.request").(*http.Request); ok {
		headers := make(map[string]string)
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		glog.V(2).Infof("Extracted %d headers from http.request context", len(headers))
		return headers
	}

	// Fallback: Check for request context pattern used by some HTTP frameworks
	if req, ok := ctx.Value("request").(*http.Request); ok {
		headers := make(map[string]string)
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		glog.V(2).Infof("Extracted %d headers from request context", len(headers))
		return headers
	}

	glog.V(2).Infof("No headers found in any context format")
	return nil
}
