package jsonapi

// Error represents the main error structure.
type Error struct {
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `json:"id"`

	// Links contains links related to the error.
	Links *ErrorLinks `json:"links,omitempty"`

	// Status is the HTTP status code applicable to this problem, expressed as a string value.
	Status string `json:"status"`

	// Code is an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`

	// Title is a short, human-readable summary of the problem.
	// It should not change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title"`

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
