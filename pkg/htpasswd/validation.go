package htpasswd

import (
	"fmt"
	"regexp"
	"strings"

	passwordValidator "github.com/openshift-online/ocm-common/pkg/idp/validations"
)

const ClusterAdminUsername = "cluster-admin"

var idRE = regexp.MustCompile(`(?i)^[0-9a-z]+([-_][0-9a-z]+)*$`)

// UsernameValidator - copied from rosa/cmd/create/idp/htpasswd.go:259
func UsernameValidator(val interface{}) error {
	if username, ok := val.(string); ok {
		if strings.ContainsAny(username, "/:%") {
			return fmt.Errorf("invalid username '%s': "+
				"username must not contain /, :, or %%", username)
		}
		return nil
	}
	return fmt.Errorf("can only validate strings, got '%v'", val)
}

// clusterAdminValidator - copied from rosa/cmd/create/idp/htpasswd.go:270
func clusterAdminValidator(val interface{}) error {
	if username, ok := val.(string); ok {
		if username == ClusterAdminUsername {
			return fmt.Errorf("username '%s' is not allowed. It is preserved for cluster admin creation", username)
		}
		return nil
	}
	return fmt.Errorf("can only validate strings, got '%v'", val)
}

// ValidateIdpName - copied from rosa/cmd/create/idp/cmd.go:448
func ValidateIdpName(idpName interface{}) error {
	name, ok := idpName.(string)
	if !ok {
		return fmt.Errorf("Invalid type for identity provider name. Expected a string, got %T", idpName)
	}

	if !idRE.MatchString(name) {
		return fmt.Errorf("Invalid identifier '%s' for 'name'", idpName)
	}

	if strings.EqualFold(name, "cluster-admin") {
		return fmt.Errorf("The name \"cluster-admin\" is reserved for admin user IDP")
	}
	return nil
}

// validateHtUsernameAndPassword - copied from rosa/cmd/create/idp/htpasswd.go:318
func validateHtUsernameAndPassword(username, password string) error {
	err := UsernameValidator(username)
	if err != nil {
		return err
	}
	err = clusterAdminValidator(username)
	if err != nil {
		return err
	}
	err = passwordValidator.PasswordValidator(password)
	if err != nil {
		return err
	}
	return nil
}

// ProcessUserInput - processes MCP tool parameters using ROSA CLI patterns (simplified to users array only)
func ProcessUserInput(userInput map[string]interface{}) (map[string]string, error) {
	userList := make(map[string]string)

	// Only support users array format
	users, ok := userInput["users"].([]interface{})
	if !ok || len(users) == 0 {
		return nil, fmt.Errorf("'users' parameter is required: provide list of username:password pairs")
	}

	for _, user := range users {
		userStr, ok := user.(string)
		if !ok {
			return nil, fmt.Errorf("invalid user format, expected string")
		}
		username, password, found := strings.Cut(userStr, ":")
		if !found || username == "" || password == "" {
			return nil, fmt.Errorf("users should be provided in format username:password")
		}
		userList[username] = password
	}
	
	return userList, nil
}

// ValidateUserCredentials - validates username and password using ROSA CLI validation
func ValidateUserCredentials(username, password string) error {
	return validateHtUsernameAndPassword(username, password)
}