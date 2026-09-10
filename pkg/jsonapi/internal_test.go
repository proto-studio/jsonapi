package jsonapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// TestDoNotExtend_ResourceLinkage ensures doNotExtend is callable (for coverage).
func TestDoNotExtend_ResourceLinkage(t *testing.T) {
	ResourceIdentifierLinkage{}.doNotExtend()
	NilResourceLinkage{}.doNotExtend()
	ResourceLinkageCollection(nil).doNotExtend()
}

// TestDoNotExtend_ValueList ensures ValueList.doNotExtend is callable (for coverage).
func TestDoNotExtend_ValueList(t *testing.T) {
	v := NewFieldList("a", "b")
	v.doNotExtend()
}

// TestResourceLinkageRuleSet_InterfaceMethods covers Evaluate, Any, String, Replaces, Required.
func TestResourceLinkageRuleSet_InterfaceMethods(t *testing.T) {
	ctx := context.Background()
	_ = ResourceLinkageRuleSet.Evaluate(ctx, NilResourceLinkage{})
	anyRS := ResourceLinkageRuleSet.Any()
	if anyRS == nil {
		t.Error("ResourceLinkageRuleSet.Any() should not be nil")
	}
	if s := ResourceLinkageRuleSet.String(); s != "ResourceLinkageRuleSet" {
		t.Errorf("String(): got %q", s)
	}
	_ = ResourceLinkageRuleSet.Replaces(nil)
	if ResourceLinkageRuleSet.Required() {
		t.Error("ResourceLinkageRuleSet.Required() should be false")
	}
}

// TestRelationshipRuleSet_InterfaceMethods covers Evaluate, Any, String, Replaces, Required.
func TestRelationshipRuleSet_InterfaceMethods(t *testing.T) {
	ctx := context.Background()
	rel := Relationship{Data: NilResourceLinkage{}}
	_ = RelationshipRuleSet.Evaluate(ctx, rel)
	anyRS := RelationshipRuleSet.Any()
	if anyRS == nil {
		t.Error("RelationshipRuleSet.Any() should not be nil")
	}
	if s := RelationshipRuleSet.String(); s != "RelationshipRuleSet" {
		t.Errorf("String(): got %q", s)
	}
	_ = RelationshipRuleSet.Replaces(nil)
	if RelationshipRuleSet.Required() {
		t.Error("RelationshipRuleSet.Required() should be false")
	}
}

// TestQueryParamAdapter covers the adapter's Apply (string branch), Evaluate, Required, String, Replaces, Any.
func TestQueryParamAdapter(t *testing.T) {
	ctx := context.Background()
	adapter := &queryParamAdapter{inner: sortRuleSet.Any()}
	out, errs := adapter.Apply(ctx, "-name")
	if errs != nil {
		t.Fatalf("adapter Apply: %s", errs)
	}
	outSlice := out.([]SortParam)
	if len(outSlice) != 1 || outSlice[0].Field != "name" || !outSlice[0].Descending {
		t.Errorf("expected [{name desc}], got %v", out)
	}
	_ = adapter.Evaluate(ctx, "x")
	_ = adapter.Required()
	_ = adapter.String()
	_ = adapter.Replaces(nil)
	anyRS := adapter.Any()
	if anyRS == nil {
		t.Error("Any() should not be nil")
	}
}

// TestLinksRuleSet_LinkCast covers linkCast with map and invalid type (for coverage).
func TestLinksRuleSet_LinkCast(t *testing.T) {
	ctx := context.Background()
	// linkCast with map (full link)
	out, errs := LinksRuleSet.Apply(ctx, map[string]any{
		"self": map[string]any{"href": "https://example.com"},
	})
	if errs != nil {
		t.Fatalf("LinksRuleSet Apply with full link map: %s", errs)
	}
	if out["self"].(*FullLink).HrefValue != "https://example.com" {
		t.Errorf("expected href, got %v", out["self"])
	}
	// linkCast with invalid type
	_, errs = LinksRuleSet.Apply(ctx, map[string]any{"self": 123})
	if errs == nil {
		t.Error("expected error for invalid link type")
	}
}

// TestJSONAPIErrorWrapper_InterfaceMethods covers wrapper methods used when errors are converted.
func TestJSONAPIErrorWrapper_InterfaceMethods(t *testing.T) {
	ctx := context.Background()
	// Create a validation error and wrap it so we get a jsonAPIErrorWrapper.
	err := errors.Errorf(errors.CodeRequired, ctx, "short", "detail")
	wrapped := ToJSONAPIErrors(err, SourcePointer)
	unwrap := errors.Unwrap(wrapped)
	if len(unwrap) != 1 {
		t.Fatalf("expected 1 unwrapped error, got %d", len(unwrap))
	}
	ve, ok := unwrap[0].(errors.ValidationError)
	if !ok {
		t.Fatal("expected ValidationError")
	}
	// PathAs (with JSON pointer for body errors)
	_ = ve.PathAs(jsonPointerSerializer)
	// Params
	_ = ve.Params()
	// Internal/Validation/Permission depend on error code
	_ = ve.Internal()
	_ = ve.Validation()
	_ = ve.Permission()
}

// TestJSONAPIErrorWrapper_InternalValidationPermission covers error type classification.
func TestJSONAPIErrorWrapper_InternalValidationPermission(t *testing.T) {
	ctx := context.Background()
	// Internal-type code
	internalErr := errors.Errorf(errors.CodeInternal, ctx, "internal", "internal error")
	wrapped := ToJSONAPIErrors(internalErr, SourceParameter)
	unwrap := errors.Unwrap(wrapped)
	ve := unwrap[0].(errors.ValidationError)
	if !ve.Internal() {
		t.Error("expected Internal() true for CodeInternal")
	}
	// Validation-type (CodeRequired)
	valErr := errors.Errorf(errors.CodeRequired, ctx, "required", "required")
	wrapped = ToJSONAPIErrors(valErr, SourceParameter)
	unwrap = errors.Unwrap(wrapped)
	ve = unwrap[0].(errors.ValidationError)
	if !ve.Validation() {
		t.Error("expected Validation() true for CodeRequired")
	}
}

// TestJSONAPIErrorWrapper_PathSource covers Path() with Source.Header and Source.Parameter.
func TestJSONAPIErrorWrapper_PathSource(t *testing.T) {
	e := &Error{
		Detail: "header error",
		Source: &Source{Header: "Content-Type"},
	}
	w := &jsonAPIErrorWrapper{err: e}
	if w.Path() != "Content-Type" {
		t.Errorf("Path() with Header: got %q", w.Path())
	}
	e.Source = &Source{Parameter: "sort"}
	if w.Path() != "sort" {
		t.Errorf("Path() with Parameter: got %q", w.Path())
	}
	e.Source = &Source{Pointer: "/data/attributes/name"}
	if w.Path() != "/data/attributes/name" {
		t.Errorf("Path() with Pointer: got %q", w.Path())
	}
}

// TestJSONAPIErrorWrapper_DocsURITraceURIMeta covers Links and Meta.
func TestJSONAPIErrorWrapper_DocsURITraceURIMeta(t *testing.T) {
	e := &Error{
		Links: &ErrorLinks{About: "https://docs.example.com", Type: "https://trace.example.com"},
		Meta:  &MetaInfo{"key": "value"},
	}
	w := &jsonAPIErrorWrapper{err: e}
	if w.DocsURI() != "https://docs.example.com" {
		t.Errorf("DocsURI: got %q", w.DocsURI())
	}
	if w.TraceURI() != "https://trace.example.com" {
		t.Errorf("TraceURI: got %q", w.TraceURI())
	}
	if m := w.Meta(); len(m) != 1 || m["key"] != "value" {
		t.Errorf("Meta: got %v", m)
	}
}

func TestErrorsFromValidationError_NilAndNonJSONAPI(t *testing.T) {
	if got := ErrorsFromValidationError(nil, SourcePointer); got != nil {
		t.Errorf("nil error: got %v, want nil", got)
	}

	ctx := context.Background()
	ve := errors.Errorf(errors.CodeRequired, ctx, "required", "field is required")
	list := ErrorsFromValidationError(ve, SourcePointer)
	if len(list) != 1 {
		t.Fatalf("plain ValidationError: got %d, want 1", len(list))
	}
	if list[0].Code != string(errors.CodeRequired) {
		t.Errorf("code: got %q", list[0].Code)
	}
}

type mixedUnwrapError struct {
	ve    errors.ValidationError
	inner []error
}

func (m mixedUnwrapError) Error() string                       { return m.ve.Error() }
func (m mixedUnwrapError) Code() errors.ErrorCode              { return m.ve.Code() }
func (m mixedUnwrapError) Path() string                        { return m.ve.Path() }
func (m mixedUnwrapError) PathAs(errors.PathSerializer) string { return m.ve.Path() }
func (m mixedUnwrapError) ShortError() string                  { return m.ve.ShortError() }
func (m mixedUnwrapError) DocsURI() string                     { return m.ve.DocsURI() }
func (m mixedUnwrapError) TraceURI() string                    { return m.ve.TraceURI() }
func (m mixedUnwrapError) Meta() map[string]any                { return m.ve.Meta() }
func (m mixedUnwrapError) Params() []any                       { return nil }
func (m mixedUnwrapError) Internal() bool                      { return false }
func (m mixedUnwrapError) Validation() bool                    { return true }
func (m mixedUnwrapError) Permission() bool                    { return false }
func (m mixedUnwrapError) Unwrap() []error                     { return m.inner }

func TestErrorsFromValidationError_SkipsNonValidationError(t *testing.T) {
	ctx := context.Background()
	ve := errors.Errorf(errors.CodeRequired, ctx, "required", "field is required")
	holder := mixedUnwrapError{
		ve:    ve,
		inner: []error{fmt.Errorf("plain"), ve},
	}
	list := ErrorsFromValidationError(holder, SourcePointer)
	if len(list) != 1 {
		t.Fatalf("got %d errors, want 1 (plain error skipped)", len(list))
	}
}

func TestLinkCast_NilMarshalAndInvalidObject(t *testing.T) {
	ctx := context.Background()

	out, errs := linkCast(ctx, nil)
	if errs != nil {
		t.Fatalf("nil link: %s", errs)
	}
	if _, ok := out.(NilLink); !ok {
		t.Errorf("expected NilLink, got %T", out)
	}

	_, errs = LinkRuleSet.Apply(ctx, map[string]any{"href": make(chan int)})
	if errs == nil {
		t.Fatal("expected marshal error for channel in link object")
	}

	_, errs = LinkRuleSet.Apply(ctx, map[string]any{"href": []any{"not", "a", "string"}})
	if errs == nil {
		t.Fatal("expected unmarshal error for non-string href")
	}
}

func TestResourceLinkageCast_InvalidJSON(t *testing.T) {
	_, errs := ResourceLinkageRuleSet.Apply(context.Background(), "{not json")
	if errs == nil {
		t.Fatal("expected error for invalid JSON string")
	}
}

func TestRelationshipRuleSet_ApplyErrorAndNilDataCast(t *testing.T) {
	ctx := context.Background()
	_, errs := RelationshipRuleSet.Apply(ctx, map[string]any{"links": 123})
	if errs == nil {
		t.Fatal("expected error for invalid links value")
	}

	out, errs := relationshipDataCast(ctx, nil)
	if errs != nil {
		t.Fatalf("nil relationship data: %s", errs)
	}
	if _, ok := out.(NilResourceLinkage); !ok {
		t.Errorf("expected NilResourceLinkage, got %T", out)
	}
}

func TestDatumUnmarshalJSON_DecodeErrors(t *testing.T) {
	var d Datum[map[string]any]
	if err := d.UnmarshalJSON([]byte("not-json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	orig := jsonUnmarshal
	t.Cleanup(func() { jsonUnmarshal = orig })

	t.Run("at-member", func(t *testing.T) {
		jsonUnmarshal = func(data []byte, v any) error {
			if _, ok := v.(*any); ok {
				return fmt.Errorf("forced at-member decode error")
			}
			return orig(data, v)
		}
		var datum Datum[map[string]any]
		if err := datum.UnmarshalJSON([]byte(`{"id":"1","type":"t","@foo":"bar"}`)); err == nil {
			t.Fatal("expected forced at-member decode error")
		}
	})

	t.Run("extension-member", func(t *testing.T) {
		jsonUnmarshal = func(data []byte, v any) error {
			if _, ok := v.(*any); ok {
				return fmt.Errorf("forced extension decode error")
			}
			return orig(data, v)
		}
		var datum Datum[map[string]any]
		if err := datum.UnmarshalJSON([]byte(`{"id":"1","type":"t","ext:foo":"bar"}`)); err == nil {
			t.Fatal("expected forced extension decode error")
		}
	})
}

func TestApplyDatum_Nil(t *testing.T) {
	type testDatum struct {
		Name string
	}
	attrs := rules.Struct[testDatum]().WithKey("Name", rules.String().Any())
	out, errs := applyDatum(context.Background(), NewDatumRuleSet[testDatum]("tests", attrs), nil)
	if errs != nil {
		t.Fatalf("nil datum: %s", errs)
	}
	if out.Type != "" {
		t.Errorf("expected zero datum, got %+v", out)
	}
}

func TestSortRuleSet_MultipleValues(t *testing.T) {
	_, errs := sortRuleSet.Apply(context.Background(), []string{"a", "b"})
	if errs == nil {
		t.Fatal("expected error when sort has more than one value")
	}
}

func TestHeaderRuleSet_NilHeaderRulesAndInvalidContentType(t *testing.T) {
	ctx := context.Background()

	rs := (&HeaderRuleSet{}).WithHeader("X-Request-Id", rules.String().Any())
	h := http.Header{}
	h.Set("Content-Type", MediaTypeJSONAPI)
	h.Set("X-Request-Id", "abc")
	if _, err := rs.Apply(ctx, h); err != nil {
		t.Fatalf("WithHeader on empty HeaderRuleSet: %v", err)
	}

	cloned := Headers().WithHeader("X-Request-Id", rules.String().WithMinLen(1).Any()).WithExt(rules.String().Any())
	h.Set("X-Request-Id", "ok")
	if _, err := cloned.Apply(ctx, h); err != nil {
		t.Fatalf("clone with header rules: %v", err)
	}

	withNilRule := &HeaderRuleSet{
		contentRequired: false,
		headerRules:     map[string]rules.RuleSet[any]{"X-Ignored": nil},
	}
	if err := withNilRule.Evaluate(ctx, http.Header{}); err != nil {
		t.Fatalf("nil header rule should be skipped: %v", err)
	}

	invalid := http.Header{}
	invalid.Set("Content-Type", `application/vnd.api+json; charset="`)
	if _, err := Headers().Apply(ctx, invalid); err == nil {
		t.Fatal("expected error for invalid Content-Type parameters")
	}

	empty := httpHeaderToHeader(http.Header{})
	if empty.Version != Version_1_1 {
		t.Errorf("empty headers version: got %q", empty.Version)
	}

	parsed := httpHeaderToHeader(invalid)
	if parsed.Version != Version_1_1 {
		t.Errorf("invalid Content-Type version: got %q", parsed.Version)
	}
}
