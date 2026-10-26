package errors

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
)

func init() {
	SetInlineEncoder()
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Type(err error) string {
	if err == nil {
		return ""
	}

	if err, ok := err.(interface{ Type() string }); ok {
		return err.Type()
	}

	return reflect.TypeOf(err).String()
}

func HasType(err error, v string) bool {
	for {
		if v == Type(err) {
			return true
		}

		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func UIError(err error) string {
	if err == nil {
		return ""
	}

	for current := err; current != nil; current = Unwrap(current) {
		if err, ok := current.(interface{ UIError() string }); ok {
			return err.UIError()
		}
	}
	return err.Error()
}

func Tag(err error, k string) any {
	for {
		if err, ok := err.(interface{ Tag(string) any }); ok {
			if v := err.Tag(k); v != nil {
				return v
			}
		}

		if err = Unwrap(err); err == nil {
			return nil
		}
	}
}

type Error struct {
	Line        string
	Message     string
	DefinedType string
	UIMessage   string
	Tags        map[string]any
	WrappedErr  error
}

func New(msg string) Error {
	return makeError(msg)
}

func Newf(msgFormat string, v ...any) Error {
	return makeError(fmt.Sprintf(msgFormat, v...))
}

func makeError(v string) Error {
	_, filename, line, _ := runtime.Caller(2)

	err := Error{
		Line:    filepath.Base(filename) + ":" + strconv.Itoa(line),
		Message: v,
	}
	return err
}

func (e Error) WithType(v string) Error {
	e.DefinedType = v
	return e
}

func (e Error) Type() string {
	if e.DefinedType != "" {
		return e.DefinedType
	}

	if e.WrappedErr != nil {
		return Type(e.WrappedErr)
	}

	return reflect.TypeOf(e).String()
}

func (e Error) WithTag(k string, v any) Error {
	if e.Tags == nil {
		e.Tags = make(map[string]any)
	}

	e.Tags[k] = v
	return e
}

func (e Error) Tag(k string) any {
	return e.Tags[k]
}

func (e Error) WithUIError(msg string) Error {
	e.UIMessage = msg
	return e
}

func (e Error) UIError() string {
	if e.UIMessage == "" {
		return e.Error()
	}
	return e.UIMessage
}

func (e Error) Wrap(err error) Error {
	e.WrappedErr = err
	return e
}

func (e Error) Unwrap() error {
	return e.WrappedErr
}

func (e Error) Error() string {
	s, err := getEncoder()(makeJSONError(e))
	if err != nil {
		return fmt.Sprintf(`{"message": "encoding error failed: %s"}`, err)
	}
	return string(s)
}

func (e Error) MarshalJSON() ([]byte, error) {
	return getEncoder()(makeJSONError(e))
}

func (e Error) Is(err error) bool {
	rerr, ok := err.(Error)
	if !ok {
		return false
	}

	return rerr.Line == e.Line &&
		rerr.Message == e.Message &&
		rerr.DefinedType == e.DefinedType &&
		reflect.DeepEqual(rerr.Tags, e.Tags) &&
		isSameErr(rerr.WrappedErr, e.WrappedErr)
}

func isSameErr(a, b error) bool {
	if a == nil || b == nil {
		return a == b
	}

	ta := reflect.TypeOf(a)
	tb := reflect.TypeOf(b)
	if ta != tb || !ta.Comparable() {
		return false
	}

	return a == b
}

type jsonError struct {
	Line        string         `json:"line,omitempty"`
	Message     string         `json:"message"`
	UIMessage   string         `json:"ui,omitempty"`
	DefinedType string         `json:"type,omitempty"`
	Tags        map[string]any `json:"tags,omitempty"`
	WrappedErr  any            `json:"wrap,omitempty"`
}

func makeJSONError(err Error) jsonError {
	tags := makeJSONTags(err.Tags)

	return jsonError{
		Line:        err.Line,
		Message:     err.Message,
		UIMessage:   err.UIMessage,
		DefinedType: err.DefinedType,
		Tags:        tags,
		WrappedErr:  makeJSONWrappedErr(err.WrappedErr),
	}
}

func makeJSONWrappedErr(err error) any {
	switch err := err.(type) {
	case nil:
		return nil

	case Error:
		return makeJSONError(err)

	case *Error:
		if err == nil {
			return nil
		}
		return makeJSONError(*err)

	default:
		return err.Error()
	}
}

func makeJSONTags(tags map[string]any) map[string]any {
	if len(tags) == 0 {
		return nil
	}

	requiresNormalization := false
normalizationLoop:
	for _, v := range tags {
		switch v.(type) {
		case reflect.Type:
			requiresNormalization = true
			break normalizationLoop
		}
	}
	if !requiresNormalization {
		return tags
	}

	jsonTags := make(map[string]any, len(tags))
	for k, v := range tags {
		switch v := v.(type) {
		case reflect.Type:
			jsonTags[k] = v.String()

		default:
			jsonTags[k] = v
		}
	}
	return jsonTags
}
