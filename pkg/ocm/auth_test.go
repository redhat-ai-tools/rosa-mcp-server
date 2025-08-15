package ocm

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTokenFromSSELegacy(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		expectedType string
		wantErr  bool
	}{
		{
			name: "valid offline token in header",
			headers: map[string]string{
				SSETokenHeader: "test-token-123",
			},
			expected: "test-token-123",
			expectedType: "offline",
			wantErr:  false,
		},
		{
			name:     "missing header",
			headers:  map[string]string{},
			expected: "",
			wantErr:  true,
		},
		{
			name: "empty token in header",
			headers: map[string]string{
				SSETokenHeader: "",
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "header with different case",
			headers: map[string]string{
				"x-ocm-offline-token": "test-token-789",
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenInfo, err := ExtractTokenInfoFromSSE(tt.headers)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tokenInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenInfo)
				assert.Equal(t, tt.expected, tokenInfo.Token)
				assert.Equal(t, tt.expectedType, tokenInfo.TokenType)
			}
		})
	}
}

func TestExtractTokenFromStdio(t *testing.T) {
	// Save original environment variable
	originalToken := os.Getenv(StdioTokenEnv)
	defer func() {
		if originalToken != "" {
			os.Setenv(StdioTokenEnv, originalToken)
		} else {
			os.Unsetenv(StdioTokenEnv)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid token in environment",
			envValue: "env-token-123",
			expected: "env-token-123",
			wantErr:  false,
		},
		{
			name:     "empty environment variable",
			envValue: "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(StdioTokenEnv, tt.envValue)
			} else {
				os.Unsetenv(StdioTokenEnv)
			}

			token, err := ExtractTokenFromStdio()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTokenFromStdio() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token != tt.expected {
				t.Errorf("ExtractTokenFromStdio() = %v, expected %v", token, tt.expected)
			}
		})
	}
}

func TestExtractHeadersFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected map[string]string
	}{
		{
			name: "headers stored in context",
			ctx: context.WithValue(context.Background(), "headers", map[string]string{
				"X-OCM-OFFLINE-TOKEN": "context-token",
				"Content-Type":        "application/json",
			}),
			expected: map[string]string{
				"X-OCM-OFFLINE-TOKEN": "context-token",
				"Content-Type":        "application/json",
			},
		},
		{
			name: "http.request in context",
			ctx: func() context.Context {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("X-OCM-OFFLINE-TOKEN", "request-token")
				req.Header.Set("Authorization", "Bearer xyz")
				return context.WithValue(context.Background(), "http.request", req)
			}(),
			expected: map[string]string{
				"X-Ocm-Offline-Token": "request-token",
				"Authorization":       "Bearer xyz",
			},
		},
		{
			name: "request in context",
			ctx: func() context.Context {
				req, _ := http.NewRequest("POST", "/api", nil)
				req.Header.Set("X-OCM-OFFLINE-TOKEN", "api-token")
				return context.WithValue(context.Background(), "request", req)
			}(),
			expected: map[string]string{
				"X-Ocm-Offline-Token": "api-token",
			},
		},
		{
			name:     "no headers in context",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name:     "invalid type in context",
			ctx:      context.WithValue(context.Background(), "headers", "not-a-map"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := extractHeadersFromContext(tt.ctx)
			if headers == nil && tt.expected == nil {
				return
			}
			if headers == nil || tt.expected == nil {
				t.Errorf("extractHeadersFromContext() = %v, expected %v", headers, tt.expected)
				return
			}
			for key, expectedValue := range tt.expected {
				if actualValue, exists := headers[key]; !exists || actualValue != expectedValue {
					t.Errorf("extractHeadersFromContext()[%s] = %v, expected %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	// Create realistic JWT tokens for testing
	// Valid token: exp=1765774800 (Dec 15, 2025), iat=1755230400 (Aug 15, 2025)
	validAccessToken := "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJoVGxRM0JnRVFqNFJvR05xc1JlZkE1VER6Z3EifQ.eyJleHAiOjE3NjU3NzQ4MDAsImlhdCI6MTc1NTIzMDQwMCwiYXV0aF90aW1lIjoxNzU1MjMwNDAwLCJqdGkiOiI2N2ZlNTE4MS05NTkyLTQzOTAtOWJmOC04ZGIxYTc1OTQzYzciLCJpc3MiOiJodHRwczovL3Nzby5yZWRoYXQuY29tL2F1dGgvcmVhbG1zL3JlZGhhdGV4dGVybmFsIiwiYXVkIjoiY2xvdWQtc2VydmljZXMiLCJzdWIiOiJmOjRjNjE5YWI0LTI4NWMtNGFhOC1hZGQ4LTY5ZTcyN2M4YmM3NDp0ZXN0dXNlciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImNsb3VkLXNlcnZpY2VzIiwic2Vzc2lvbl9zdGF0ZSI6IjY3ZmU1MTgxLTk1OTItNDM5MC05YmY4LThkYjFhNzU5NDNjNyIsImFjciI6IjEiLCJzY29wZSI6Im9wZW5pZCBhcGkuaWFtLnNlcnZpY2VfYWNjb3VudHMifQ.fakesignature"
	// Expired token: exp=1634171069 (Oct 14, 2021), iat=1634167469 (Oct 14, 2021)
	expiredAccessToken := "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJoVGxRM0JnRVFqNFJvR05xc1JlZkE1VER6Z3EifQ.eyJleHAiOjE2MzQxNzEwNjksImlhdCI6MTYzNDE2NzQ2OSwiYXV0aF90aW1lIjoxNjM0MTY3NDY5LCJqdGkiOiI2N2ZlNTE4MS05NTkyLTQzOTAtOWJmOC04ZGIxYTc1OTQzYzciLCJpc3MiOiJodHRwczovL3Nzby5yZWRoYXQuY29tL2F1dGgvcmVhbG1zL3JlZGhhdGV4dGVybmFsIiwiYXVkIjoiY2xvdWQtc2VydmljZXMiLCJzdWIiOiJmOjRjNjE5YWI0LTI4NWMtNGFhOC1hZGQ4LTY5ZTcyN2M4YmM3NDp0ZXN0dXNlciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImNsb3VkLXNlcnZpY2VzIiwic2Vzc2lvbl9zdGF0ZSI6IjY3ZmU1MTgxLTk1OTItNDM5MC05YmY4LThkYjFhNzU5NDNjNyIsImFjciI6IjEiLCJzY29wZSI6Im9wZW5pZCBhcGkuaWFtLnNlcnZpY2VfYWNjb3VudHMifQ.fakesignature"

	tests := []struct {
		name        string
		headers     map[string]string
		expected    string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			headers:     map[string]string{"Authorization": "Bearer " + validAccessToken},
			expected:    validAccessToken,
			expectError: false,
		},
		{
			name:        "case insensitive bearer",
			headers:     map[string]string{"Authorization": "bearer " + expiredAccessToken},
			expected:    expiredAccessToken,
			expectError: false,
		},
		{
			name:        "missing authorization header",
			headers:     map[string]string{},
			expectError: true,
		},
		{
			name:        "invalid scheme",
			headers:     map[string]string{"Authorization": "Basic " + validAccessToken},
			expectError: true,
		},
		{
			name:        "empty token",
			headers:     map[string]string{"Authorization": "Bearer "},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractBearerToken(tt.headers)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTokenPriority(t *testing.T) {
	// Use realistic access token for priority testing
	validAccessToken := "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJoVGxRM0JnRVFqNFJvR05xc1JlZkE1VER6Z3EifQ.eyJleHAiOjE3MzQxNzEwNjksImlhdCI6MTczNDE2NzQ2OSwiYXV0aF90aW1lIjoxNzM0MTY3NDY5LCJqdGkiOiI2N2ZlNTE4MS05NTkyLTQzOTAtOWJmOC04ZGIxYTc1OTQzYzciLCJpc3MiOiJodHRwczovL3Nzby5yZWRoYXQuY29tL2F1dGgvcmVhbG1zL3JlZGhhdGV4dGVybmFsIiwiYXVkIjoiY2xvdWQtc2VydmljZXMiLCJzdWIiOiJmOjRjNjE5YWI0LTI4NWMtNGFhOC1hZGQ4LTY5ZTcyN2M4YmM3NDp0ZXN0dXNlciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImNsb3VkLXNlcnZpY2VzIiwic2Vzc2lvbl9zdGF0ZSI6IjY3ZmU1MTgxLTk1OTItNDM5MC05YmY4LThkYjFhNzU5NDNjNyIsImFjciI6IjEiLCJzY29wZSI6Im9wZW5pZCBhcGkuaWFtLnNlcnZpY2VfYWNjb3VudHMifQ.fakesignature"
	offlineToken := "OCM_OFFLINE_TOKEN_abcd1234567890efghijklmnopqrstuvwxyz"

	headers := map[string]string{
		"Authorization":       "Bearer " + validAccessToken,
		"X-Ocm-Offline-Token": offlineToken,
	}

	tokenInfo, err := ExtractTokenInfoFromSSE(headers)
	assert.NoError(t, err)
	assert.Equal(t, validAccessToken, tokenInfo.Token)
	assert.Equal(t, "access", tokenInfo.TokenType)
}

func TestExtractTokenInfoFromSSE(t *testing.T) {
	validAccessToken := "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJoVGxRM0JnRVFqNFJvR05xc1JlZkE1VER6Z3EifQ.eyJleHAiOjE3NjU3NzQ4MDAsImlhdCI6MTc1NTIzMDQwMCwiYXV0aF90aW1lIjoxNzU1MjMwNDAwLCJqdGkiOiI2N2ZlNTE4MS05NTkyLTQzOTAtOWJmOC04ZGIxYTc1OTQzYzciLCJpc3MiOiJodHRwczovL3Nzby5yZWRoYXQuY29tL2F1dGgvcmVhbG1zL3JlZGhhdGV4dGVybmFsIiwiYXVkIjoiY2xvdWQtc2VydmljZXMiLCJzdWIiOiJmOjRjNjE5YWI0LTI4NWMtNGFhOC1hZGQ4LTY5ZTcyN2M4YmM3NDp0ZXN0dXNlciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImNsb3VkLXNlcnZpY2VzIiwic2Vzc2lvbl9zdGF0ZSI6IjY3ZmU1MTgxLTk1OTItNDM5MC05YmY4LThkYjFhNzU5NDNjNyIsImFjciI6IjEiLCJzY29wZSI6Im9wZW5pZCBhcGkuaWFtLnNlcnZpY2VfYWNjb3VudHMifQ.fakesignature"
	offlineToken := "OCM_OFFLINE_TOKEN_abcd1234567890efghijklmnopqrstuvwxyz"

	tests := []struct {
		name            string
		headers         map[string]string
		expectedToken   string
		expectedType    string
		expectError     bool
	}{
		{
			name:            "access token only",
			headers:         map[string]string{"Authorization": "Bearer " + validAccessToken},
			expectedToken:   validAccessToken,
			expectedType:    "access",
			expectError:     false,
		},
		{
			name:            "offline token only",
			headers:         map[string]string{"X-Ocm-Offline-Token": offlineToken},
			expectedToken:   offlineToken,
			expectedType:    "offline",
			expectError:     false,
		},
		{
			name: "both tokens - access takes priority",
			headers: map[string]string{
				"Authorization":       "Bearer " + validAccessToken,
				"X-Ocm-Offline-Token": offlineToken,
			},
			expectedToken: validAccessToken,
			expectedType:  "access",
			expectError:   false,
		},
		{
			name:        "no tokens",
			headers:     map[string]string{},
			expectError: true,
		},
		{
			name:        "invalid authorization header",
			headers:     map[string]string{"Authorization": "Invalid format"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractTokenInfoFromSSE(tt.headers)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedToken, result.Token)
				assert.Equal(t, tt.expectedType, result.TokenType)
			}
		})
	}
}

func TestExtractTokenInfoFromContext(t *testing.T) {
	// Save original environment variable
	originalToken := os.Getenv(StdioTokenEnv)
	defer func() {
		if originalToken != "" {
			os.Setenv(StdioTokenEnv, originalToken)
		} else {
			os.Unsetenv(StdioTokenEnv)
		}
	}()

	tests := []struct {
		name             string
		ctx              context.Context
		transport        string
		envValue         string
		expectedToken    string
		expectedType     string
		wantErr          bool
	}{
		{
			name:          "stdio transport with env token",
			ctx:           context.Background(),
			transport:     "stdio",
			envValue:      "stdio-env-token",
			expectedToken: "stdio-env-token",
			expectedType:  "offline",
			wantErr:       false,
		},
		{
			name:      "stdio transport without env token",
			ctx:       context.Background(),
			transport: "stdio",
			envValue:  "",
			wantErr:   true,
		},
		{
			name: "sse transport with authorization header",
			ctx: func() context.Context {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.fakesig")
				return context.WithValue(context.Background(), "http.request", req)
			}(),
			transport:     "sse",
			envValue:      "",
			expectedToken: "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.fakesig",
			expectedType:  "access",
			wantErr:       false,
		},
		{
			name: "sse transport with offline token header",
			ctx: func() context.Context {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("X-OCM-OFFLINE-TOKEN", "sse-request-token")
				return context.WithValue(context.Background(), "http.request", req)
			}(),
			transport:     "sse",
			envValue:      "",
			expectedToken: "sse-request-token",
			expectedType:  "offline",
			wantErr:       false,
		},
		{
			name:          "sse transport fallback to env",
			ctx:           context.Background(),
			transport:     "sse",
			envValue:      "sse-fallback-token",
			expectedToken: "sse-fallback-token",
			expectedType:  "offline",
			wantErr:       false,
		},
		{
			name:      "sse transport no token available",
			ctx:       context.Background(),
			transport: "sse",
			envValue:  "",
			wantErr:   true,
		},
		{
			name:      "unsupported transport",
			ctx:       context.Background(),
			transport: "websocket",
			envValue:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(StdioTokenEnv, tt.envValue)
			} else {
				os.Unsetenv(StdioTokenEnv)
			}

			tokenInfo, err := ExtractTokenInfoFromContext(tt.ctx, tt.transport)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tokenInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenInfo)
				assert.Equal(t, tt.expectedToken, tokenInfo.Token)
				assert.Equal(t, tt.expectedType, tokenInfo.TokenType)
			}
		})
	}
}
