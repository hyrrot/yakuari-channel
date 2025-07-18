package errors

import "fmt"

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION"
	ErrorTypeFileNotFound ErrorType = "FILE_NOT_FOUND"
	ErrorTypeAPIError     ErrorType = "API_ERROR"
	ErrorTypeCyclicRef    ErrorType = "CYCLIC_REFERENCE"
	ErrorTypeParser       ErrorType = "PARSER"
)

// CompilerError represents a ymmp-compiler specific error
type CompilerError struct {
	Type     ErrorType
	Message  string
	Location string // File path and line number
	Cause    error
}

// Error implements the error interface
func (e *CompilerError) Error() string {
	if e.Location != "" {
		return fmt.Sprintf("[%s] %s\nLocation: %s", e.Type, e.Message, e.Location)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string, location string) *CompilerError {
	return &CompilerError{
		Type:     ErrorTypeValidation,
		Message:  message,
		Location: location,
	}
}

// NewParserError creates a new parser error
func NewParserError(message string, cause error) *CompilerError {
	return &CompilerError{
		Type:    ErrorTypeParser,
		Message: message,
		Cause:   cause,
	}
}

// NewFileNotFoundError creates a new file not found error
func NewFileNotFoundError(filePath string) *CompilerError {
	return &CompilerError{
		Type:     ErrorTypeFileNotFound,
		Message:  fmt.Sprintf("file not found: %s", filePath),
		Location: filePath,
	}
}