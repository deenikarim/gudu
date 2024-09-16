package gudu

import (
	"context"
	"fmt"
	"image"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

//  ========================== utility functions ===========================

// isValidEmail checks if a value is a validate email address.
func (v *Validator) isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// isMin checks if a value's length is at least the specified minimum length.
func (v *Validator) isMin(value, min string) bool {
	minLength, err := strconv.Atoi(min)
	if err != nil {
		return false
	}
	return len(value) >= minLength
}

// isMax checks if a value's length does not exceed the specified maximum length.
func (v *Validator) isMax(value, max string) bool {
	maxLength, err := strconv.Atoi(max)
	if err != nil {
		return false
	}
	return len(value) <= maxLength
}

// matchesRegex checks if a value matches a regular expression pattern.
func (v *Validator) matchesRegex(value, pattern string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(value)
}

// isNumeric checks if a value is numeric.
func (v *Validator) isNumeric(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

// IsValidDateFormat checks if a string is a validate date in YYYY-MM-DD format.
func (v *Validator) isValidDateFormat(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

// isConfirmed checks if a field value matches its confirmation field value.
func (v *Validator) isConfirmed(field, value string) bool {
	confirmField := field + "_confirmation"
	confirmValue, exists := v.getFieldValue(confirmField)
	if !exists {
		return false
	}
	if confirmStrValue, ok := confirmValue.(string); ok {
		return value == confirmStrValue
	}
	return false
}

// isIn checks if a field value is in a list of validate options.
func (v *Validator) isIn(value, param string) bool {
	options := strings.Split(param, ",")
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

// tip: Use a mock database or data source to check for uniqueness and existence.

// isUnique checks if a field value is unique in the mock database.
func (v *Validator) isUnique(field, value, tableName string) bool {
	//This line builds an SQL query to check how many rows in the table tableName have
	//the given field equal to the value.
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s = $1", tableName, field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := v.DBPool.QueryRowContext(ctx, query, value).Scan(&count)
	if err != nil {
		v.addError(field, "Database error during uniqueness check")
		return false
	}

	//If count == 0, it means the value is unique (because no rows were found in the database
	//with that value), so the function returns true.
	//Otherwise, if the value already exists, it returns false
	return count == 0
}

// exists checks if a field value exists in the mock database.
func (v *Validator) exists(field, value, tableName string) bool {
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE %s = $1)", tableName, field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exist bool
	err := v.DBPool.QueryRowContext(ctx, query, value).Scan(&exist)
	if err != nil {
		v.addError(field, "Database error during existence check")
		return false
	}

	//f count > 0, it means the value exists in the database, so the function returns true.
	//Otherwise, it returns false (i.e., the value does not exist).
	return exist
}

// return at most one row.

// isValidMimeType checks if a file's MIME type is validate.
func (v *Validator) isValidMimeType(file *multipart.FileHeader, mimes string) bool {
	options := strings.Split(mimes, ",")
	for _, option := range options {
		if file.Header.Get("Content-Type") == option {
			return true
		}
	}
	return false
}

// isValidFileSize checks if a file's size is within the maximum allowed size.
func (v *Validator) isValidFileSize(file *multipart.FileHeader, maxSizeStr string) bool {
	// Implementation for existence check
	maxSize, err := strconv.Atoi(maxSizeStr)
	if err != nil {
		return false
	}
	return file.Size <= int64(maxSize*1024)
}

// isValidImageDimensions checks if a file's image dimensions are within the allowed size.
func (v *Validator) isValidImageDimensions(file *multipart.FileHeader, minWidth, minHeight int) bool {
	f, err := file.Open()
	if err != nil {
		return false
	}
	defer func(f multipart.File) {
		_ = f.Close()
	}(f)

	img, _, err := image.Decode(f)
	if err != nil {
		return false
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	return width >= minWidth && height >= minHeight
}

// addErrorForCrossFieldValidation validation where the value of one field depends on another.
func (v *Validator) addErrorForCrossFieldValidation(field1, field2, rule, defaultMsg string) {
	alias1 := field1
	if customAlias, exists := v.AttributeAliases[field1]; exists {
		alias1 = customAlias
	}
	alias2 := field2
	if customAlias, exists := v.AttributeAliases[field2]; exists {
		alias2 = customAlias
	}
	message, ok := v.CustomMessages[field1+"."+field2+"."+rule]
	if !ok {
		message = defaultMsg
	}

	formattedMessage := strings.Replace(message, "%s", alias1, 1)
	formattedMessage = strings.Replace(formattedMessage, "%s", alias2, 1)

	v.Errors[field1] = append(v.Errors[field1], formattedMessage)
}

// ExecAfterHooks runs the after hooks if validation passes.
func (v *Validator) execAfterHooks() bool {
	// Execute each after hook with the validated data
	for _, hook := range v.AfterHooks {
		if err := hook(v.Data); err != nil {
			return false
		}
	}
	return true
}

// execPreHooks runs the pre-hooks before validation starts
func (v *Validator) execPreHooks() bool {
	// Execute each before hook with the validated data
	for _, hook := range v.PreHooks {
		if err := hook(v.Data); err != nil {
			return false
		}
	}
	return true
}

// password checking methods

// isMixedCase checks if a password contains both uppercase and lowercase letters.
func (v *Validator) isMixedCase(value string) bool {
	hasLower := false
	hasUpper := false

	for _, char := range value {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		}

		if hasUpper && hasLower {
			return true
		}

	}
	return false
}

// hasSymbol checks if a password contains at least one symbol.
func (v *Validator) hasSymbol(value string) bool {
	for _, char := range value {
		if strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:'\\\",.<>?/`~", char) {
			return true
		}
	}
	return false
}

// hasNumber checks if a password contains at least one number.
func (v *Validator) hasNumber(value string) bool {
	for _, char := range value {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

func (v *Validator) hasLetter(s string) bool {
	re := regexp.MustCompile(`[a-zA-Z]`)
	return re.MatchString(s)
}
