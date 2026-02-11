package jsonapi

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

func TestError_JSONSerialization(t *testing.T) {
	resp := ErrorResponse{
		Errors: []Error{
			{
				Status: "422",
				Code:   "REQUIRED",
				Title:  "member name required",
				Detail: "member name must not be empty",
				Source: &Source{Pointer: "/data/attributes/name"},
			},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded struct {
		Errors []struct {
			Status string  `json:"status"`
			Code   string  `json:"code"`
			Title  string  `json:"title"`
			Detail string  `json:"detail"`
			Source *Source `json:"source"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Errors) != 1 {
		t.Fatalf("len(errors) = %d, want 1", len(decoded.Errors))
	}
	e := decoded.Errors[0]
	if e.Status != "422" || e.Code != "REQUIRED" || e.Title != "member name required" || e.Detail != "member name must not be empty" {
		t.Errorf("error fields: status=%q code=%q title=%q detail=%q", e.Status, e.Code, e.Title, e.Detail)
	}
	if e.Source == nil || e.Source.Pointer != "/data/attributes/name" {
		t.Errorf("source.pointer: got %v", e.Source)
	}
	// JSON:API format: top-level "errors" array
	if len(data) == 0 || string(data) == "{}" {
		t.Error("expected non-empty JSON with errors key")
	}
}

func TestError_QueryUsesParameterBodyUsesPointer(t *testing.T) {
	ctx := context.Background()

	// Query: should use source.parameter (path in context becomes parameter name)
	queryCtx := rulecontext.WithPathString(ctx, "sort")
	queryErr := errors.Errorf(errors.CodeUnexpected, queryCtx, "reserved", "query param %q reserved", "sort")
	queryWrapped := ToJSONAPIErrors(queryErr, SourceParameter)
	queryList := ErrorsFromValidationError(queryWrapped, SourceParameter)
	if len(queryList) != 1 {
		t.Fatalf("query errors: got %d, want 1", len(queryList))
	}
	if queryList[0].Source == nil {
		t.Fatal("query error Source is nil")
	}
	if queryList[0].Source.Parameter == "" {
		t.Errorf("query error should set source.parameter, got pointer=%q parameter=%q",
			queryList[0].Source.Pointer, queryList[0].Source.Parameter)
	}
	if queryList[0].Source.Pointer != "" {
		t.Errorf("query error should not set source.pointer, got %q", queryList[0].Source.Pointer)
	}

	// Body: should use source.pointer
	bodyCtx := rulecontext.WithPathString(ctx, "data")
	bodyCtx = rulecontext.WithPathString(bodyCtx, "attributes")
	bodyCtx = rulecontext.WithPathString(bodyCtx, "name")
	bodyErr := errors.Errorf(errors.CodeRequired, bodyCtx, "required", "field is required")
	bodyWrapped := ToJSONAPIErrors(bodyErr, SourcePointer)
	bodyList := ErrorsFromValidationError(bodyWrapped, SourcePointer)
	if len(bodyList) != 1 {
		t.Fatalf("body errors: got %d, want 1", len(bodyList))
	}
	if bodyList[0].Source == nil {
		t.Fatal("body error Source is nil")
	}
	if bodyList[0].Source.Pointer == "" {
		t.Errorf("body error should set source.pointer, got pointer=%q parameter=%q",
			bodyList[0].Source.Pointer, bodyList[0].Source.Parameter)
	}
	if bodyList[0].Source.Parameter != "" {
		t.Errorf("body error should not set source.parameter, got %q", bodyList[0].Source.Parameter)
	}

	// Header: should use source.header
	headerCtx := rulecontext.WithPathString(ctx, "X-Request-Id")
	headerErr := errors.Errorf(errors.CodeUnexpected, headerCtx, "invalid header", "header value invalid")
	headerWrapped := ToJSONAPIErrors(headerErr, SourceHeader)
	headerList := ErrorsFromValidationError(headerWrapped, SourceHeader)
	if len(headerList) != 1 {
		t.Fatalf("header errors: got %d, want 1", len(headerList))
	}
	if headerList[0].Source == nil {
		t.Fatal("header error Source is nil")
	}
	if headerList[0].Source.Header == "" {
		t.Errorf("header error should set source.header, got pointer=%q parameter=%q header=%q",
			headerList[0].Source.Pointer, headerList[0].Source.Parameter, headerList[0].Source.Header)
	}
	if headerList[0].Source.Pointer != "" || headerList[0].Source.Parameter != "" {
		t.Errorf("header error should not set pointer or parameter")
	}
}

func TestErrorsFromValidationError_UnwrapReturnsAllErrors(t *testing.T) {
	ctx := context.Background()
	e1 := errors.Errorf(errors.CodeRequired, ctx, "first", "first error")
	e2 := errors.Errorf(errors.CodeUnexpected, ctx, "second", "second error")
	e3 := errors.Errorf(errors.CodePattern, ctx, "third", "third error")
	joined := errors.Join(e1, e2, e3)

	wrapped := ToJSONAPIErrors(joined, SourceParameter)
	list := ErrorsFromValidationError(wrapped, SourceParameter)

	if len(list) != 3 {
		t.Fatalf("ErrorsFromValidationError: got %d errors, want 3 (Unwrap must return all)", len(list))
	}
	// Verify we didn't just return the first
	if list[0].Detail == list[1].Detail && list[1].Detail == list[2].Detail {
		t.Error("all details identical; expected distinct errors from Unwrap")
	}
	// Response should serialize to 3 errors
	resp := ErrorResponse{Errors: list}
	data, _ := json.Marshal(resp)
	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if len(decoded.Errors) != 3 {
		t.Fatalf("round-trip len(Errors) = %d, want 3", len(decoded.Errors))
	}
}

func TestQueryRuleSet_ReturnsParameterInSource(t *testing.T) {
	// Validate query rule set wraps errors with parameter (reserved "filter" triggers error)
	rs := QueryStringBaseRuleSet
	values := url.Values{}
	values.Set("filter", "x") // all-lowercase reserved param triggers validation error
	var out url.Values
	errs := rs.Apply(context.Background(), values, &out)
	if errs == nil {
		t.Fatal("expected validation errors")
	}
	list := ErrorsFromValidationError(errs, SourceParameter)
	if len(list) == 0 {
		t.Fatal("expected at least one error")
	}
	for i, e := range list {
		if e.Source != nil && e.Source.Parameter == "" && e.Source.Pointer != "" {
			t.Errorf("query error[%d]: expected source.parameter, got pointer=%q", i, e.Source.Pointer)
		}
	}
}

func TestSingleRuleSet_ReturnsPointerInSource(t *testing.T) {
	// Validate body rule set wraps errors with pointer (e.g. invalid JSON)
	rs := NewSingleRuleSet[map[string]any]("articles", Attributes())
	var out SingleDatumEnvelope[map[string]any]
	errs := rs.Apply(context.Background(), "not json at all", &out)
	if errs == nil {
		t.Fatal("expected validation errors")
	}
	list := ErrorsFromValidationError(errs, SourcePointer)
	if len(list) == 0 {
		t.Fatal("expected at least one error")
	}
	for i, e := range list {
		if e.Source != nil && e.Source.Pointer != "" && e.Source.Parameter != "" {
			t.Errorf("body error[%d]: should use either pointer or parameter, not both", i)
		}
	}
}

func TestErrorFromValidationError_PointerToExactMember(t *testing.T) {
	// Source pointer points to the exact member (full path).
	ctx := context.Background()
	ctx = rulecontext.WithPathString(ctx, "data")
	ctx = rulecontext.WithPathString(ctx, "attributes")
	ctx = rulecontext.WithPathString(ctx, "unexpectedKey")
	unexpectedErr := errors.Errorf(errors.CodeUnexpected, ctx, "value was not expected", "value was not expected")

	e := ErrorFromValidationError(unexpectedErr, SourcePointer)
	if e.Source == nil {
		t.Fatal("expected Source to be set")
	}
	if e.Source.Pointer != "/data/attributes/unexpectedKey" {
		t.Errorf("pointer: got %q, want /data/attributes/unexpectedKey (exact member)", e.Source.Pointer)
	}
}

// mockValidationError implements errors.ValidationError for testing docs URI, trace URI, meta, and source.
type mockValidationError struct {
	code     errors.ErrorCode
	title    string
	detail   string
	path     string
	docsURI  string
	traceURI string
	meta     map[string]any
}

func (m *mockValidationError) Error() string                              { return m.detail }
func (m *mockValidationError) Code() errors.ErrorCode                      { return m.code }
func (m *mockValidationError) Path() string                                { return m.path }
func (m *mockValidationError) PathAs(errors.PathSerializer) string        { return m.path }
func (m *mockValidationError) ShortError() string                          { return m.title }
func (m *mockValidationError) DocsURI() string                             { return m.docsURI }
func (m *mockValidationError) TraceURI() string                            { return m.traceURI }
func (m *mockValidationError) Meta() map[string]any                        { return m.meta }
func (m *mockValidationError) Params() []any                               { return nil }
func (m *mockValidationError) Internal() bool                              { return false }
func (m *mockValidationError) Validation() bool                            { return true }
func (m *mockValidationError) Permission() bool                            { return false }
func (m *mockValidationError) Unwrap() []error                             { return nil }

func TestErrorFromValidationError_IncludesDocsURITraceURIAndMeta(t *testing.T) {
	docsURI := "https://docs.example.com/errors/REQUIRED"
	traceURI := "https://trace.example.com/abc123"
	meta := map[string]any{"requestId": "req-1", "field": "name"}
	ve := &mockValidationError{
		code:     errors.CodeRequired,
		title:    "required",
		detail:   "name is required",
		path:     "/data/attributes/name",
		docsURI:  docsURI,
		traceURI: traceURI,
		meta:     meta,
	}

	e := ErrorFromValidationError(ve, SourcePointer)

	// Links: docs URI (about) and trace URI (type) must be present
	if e.Links == nil {
		t.Fatal("expected Links to be set when DocsURI and TraceURI are non-empty")
	}
	if e.Links.About != docsURI {
		t.Errorf("Links.About (docs URI): got %q, want %q", e.Links.About, docsURI)
	}
	if e.Links.Type != traceURI {
		t.Errorf("Links.Type (trace URI): got %q, want %q", e.Links.Type, traceURI)
	}

	// Meta must be present and equal
	if e.Meta == nil {
		t.Fatal("expected Meta to be set when ValidationError.Meta() is non-empty")
	}
	if (*e.Meta)["requestId"] != "req-1" || (*e.Meta)["field"] != "name" {
		t.Errorf("Meta: got %v, want requestId=req-1, field=name", *e.Meta)
	}

	// Other standard fields
	if e.Status != "422" {
		t.Errorf("Status: got %q, want 422", e.Status)
	}
	if e.Code != string(errors.CodeRequired) {
		t.Errorf("Code: got %q", e.Code)
	}
	if e.Title != ve.title || e.Detail != ve.detail {
		t.Errorf("Title/Detail: got %q / %q", e.Title, e.Detail)
	}
	if e.Source == nil || e.Source.Pointer != ve.path {
		t.Errorf("Source.Pointer: got %v, want pointer %q", e.Source, ve.path)
	}
}

func TestErrorFromValidationError_OmitsLinksWhenBothURIsEmpty(t *testing.T) {
	ve := &mockValidationError{
		code:    errors.CodeRequired,
		title:   "required",
		detail:  "field required",
		docsURI: "",
		traceURI: "",
	}
	e := ErrorFromValidationError(ve, SourcePointer)
	if e.Links != nil {
		t.Errorf("expected Links to be nil when both DocsURI and TraceURI are empty, got %+v", e.Links)
	}
}

func TestErrorFromValidationError_OmitsMetaWhenEmpty(t *testing.T) {
	ve := &mockValidationError{
		code:  errors.CodeRequired,
		title: "required",
		detail: "field required",
		meta:  nil,
	}
	e := ErrorFromValidationError(ve, SourcePointer)
	if e.Meta != nil {
		t.Errorf("expected Meta to be nil when ValidationError.Meta() is nil, got %+v", e.Meta)
	}
}

// TestError_JSONIncludesAllErrorFields ensures the serialized JSON:API error object
// includes id, links (about, type), status, code, title, detail, source, and meta.
func TestError_JSONIncludesAllErrorFields(t *testing.T) {
	resp := ErrorResponse{
		Errors: []Error{
			{
				ID:     "err-1",
				Links:  &ErrorLinks{About: "https://docs.example.com/err", Type: "https://trace.example.com/abc"},
				Status: "422",
				Code:   "REQUIRED",
				Title:  "member required",
				Detail: "The 'name' attribute is required.",
				Source: &Source{Pointer: "/data/attributes/name"},
				Meta:   &MetaInfo{"requestId": "req-1", "timestamp": "2025-01-01T00:00:00Z"},
			},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded struct {
		Errors []struct {
			ID     string     `json:"id"`
			Links  *ErrorLinks `json:"links"`
			Status string     `json:"status"`
			Code   string     `json:"code"`
			Title  string     `json:"title"`
			Detail string     `json:"detail"`
			Source *Source    `json:"source"`
			Meta   *MetaInfo  `json:"meta"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Errors) != 1 {
		t.Fatalf("len(errors) = %d, want 1", len(decoded.Errors))
	}
	e := decoded.Errors[0]

	if e.ID != "err-1" {
		t.Errorf("id: got %q, want err-1", e.ID)
	}
	if e.Links == nil {
		t.Fatal("links: expected non-nil (docs URI and trace URI)")
	}
	if e.Links.About != "https://docs.example.com/err" {
		t.Errorf("links.about (docs URI): got %q", e.Links.About)
	}
	if e.Links.Type != "https://trace.example.com/abc" {
		t.Errorf("links.type (trace URI): got %q", e.Links.Type)
	}
	if e.Status != "422" {
		t.Errorf("status: got %q", e.Status)
	}
	if e.Code != "REQUIRED" {
		t.Errorf("code: got %q", e.Code)
	}
	if e.Title != "member required" {
		t.Errorf("title: got %q", e.Title)
	}
	if e.Detail != "The 'name' attribute is required." {
		t.Errorf("detail: got %q", e.Detail)
	}
	if e.Source == nil || e.Source.Pointer != "/data/attributes/name" {
		t.Errorf("source: got %v", e.Source)
	}
	if e.Meta == nil {
		t.Fatal("meta: expected non-nil")
	}
	if (*e.Meta)["requestId"] != "req-1" || (*e.Meta)["timestamp"] != "2025-01-01T00:00:00Z" {
		t.Errorf("meta: got %v", *e.Meta)
	}
}

// TestError_SerializationOmitsEmptyOptionalFields ensures that when optional error fields
// are empty, they are omitted from JSON output entirely (no empty string values).
func TestError_SerializationOmitsEmptyOptionalFields(t *testing.T) {
	t.Run("links_with_only_about", func(t *testing.T) {
		resp := ErrorResponse{Errors: []Error{{
			Status: "422",
			Links:  &ErrorLinks{About: "https://docs.example.com", Type: ""},
		}}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		// Must not contain empty "type" value
		if strings.Contains(string(data), `"type":""`) || strings.Contains(string(data), `"type": ""`) {
			t.Errorf("JSON must not contain empty type; got %s", data)
		}
		if !strings.Contains(string(data), `"about":"https://docs.example.com"`) {
			t.Errorf("JSON should contain about link; got %s", data)
		}
	})

	t.Run("links_with_only_type", func(t *testing.T) {
		resp := ErrorResponse{Errors: []Error{{
			Status: "422",
			Links:  &ErrorLinks{About: "", Type: "https://trace.example.com"},
		}}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(data), `"about":""`) || strings.Contains(string(data), `"about": ""`) {
			t.Errorf("JSON must not contain empty about; got %s", data)
		}
		if !strings.Contains(string(data), `"type":"https://trace.example.com"`) {
			t.Errorf("JSON should contain type link; got %s", data)
		}
	})

	t.Run("empty_optional_top_level_fields_omitted", func(t *testing.T) {
		resp := ErrorResponse{Errors: []Error{{
			ID:     "",
			Status: "422",
			Code:   "",
			Title:  "",
			Detail: "",
		}}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		raw := string(data)
		for _, key := range []string{`"id":""`, `"code":""`, `"title":""`, `"detail":""`} {
			if strings.Contains(raw, key) {
				t.Errorf("JSON must not contain empty optional field %s; got %s", key, data)
			}
		}
		// Should only have status (required)
		if !strings.Contains(raw, `"status":"422"`) {
			t.Errorf("JSON should contain status; got %s", data)
		}
	})

	t.Run("source_with_only_parameter_omits_pointer_and_header", func(t *testing.T) {
		resp := ErrorResponse{Errors: []Error{{
			Status: "422",
			Source: &Source{Parameter: "sort", Pointer: "", Header: ""},
		}}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		raw := string(data)
		if strings.Contains(raw, `"pointer":""`) || strings.Contains(raw, `"header":""`) {
			t.Errorf("JSON must not contain empty pointer or header; got %s", data)
		}
		if !strings.Contains(raw, `"parameter":"sort"`) {
			t.Errorf("JSON should contain parameter; got %s", data)
		}
	})

	t.Run("error_from_validation_error_omits_empty_links_and_meta", func(t *testing.T) {
		ve := &mockValidationError{
			code:     errors.CodeRequired,
			title:    "required",
			detail:   "field required",
			docsURI:  "",
			traceURI: "",
			meta:     nil,
		}
		e := ErrorFromValidationError(ve, SourcePointer)
		resp := ErrorResponse{Errors: []Error{*e}}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		raw := string(data)
		if strings.Contains(raw, `"links"`) {
			t.Errorf("JSON must not include links when both docs and trace URI are empty; got %s", data)
		}
		if strings.Contains(raw, `"meta"`) {
			t.Errorf("JSON must not include meta when empty; got %s", data)
		}
	})
}
