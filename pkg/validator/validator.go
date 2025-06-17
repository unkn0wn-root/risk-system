package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

type Validator struct {
	errors ValidationErrors
}

func New() *Validator {
	return &Validator{errors: make(ValidationErrors, 0)}
}

func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "is required",
		})
	}
	return v
}

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

func (v *Validator) MinLength(field, value string, length int) *Validator {
	if len(value) < length {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", length),
		})
	}
	return v
}

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

func (v *Validator) Min(field string, value float64, min float64) *Validator {
	if value < min {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be greater than or equal to %.0f", min),
		})
	}
	return v
}

func (v *Validator) Max(field string, value float64, max float64) *Validator {
	if value > max {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be less than or equal to %.0f", max),
		})
	}
	return v
}

func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

func (v *Validator) Errors() ValidationErrors {
	return v.errors
}
