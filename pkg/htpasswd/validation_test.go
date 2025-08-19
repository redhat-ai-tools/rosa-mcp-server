package htpasswd

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsernameValidator(t *testing.T) {
	testCases := []struct {
		username    interface{}
		expectError bool
		description string
	}{
		{"valid-user", false, "valid username with dash"},
		{"valid_user", false, "valid username with underscore"},
		{"validuser123", false, "valid alphanumeric username"},
		{"user:with:colons", true, "username with colons"},
		{"user/with/slashes", true, "username with slashes"},
		{"user%with%percent", true, "username with percent"},
		{"cluster-admin", false, "cluster-admin should pass username validation"},
		{123, true, "non-string input"},
		{"", false, "empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := UsernameValidator(tc.username)
			if tc.expectError {
				assert.Error(t, err, "Expected error for %v", tc.username)
			} else {
				assert.NoError(t, err, "Expected no error for %v", tc.username)
			}
		})
	}
}

func TestClusterAdminValidator(t *testing.T) {
	testCases := []struct {
		username    interface{}
		expectError bool
		description string
	}{
		{"valid-user", false, "valid non-admin username"},
		{"cluster-admin", true, "cluster-admin username"},
		{"CLUSTER-ADMIN", false, "case sensitive check"},
		{123, true, "non-string input"},
		{"", false, "empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := clusterAdminValidator(tc.username)
			if tc.expectError {
				assert.Error(t, err, "Expected error for %v", tc.username)
			} else {
				assert.NoError(t, err, "Expected no error for %v", tc.username)
			}
		})
	}
}

func TestValidateIdpName(t *testing.T) {
	testCases := []struct {
		name        interface{}
		expectError bool
		description string
	}{
		{"valid-name", false, "valid IDP name"},
		{"valid_name", false, "valid IDP name with underscore"},
		{"validname123", false, "valid alphanumeric IDP name"},
		{"123validname", false, "valid name starting with number"},
		{"invalid-name-", true, "name ending with dash"},
		{"invalid_name_", true, "name ending with underscore"},
		{"-invalid", true, "name starting with dash"},
		{"_invalid", true, "name starting with underscore"},
		{"invalid name", true, "name with space"},
		{"cluster-admin", true, "reserved cluster-admin name"},
		{"CLUSTER-ADMIN", true, "reserved cluster-admin name (case insensitive)"},
		{123, true, "non-string input"},
		{"", true, "empty string"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := ValidateIdpName(tc.name)
			if tc.expectError {
				assert.Error(t, err, "Expected error for %v", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for %v", tc.name)
			}
		})
	}
}

func TestValidateUserCredentials(t *testing.T) {
	testCases := []struct {
		username    string
		password    string
		expectError bool
		description string
	}{
		{"validuser", "ValidPassword123!", false, "valid user credentials"},
		{"user:invalid", "ValidPassword123!", true, "invalid username with colon"},
		{"cluster-admin", "ValidPassword123!", true, "reserved username"},
		{"validuser", "short", true, "password too short"},
		{"validuser", "", true, "empty password"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := ValidateUserCredentials(tc.username, tc.password)
			if tc.expectError {
				assert.Error(t, err, "Expected error for username: %s, password: %s", tc.username, tc.password)
			} else {
				assert.NoError(t, err, "Expected no error for username: %s, password: %s", tc.username, tc.password)
			}
		})
	}
}

func TestParseHtpasswordFile(t *testing.T) {
	testCases := []struct {
		content     string
		expected    map[string]string
		expectError bool
		description string
	}{
		{
			"user1:password1\nuser2:password2",
			map[string]string{"user1": "password1", "user2": "password2"},
			false,
			"valid htpasswd content",
		},
		{
			"user1:password1\n\nuser2:password2\n",
			map[string]string{"user1": "password1", "user2": "password2"},
			false,
			"htpasswd content with empty lines",
		},
		{
			"user1:$apr1$hRY7OJWH$km1EYH.UIRjp6CzfZQz/g1",
			map[string]string{"user1": "$apr1$hRY7OJWH$km1EYH.UIRjp6CzfZQz/g1"},
			false,
			"htpasswd with hashed password",
		},
		{
			"invalid_line_without_colon",
			nil,
			true,
			"malformed line without colon",
		},
		{
			"user1:\nuser2:password2",
			nil,
			true,
			"empty password",
		},
		{
			":password1",
			nil,
			true,
			"empty username",
		},
		{
			"",
			map[string]string{},
			false,
			"empty content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			usersList := make(map[string]string)
			err := parseHtpasswordFile(&usersList, tc.content)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for content: %s", tc.content)
			} else {
				assert.NoError(t, err, "Expected no error for content: %s", tc.content)
				assert.Equal(t, tc.expected, usersList, "Expected parsed users to match")
			}
		})
	}
}

func TestProcessUserInput(t *testing.T) {
	testCases := []struct {
		input           map[string]interface{}
		expectedUsers   map[string]string
		expectedHashed  bool
		expectError     bool
		description     string
	}{
		{
			map[string]interface{}{
				"users": []interface{}{"user1:password1", "user2:password2"},
			},
			map[string]string{"user1": "password1", "user2": "password2"},
			false,
			false,
			"users array input",
		},
		{
			map[string]interface{}{
				"username": "testuser",
				"password": "testpassword",
			},
			map[string]string{"testuser": "testpassword"},
			false,
			false,
			"single username/password input",
		},
		{
			map[string]interface{}{
				"htpasswd_file_content": base64.StdEncoding.EncodeToString([]byte("user1:$apr1$hash1\nuser2:$apr1$hash2")),
			},
			map[string]string{"user1": "$apr1$hash1", "user2": "$apr1$hash2"},
			true,
			false,
			"htpasswd file content input",
		},
		{
			map[string]interface{}{
				"username": "testuser",
				// missing password
			},
			nil,
			false,
			true,
			"username without password",
		},
		{
			map[string]interface{}{
				"users": []interface{}{"invalid_user_without_colon"},
			},
			nil,
			false,
			true,
			"invalid users array format",
		},
		{
			map[string]interface{}{
				"htpasswd_file_content": "invalid_base64",
			},
			nil,
			false,
			true,
			"invalid base64 htpasswd content",
		},
		{
			map[string]interface{}{},
			nil,
			false,
			true,
			"no user input provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			users, isHashed, err := ProcessUserInput(tc.input)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for input: %v", tc.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %v", tc.input)
				assert.Equal(t, tc.expectedUsers, users, "Expected users to match")
				assert.Equal(t, tc.expectedHashed, isHashed, "Expected hashed flag to match")
			}
		})
	}
}