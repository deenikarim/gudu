package gudu

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
