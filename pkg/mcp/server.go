package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tiwillia/rosa-mcp-go/pkg/config"
	"github.com/tiwillia/rosa-mcp-go/pkg/ocm"
)

// Server represents the MCP server
type Server struct {
	mcpServer *server.MCPServer
	ocmClient *ocm.Client
	config    *config.Configuration
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Configuration) *Server {
	s := &Server{
		config: cfg,
	}

	// Create MCP server following OpenShift MCP patterns
	mcpServer := server.NewMCPServer(
		"rosa-mcp-server",
		"0.1.0",
		server.WithLogging(),
	)

	s.mcpServer = mcpServer
	s.registerTools()

	return s
}

// Start starts the MCP server
func (s *Server) Start() error {
	glog.Infof("Starting ROSA MCP Server with transport: %s", s.config.Transport)

	switch s.config.Transport {
	case "stdio":
		return s.ServeStdio()
	case "sse":
		return s.ServeSSE()
	default:
		return fmt.Errorf("unsupported transport mode: %s", s.config.Transport)
	}
}

// ServeStdio serves the MCP server via stdio transport
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// ServeSSE serves the MCP server via SSE transport
func (s *Server) ServeSSE() error {
	glog.Infof("Starting SSE server on %s:%d", s.config.Host, s.config.Port)
	
	// Create SSE server using mcp-go library
	options := []server.SSEOption{}
	if s.config.SSEBaseURL != "" {
		options = append(options, server.WithBaseURL(s.config.SSEBaseURL))
	}
	
	// Add context function to extract headers from HTTP request
	options = append(options, server.WithSSEContextFunc(s.extractHeadersToContext))
	
	sseServer := server.NewSSEServer(s.mcpServer, options...)
	return sseServer.Start(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
}


// getAuthenticatedOCMClient extracts token from context and creates authenticated OCM client
func (s *Server) getAuthenticatedOCMClient(ctx context.Context) (*ocm.Client, error) {
	// Extract token based on transport mode
	token, err := ocm.ExtractTokenFromContext(ctx, s.config.Transport)
	if err != nil {
		glog.Errorf("Failed to extract OCM token: %v", err)
		return nil, err
	}

	// Create OCM client and authenticate
	baseClient := ocm.NewClient(s.config.OCMBaseURL, s.config.OCMClientID)
	authenticatedClient, err := baseClient.WithToken(token)
	if err != nil {
		authErr := fmt.Errorf("OCM authentication failed: %w", err)
		glog.Errorf("OCM client authentication failed: %v", authErr)
		return nil, authErr
	}

	return authenticatedClient, nil
}

// extractHeadersToContext is an SSE context function that extracts HTTP headers 
// from the request and stores them in the context for later authentication use
func (s *Server) extractHeadersToContext(ctx context.Context, r *http.Request) context.Context {
	glog.V(2).Infof("SSE context function: extracting headers from HTTP request")
	glog.V(3).Infof("Request headers: %+v", r.Header)
	
	// Store the HTTP headers in the context using the same key type that our auth code expects
	// We need to use the contextKey type defined in our auth package
	return context.WithValue(ctx, ocm.RequestHeaderKey(), r.Header)
}

// logToolCall logs tool execution with structured logging
func (s *Server) logToolCall(toolName string, params map[string]interface{}) {
	glog.V(2).Infof("Tool called: %s with params: %v", toolName, params)
}

// convertParamsToMap converts tool parameters to map for logging
func convertParamsToMap(params ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i, param := range params {
		result[fmt.Sprintf("param_%d", i)] = param
	}
	return result
}
