package colorkit

import "errors"

var (
	// ErrSyntax indicates unsupported or malformed color input.
	ErrSyntax = errors.New("invalid color syntax")
	// ErrRange indicates a channel outside its documented range.
	ErrRange = errors.New("color channel out of range")
	// ErrAlpha indicates a color that is not fully opaque.
	ErrAlpha = errors.New("color must be fully opaque")
)

// Error identifies the operation and, when applicable, the invalid field.
// Err can be inspected with errors.Is; Error can be inspected with errors.As.
type Error struct {
	Op    string
	Field string
	Err   error
}

// Error returns a description of the failed operation.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	prefix := e.Op
	if e.Field != "" {
		prefix += "." + e.Field
	}
	if e.Err == nil {
		return prefix
	}
	return prefix + ": " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func invalid(op, field string, err error) error {
	return &Error{Op: op, Field: field, Err: err}
}
