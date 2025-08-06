package ocm

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestExtractTokenFromSSE(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		wantErr  bool
	}{
		{
			name: "valid token in header",
			headers: map[string]string{
				SSETokenHeader: "test-token-123",
			},
			expected: "test-token-123",
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
			token, err := ExtractTokenFromSSE(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTokenFromSSE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token != tt.expected {
				t.Errorf("ExtractTokenFromSSE() = %v, expected %v", token, tt.expected)
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

func TestExtractTokenFromContext(t *testing.T) {
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
		name      string
		ctx       context.Context
		transport string
		envValue  string
		expected  string
		wantErr   bool
	}{
		{
			name:      "stdio transport with env token",
			ctx:       context.Background(),
			transport: "stdio",
			envValue:  "stdio-env-token",
			expected:  "stdio-env-token",
			wantErr:   false,
		},
		{
			name:      "stdio transport without env token",
			ctx:       context.Background(),
			transport: "stdio",
			envValue:  "",
			expected:  "",
			wantErr:   true,
		},
		{
			name: "sse transport with http.request",
			ctx: func() context.Context {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("X-OCM-OFFLINE-TOKEN", "sse-request-token")
				return context.WithValue(context.Background(), "http.request", req)
			}(),
			transport: "sse",
			envValue:  "",
			expected:  "sse-request-token",
			wantErr:   false,
		},
		{
			name:      "sse transport fallback to env",
			ctx:       context.Background(),
			transport: "sse",
			envValue:  "sse-fallback-token",
			expected:  "sse-fallback-token",
			wantErr:   false,
		},
		{
			name:      "sse transport no token available",
			ctx:       context.Background(),
			transport: "sse",
			envValue:  "",
			expected:  "",
			wantErr:   true,
		},
		{
			name:      "unsupported transport",
			ctx:       context.Background(),
			transport: "websocket",
			envValue:  "",
			expected:  "",
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

			token, err := ExtractTokenFromContext(tt.ctx, tt.transport)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTokenFromContext() name = %v error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if token != tt.expected {
				t.Errorf("ExtractTokenFromContext() name = %v, token = %v, expected %v", tt.name, token, tt.expected)
			}
		})
	}
}
