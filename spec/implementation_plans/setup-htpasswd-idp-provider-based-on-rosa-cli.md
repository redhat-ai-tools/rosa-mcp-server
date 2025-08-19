# MCP Server Implementation Plan: HTPasswd Identity Provider Setup (Using ROSA CLI Libraries)

## Overview

Adding an HTPasswd identity provider is one of the most common post-installation tasks for ROSA HCP clusters, especially for development and testing environments. This implementation leverages proven ROSA CLI libraries while adapting for access token authentication used by the MCP server.

## ROSA CLI Library Integration Strategy

This implementation reuses **60-70%** of ROSA CLI functionality while adapting for access token lifecycle management. Key libraries to integrate:

### External Dependencies to Add
```go
// Add to go.mod
require (
    github.com/openshift-online/ocm-common v0.0.25
)

// Import statements for reused utilities
import (
    idputils "github.com/openshift-online/ocm-common/pkg/idp/utils"
    passwordValidator "github.com/openshift-online/ocm-common/pkg/idp/validations"
)
```

### ROSA CLI Functions to Copy and Adapt
- **Validation**: `UsernameValidator`, `clusterAdminValidator`, `ValidateIdpName`, `validateHtUsernameAndPassword`
- **Parsing**: `parseHtpasswordFile` (htpasswd file support)
- **OCM Methods**: `GetIdentityProviders`, `CreateIdentityProvider`
- **Core Logic**: HTPasswd user list construction pattern

## Implementation Plan

### 1. Tool Definition (`pkg/mcp/tools.go`)

**Tool Name**: `setup_htpasswd_identity_provider`

**Required Parameters**:
- `cluster_id` (string): Target cluster identifier
- `name` (string): Identity provider name (default: "htpasswd")
- `mapping_method` (string): User mapping method - options: "add", "claim", "generate", "lookup" (default: "claim")

**User Input Parameters** (one of the following required):
- `users` (array): List of username:password pairs ["user1:password1", "user2:password2"]
- `username` (string) + `password` (string): Single user credentials (backward compatibility)
- `htpasswd_file_content` (string): Base64-encoded htpasswd file content

**Optional Parameters**:
- `overwrite_existing` (boolean): Whether to overwrite if IDP with same name exists (default: false)

### 2. Validation Functions (Copied from ROSA CLI)

**Source**: Copy from `rosa/cmd/create/idp/htpasswd.go` and `rosa/cmd/create/idp/cmd.go`

```go
// pkg/htpasswd/validation.go - Copied and adapted from ROSA CLI

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
```

### 3. OCM Client Methods (Adapted from ROSA CLI)

**Source**: Copy and adapt from `rosa/pkg/ocm/idps.go`

```go
// pkg/ocm/htpasswd.go - Methods adapted from ROSA CLI

import (
    idputils "github.com/openshift-online/ocm-common/pkg/idp/utils"
    cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// GetIdentityProviders - copied from rosa/pkg/ocm/idps.go:41
func (c *Client) GetIdentityProviders(clusterID string) ([]*cmv1.IdentityProvider, error) {
    response, err := c.connection.ClustersMgmt().V1().
        Clusters().Cluster(clusterID).
        IdentityProviders().
        List().Page(1).Size(-1).
        Send()
    if err != nil {
        return nil, HandleOCMError(err)
    }
    return response.Items().Slice(), nil
}

// CreateIdentityProvider - copied from rosa/pkg/ocm/idps.go:54
func (c *Client) CreateIdentityProvider(clusterID string, idp *cmv1.IdentityProvider) (*cmv1.IdentityProvider, error) {
    response, err := c.connection.ClustersMgmt().V1().
        Clusters().Cluster(clusterID).
        IdentityProviders().
        Add().Body(idp).
        Send()
    if err != nil {
        return nil, HandleOCMError(err)
    }
    return response.Body(), nil
}

// SetupHTPasswdIdentityProvider - main implementation using ROSA CLI patterns
func (c *Client) SetupHTPasswdIdentityProvider(
    clusterID string,
    name string,
    mappingMethod string,
    userInput map[string]interface{}, // MCP parameter input
    overwriteExisting bool,
) (*cmv1.IdentityProvider, error) {

    // Step 1: Validate cluster exists - reusing existing MCP pattern
    cluster, err := c.GetCluster(clusterID)
    if err != nil {
        return nil, fmt.Errorf("cluster not accessible: %w", err)
    }

    // Step 2: Validate IDP name using ROSA CLI validation
    if err := ValidateIdpName(name); err != nil {
        return nil, fmt.Errorf("invalid identity provider name: %w", err)
    }

    // Step 3: Check existing IDPs using ROSA CLI method
    existingIDPs, err := c.GetIdentityProviders(clusterID)
    if err != nil {
        return nil, fmt.Errorf("failed to check existing identity providers: %w", err)
    }

    if !overwriteExisting {
        for _, idp := range existingIDPs {
            if idp.Name() == name {
                return nil, fmt.Errorf("identity provider with name '%s' already exists", name)
            }
        }
    }

    // Step 4: Process user input using ROSA CLI patterns
    userList, isHashedPassword, err := c.getUserListFromInput(userInput)
    if err != nil {
        return nil, fmt.Errorf("failed to process user input: %w", err)
    }

    // Step 5: Build HTPasswd user list using ROSA CLI pattern
    htpasswdUsers := []*cmv1.HTPasswdUserBuilder{}
    for username, password := range userList {
        // Validate each user using ROSA CLI validation
        if err := validateHtUsernameAndPassword(username, password); err != nil {
            return nil, fmt.Errorf("invalid user credentials for '%s': %w", username, err)
        }

        userBuilder := cmv1.NewHTPasswdUser().Username(username)
        if isHashedPassword {
            userBuilder.HashedPassword(password)
        } else {
            // Use ROSA CLI password hashing
            hashedPwd, err := idputils.GenerateHTPasswdCompatibleHash(password)
            if err != nil {
                return nil, fmt.Errorf("failed to hash password for user '%s': %w", username, err)
            }
            userBuilder.HashedPassword(hashedPwd)
        }
        htpasswdUsers = append(htpasswdUsers, userBuilder)
    }

    htpassUserList := cmv1.NewHTPasswdUserList().Items(htpasswdUsers...)

    // Step 6: Build IDP using ROSA CLI pattern
    idpBuilder := cmv1.NewIdentityProvider().
        Type(cmv1.IdentityProviderTypeHtpasswd).
        Name(name).
        MappingMethod(cmv1.IdentityProviderMappingMethod(mappingMethod)).
        Htpasswd(cmv1.NewHTPasswdIdentityProvider().Users(htpassUserList))

    idp, err := idpBuilder.Build()
    if err != nil {
        return nil, fmt.Errorf("failed to build identity provider: %w", err)
    }

    // Step 7: Create IDP using ROSA CLI method
    createdIdp, err := c.CreateIdentityProvider(clusterID, idp)
    if err != nil {
        return nil, fmt.Errorf("failed to create identity provider: %w", err)
    }

    return createdIdp, nil
}

// getUserListFromInput - processes MCP tool parameters using ROSA CLI patterns
func (c *Client) getUserListFromInput(userInput map[string]interface{}) (map[string]string, bool, error) {
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

```

### 4. Response Formatter (`pkg/mcp/formatters.go`)

**Enhanced formatter using ROSA CLI patterns for consistency**

```go
// FormatHTPasswdIdentityProviderResult - Enhanced with ROSA CLI patterns
func FormatHTPasswdIdentityProviderResult(
    idp *cmv1.IdentityProvider,
    cluster *cmv1.Cluster,
    userCount int,
) string {
    var output strings.Builder

    output.WriteString("HTPasswd Identity Provider Setup Complete\n\n")

    // Provider Details (similar to ROSA CLI output format)
    output.WriteString("Provider Details:\n")
    output.WriteString(fmt.Sprintf("- Name: %s\n", idp.Name()))
    output.WriteString("- Type: HTPasswd\n")
    output.WriteString(fmt.Sprintf("- Mapping Method: %s\n", idp.MappingMethod()))
    output.WriteString(fmt.Sprintf("- Users Created: %d\n", userCount))
    output.WriteString(fmt.Sprintf("- Cluster: %s (%s)\n", cluster.Name(), cluster.ID()))
    output.WriteString(fmt.Sprintf("- Status: %s\n", idp.Type()))

    // Next Steps (adapted from ROSA CLI messaging)
    output.WriteString("\nNext Steps:\n")
    output.WriteString("1. Users can now log in using their credentials\n")

    // Add console URL if available (ROSA CLI pattern)
    if cluster.Console() != nil && cluster.Console().URL() != "" {
        output.WriteString(fmt.Sprintf("2. Access cluster console at: %s\n", cluster.Console().URL()))
        output.WriteString(fmt.Sprintf("3. Click on '%s' to log in\n", idp.Name()))
        output.WriteString("4. Use 'oc login' with htpasswd credentials\n")
        output.WriteString("5. Consider setting up RBAC for user permissions\n")
    } else {
        output.WriteString("2. Use 'oc login' with htpasswd credentials\n")
        output.WriteString("3. Console URL will be available when cluster networking is ready\n")
        output.WriteString("4. Consider setting up RBAC for user permissions\n")
    }

    // ROSA CLI compatibility note
    output.WriteString(fmt.Sprintf("\nROSA CLI Equivalent:\n"))
    output.WriteString(fmt.Sprintf("rosa create idp htpasswd --cluster=%s --name=%s --mapping-method=%s\n",
        cluster.ID(), idp.Name(), idp.MappingMethod()))

    return output.String()
}
```

## Enhanced Error Handling (ROSA CLI Compatible)

**Error Categories and ROSA CLI Compatibility**:

### 1. Standard OCM Errors (Using ROSA CLI Error Patterns)
```go
// Cluster not found - adapted from ROSA CLI error handling
if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
    return NewTextResult(fmt.Sprintf(`HTPasswd Identity Provider Setup Failed

Error: Cluster not found or not accessible
Cluster: %s
Operation ID: %s

Troubleshooting:
- Verify cluster ID is correct: %s
- Check OCM permissions for cluster access
- Ensure cluster is in 'Ready' state
- Verify you have access to this cluster

ROSA CLI Check: rosa describe cluster %s`, clusterID, operationID, clusterID, clusterID))
}

// Permission denied - using ROSA CLI messaging patterns
if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "forbidden") {
    return NewTextResult(fmt.Sprintf(`HTPasswd Identity Provider Setup Failed

Error: Insufficient permissions
Cluster: %s
Operation ID: %s

Required Permissions:
- Cluster Editor role in OCM
- Identity Provider management permissions
- Cluster access permissions

Troubleshooting:
- Contact your Red Hat account administrator
- Verify your OCM organization membership
- Check cluster access permissions

ROSA CLI Check: rosa whoami`, clusterID, operationID))
}
```

### 2. Validation Errors (Using ROSA CLI Validation)
```go
// Username validation errors - leveraging copied ROSA CLI functions
if validationErr := validateHtUsernameAndPassword(username, password); validationErr != nil {
    return NewTextResult(fmt.Sprintf(`HTPasswd Identity Provider Setup Failed

Error: Invalid user credentials
Username: %s
Validation Error: %s

Username Requirements (ROSA CLI Compatible):
- Must not contain /, :, or %% characters
- Cannot be 'cluster-admin' (reserved)
- Must be ASCII characters only

Password Requirements:
- Minimum 14 characters
- Include uppercase, lowercase, and numbers/symbols
- No whitespace characters
- ASCII-standard characters only

Fix and retry with valid credentials.`, username, validationErr.Error()))
}

// IDP name validation - using ROSA CLI ValidateIdpName function
if err := ValidateIdpName(name); err != nil {
    return NewTextResult(fmt.Sprintf(`HTPasswd Identity Provider Setup Failed

Error: Invalid identity provider name
Name: %s
Validation Error: %s

Name Requirements (ROSA CLI Compatible):
- Format: ^[0-9a-z]+([-_][0-9a-z]+)*$
- Cannot be 'cluster-admin' (reserved)
- Must be unique within the cluster

ROSA CLI Name Check: rosa list idp --cluster=%s`, name, err.Error(), clusterID))
}
```

### 3. Duplicate IDP Error (ROSA CLI Pattern)
```go
// Existing IDP check - using ROSA CLI GetIdentityProviders method
existingIDPs, err := c.GetIdentityProviders(clusterID)
for _, existingIDP := range existingIDPs {
    if existingIDP.Name() == name {
        return NewTextResult(fmt.Sprintf(`HTPasswd Identity Provider Setup Failed

Error: Identity provider with name '%s' already exists
Cluster: %s
Existing Type: %s

Resolution Options:
1. Choose a different name for the new IDP
2. Use --overwrite_existing=true to replace the existing IDP
3. Delete existing IDP first: rosa delete idp %s --cluster=%s

Existing Identity Providers:
%s

ROSA CLI List: rosa list idp --cluster=%s`,
            name, clusterID, existingIDP.Type(), name, clusterID,
            formatExistingIDPs(existingIDPs), clusterID))
    }
}
```

### 4. General OCM API Errors (Standard Handling)
```go
func handleOCMError(err error, clusterID, operationID string) string {
    ocmErr := HandleOCMError(err) // Use existing MCP error handling

    return fmt.Sprintf(`HTPasswd Identity Provider Setup Failed

Error: %s
Cluster: %s
Operation ID: %s

General Troubleshooting:
- Check cluster status: rosa describe cluster %s
- Verify OCM connectivity: rosa whoami
- Review cluster permissions
- Check Red Hat service status

ROSA CLI Alternative:
rosa create idp htpasswd --cluster=%s --interactive`,
        ocmErr.Error(), clusterID, operationID, clusterID, clusterID)
}
```

## Security Considerations (ROSA CLI Compatible)

### 1. Password Handling (Using ROSA CLI Libraries)
```go
// Use ROSA CLI's proven password hashing from ocm-common
import idputils "github.com/openshift-online/ocm-common/pkg/idp/utils"

// Same bcrypt hashing used by ROSA CLI and Red Hat Cluster Services
hashedPwd, err := idputils.GenerateHTPasswdCompatibleHash(password)
// Uses bcrypt.DefaultCost for consistency with ROSA CLI
```

### 2. Credential Validation (ROSA CLI Standards)
```go
// Use ROSA CLI validation functions for consistency
import passwordValidator "github.com/openshift-online/ocm-common/pkg/idp/validations"

// Same password requirements as ROSA CLI:
// - Minimum 14 characters, uppercase, lowercase, numbers/symbols
// - No whitespace, ASCII-standard characters only
err := passwordValidator.PasswordValidator(password)

// Same username validation as ROSA CLI:
// - No /, :, or % characters
// - Cannot be 'cluster-admin'
err := validateHtUsernameAndPassword(username, password)
```

### 3. Token Security (Standard)
- **Secure Error Messages**: Never log token values, only token types
- **Standard OCM SDK Handling**: Let OCM SDK handle token lifecycle
- **Token Type Detection**: Distinguish between access and offline tokens for logging

### 4. Permission Checks (OCM Standard)
- **OCM Authentication**: Same OCM permissions as ROSA CLI
- **Audit Logging**: All operations logged via OCM SDK
- **Operation IDs**: Track operations for support and debugging

## Enhanced Testing Strategy (ROSA CLI Integration)

### 1. Unit Tests with ROSA CLI Function Coverage
```go
// Test all copied ROSA CLI validation functions
func TestUsernameValidator(t *testing.T) {
    // Test cases from ROSA CLI test suite
    testCases := []struct{
        username string
        expectError bool
    }{
        {"valid-user", false},
        {"user:with:colons", true},
        {"user/with/slashes", true},
        {"cluster-admin", true},
    }
    // ... test implementation
}

func TestPasswordHashing(t *testing.T) {
    // Verify compatibility with ROSA CLI password hashing
    password := "ValidPassword123!"
    hash1, _ := idputils.GenerateHTPasswdCompatibleHash(password)
    hash2, _ := idputils.GenerateHTPasswdCompatibleHash(password)

    // Hashes should be different (bcrypt includes salt)
    assert.NotEqual(t, hash1, hash2)

    // But both should verify against original password
    assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password)))
    assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password)))
}
```

### Library Dependencies Summary
**External Libraries Added:**
```go
require (
    github.com/openshift-online/ocm-common v0.0.25  // Password utilities & validation
)
```

**ROSA CLI Functions Copied and Adapted:**
- `UsernameValidator()` - Username format validation
- `clusterAdminValidator()` - Reserved username check
- `ValidateIdpName()` - IDP name validation
- `validateHtUsernameAndPassword()` - Combined credential validation
- `parseHtpasswordFile()` - HTPasswd file parsing
- `GetIdentityProviders()` - OCM IDP listing
- `CreateIdentityProvider()` - OCM IDP creation
- HTPasswd user list construction patterns

**Code Reuse Percentage**: ~70% (high reuse of ROSA CLI validation and OCM patterns)
