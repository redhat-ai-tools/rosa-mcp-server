package ocm

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/errors"
	accountsmgmt "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	clustersmgmt "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// Client wraps the OCM SDK client
type Client struct {
	connection *sdk.Connection
	baseURL    string
}

// NewClient creates a new OCM client wrapper
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

// WithToken creates a new client with authentication token
func (c *Client) WithToken(token string) (*Client, error) {
	// Build OCM SDK connection with offline token
	// The OCM SDK uses TokenURL for offline token refresh flow
	builder := sdk.NewConnectionBuilder().
		URL(c.baseURL).
		Tokens(token)

	connection, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build OCM connection: %w", err)
	}

	return &Client{
		connection: connection,
		baseURL:    c.baseURL,
	}, nil
}

// Close closes the OCM connection
func (c *Client) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// OCMError represents an OCM API error with preserved details
type OCMError struct {
	Code        string
	Reason      string
	OperationID string
}

func (e *OCMError) Error() string {
	if e.OperationID != "" {
		return fmt.Sprintf("OCM API Error [%s]: %s (Operation ID: %s)", e.Code, e.Reason, e.OperationID)
	}
	return fmt.Sprintf("OCM API Error [%s]: %s", e.Code, e.Reason)
}

// HandleOCMError converts an OCM SDK error to our error type
func HandleOCMError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's an OCM SDK error with structured details
	if ocmErr, ok := err.(*errors.Error); ok {
		return &OCMError{
			Code:        ocmErr.Code(),
			Reason:      ocmErr.Reason(),
			OperationID: ocmErr.OperationID(),
		}
	}

	// Return original error if not an OCM error
	return err
}

// GetCurrentAccount returns the current authenticated account
func (c *Client) GetCurrentAccount() (*accountsmgmt.Account, error) {
	if c.connection == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	response, err := c.connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		return nil, HandleOCMError(err)
	}

	return response.Body(), nil
}

// GetClusters returns a list of clusters filtered by state
func (c *Client) GetClusters(state string) ([]*clustersmgmt.Cluster, error) {
	if c.connection == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	request := c.connection.ClustersMgmt().V1().Clusters().List()
	
	// Add state filter if provided
	if state != "" {
		request = request.Search(fmt.Sprintf("state = '%s'", state))
	}

	response, err := request.Send()
	if err != nil {
		return nil, HandleOCMError(err)
	}

	clusters := make([]*clustersmgmt.Cluster, 0, response.Size())
	response.Items().Each(func(cluster *clustersmgmt.Cluster) bool {
		clusters = append(clusters, cluster)
		return true
	})

	return clusters, nil
}

// GetCluster returns a single cluster by ID
func (c *Client) GetCluster(clusterID string) (*clustersmgmt.Cluster, error) {
	if c.connection == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	response, err := c.connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
	if err != nil {
		return nil, HandleOCMError(err)
	}

	return response.Body(), nil
}

// CreateROSAHCPCluster creates a new ROSA HCP cluster
func (c *Client) CreateROSAHCPCluster(
	clusterName, awsAccountID, billingAccountID, roleArn,
	operatorRolePrefix, oidcConfigID string,
	subnetIDs []string, region string,
) (*clustersmgmt.Cluster, error) {
	if c.connection == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	// Build ROSA HCP cluster payload following the example structure
	clusterBuilder := clustersmgmt.NewCluster().
		Name(clusterName).
		Product(clustersmgmt.NewProduct().ID("rosa")).
		Region(clustersmgmt.NewCloudRegion().ID(region)).
		AWS(clustersmgmt.NewAWS().
			AccountID(awsAccountID).
			BillingAccountID(billingAccountID).
			STS(clustersmgmt.NewSTS().
				AutoMode(true).
				RoleARN(roleArn).
				OperatorRolePrefix(operatorRolePrefix).
				OidcConfig(clustersmgmt.NewOidcConfig().ID(oidcConfigID))).
			SubnetIDs(subnetIDs...)).
		CCS(clustersmgmt.NewCCS().Enabled(true)).
		Hypershift(clustersmgmt.NewHypershift().Enabled(true)).
		BillingModel("marketplace-aws")

	cluster, err := clusterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build cluster payload: %w", err)
	}

	response, err := c.connection.ClustersMgmt().V1().Clusters().Add().Body(cluster).Send()
	if err != nil {
		return nil, HandleOCMError(err)
	}

	return response.Body(), nil
}