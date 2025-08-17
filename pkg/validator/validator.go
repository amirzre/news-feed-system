package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// NewValidator creates a new custom validator instance
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom tag name function to use json tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{validator: v}
}

// Validate validates a struct and returns detailed error information
func (cv *CustomValidator) Validate(i interface{}) error {
	err := cv.validator.Struct(i)
	if err == nil {
		return nil
	}

	// Handle validation errors
	var validationErrors []ValidationError

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, err := range errs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: getErrorMessage(err),
			})
		}
	}

	return ValidationErrors{Errors: validationErrors}
}

// getErrorMessage returns a human-readable error message based on the validation tag
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()
	value := fmt.Sprintf("%v", err.Value())

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if err.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters long", field, param)
		}
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		if err.Kind() == reflect.String {
			return fmt.Sprintf("%s must not exceed %s characters", field, param)
		}
		return fmt.Sprintf("%s must not exceed %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, param)
	case "unique":
		return fmt.Sprintf("%s must contain unique values", field)
	case "dive":
		return fmt.Sprintf("%s contains invalid nested values", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime in format %s", field, param)
	case "eqfield":
		return fmt.Sprintf("%s must equal %s", field, param)
	case "nefield":
		return fmt.Sprintf("%s must not equal %s", field, param)
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of the following characters: %s", field, param)
	case "excludesall":
		return fmt.Sprintf("%s cannot contain any of the following characters: %s", field, param)
	case "startswith":
		return fmt.Sprintf("%s must start with '%s'", field, param)
	case "endswith":
		return fmt.Sprintf("%s must end with '%s'", field, param)
	default:
		// Fallback for unknown tags
		if param != "" {
			return fmt.Sprintf("%s failed validation for tag '%s' with parameter '%s' (current value: '%s')", field, tag, param, value)
		}
		return fmt.Sprintf("%s failed validation for tag '%s' (current value: '%s')", field, tag, value)
	}
}
