package execution

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"workflow-code-test/api/internal/models"
)

// DefaultInputValidator provides input validation for form data
type DefaultInputValidator struct{}

// NewDefaultInputValidator creates a new input validator
func NewDefaultInputValidator() *DefaultInputValidator {
	return &DefaultInputValidator{}
}

// ValidateFormData validates form input data against node field definitions
func (v *DefaultInputValidator) ValidateFormData(formData map[string]interface{}, nodeData *models.FormNodeData) error {
	if nodeData == nil || len(nodeData.Metadata.InputFields) == 0 {
		// No validation rules defined, accept all data
		return nil
	}

	var errors []string

	// Validate each input field (simplified validation for strongly typed data)
	for _, fieldName := range nodeData.Metadata.InputFields {
		value, exists := formData[fieldName]

		// For now, just check if required fields exist
		// TODO: Extend FormNodeData to include detailed field validation rules
		if !exists || isEmptyValue(value) {
			errors = append(errors, fmt.Sprintf("field '%s' is required", fieldName))
			continue
		}

		// Basic validation for known field types
		if err := v.validateBasicFieldValue(fieldName, value); err != nil {
			errors = append(errors, fmt.Sprintf("field '%s': %s", fieldName, err.Error()))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateBasicFieldValue provides basic validation for common field types
func (v *DefaultInputValidator) validateBasicFieldValue(fieldName string, value interface{}) error {
	// Basic validation based on field name patterns
	switch fieldName {
	case "email":
		if strValue, ok := value.(string); ok {
			if _, err := mail.ParseAddress(strValue); err != nil {
				return fmt.Errorf("must be a valid email address")
			}
		} else {
			return fmt.Errorf("must be a string")
		}
	case "name":
		if strValue, ok := value.(string); ok {
			if strings.TrimSpace(strValue) == "" {
				return fmt.Errorf("cannot be empty")
			}
		} else {
			return fmt.Errorf("must be a string")
		}
	default:
		// No specific validation for other fields
		return nil
	}
	return nil
}

// validateFieldValue validates a single field value based on its type and validation rules
func (v *DefaultInputValidator) validateFieldValue(field models.FormField, value interface{}) error {
	// Type validation
	switch field.Type {
	case "text", "email":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		return v.validateStringField(field, strValue)

	case "number":
		switch numValue := value.(type) {
		case float64, int, int64:
			return v.validateNumberField(field, numValue)
		default:
			return fmt.Errorf("must be a number")
		}

	case "select":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		return v.validateSelectField(field, strValue)

	default:
		// Unknown field type, no specific validation
		return nil
	}
}

// validateStringField validates string fields (text, email)
func (v *DefaultInputValidator) validateStringField(field models.FormField, value string) error {
	// Check empty values
	if strings.TrimSpace(value) == "" && field.Required {
		return fmt.Errorf("cannot be empty")
	}

	// Email validation
	if field.Type == "email" {
		if _, err := mail.ParseAddress(value); err != nil {
			return fmt.Errorf("must be a valid email address")
		}
	}

	// Custom validation rules
	if field.Validation != "" {
		if err := v.validateCustomRule(field.Validation, value); err != nil {
			return err
		}
	}

	return nil
}

// validateNumberField validates number fields
func (v *DefaultInputValidator) validateNumberField(field models.FormField, value interface{}) error {
	var numValue float64

	switch v := value.(type) {
	case float64:
		numValue = v
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	default:
		return fmt.Errorf("invalid number type")
	}

	// Custom validation rules for numbers
	if field.Validation != "" {
		if err := v.validateNumberRule(field.Validation, numValue); err != nil {
			return err
		}
	}

	return nil
}

// validateSelectField validates select/dropdown fields
func (v *DefaultInputValidator) validateSelectField(field models.FormField, value string) error {
	if len(field.Options) == 0 {
		// No options defined, accept any value
		return nil
	}

	// Check if value is in allowed options
	for _, option := range field.Options {
		if option == value {
			return nil
		}
	}

	return fmt.Errorf("must be one of: %s", strings.Join(field.Options, ", "))
}

// validateCustomRule validates custom validation rules
func (v *DefaultInputValidator) validateCustomRule(rule, value string) error {
	switch {
	case strings.HasPrefix(rule, "min_length:"):
		return v.validateMinLength(rule, value)
	case strings.HasPrefix(rule, "max_length:"):
		return v.validateMaxLength(rule, value)
	case strings.HasPrefix(rule, "regex:"):
		return v.validateRegex(rule, value)
	case rule == "no_spaces":
		if strings.Contains(value, " ") {
			return fmt.Errorf("cannot contain spaces")
		}
	case rule == "alpha_only":
		if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(value) {
			return fmt.Errorf("must contain only letters")
		}
	case rule == "alphanumeric":
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
			return fmt.Errorf("must contain only letters and numbers")
		}
	default:
		// Unknown rule, skip validation
		return nil
	}

	return nil
}

// validateNumberRule validates number-specific rules
func (v *DefaultInputValidator) validateNumberRule(rule string, value float64) error {
	switch {
	case strings.HasPrefix(rule, "min:"):
		var min float64
		if _, err := fmt.Sscanf(rule, "min:%f", &min); err != nil {
			return nil // Invalid rule format, skip
		}
		if value < min {
			return fmt.Errorf("must be at least %.1f", min)
		}
	case strings.HasPrefix(rule, "max:"):
		var max float64
		if _, err := fmt.Sscanf(rule, "max:%f", &max); err != nil {
			return nil // Invalid rule format, skip
		}
		if value > max {
			return fmt.Errorf("must be at most %.1f", max)
		}
	case strings.HasPrefix(rule, "range:"):
		var min, max float64
		if _, err := fmt.Sscanf(rule, "range:%f,%f", &min, &max); err != nil {
			return nil // Invalid rule format, skip
		}
		if value < min || value > max {
			return fmt.Errorf("must be between %.1f and %.1f", min, max)
		}
	}

	return nil
}

// validateMinLength validates minimum string length
func (v *DefaultInputValidator) validateMinLength(rule, value string) error {
	var minLen int
	if _, err := fmt.Sscanf(rule, "min_length:%d", &minLen); err != nil {
		return nil // Invalid rule format, skip
	}

	if len(value) < minLen {
		return fmt.Errorf("must be at least %d characters", minLen)
	}

	return nil
}

// validateMaxLength validates maximum string length
func (v *DefaultInputValidator) validateMaxLength(rule, value string) error {
	var maxLen int
	if _, err := fmt.Sscanf(rule, "max_length:%d", &maxLen); err != nil {
		return nil // Invalid rule format, skip
	}

	if len(value) > maxLen {
		return fmt.Errorf("must be at most %d characters", maxLen)
	}

	return nil
}

// validateRegex validates against a regular expression
func (v *DefaultInputValidator) validateRegex(rule, value string) error {
	pattern := strings.TrimPrefix(rule, "regex:")
	if pattern == "" {
		return nil // Empty pattern, skip
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil // Invalid regex, skip validation
	}

	if !regex.MatchString(value) {
		return fmt.Errorf("does not match required pattern")
	}

	return nil
}

// isEmptyValue checks if a value is considered empty
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int, int64, float64:
		return false // Numbers are not considered empty
	default:
		return false
	}
}
