package ocm

// Client wraps the OCM SDK client
type Client struct {
	baseURL string
}

// NewClient creates a new OCM client wrapper
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

// WithToken creates a new client with authentication token
func (c *Client) WithToken(token string) *Client {
	// TODO: Implement OCM SDK integration with token
	return c
}