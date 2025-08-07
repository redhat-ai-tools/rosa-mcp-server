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
	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	// Create SSE server similar
	sseServer := s.ServeSse(s.config.SSEBaseURL, httpServer)

	// Register SSE endpoints
	mux.Handle("/sse", sseServer)
	mux.Handle("/message", sseServer)

	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	glog.Infof("Starting SSE server on port %d", s.config.Port)
	return httpServer.ListenAndServe()
}

// ServeSse creates SSE server
func (s *Server) ServeSse(baseURL string, httpServer *http.Server) http.Handler {
	options := []server.SSEOption{
		server.WithHTTPServer(httpServer),
	}
	if baseURL != "" {
		options = append(options, server.WithBaseURL(baseURL))
	}
	return server.NewSSEServer(s.mcpServer, options...)
}

// getAuthenticatedOCMClient extracts token from context and creates authenticated OCM client
func (s *Server) getAuthenticatedOCMClient(ctx context.Context) (*ocm.Client, error) {
	// Extract token based on transport mode
	token, err := ocm.ExtractTokenFromContext(ctx, s.config.Transport)
	if err != nil {
		return nil, err
	}

	// Create OCM client and authenticate
	baseClient := ocm.NewClient(s.config.OCMBaseURL, s.config.OCMClientID)
	authenticatedClient, err := baseClient.WithToken(token)
	if err != nil {
		return nil, fmt.Errorf("OCM authentication failed: %w", err)
	}

	return authenticatedClient, nil
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
