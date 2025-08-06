package version

// Version information
var (
	Version = "0.1.0"
	Build   = "dev"
)

// GetVersion returns the version string
func GetVersion() string {
	return Version + "-" + Build
}