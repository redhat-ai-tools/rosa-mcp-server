package htpasswd

import (
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

func TestProcessUserInput(t *testing.T) {
	testCases := []struct {
		input           map[string]interface{}
		expectedUsers   map[string]string
		expectError     bool
		description     string
	}{
		{
			map[string]interface{}{
				"users": []interface{}{"user1:password1", "user2:password2"},
			},
			map[string]string{"user1": "password1", "user2": "password2"},
			false,
			"users array input",
		},
		{
			map[string]interface{}{
				"users": []interface{}{"invalid_user_without_colon"},
			},
			nil,
			true,
			"invalid users array format",
		},
		{
			map[string]interface{}{},
			nil,
			true,
			"no users parameter provided",
		},
		{
			map[string]interface{}{
				"users": []interface{}{},
			},
			nil,
			true,
			"empty users array",
		},
		{
			map[string]interface{}{
				"users": []interface{}{123},
			},
			nil,
			true,
			"invalid user type in array",
		},
		{
			map[string]interface{}{
				"users": []interface{}{"user1:password1", "user2:"},
			},
			nil,
			true,
			"user with empty password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			users, err := ProcessUserInput(tc.input)
			
			if tc.expectError {
				assert.Error(t, err, "Expected error for input: %v", tc.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %v", tc.input)
				assert.Equal(t, tc.expectedUsers, users, "Expected users to match")
			}
		})
	}
}