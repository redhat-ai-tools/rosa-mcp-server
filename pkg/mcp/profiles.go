package mcp

import (
	"github.com/mark3labs/mcp-go/server"
)

// Profile represents a tool profile following OpenShift MCP patterns
type Profile interface {
	GetName() string
	GetDescription() string
	GetTools(s *Server) []server.ServerTool
}

// DefaultProfile implements the default profile with all ROSA HCP tools enabled
type DefaultProfile struct{}

func (p *DefaultProfile) GetName() string {
	return "default"
}

func (p *DefaultProfile) GetDescription() string {
	return "Default profile with all ROSA HCP tools enabled"
}

func (p *DefaultProfile) GetTools(s *Server) []server.ServerTool {
	return s.initTools()
}

// GetDefaultProfile returns the default profile instance
func GetDefaultProfile() Profile {
	return &DefaultProfile{}
}