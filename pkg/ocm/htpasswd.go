package ocm

import (
	"fmt"

	idputils "github.com/openshift-online/ocm-common/pkg/idp/utils"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	
	"github.com/tiwillia/rosa-mcp-go/pkg/htpasswd"
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
	_, err := c.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not accessible: %w", err)
	}

	// Step 2: Validate IDP name using ROSA CLI validation
	if err := htpasswd.ValidateIdpName(name); err != nil {
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

	// Step 4: Process user input using simplified validation
	userList, err := htpasswd.ProcessUserInput(userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to process user input: %w", err)
	}

	// Step 5: Build HTPasswd user list (always hash passwords)
	htpasswdUsers := []*cmv1.HTPasswdUserBuilder{}
	for username, password := range userList {
		// Validate each user using ROSA CLI validation
		if err := htpasswd.ValidateUserCredentials(username, password); err != nil {
			return nil, fmt.Errorf("invalid user credentials for '%s': %w", username, err)
		}

		// Always hash passwords using ROSA CLI method
		hashedPwd, err := idputils.GenerateHTPasswdCompatibleHash(password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password for user '%s': %w", username, err)
		}
		
		userBuilder := cmv1.NewHTPasswdUser().
			Username(username).
			HashedPassword(hashedPwd)
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