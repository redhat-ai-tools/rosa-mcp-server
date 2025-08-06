package mcp

import (
	"github.com/redhat-ai-tools/rosa-mcp-go/pkg/config"
	"github.com/redhat-ai-tools/rosa-mcp-go/pkg/ocm"
)

// Server represents the MCP server
type Server struct {
	ocmClient *ocm.Client
	config    *config.Configuration
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Configuration) *Server {
	return &Server{
		config: cfg,
	}
}

// Start starts the MCP server
func (s *Server) Start() error {
	// TODO: Implement MCP server startup
	return nil
}