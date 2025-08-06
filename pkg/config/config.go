package config

// Configuration holds the server configuration
type Configuration struct {
	// OCM API base URL (default: https://api.openshift.com)
	OCMBaseURL string
	
	// Transport mode selection (stdio/SSE)
	Transport string
	
	// Optional: Port for SSE/HTTP transport
	Port int
	
	// Optional: SSE base URL for public endpoints
	SSEBaseURL string
}

// NewConfiguration creates a new configuration with defaults
func NewConfiguration() *Configuration {
	return &Configuration{
		OCMBaseURL: "https://api.openshift.com",
		Transport:  "stdio",
		Port:       8080,
	}
}