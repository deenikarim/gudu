package validate

import (
	"image"
	"mime/multipart"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CustomValidationFunc func(value string, params ...string) bool
type ConditionalValidationFunc func(data url.Values) bool
type AfterHookFunc func(data url.Values) error
type PreHookFunc func(data url.Values) error
type ValidationErrors map[string][]string

// Validator struct holds the data to be validated and the validation rules.
type Validator struct {
	Data             url.Values
	Errors           ValidationErrors
	Rules            map[string][]string
	CustomValidation map[string]CustomValidationFunc
	Conditionals     map[string]ConditionalValidationFunc
	CustomMessages   map[string]string
	AttributeAliases map[string]string
	AfterHooks       []AfterHookFunc
	PreHooks         []PreHookFunc
	FileData         map[string]*multipart.FileHeader
	DIContainer      map[string]interface{}
	StopOnFirstFail  bool
}

// NewValidator creates a new Validator instance.
func NewValidator(data url.Values, FileData map[string]*multipart.FileHeader, rules map[string][]string) *Validator {
	return &Validator{
		Data:             data,
		Errors:           ValidationErrors{},
		Rules:            rules,
		CustomValidation: make(map[string]CustomValidationFunc),
		Conditionals:     make(map[string]ConditionalValidationFunc),
		CustomMessages:   make(map[string]string),
		AttributeAliases: make(map[string]string),
		AfterHooks:       []AfterHookFunc{},
		PreHooks:         []PreHookFunc{},
		FileData:         FileData,
		DIContainer:      map[string]interface{}{},
		StopOnFirstFail:  true, // Set this to true by default to enable stopping on first failure
	}
}

// ============ main functionalities and features definitions

// Validate runs the validation rules on the data.
func (v *Validator) Validate() bool {
	// execPreHooks
	if !v.execPreHooks() {
		return false
	}

	// Iterate over each field and its associated rules
	for field, fieldRules := range v.Rules {
		// Get the value of the field
		value, exists := v.getFieldValue(field)
		if !exists {
			value = ""
		}
		// Apply each rule to the field's value
		for _, rule := range fieldRules {
			if v.shouldApplyRule(field, rule) {
				if !v.applyRule(field, value, rule) && v.StopOnFirstFail {
					break
				}
			}
		}
	}
	// If no errors, execute after hooks
	if len(v.Errors) == 0 {
		return v.execAfterHooks()
	}

	return false
}

// getFieldValue retrieves the value of a field from the data.
func (v *Validator) getFieldValue(field string) (interface{}, bool) {
	// Check if the field is in the file data
	if fileValue, exists := v.FileData[field]; exists {
		return fileValue, true
	}
	// Check if the field is in the URL values
	if value, exists := v.Data[field]; exists && len(value) > 0 {
		return value[0], true
	}
	return nil, false
}

// shouldApplyRule checks if a rule should be applied based on conditional validations.
func (v *Validator) shouldApplyRule(field, rule string) bool {
	parts := strings.Split(rule, "|")
	if len(parts) > 1 {
		conditionalName := parts[1]
		if conditionalFuncs, ok := v.Conditionals[conditionalName]; ok {
			return conditionalFuncs(v.Data)
		}
		// If the conditional name exists but no function is found, do not apply the rule
		return false
	}
	return true
}

// addError adds an error message for a field.
func (v *Validator) addError(field, defaultMsg string, params ...string) {
	// Retrieve the custom message if it exists, otherwise use the default message
	message, ok := v.CustomMessages[field]
	if !ok {
		message = defaultMsg
	}
	// Use the attribute alias if it exists, otherwise use the field name
	alias := field
	if customAlias, exists := v.AttributeAliases[field]; exists {
		alias = customAlias
	}

	// Format message with alias and params

	// Replace the first %s with the alias
	formattedMessage := strings.Replace(message, "%s", alias, 1)
	// Replace any remaining %s with the params
	for _, param := range params {
		formattedMessage = strings.Replace(formattedMessage, "%s", param, 1)
	}

	// Store formatted message in the errors map
	v.Errors[field] = append(v.Errors[field], formattedMessage)
}

// ApplyRule applies a single validation rule to a field value.
func (v *Validator) applyRule(field string, value interface{}, rule string) bool {
	// Split the rule into its name and parameter
	parts := strings.Split(rule, ":")
	ruleName := parts[0]
	var ruleParams string

	if len(parts) > 1 {
		ruleParams = parts[1]
	}

	// Apply the appropriate validation logic based on the rule name
	switch ruleName {
	case "required":
		if strValue, ok := value.(string); ok && strValue == "" {
			v.addError(field, "This %s is required")
			return false
		} else if fileValue, ok := value.(*multipart.FileHeader); ok && fileValue == nil {
			v.addError(field, "This %s is required")
			return false
		}

	case "email":
		if strValue, ok := value.(string); ok && !v.isValidEmail(strValue) {
			v.addError(field, "The %s field must be a validate email address")
			return false
		}
	case "min":
		if strValue, ok := value.(string); ok && !v.isMin(strValue, ruleParams) {
			v.addError(field, "The %s field must be at least %s characters.", ruleParams)
			return false
		}
	case "max":
		if strValue, ok := value.(string); ok && !v.isMax(strValue, ruleParams) {
			v.addError(field, "The %s field must not exceed %s characters", ruleParams)
			return false
		}
	case "regexp":
		if strValue, ok := value.(string); ok && !v.matchesRegex(strValue, ruleParams) {
			v.addError(field, "The %s field format is invalid")
			return false
		}
	case "numeric":
		if strValue, ok := value.(string); ok && !v.isNumeric(strValue) {
			v.addError(field, "The %s field must be a number")
			return false
		}
	case "date":
		if strValue, ok := value.(string); ok && !v.isValidDateFormat(strValue) {
			v.addError(field, "The %s field must be a date in the form of YYYY-MM-DD")
			return false
		}
	case "confirmed":
		if strValue, ok := value.(string); ok && !v.isConfirmed(field, strValue) {
			v.addError(field, "The %s field confirmation does not match")
			return false
		}
	case "unique":
		if strValue, ok := value.(string); ok && !v.isUnique(strValue, ruleParams) {
			v.addError(field, "The %s field must be unique")
			return false
		}
	case "exists":
		if strValue, ok := value.(string); ok && !v.exists(strValue, ruleParams) {
			v.addError(field, "The %s field already exists")
			return false
		}
	case "in":
		if strValue, ok := value.(string); ok && !v.isIn(strValue, ruleParams) {
			v.addError(field, "The %s field must be one of %s", ruleParams)
			return false
		}
	case "not_in":
		if strValue, ok := value.(string); ok && v.isIn(strValue, ruleParams) {
			v.addError(field, "The %s field must not be one of %s", ruleParams)
			return false
		}
	case "file":
		if fileValue, ok := value.(*multipart.FileHeader); !ok && fileValue == nil {
			v.addError(field, "The %s field must be a validate file")
			return false
		}
	case "mimes":
		if fileValue, ok := value.(*multipart.FileHeader); ok && !v.isValidMimeType(fileValue, ruleParams) {
			v.addError(field, "The %s field must be a file of type: %s", ruleParams)
			return false
		}
	case "max_size":
		if fileValue, ok := value.(*multipart.FileHeader); ok && !v.isValidFileSize(fileValue, ruleParams) {
			v.addError(field, "The %s field must not exceed %s kilobytes", ruleParams)
			return false
		}
	case "image-dimensions":
		dims := strings.Split(ruleParams, ",")
		minWidth, _ := strconv.Atoi(dims[0])
		minHeight, _ := strconv.Atoi(dims[0])
		if fileValue, ok := value.(*multipart.FileHeader); ok && !v.isValidImageDimensions(fileValue, minWidth, minHeight) {
			v.addError(field, "The %s must be at least %s pixels wide and %s pixels tall.", strconv.Itoa(minWidth), strconv.Itoa(minHeight))
			return false
		}

	default:
		if customFuncs, ok := v.CustomValidation[ruleName]; ok {
			if strValue, ok := value.(string); ok && !customFuncs(strValue, ruleParams) {
				v.addError(field, "The %s field failed custom validation for rule %s", ruleParams)
				return false
			}
		}
	}

	return true
}

// ValidateDateOrder  checks if the end date is after the start date.
func (v *Validator) ValidateDateOrder(startField, endField string) {
	startDate, startExist := v.getFieldValue(startField)
	endDate, endExist := v.getFieldValue(endField)
	if !startExist || !endExist {
		return
	}

	start, err1 := time.Parse("2006-01-02", startDate.(string))
	end, err2 := time.Parse("2006-01-02", endDate.(string))
	if err1 != nil || err2 != nil {
		return
	}

	if end.Before(start) {
		v.addErrorForCrossFieldValidation(startField, endField, "date_order", "The %s must be before %s.")
	}

}

//  ========================== utility functions ===========================

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
func (v *Validator) isUnique(value, param string) bool {
	// Implementation for uniqueness check
	return true
}

// exists checks if a field value exists in the mock database.
func (v *Validator) exists(value, param string) bool {
	// Implementation for existence check
	return true
}

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

// ============================== User Methods ===========================

// Errorer returns the validation errors.
func (v *Validator) Errorer() ValidationErrors {
	return v.Errors
}

// AddCustomValidation adds a custom validation function.
func (v *Validator) AddCustomValidation(name string, fn CustomValidationFunc) {
	v.CustomValidation[name] = fn
}

// AddConditionalValidation adds a conditional validation function.
func (v *Validator) AddConditionalValidation(name string, Cond ConditionalValidationFunc) {
	v.Conditionals[name] = Cond
}

// SetCustomMessage sets a custom error message for a field.
func (v *Validator) SetCustomMessage(fieldName, msg string) {
	v.CustomMessages[fieldName] = msg
}

// SetAttributeAlias sets an attribute alias for a field.
func (v *Validator) SetAttributeAlias(fieldName, alias string) {
	v.AttributeAliases[fieldName] = alias
}

// AddAfterHook adds an after hook to be executed after validation passes.
func (v *Validator) AddAfterHook(hook AfterHookFunc) {
	v.AfterHooks = append(v.AfterHooks, hook)
}

// AddPreHook adds a pre-hook to be executed before validation starts.
func (v *Validator) AddPreHook(hook PreHookFunc) {
	v.PreHooks = append(v.PreHooks, hook)
}

// AddCompositeRule adds a composite rule that combines multiple simple rules
func (v *Validator) AddCompositeRule(name string, rules []string) {
	v.Rules[name] = rules
}

// AddRule dynamically adds a rule to a field.
func (v *Validator) AddRule(field, rule string) {
	v.Rules[field] = append(v.Rules[field], rule)
}

// SetDependency sets a dependency in the DI container.
func (v *Validator) SetDependency(key string, value interface{}) {
	v.DIContainer[key] = value
}

// GetDependency retrieves a dependency from the DI container.
func (v *Validator) GetDependency(key string) (interface{}, bool) {
	value, exists := v.DIContainer[key]
	return value, exists
}
