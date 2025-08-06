package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/redhat-ai-tools/rosa-mcp-go/pkg/config"
	"github.com/redhat-ai-tools/rosa-mcp-go/pkg/mcp"
	"github.com/redhat-ai-tools/rosa-mcp-go/pkg/version"
)

var (
	configFile   string
	transport    string
	ocmBaseURL   string
	port         int
	sseBaseURL   string
)

var rootCmd = &cobra.Command{
	Use:   "rosa-mcp-server",
	Short: "ROSA HCP MCP Server",
	Long:  "A Model Context Protocol server for ROSA HCP (Red Hat OpenShift on AWS using Hosted Control Planes)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg *config.Configuration
		var err error

		if configFile != "" {
			cfg, err = config.LoadFromFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
		} else {
			cfg = config.NewConfiguration()
		}

		// Override config with command line flags
		if cmd.Flags().Changed("transport") {
			cfg.Transport = transport
		}
		if cmd.Flags().Changed("ocm-base-url") {
			cfg.OCMBaseURL = ocmBaseURL
		}
		if cmd.Flags().Changed("port") {
			cfg.Port = port
		}
		if cmd.Flags().Changed("sse-base-url") {
			cfg.SSEBaseURL = sseBaseURL
		}

		// Create and start MCP server
		server := mcp.NewServer(cfg)
		return server.Start()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ROSA MCP Server %s\n", version.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().StringVar(&configFile, "config", "", "path to configuration file")
	rootCmd.Flags().StringVar(&transport, "transport", "stdio", "transport mode (stdio/sse)")
	rootCmd.Flags().StringVar(&ocmBaseURL, "ocm-base-url", "https://api.openshift.com", "OCM API base URL")
	rootCmd.Flags().IntVar(&port, "port", 8080, "port for SSE transport")
	rootCmd.Flags().StringVar(&sseBaseURL, "sse-base-url", "", "SSE base URL for public endpoints")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}