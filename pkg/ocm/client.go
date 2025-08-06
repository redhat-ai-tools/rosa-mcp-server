package ocm

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/errors"
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