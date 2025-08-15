package ocm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	// Authorization header for access tokens (standard OAuth2)
	AuthorizationHeader = "Authorization"
	
	// SSE transport header for OCM offline token
	// NOTE: http.Header keys are stored in canonical format, hence the different casing required here.
	// The provided X-OCM-OFFLINE-TOKEN header key will be translated to X-Ocm-Offline-Token
	SSETokenHeader = "X-Ocm-Offline-Token"

	// Stdio transport environment variable for OCM offline token
	StdioTokenEnv = "OCM_OFFLINE_TOKEN"
)

// TokenInfo represents extracted token information
type TokenInfo struct {
	Token     string
	TokenType string // "access" or "offline"
}

// ExtractBearerToken extracts access token from Authorization header
func ExtractBearerToken(headers map[string]string) (string, error) {
	authHeader, exists := headers[AuthorizationHeader]
	if !exists || authHeader == "" {
		return "", fmt.Errorf("missing or empty %s header", AuthorizationHeader)
	}
	
	// Check for "Bearer " prefix (case-insensitive)
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	
	if !strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
		return "", fmt.Errorf("Authorization header must use Bearer scheme")
	}
	
	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", fmt.Errorf("empty Bearer token")
	}
	
	return token, nil
}

// ExtractTokenInfoFromSSE extracts token info from SSE headers, preferring access tokens
func ExtractTokenInfoFromSSE(headers map[string]string) (*TokenInfo, error) {
	// Log only header keys for security (never log header values which may contain tokens)
	headerKeys := make([]string, 0, len(headers))
	for key := range headers {
		headerKeys = append(headerKeys, key)
	}
	glog.V(3).Infof("SSE headers received (keys only): %v", headerKeys)
	
	// Try Authorization header first (access token)
	if token, err := ExtractBearerToken(headers); err == nil {
		glog.V(3).Info("Found access token in Authorization header")
		return &TokenInfo{Token: token, TokenType: "access"}, nil
	}
	
	// Fallback to offline token header
	token, exists := headers[SSETokenHeader]
	if exists && token != "" {
		glog.V(3).Infof("Found offline token in header %s", SSETokenHeader)
		return &TokenInfo{Token: token, TokenType: "offline"}, nil
	}

	glog.Warningf("No valid tokens found in SSE headers. Available header keys: %v", headerKeys)
	return nil, fmt.Errorf("missing valid authentication token")
}

// ExtractTokenFromStdio extracts OCM offline token from OCM_OFFLINE_TOKEN environment variable
func ExtractTokenFromStdio() (string, error) {
	token := os.Getenv(StdioTokenEnv)
	if token == "" {
		return "", fmt.Errorf("missing or empty %s environment variable", StdioTokenEnv)
	}
	return token, nil
}

// ExtractTokenInfoFromContext extracts token info from context based on transport mode
func ExtractTokenInfoFromContext(ctx context.Context, transport string) (*TokenInfo, error) {
	glog.V(2).Infof("Extracting token for transport mode: %s", transport)
	
	switch transport {
	case "stdio":
		// Stdio only supports offline tokens via environment variable
		token, err := ExtractTokenFromStdio()
		if err != nil {
			return nil, err
		}
		return &TokenInfo{Token: token, TokenType: "offline"}, nil
		
	case "sse":
		// For SSE transport, extract from HTTP headers
		headers := extractHeadersFromContext(ctx)
		if headers != nil {
			if tokenInfo, err := ExtractTokenInfoFromSSE(headers); err == nil {
				return tokenInfo, nil
			} else {
				glog.Warningf("Failed to extract token from SSE headers: %v", err)
			}
		}

		// Fallback to environment variable
		token := os.Getenv(StdioTokenEnv)
		if token == "" {
			err := fmt.Errorf("SSE transport requires %s header or %s environment variable", 
				AuthorizationHeader, StdioTokenEnv)
			glog.Errorf("Authentication failed: %v", err)
			return nil, err
		}
		glog.V(2).Infof("Using fallback environment variable for SSE transport")
		return &TokenInfo{Token: token, TokenType: "offline"}, nil
		
	default:
		err := fmt.Errorf("unsupported transport mode: %s", transport)
		glog.Errorf("Authentication failed: %v", err)
		return nil, err
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
