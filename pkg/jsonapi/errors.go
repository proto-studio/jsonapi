package jsonapi

import (
	"strings"

	"proto.zip/studio/validate/pkg/errors"
)

// ErrorSourceKind is the type of JSON:API error source (pointer, parameter, or header).
type ErrorSourceKind string

const (
	// SourcePointer is for body/document errors (JSON Pointer to the value in the request document).
	SourcePointer ErrorSourceKind = "pointer"
	// SourceParameter is for query string errors (URI query parameter name).
	SourceParameter ErrorSourceKind = "parameter"
	// SourceHeader is for request header errors (header name).
	SourceHeader ErrorSourceKind = "header"
)

// jsonAPIErrorWrapper wraps *Error to implement errors.ValidationError without
// method/field name conflicts (Error has fields Code and Meta).
type jsonAPIErrorWrapper struct{ err *Error }

var _ errors.ValidationError = (*jsonAPIErrorWrapper)(nil)

// JSONAPIError returns the underlying JSON:API Error for serialization.
func (w *jsonAPIErrorWrapper) JSONAPIError() *Error { return w.err }

func (w *jsonAPIErrorWrapper) Code() errors.ErrorCode { return errors.ErrorCode(w.err.Code) }
func (w *jsonAPIErrorWrapper) Path() string {
	if w.err.Source == nil {
		return ""
	}
	if w.err.Source.Pointer != "" {
		return w.err.Source.Pointer
	}
	if w.err.Source.Parameter != "" {
		return w.err.Source.Parameter
	}
	return w.err.Source.Header
}
func (w *jsonAPIErrorWrapper) PathAs(errors.PathSerializer) string { return w.Path() }
func (w *jsonAPIErrorWrapper) Error() string                         { return w.err.Detail }
func (w *jsonAPIErrorWrapper) ShortError() string                    { return w.err.Title }
func (w *jsonAPIErrorWrapper) DocsURI() string {
	if w.err.Links != nil {
		return w.err.Links.About
	}
	return ""
}
func (w *jsonAPIErrorWrapper) TraceURI() string {
	if w.err.Links != nil {
		return w.err.Links.Type
	}
	return ""
}
func (w *jsonAPIErrorWrapper) Meta() map[string]any {
	if w.err.Meta == nil {
		return nil
	}
	return map[string]any(*w.err.Meta)
}
func (w *jsonAPIErrorWrapper) Params() []any { return nil }
func (w *jsonAPIErrorWrapper) Internal() bool {
	return errors.DefaultDict().ErrorType(errors.ErrorCode(w.err.Code)) == errors.ErrorTypeInternal
}
func (w *jsonAPIErrorWrapper) Validation() bool {
	return errors.DefaultDict().ErrorType(errors.ErrorCode(w.err.Code)) == errors.ErrorTypeValidation
}
func (w *jsonAPIErrorWrapper) Permission() bool {
	return errors.DefaultDict().ErrorType(errors.ErrorCode(w.err.Code)) == errors.ErrorTypePermission
}

// Unwrap returns nil (single error, no wrapped errors). Required by ValidationError.
func (w *jsonAPIErrorWrapper) Unwrap() []error { return nil }

// Error represents the main error structure for JSON:API responses.
type Error struct {
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `json:"id,omitempty"`

	// Links contains links related to the error.
	Links *ErrorLinks `json:"links,omitempty"`

	// Status is the HTTP status code applicable to this problem, expressed as a string value.
	Status string `json:"status"`

	// Code is an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`

	// Title is a short, human-readable summary of the problem.
	// It should not change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title,omitempty"`

	// Detail is a human-readable explanation specific to this occurrence of the problem.
	// Like Title, this field’s value can be localized.
	Detail string `json:"detail,omitempty"`

	// Source represents the primary source of the error.
	Source *Source `json:"source,omitempty"`

	// Meta contains non-standard meta-information about the error.
	Meta *MetaInfo `json:"meta,omitempty"`
}

// Links contains links related to the error.
type ErrorLinks struct {
	// About is a link that leads to further details about this particular occurrence of the problem.
	About string `json:"about,omitempty"`

	// Type identifies the type of error that this particular error is an instance of.
	Type string `json:"type,omitempty"`
}

// Source represents the primary source of the error.
type Source struct {
	// Pointer is a JSON Pointer to the value in the request document that caused the error.
	// This must point to a value in the request document that exists.
	Pointer string `json:"pointer,omitempty"`

	// Parameter indicates which URI query parameter caused the error.
	Parameter string `json:"parameter,omitempty"`

	// Header indicates the name of a single request header which caused the error.
	Header string `json:"header,omitempty"`
}

// MetaInfo contains non-standard meta-information about the error.
type MetaInfo map[string]any

// ErrorResponse represents the structure of the error response.
type ErrorResponse struct {
	Errors []Error `json:"errors,omitempty"`
}

// ValidationError interface methods (errors.ValidationError).

// jsonAPIErrorHolder is used to extract *Error from our wrapper in ErrorsFromCollection.
type jsonAPIErrorHolder interface {
	JSONAPIError() *Error
}

// jsonPointerSerializer is used for source.pointer so it follows RFC 6901 (JSON Pointer) as required by JSON:API.
var jsonPointerSerializer errors.JSONPointerSerializer

// queryParamName returns the JSON:API source.parameter value (query param name only).
// Path from validation may be "query[key]" or "/query[key]" (e.g. "query[filter]", "query[page[size]]"); we return just the param name.
func queryParamName(path string) string {
	path = strings.TrimPrefix(path, "/")
	const prefix = "query["
	if strings.HasPrefix(path, prefix) && strings.HasSuffix(path, "]") && len(path) > len(prefix)+1 {
		return path[len(prefix) : len(path)-1]
	}
	return path
}

// ErrorFromValidationError builds a JSON:API Error from a ValidationError.
// kind selects which source field to set: SourcePointer (body), SourceParameter (query), or SourceHeader.
// When kind is SourcePointer, the path is serialized with JSON Pointer (RFC 6901) per JSON:API; other kinds use the default path.
// Query string errors use HTTP status 400 per JSON:API; body validation errors use 422.
func ErrorFromValidationError(ve errors.ValidationError, kind ErrorSourceKind) *Error {
	status := "422"
	if kind == SourceParameter {
		status = "400"
	}
	e := &Error{
		Status: status,
		Code:  string(ve.Code()),
		Title: ve.ShortError(),
		Detail: ve.Error(),
	}
	// Only set Links when at least one URI is non-empty; only include non-empty values so serialization omits empty strings.
	if docs, trace := ve.DocsURI(), ve.TraceURI(); docs != "" || trace != "" {
		e.Links = &ErrorLinks{About: docs, Type: trace}
	}
	if m := ve.Meta(); len(m) > 0 {
		meta := make(MetaInfo, len(m))
		for k, v := range m {
			meta[k] = v
		}
		e.Meta = &meta
	}
	var path string
	switch kind {
	case SourcePointer:
		// JSON:API source.pointer MUST be a JSON Pointer [RFC6901].
		path = ve.PathAs(jsonPointerSerializer)
	default:
		path = ve.Path()
	}
	if path != "" {
		e.Source = &Source{}
		switch kind {
		case SourcePointer:
			e.Source.Pointer = path
		case SourceParameter:
			// JSON:API source.parameter is the query parameter name only, not a path.
			e.Source.Parameter = queryParamName(path)
		case SourceHeader:
			// JSON:API source.header is the header name; path may have a leading slash from serialization
			e.Source.Header = strings.TrimPrefix(path, "/")
		}
	}
	return e
}

// ToJSONAPIErrors wraps each error in err as a JSON:API Error and returns a single ValidationError.
// kind selects the source field: SourcePointer (body), SourceParameter (query), or SourceHeader.
func ToJSONAPIErrors(err errors.ValidationError, kind ErrorSourceKind) errors.ValidationError {
	unwrapped := errors.Unwrap(err)
	if len(unwrapped) == 0 {
		return nil
	}
	var out []error
	for _, e := range unwrapped {
		ve := e.(errors.ValidationError)
		out = append(out, &jsonAPIErrorWrapper{err: ErrorFromValidationError(ve, kind)})
	}
	return errors.Join(out...)
}

// ErrorsFromValidationError builds a slice of JSON:API Error for response bodies.
// It uses errors.Unwrap() to obtain all errors (not just the first).
// kind is only used when converting non-JSON:API ValidationErrors.
// Any unwrapped error that does not implement ValidationError is skipped.
func ErrorsFromValidationError(err errors.ValidationError, kind ErrorSourceKind) []Error {
	unwrapped := errors.Unwrap(err)
	if len(unwrapped) == 0 {
		return nil
	}
	out := make([]Error, 0, len(unwrapped))
	for _, e := range unwrapped {
		ve, ok := e.(errors.ValidationError)
		if !ok {
			continue
		}
		if h, ok := ve.(jsonAPIErrorHolder); ok {
			out = append(out, *h.JSONAPIError())
		} else {
			out = append(out, *ErrorFromValidationError(ve, kind))
		}
	}
	return out
}
