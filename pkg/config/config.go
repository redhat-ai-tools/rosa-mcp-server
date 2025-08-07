package config

import (
	"github.com/BurntSushi/toml"
)

// Configuration holds the server configuration
type Configuration struct {
	// OCM API base URL (default: https://api.openshift.com)
	OCMBaseURL string `toml:"ocm_base_url"`
	
	// OCM Client ID (default: cloud-services)
	OCMClientID string `toml:"ocm_client_id"`
	
	// Transport mode selection (stdio/SSE)
	Transport string `toml:"transport"`
	
	// Optional: Host for SSE/HTTP transport
	Host string `toml:"host"`
	
	// Optional: Port for SSE/HTTP transport
	Port int `toml:"port"`
	
	// Optional: SSE base URL for public endpoints
	SSEBaseURL string `toml:"sse_base_url"`
}

// NewConfiguration creates a new configuration with defaults
func NewConfiguration() *Configuration {
	return &Configuration{
		OCMBaseURL:  "https://api.openshift.com",
		OCMClientID: "cloud-services",
		Transport:   "stdio",
		Host:        "0.0.0.0",
		Port:        8080,
	}
}

// LoadFromFile loads configuration from a TOML file
func LoadFromFile(path string) (*Configuration, error) {
	config := NewConfiguration()
	if _, err := toml.DecodeFile(path, config); err != nil {
		return nil, err
	}
	return config, nil
}