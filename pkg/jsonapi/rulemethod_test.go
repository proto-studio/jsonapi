package jsonapi_test

import (
	"context"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
)

func TestHTTPMethodRule_String(t *testing.T) {
	rule := jsonapi.HTTPMethodRule[[]string, string]("GET", "POST")

	str := rule.String()
	expected := "HttpMethod([GET POST])"
	if str != expected {
		t.Errorf("Expected String() to return %q, got %q", expected, str)
	}

	// Test with single method
	rule2 := jsonapi.HTTPMethodRule[[]string, string]("GET")
	str2 := rule2.String()
	expected2 := "HttpMethod([GET])"
	if str2 != expected2 {
		t.Errorf("Expected String() to return %q, got %q", expected2, str2)
	}
}

func TestHTTPMethodRule_KeyRules(t *testing.T) {
	rule := jsonapi.HTTPMethodRule[[]string, string]("GET", "POST")

	keyRules := rule.KeyRules()
	if len(keyRules) != 0 {
		t.Errorf("Expected KeyRules() to return empty slice, got %d rules", len(keyRules))
	}
}

func TestIndexRule_String(t *testing.T) {
	rule := jsonapi.IndexRule[[]string, string]()

	// IndexRule behavior depends on context (method/ID); testhelpers use context.TODO(), so test manually.
	ctx := context.Background()
	value := []string{"test"}

	errs := rule.Evaluate(ctx, value)
	if errs != nil {
		t.Errorf("Expected no errors for index rule with no method/ID, got: %s", errs)
	}

	ctx = jsonapi.WithMethod(ctx, "GET")
	errs = rule.Evaluate(ctx, value)
	if errs != nil {
		t.Errorf("Expected no errors for index rule with GET and no ID, got: %s", errs)
	}

	ctx = jsonapi.WithId(ctx, "123")
	errs = rule.Evaluate(ctx, value)
	if errs == nil {
		t.Error("Expected errors for index rule with GET and ID")
	}
}

func TestHTTPMethodRule_Conflict(t *testing.T) {
	rule1 := jsonapi.HTTPMethodRule[[]string, string]("GET", "POST")
	rule2 := jsonapi.HTTPMethodRule[[]string, string]("GET")
	if !rule1.Conflict(rule2) {
		t.Error("expected Conflict true for two HTTPMethodRules")
	}
}

