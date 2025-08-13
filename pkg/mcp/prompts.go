package mcp

import (
	"context"
	_ "embed"

	"github.com/mark3labs/mcp-go/mcp"
)

//go:embed prompts/rosa-hcp-prerequisites-guide.md
var prereqsGuide string

// registerPrompts registers all MCP prompts with the server
func (s *Server) registerPrompts() {
	// ROSA HCP Prerequisites Guide prompt
	s.mcpServer.AddPrompt(mcp.NewPrompt("rosa_hcp_prerequisites_guide",
		mcp.WithPromptDescription("Comprehensive guidance on ROSA HCP cluster creation prerequisites and setup steps"),
	), s.handleROSAHCPPrereqsPrompt)
}

// handleROSAHCPPrereqsPrompt provides comprehensive ROSA HCP prerequisites guidance as a prompt
func (s *Server) handleROSAHCPPrereqsPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return mcp.NewGetPromptResult(
		"ROSA HCP Prerequisites Guide",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(prereqsGuide),
			),
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewResourceLink(
					"https://cloud.redhat.com/learning/learn:getting-started-red-hat-openshift-service-aws-rosa/resource/resources:creating-rosa-hcp-clusters-using-default-options#page-title",
					"ROSA HCP Documentation",
					"Official Red Hat documentation for creating ROSA HCP clusters using default options",
					"text/html",
				),
			),
		},
	), nil
}
