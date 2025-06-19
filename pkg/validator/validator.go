// Package validator provides a fluent validation framework for input data validation.
// It supports common validation rules and collects multiple validation errors.
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a single validation failure for a specific field.
type ValidationError struct {
	Field   string `json:"field"`   // Name of the field that failed validation
	Message string `json:"message"` // Human-readable validation error message
}

// ValidationErrors is a collection of validation errors that implements the error interface.
type ValidationErrors []ValidationError

// Error implements the error interface by combining all validation errors into a single message.
func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

// Validator provides a fluent interface for validating data and collecting errors.
type Validator struct {
	errors ValidationErrors // Collection of validation errors encountered
}

// New creates a new validator instance with an empty error collection.
func New() *Validator {
	return &Validator{errors: make(ValidationErrors, 0)}
}

// Required validates that a string field is not empty after trimming whitespace.
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "is required",
		})
	}
	return v
}

// Email validates that a string field contains a valid email address format.
// Skips validation if the value is empty. Returns the validator for method chaining.
func (v *Validator) Email(field, value string) *Validator {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if value != "" && !emailRegex.MatchString(strings.ToLower(value)) {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "must be a valid email address",
		})
	}
	return v
}

// MinLength validates that a string field meets the minimum length requirement.
func (v *Validator) MinLength(field, value string, length int) *Validator {
	if len(value) < length {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", length),
		})
	}
	return v
}

// Phone validates that a string field contains a valid phone number format.
func (v *Validator) Phone(field, value string) *Validator {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if value != "" && !phoneRegex.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "must be a valid phone number",
		})
	}
	return v
}

// Min validates that a numeric field meets the minimum value requirement.
func (v *Validator) Min(field string, value float64, min float64) *Validator {
	if value < min {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be greater than or equal to %.0f", min),
		})
	}
	return v
}

// Max validates that a numeric field does not exceed the maximum value.
func (v *Validator) Max(field string, value float64, max float64) *Validator {
	if value > max {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be less than or equal to %.0f", max),
		})
	}
	return v
}

// IsValid returns true if no validation errors have been collected.
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// Errors returns the collection of validation errors encountered during validation.
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}
