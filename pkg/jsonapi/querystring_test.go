package jsonapi_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
// - Returns a field list with the appropriate fields set.
// - Errors on fields with more than one value.
// - Requires Related to be registered.
// - Errors if field is not recognized.
func TestQueryStringFields(t *testing.T) {
	qs := `fields[articles]=abc,xyz`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := context.Background()

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if v := vals.Get("fields[articles]"); v != "abc,xyz" {
		t.Fatalf("Expected fields[articles]=abc,xyz, got %q", v)
	}
}

// Requirements:
// - Returns a sort list.
// - The sort list has ascending and descending set appropriately.
// - Sort fields are in the correct order.
func TestQueryStringSort(t *testing.T) {
	qs := `sort=abc,-xyz`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := context.Background()

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if v := vals.Get("sort"); v != "abc,-xyz" {
		t.Errorf("Expected sort=abc,-xyz, got %q", v)
	}
}

// TestQueryRuleSet_WithParam_legal verifies that Query() and WithParam with legal keys work.
func TestQueryRuleSet_WithParam_legal(t *testing.T) {
	rs := jsonapi.Query().WithParam("sort", rules.String().Any())
	var out url.Values
	err := rs.Apply(context.Background(), "sort=title", &out)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if out.Get("sort") != "title" {
		t.Errorf("expected sort=title, got %q", out.Get("sort"))
	}
}

func TestQueryRuleSet_EvaluateRequiredStringReplacesAny(t *testing.T) {
	rs := jsonapi.Query().WithParam("sort", rules.String().Any())
	ctx := context.Background()
	vals := url.Values{}
	vals.Set("sort", "title")
	_ = rs.Evaluate(ctx, vals)
	if rs.Required() {
		t.Error("expected Required false for default query rule set")
	}
	if s := rs.String(); s == "" {
		t.Error("expected non-empty String()")
	}
	_ = rs.Replaces(nil)
	anyRS := rs.Any()
	if anyRS == nil {
		t.Error("Any() should not be nil")
	}
	var out url.Values
	errs := anyRS.Apply(ctx, vals, &out)
	if errs != nil {
		t.Fatalf("Apply via Any: %s", errs)
	}
}

// TestQueryRuleSet_MustApplyTypes verifies Apply supports expected output types (protovalidate test helper).
func TestQueryRuleSet_MustApplyTypes(t *testing.T) {
	parsed, _ := url.ParseQuery("sort=title")
	rs := jsonapi.Query().WithParam("sort", rules.String().Any())
	testhelpers.MustApplyTypes(t, rs, parsed)
}

// TestQueryRuleSet_WithParam_illegal verifies that WithParam panics for all-lowercase reserved keys.
func TestQueryRuleSet_WithParam_illegal(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for illegal param name")
		} else if msg, ok := r.(string); !ok || !strings.Contains(msg, "illegal per JSON:API") {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()
	jsonapi.Query().WithParam("unknownparam", rules.String().Any())
}

// TestQueryRuleSet_WithParamUnsafe verifies that WithParamUnsafe does not panic for any key.
func TestQueryRuleSet_WithParamUnsafe(t *testing.T) {
	rs := jsonapi.Query().WithParamUnsafe("unknownparam", rules.String().Any())
	var out url.Values
	err := rs.Apply(context.Background(), "unknownparam=value", &out)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if out.Get("unknownparam") != "value" {
		t.Errorf("expected unknownparam=value, got %q", out.Get("unknownparam"))
	}
}

// TestQueryRuleSet_standardParams verifies that standard JSON:API param names are legal.
func TestQueryRuleSet_standardParams(t *testing.T) {
	for _, name := range []string{"sort", "include"} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("WithParam(%q) should not panic: %v", name, r)
				}
			}()
			_ = jsonapi.Query().WithParam(name, rules.String().Any())
		}()
	}
}

// Requirements:
// - JSON:API allows undefined query params that contain non-lowercase (implementation-specific).
// - All-lowercase unknown params are reserved and rejected.
func TestQueryUnexpected(t *testing.T) {
	qs := `Foo=bar`
	query, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(context.Background(), query, &vals)
	if verrs != nil {
		t.Errorf("Implementation-specific query param (with non-lowercase) should be allowed; got: %v", verrs)
	}
}

// TestQueryParamNameRule verifies the exported rule so callers can validate before WithParam.
func TestQueryParamNameRule(t *testing.T) {
	rule := jsonapi.QueryParamNameRule{}

	valid := []string{"sort", "include", "page[size]", "fields[articles]", "filter[x]", "ext:foo", "camelCase", "my_param"}
	for _, name := range valid {
		testhelpers.MustEvaluate(t, rule, name)
	}
	invalid := []string{"foo", "unknownparam", "bar", "filter", "page", "fields"}
	for _, name := range invalid {
		testhelpers.MustNotEvaluate(t, rule, name, errors.CodeUnexpected)
	}

	if rule.Replaces(nil) {
		t.Error("QueryParamNameRule.Replaces should be false")
	}
	if s := rule.String(); s != "QueryParamNameRule" {
		t.Errorf("String(): got %q", s)
	}
}

func TestFieldListMap_doNotExtend(t *testing.T) {
	// This is a marker method, just verify it exists and doesn't panic
	// We can't directly test the unexported fieldListMap type, but we can test
	// that NewFieldList returns a ValueList that implements the interface
	fl := jsonapi.NewFieldList("field1", "field2")

	// Verify it works as expected
	if !fl.Contains("field1") {
		t.Error("Expected field1 to be contained")
	}
	if !fl.Contains("field2") {
		t.Error("Expected field2 to be contained")
	}
	if fl.Contains("field3") {
		t.Error("Expected field3 to not be contained")
	}

	values := fl.Values()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

func TestQueryStringFields_DELETE_Forbidden(t *testing.T) {
	qs := `fields[articles]=abc,xyz`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "DELETE")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for DELETE method, got nil")
	}

	// Check that the error is about fields not being allowed on DELETE
	found := false
	for _, err := range errors.Unwrap(verrs) {
		if ve, ok := err.(errors.ValidationError); ok && ve.Code() == errors.CodeForbidden {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected CodeForbidden error, got: %s", verrs)
	}
}

func TestQueryStringSort_IndexOnly(t *testing.T) {
	// Test that sort works on index GET requests
	qs := `sort=abc,-xyz`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")
	// No ID means it's an index request

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if vals.Get("sort") != "abc,-xyz" {
		t.Fatalf("Expected sort=abc,-xyz, got %q", vals.Get("sort"))
	}
}

func TestQueryStringSort_WithId_Forbidden(t *testing.T) {
	qs := `sort=abc`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")
	ctx = jsonapi.WithId(ctx, "123") // Has ID, so not an index request

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for sort with ID, got nil")
	}

	found := false
	for _, err := range errors.Unwrap(verrs) {
		if ve, ok := err.(errors.ValidationError); ok && ve.Code() == errors.CodeForbidden {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected CodeForbidden error, got: %s", verrs)
	}
}

func TestQueryStringSort_NonGET_Forbidden(t *testing.T) {
	qs := `sort=abc`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "POST")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for sort on POST, got nil")
	}

	found := false
	for _, err := range errors.Unwrap(verrs) {
		if ve, ok := err.(errors.ValidationError); ok && ve.Code() == errors.CodeForbidden {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected CodeForbidden error, got: %s", verrs)
	}
}

func TestQueryStringSort_HEAD_Allowed(t *testing.T) {
	qs := `sort=abc`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "HEAD")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil for HEAD, got: %s", verrs)
	}
}

func TestQueryStringFilter_IndexOnly(t *testing.T) {
	qs := `filter[status]=active`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if vals.Get("filter[status]") != "active" {
		t.Errorf("Expected filter[status]=active, got %q", vals.Get("filter[status]"))
	}
}

func TestQueryStringFilter_WithId_Forbidden(t *testing.T) {
	qs := `filter[status]=active`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")
	ctx = jsonapi.WithId(ctx, "123")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for filter with ID, got nil")
	}
}

func TestQueryStringFilter_NonGET_Forbidden(t *testing.T) {
	qs := `filter[status]=active`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "POST")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for filter on POST, got nil")
	}
}

func TestQueryStringPageSize_IndexOnly(t *testing.T) {
	qs := `page[size]=10`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if vals.Get("page[size]") != "10" {
		t.Errorf("Expected page[size]=10, got %q", vals.Get("page[size]"))
	}
}

func TestQueryStringPageSize_WithId_Forbidden(t *testing.T) {
	qs := `page[size]=10`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")
	ctx = jsonapi.WithId(ctx, "123")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for page[size] with ID, got nil")
	}
}

func TestQueryStringPageSize_InvalidRange(t *testing.T) {
	qs := `page[size]=0`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")

	var vals url.Values
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for page[size]=0, got nil")
	}

	qs = `page[size]=101`
	parsed, err = url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	verrs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if verrs == nil {
		t.Fatalf("Expected validation error for page[size]=101, got nil")
	}
}
