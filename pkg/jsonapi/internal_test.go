package jsonapi

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
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
	var out []SortParam
	errs := adapter.Apply(ctx, "-name", &out)
	if errs != nil {
		t.Fatalf("adapter Apply: %s", errs)
	}
	if len(out) != 1 || out[0].Field != "name" || !out[0].Descending {
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
	var out map[string]Link
	errs := LinksRuleSet.Apply(ctx, map[string]any{
		"self": map[string]any{"href": "https://example.com"},
	}, &out)
	if errs != nil {
		t.Fatalf("LinksRuleSet Apply with full link map: %s", errs)
	}
	if out["self"].(*FullLink).HrefValue != "https://example.com" {
		t.Errorf("expected href, got %v", out["self"])
	}
	// linkCast with invalid type
	var out2 map[string]Link
	errs = LinksRuleSet.Apply(ctx, map[string]any{"self": 123}, &out2)
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
