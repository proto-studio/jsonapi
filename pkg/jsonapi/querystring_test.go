package jsonapi_test

import (
	"context"
	"net/url"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
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

	var out jsonapi.QueryData
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &out)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if out.Fields == nil {
		t.Fatalf("Expected Fields to not be nil")
	}

	articles, ok := out.Fields["fields[articles]"]

	if !ok {
		t.Fatalf("Expected fields[articles] to be set")
	}

	allFields := articles.Values()

	if len(allFields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(allFields))
	}

	if !articles.Contains("abc") {
		t.Error(`Expected fields to contain "abc"`)
	}

	if !articles.Contains("xyz") {
		t.Error(`Expected fields to contain "xyz"`)
	}

	if articles.Contains("qrs") {
		t.Error(`Expected fields to not contain "qrs"`)
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

	var out jsonapi.QueryData
	verrs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &out)
	if verrs != nil {
		t.Fatalf("Expected validation error to be nil, got: %s", verrs)
	}

	if out.Sort == nil {
		t.Fatalf("Expected Sort to not be nil")
	}

	if len(out.Sort) != 2 {
		t.Fatalf("Expected len(Sort) to be 2, got %d", len(out.Sort))
	}

	if out.Sort[0].Field != "abc" {
		t.Errorf(`Expected Sort[0].Field to be "abc", got: "%s"`, out.Sort[0].Field)
	}
	if out.Sort[0].Descending {
		t.Errorf(`Expected Sort[0].Descending to be false`)
	}

	if out.Sort[1].Field != "xyz" {
		t.Errorf(`Expected Sort[1].Field to be "xyz", got: "%s"`, out.Sort[1].Field)
	}
	if !out.Sort[1].Descending {
		t.Errorf(`Expected Sort[1].Descending to be true`)
	}
}

// Requirements:
// - Unexpected fields produce an error
func TestQueryUnexpected(t *testing.T) {
	qs := `foo=bar`
	query, err := url.ParseQuery(qs)
	if err != nil {
		t.Fatalf("Expected parse error to be nil, got: %s", err)
	}

	testhelpers.MustNotApply(t, jsonapi.QueryStringBaseRuleSet.Any(), query, errors.CodeUnexpected)
}
