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

	vals, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	vals, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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
	out, err := rs.Apply(context.Background(), "sort=title")
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
	_, errs := anyRS.Apply(ctx, vals)
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
	out, err := rs.Apply(context.Background(), "unknownparam=value")
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(context.Background(), query)
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	vals, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	vals, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if verrs == nil {
		t.Fatalf("Expected validation error for filter with ID, got nil")
	}
}

func queryParamErrors(t *testing.T, verrs errors.ValidationError) []jsonapi.Error {
	t.Helper()
	return jsonapi.ErrorsFromValidationError(verrs, jsonapi.SourceParameter)
}

func requireParamError(t *testing.T, list []jsonapi.Error, parameter string) {
	t.Helper()
	for _, e := range list {
		if e.Source != nil && e.Source.Parameter == parameter {
			return
		}
	}
	t.Fatalf("missing error for parameter %q in %+v", parameter, list)
}

// Requirements:
//   - filter[*] on a member GET is rejected once.
//   - WithParam / WithParamUnsafe stack additional param rules without re-running
//     the existing whole-query rule (the same filter error must not be duplicated).
func TestQueryStringBaseRuleSet_FilterOnMemberGet_WithParamDoesNotDuplicate(t *testing.T) {
	ctx := jsonapi.WithId(jsonapi.WithMethod(context.Background(), "GET"), "K7zPsVDfP")
	parsed, err := url.ParseQuery("filter[tenant]=GimsxMicP&include=tenant,permissions")
	if err != nil {
		t.Fatal(err)
	}

	assertOnlyFilter := func(t *testing.T, label string, verrs errors.ValidationError) {
		t.Helper()
		list := queryParamErrors(t, verrs)
		if n := len(list); n != 1 {
			t.Fatalf("%s: got %d errors, want 1: %v", label, n, verrs)
		}
		if list[0].Source == nil || list[0].Source.Parameter != "filter[tenant]" {
			t.Fatalf("%s: want source.parameter=filter[tenant], got %+v", label, list[0].Source)
		}
		if list[0].Code != string(errors.CodeForbidden) {
			t.Fatalf("%s: code=%q, want %s", label, list[0].Code, errors.CodeForbidden)
		}
	}

	assertOnlyFilter(t, "base", jsonapi.QueryStringBaseRuleSet.Evaluate(ctx, parsed))

	// Same fork protoauth-core uses: replace include, then add fields[roles].
	forked := jsonapi.QueryStringBaseRuleSet.
		WithParam("include", rules.String().Any()).
		WithParamUnsafe("fields[roles]", rules.String().Any())
	assertOnlyFilter(t, "after WithParam stack", forked.Evaluate(ctx, parsed))

	t.Run("distinct params still stack", func(t *testing.T) {
		withReserved, err := url.ParseQuery("filter[tenant]=GimsxMicP&include=tenant,permissions&unknownparam=x")
		if err != nil {
			t.Fatal(err)
		}
		list := queryParamErrors(t, forked.Evaluate(ctx, withReserved))
		if n := len(list); n != 2 {
			t.Fatalf("got %d errors, want 2 distinct (filter + reserved param), not copies of the same rule: %v", n, list)
		}
		requireParamError(t, list, "filter[tenant]")
		requireParamError(t, list, "unknownparam")
	})
}

func TestQueryStringFilter_NonGET_Forbidden(t *testing.T) {
	qs := `filter[status]=active`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "POST")

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	vals, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
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

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if verrs == nil {
		t.Fatalf("Expected validation error for page[size] with ID, got nil")
	}
}

func TestQueryStringFields_MultipleValues(t *testing.T) {
	parsed := url.Values{}
	parsed.Add("fields[articles]", "abc")
	parsed.Add("fields[articles]", "xyz")

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(context.Background(), parsed)
	if verrs == nil {
		t.Fatal("expected error when fields[articles] has more than one value")
	}
}

func TestQueryStringSort_EmptyItems(t *testing.T) {
	parsed, err := url.ParseQuery("sort=abc,,xyz")
	if err != nil {
		t.Fatalf("parse: %s", err)
	}

	vals, verrs := jsonapi.QueryStringBaseRuleSet.Apply(context.Background(), parsed)
	if verrs != nil {
		t.Fatalf("expected empty sort items to be skipped, got: %s", verrs)
	}
	if vals.Get("sort") != "abc,,xyz" {
		t.Errorf("expected sort=abc,,xyz, got %q", vals.Get("sort"))
	}
}

func TestQueryStringPageSize_InvalidRange(t *testing.T) {
	qs := `page[size]=0`
	parsed, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	ctx := jsonapi.WithMethod(context.Background(), "GET")

	_, verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if verrs == nil {
		t.Fatalf("Expected validation error for page[size]=0, got nil")
	}

	qs = `page[size]=101`
	parsed, err = url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	_, verrs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if verrs == nil {
		t.Fatalf("Expected validation error for page[size]=101, got nil")
	}
}
