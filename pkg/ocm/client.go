package ocm

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
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