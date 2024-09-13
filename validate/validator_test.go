package validate

import (
	"mime/multipart"
	"net/url"
	"testing"
)

func TestValidator_Validate(t *testing.T) {
	// Ensure the validator is reset before each test case
	resetValidator()

	// example test case
	tests := []struct {
		name        string
		data        url.Values
		fileData    map[string]*multipart.FileHeader
		rules       map[string][]string
		expectedErr bool
	}{
		{
			name: "Valid Case",
			data: url.Values{
				"username": {"john_doe"},
				"email":    {"john.doe@example.com"},
			},
			fileData:    nil,
			rules:       map[string][]string{"username": {"required"}, "email": {"required", "email"}},
			expectedErr: false,
		},
		{
			name: "Invalid Case - Missing Required Field",
			data: url.Values{
				"username": {""}, // Required field missing
				"email":    {"john.doe@example.com"},
			},
			fileData:    nil,
			rules:       map[string][]string{"username": {"required"}, "email": {"required", "email"}},
			expectedErr: true,
		},
		{
			name: "Invalid Case - Invalid Email",
			data: url.Values{
				"username": {"john_doe"},
				"email":    {"invalid_email"}, // Invalid email format
			},
			fileData:    nil,
			rules:       map[string][]string{"username": {"required"}, "email": {"required", "email"}},
			expectedErr: true,
		},
		{
			name: "Valid Case - File Upload",
			data: url.Values{
				"username": {"john_doe"},
			},
			fileData: map[string]*multipart.FileHeader{
				"avatar": {
					Filename: "avatar.jpg",
					Size:     1024, // Assuming file size in bytes
					Header:   make(map[string][]string),
				},
			},
			rules:       map[string][]string{"avatar": {"file", "mimes:image/jpeg"}},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testValidator.Data = tt.data
			testValidator.FileData = tt.fileData
			testValidator.Rules = tt.rules

			isValid := testValidator.Validate()

			if tt.expectedErr && isValid {
				t.Errorf("Expected validation to fail but succeeded")
			}

			if tt.expectedErr && !isValid {
				t.Errorf("Expected validation to pass but failed")
			}
		})
	}
}

func TestValidator_CustomValidation(t *testing.T) {
	// Ensure the validator is reset before each test case
	resetValidator()

	// Define a custom validation function
	customValidationFunc := func(value string, params ...string) bool {
		return value == "custom_value"
	}

	// Add the custom validation function to the validator
	testValidator.AddCustomValidation("custom_rule", customValidationFunc)

	// Test case where custom validation should fail
	testValidator.AddRule("custom_field", "custom_rule")
	testValidator.Data.Set("custom_field", "incorrect_value")
	isValid := testValidator.Validate()

	if isValid {
		t.Errorf("Expected custom validation to fail but succeeded")
	}

	// Test case where custom validation should pass
	testValidator.Data.Set("custom_field", "custom_value")
	isValid = testValidator.Validate()

	if !isValid {
		t.Errorf("Expected custom validation to pass but failed")
	}
}
