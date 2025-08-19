package htpasswd

import (
	"encoding/base64"
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

// parseHtpasswordFile - copied from rosa/cmd/create/idp/htpasswd.go:281
func parseHtpasswordFile(usersList *map[string]string, fileContent string) error {
	lines := strings.Split(fileContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// split "user:password" at colon
		username, password, found := strings.Cut(line, ":")
		if !found || username == "" || password == "" {
			return fmt.Errorf("Malformed line, Expected: validUsername:validPassword, Got: %s", line)
		}

		(*usersList)[username] = password
	}
	return nil
}

// ProcessUserInput - processes MCP tool parameters using ROSA CLI patterns
func ProcessUserInput(userInput map[string]interface{}) (map[string]string, bool, error) {
	userList := make(map[string]string)
	isHashedPassword := false

	// Option 1: users array (comma-separated list like ROSA CLI)
	if users, ok := userInput["users"].([]interface{}); ok && len(users) > 0 {
		for _, user := range users {
			userStr, ok := user.(string)
			if !ok {
				return nil, false, fmt.Errorf("invalid user format, expected string")
			}
			username, password, found := strings.Cut(userStr, ":")
			if !found {
				return nil, false, fmt.Errorf("users should be provided in format username:password")
			}
			userList[username] = password
		}
		return userList, false, nil
	}

	// Option 2: single username/password (ROSA CLI backward compatibility)
	if username, hasUsername := userInput["username"].(string); hasUsername {
		if password, hasPassword := userInput["password"].(string); hasPassword {
			userList[username] = password
			return userList, false, nil
		}
		return nil, false, fmt.Errorf("password required when username is provided")
	}

	// Option 3: htpasswd file content
	if fileContent, ok := userInput["htpasswd_file_content"].(string); ok && fileContent != "" {
		// Decode base64 content
		decoded, err := base64.StdEncoding.DecodeString(fileContent)
		if err != nil {
			return nil, false, fmt.Errorf("failed to decode htpasswd file content: %w", err)
		}

		// Use ROSA CLI parsing function
		if err := parseHtpasswordFile(&userList, string(decoded)); err != nil {
			return nil, false, fmt.Errorf("failed to parse htpasswd file: %w", err)
		}
		isHashedPassword = true // htpasswd files contain pre-hashed passwords
		return userList, isHashedPassword, nil
	}

	return nil, false, fmt.Errorf("no user input provided: specify 'users', 'username'+'password', or 'htpasswd_file_content'")
}

// ValidateUserCredentials - validates username and password using ROSA CLI validation
func ValidateUserCredentials(username, password string) error {
	return validateHtUsernameAndPassword(username, password)
}