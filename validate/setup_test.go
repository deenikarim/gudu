package validate

import (
	"mime/multipart"
	"net/url"
	"os"
	"testing"
)

var testValidator *Validator

func createTestValidator() *Validator {
	// Example:
	data := url.Values{
		"username": {"john_doe"},
		"email":    {"john.doe@example.com"},
	}
	fileData := map[string]*multipart.FileHeader{}
	rules := map[string][]string{"username": {"required"}, "email": {"required", "email"}}

	return NewValidator(data, fileData, rules)
}

// Reset the validator before each test
func resetValidator() {
	testValidator.Errors = ValidationErrors{}
	// Add reset logic for other fields if needed
}

func TestMain(m *testing.M) {
	// Perform setup operations here
	testValidator = createTestValidator()

	// Run tests
	exitVal := m.Run()

	// Exit with the appropriate exit code
	os.Exit(exitVal)
}
