package openfeature

import (
	"errors"
	"fmt"
)

// Error is an OpenFeature error carrying an ErrorCode.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.Cause }

// NewError returns an Error with the given code and message.
func NewError(code ErrorCode, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func codeOf(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeGeneral
}

var errorClasses = map[ErrorCode]string{
	CodeGeneral:             "GeneralError",
	CodeParseError:          "ParseError",
	CodeFlagNotFound:        "FlagNotFoundError",
	CodeTypeMismatch:        "TypeMismatchError",
	CodeTargetingKeyMissing: "TargetingKeyMissingError",
	CodeInvalidContext:      "InvalidContextError",
	CodeProviderNotReady:    "ProviderNotReadyError",
	CodeProviderFatal:       "ProviderFatalError",
}

func jsErrorString(err error) string {
	name := "Error"
	var e *Error
	if errors.As(err, &e) && e.Code != "" {
		name = "GeneralError"
		if n, ok := errorClasses[e.Code]; ok {
			name = n
		}
	}
	if msg := err.Error(); msg != "" {
		return name + ": " + msg
	}
	return name
}

type jsThrown interface {
	error
	JSString() string
	JSField(name string) (string, bool)
	JSIsError() bool
}

func asJS(v any) (jsThrown, bool) {
	err, ok := v.(error)
	if !ok {
		return nil, false
	}
	var j jsThrown
	return j, errors.As(err, &j)
}

func jsProp(err error, name, fallback string) string {
	j, ok := asJS(err)
	if !ok {
		return fallback
	}
	if s, ok := j.JSField(name); ok {
		return s
	}
	return "undefined"
}

func logField(l Logger, j jsThrown, name string) {
	if s, ok := j.JSField(name); ok {
		l.Error(s)
		return
	}
	l.Error()
}

type thrown struct{ v any }

func (t thrown) Error() string { return fmt.Sprint(t.v) }

func panicError(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	return thrown{r}
}

func errorMessage(err error) string {
	if _, ok := err.(thrown); ok {
		return ""
	}
	return err.Error()
}
