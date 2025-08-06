package mcp

// Profile represents a tool profile
type Profile struct {
	Name  string
	Tools []string
}

// GetDefaultProfile returns the default profile with all tools enabled
func GetDefaultProfile() Profile {
	return Profile{
		Name:  "default",
		Tools: []string{"whoami", "get_clusters", "get_cluster", "create_rosa_hcp_cluster"},
	}
}