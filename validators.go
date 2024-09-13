package gudu

import (
	"mime/multipart"
	"net/url"
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
func (g *Gudu) NewValidator(data url.Values, FileData map[string]*multipart.FileHeader, rules map[string][]string) *Validator {
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
	case "password":
		if strValue, ok := value.(string); ok {
			if !v.isMixedCase(strValue) {
				v.addError(field, "The %s field must contain both uppercase and lowercase letters")
				return false
			}
			if !v.hasSymbol(strValue) {
				v.addError(field, "The %s field must contain at least one symbol")
				return false
			}
			if !v.hasNumber(strValue) {
				v.addError(field, "The %s field must contain at least one number")
				return false
			}
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
