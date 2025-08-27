package mcp

import (
	"fmt"
	"strings"
	"time"

	accountsmgmt "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	clustersmgmt "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// formatAccountResponse formats account information for display
func formatAccountResponse(account *accountsmgmt.Account) string {
	if account == nil {
		return "No account information available"
	}

	var parts []string
	parts = append(parts, "=== Account Information ===")
	
	if username := account.Username(); username != "" {
		parts = append(parts, fmt.Sprintf("Username: %s", username))
	}
	
	if email := account.Email(); email != "" {
		parts = append(parts, fmt.Sprintf("Email: %s", email))
	}
	
	if firstName := account.FirstName(); firstName != "" {
		parts = append(parts, fmt.Sprintf("First Name: %s", firstName))
	}
	
	if lastName := account.LastName(); lastName != "" {
		parts = append(parts, fmt.Sprintf("Last Name: %s", lastName))
	}
	
	if orgID := account.Organization(); orgID != nil && orgID.ID() != "" {
		parts = append(parts, fmt.Sprintf("Organization ID: %s", orgID.ID()))
		if orgName := orgID.Name(); orgName != "" {
			parts = append(parts, fmt.Sprintf("Organization Name: %s", orgName))
		}
	}
	
	if id := account.ID(); id != "" {
		parts = append(parts, fmt.Sprintf("Account ID: %s", id))
	}
	
	return strings.Join(parts, "\n")
}

// formatClustersResponse formats cluster list for display  
func formatClustersResponse(clusters []*clustersmgmt.Cluster) string {
	if len(clusters) == 0 {
		return "No clusters found"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("=== Clusters (%d found) ===", len(clusters)))
	
	for i, cluster := range clusters {
		if i > 0 {
			parts = append(parts, "---")
		}
		
		parts = append(parts, fmt.Sprintf("Name: %s", cluster.Name()))
		parts = append(parts, fmt.Sprintf("ID: %s", cluster.ID()))
		parts = append(parts, fmt.Sprintf("State: %s", cluster.State()))
		
		if api := cluster.API(); api != nil && api.URL() != "" {
			parts = append(parts, fmt.Sprintf("API URL: %s", api.URL()))
		}
		
		if console := cluster.Console(); console != nil && console.URL() != "" {
			parts = append(parts, fmt.Sprintf("Console URL: %s", console.URL()))
		}
		
		if version := cluster.Version(); version != nil && version.ID() != "" {
			parts = append(parts, fmt.Sprintf("Version: %s", version.ID()))
		}
		
		if region := cluster.Region(); region != nil && region.ID() != "" {
			parts = append(parts, fmt.Sprintf("Region: %s", region.ID()))
		}
	}
	
	return strings.Join(parts, "\n")
}

// formatClusterResponse formats single cluster details for display
func formatClusterResponse(cluster *clustersmgmt.Cluster) string {
	if cluster == nil {
		return "No cluster information available"
	}

	var parts []string
	parts = append(parts, "=== Cluster Details ===")
	
	parts = append(parts, fmt.Sprintf("Name: %s", cluster.Name()))
	parts = append(parts, fmt.Sprintf("ID: %s", cluster.ID()))
	parts = append(parts, fmt.Sprintf("State: %s", cluster.State()))
	
	if api := cluster.API(); api != nil && api.URL() != "" {
		parts = append(parts, fmt.Sprintf("API URL: %s", api.URL()))
	}
	
	if console := cluster.Console(); console != nil && console.URL() != "" {
		parts = append(parts, fmt.Sprintf("Console URL: %s", console.URL()))
	}
	
	if version := cluster.Version(); version != nil && version.ID() != "" {
		parts = append(parts, fmt.Sprintf("Version: %s", version.ID()))
	}
	
	if region := cluster.Region(); region != nil && region.ID() != "" {
		parts = append(parts, fmt.Sprintf("Region: %s", region.ID()))
	}
	
	if aws := cluster.AWS(); aws != nil {
		parts = append(parts, "--- AWS Configuration ---")
		if accountID := aws.AccountID(); accountID != "" {
			parts = append(parts, fmt.Sprintf("AWS Account ID: %s", accountID))
		}
		if billingAccountID := aws.BillingAccountID(); billingAccountID != "" {
			parts = append(parts, fmt.Sprintf("Billing Account ID: %s", billingAccountID))
		}
		if sts := aws.STS(); sts != nil {
			if roleArn := sts.RoleARN(); roleArn != "" {
				parts = append(parts, fmt.Sprintf("Role ARN: %s", roleArn))
			}
			if operatorRolePrefix := sts.OperatorRolePrefix(); operatorRolePrefix != "" {
				parts = append(parts, fmt.Sprintf("Operator Role Prefix: %s", operatorRolePrefix))
			}
		}
		if len(aws.SubnetIDs()) > 0 {
			parts = append(parts, fmt.Sprintf("Subnet IDs: %s", strings.Join(aws.SubnetIDs(), ", ")))
		}
	}
	
	if hypershift := cluster.Hypershift(); hypershift != nil {
		parts = append(parts, fmt.Sprintf("Hypershift Enabled: %t", hypershift.Enabled()))
	}
	
	if ccs := cluster.CCS(); ccs != nil {
		parts = append(parts, fmt.Sprintf("CCS Enabled: %t", ccs.Enabled()))
	}
	
	if billingModel := cluster.BillingModel(); billingModel != "" {
		parts = append(parts, fmt.Sprintf("Billing Model: %s", billingModel))
	}
	
	if creationTime := cluster.CreationTimestamp(); !creationTime.IsZero() {
		parts = append(parts, fmt.Sprintf("Created: %s", creationTime.Format(time.RFC3339)))
	}
	
	return strings.Join(parts, "\n")
}

// formatClusterCreateResponse formats cluster creation response for display
func formatClusterCreateResponse(cluster *clustersmgmt.Cluster) string {
	if cluster == nil {
		return "Cluster creation failed - no cluster information returned"
	}

	var parts []string
	parts = append(parts, "=== Cluster Creation Response ===")
	
	parts = append(parts, fmt.Sprintf("✓ Cluster created successfully"))
	parts = append(parts, fmt.Sprintf("Name: %s", cluster.Name()))
	parts = append(parts, fmt.Sprintf("ID: %s", cluster.ID()))
	parts = append(parts, fmt.Sprintf("State: %s", cluster.State()))
	
	if region := cluster.Region(); region != nil && region.ID() != "" {
		parts = append(parts, fmt.Sprintf("Region: %s", region.ID()))
	}
	
	if creationTime := cluster.CreationTimestamp(); !creationTime.IsZero() {
		parts = append(parts, fmt.Sprintf("Creation Started: %s", creationTime.Format(time.RFC3339)))
	}
	
	parts = append(parts, "")
	parts = append(parts, "Note: Cluster provisioning is in progress. Use 'get_cluster' to check status.")
	
	if api := cluster.API(); api != nil && api.URL() != "" {
		parts = append(parts, fmt.Sprintf("API URL will be: %s", api.URL()))
	}
	
	if console := cluster.Console(); console != nil && console.URL() != "" {
		parts = append(parts, fmt.Sprintf("Console URL will be: %s", console.URL()))
	}
	
	return strings.Join(parts, "\n")
}

// FormatHTPasswdIdentityProviderResult - Enhanced with ROSA CLI patterns
func FormatHTPasswdIdentityProviderResult(
	idp *clustersmgmt.IdentityProvider,
	cluster *clustersmgmt.Cluster,
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