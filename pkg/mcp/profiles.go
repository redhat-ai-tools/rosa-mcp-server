/*
profiles implements a basic profile system inspired by the OpenShift MCP server.

The profiles concept allows us to selectively expose different sets of OCM tools
to avoid overloading an LLM's context window with too many tool options. For the MVP,
we implement only a basic "all tools enabled" profile, but this foundation will
allow us to support more complex OCM operations in the future without overwhelming
the AI assistant with excessive tool choices.

Future profile examples could include:
- rosa-classic: Legacy classic ROSA clusters
- rosa-full: All cluster lifecycle operations for all ROSA
- rosa-admin: Administrative, billing, and SRE operations
*/
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
