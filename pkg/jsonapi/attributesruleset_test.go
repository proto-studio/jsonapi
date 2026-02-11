package jsonapi_test

import (
	"context"
	"strings"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

func TestAttributesRuleSet_WithKey_valid(t *testing.T) {
	rs := jsonapi.Attributes().
		WithKey("title", rules.String().Any()).
		WithKey("createdAt", rules.String().Any())

	input := map[string]any{"title": "Hello", "createdAt": "2024-01-01"}
	var out map[string]any
	errs := rs.Apply(context.Background(), input, &out)
	if errs != nil {
		t.Fatalf("Apply: %s", errs)
	}
	if out["title"] != "Hello" {
		t.Errorf("expected title=Hello, got %v", out["title"])
	}
}

func TestAttributesRuleSet_WithKey_invalid_panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for invalid attribute name")
			return
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "jsonapi: attribute name") {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()
	jsonapi.Attributes().WithKey("field+name", rules.String().Any())
}

func TestAttributesRuleSet_WithKeyUnsafe(t *testing.T) {
	// WithKeyUnsafe allows names that would fail MemberNameRule
	rs := jsonapi.Attributes().WithKeyUnsafe("field.name", rules.String().Any())

	input := map[string]any{"field.name": "value"}
	var out map[string]any
	errs := rs.Apply(context.Background(), input, &out)
	if errs != nil {
		t.Fatalf("Apply: %s", errs)
	}
	if out["field.name"] != "value" {
		t.Errorf("expected field.name=value, got %v", out["field.name"])
	}
}

func TestAttributesRuleSet_WithUnknown(t *testing.T) {
	rs := jsonapi.Attributes().
		WithKey("title", rules.String().Any()).
		WithUnknown()

	input := map[string]any{"title": "Hi", "custom": "allowed"}
	var out map[string]any
	errs := rs.Apply(context.Background(), input, &out)
	if errs != nil {
		t.Fatalf("Apply: %s", errs)
	}
	if out["custom"] != "allowed" {
		t.Errorf("expected custom=allowed, got %v", out["custom"])
	}
}

func TestAttributesRuleSet_DelegatedMethods(t *testing.T) {
	rs := jsonapi.Attributes().
		WithKey("title", rules.String().Any()).
		WithRequired().
		WithJson().
		WithErrorMessage("short", "long").
		WithDocsURI("https://docs.example.com").
		WithTraceURI("https://trace.example.com").
		WithRule(rules.RuleFunc[map[string]any](func(ctx context.Context, m map[string]any) errors.ValidationError { return nil })).
		WithRuleFunc(func(ctx context.Context, m map[string]any) errors.ValidationError { return nil })

	// Cover Evaluate, Required, String, Replaces, Any
	ctx := context.Background()
	input := map[string]any{"title": "Hi"}
	_ = rs.Evaluate(ctx, input)
	if !rs.Required() {
		t.Error("expected Required true after WithRequired")
	}
	if s := rs.String(); s == "" {
		t.Error("expected non-empty String()")
	}
	_ = rs.Replaces(nil)
	anyRS := rs.Any()
	if anyRS == nil {
		t.Error("expected non-nil Any()")
	}
	var out map[string]any
	errs := anyRS.Apply(ctx, input, &out)
	if errs != nil {
		t.Fatalf("Apply via Any: %s", errs)
	}
}

func TestAttributesRuleSet_WithConditionalKeyAndDynamic(t *testing.T) {
	// Conditional: use a Rule that implements Conditional (e.g. ObjectRuleSet)
	cond := rules.StringMap[any]()
	keyRule := jsonapi.MemberNameRule{}
	rs := jsonapi.Attributes().
		WithConditionalKey("opt", cond, rules.String().Any()).
		WithConditionalKeyUnsafe("unsafeOpt", cond, rules.String().Any()).
		WithDynamicKey(keyRule, rules.String().Any()).
		WithDynamicBucket(keyRule, "bucket").
		WithConditionalDynamicBucket(keyRule, cond, "bucket2").
		WithErrorCode(errors.CodeRequired).
		WithErrorMeta("k", "v").
		WithErrorCallback(func(ctx context.Context, err errors.ValidationError) errors.ValidationError { return err })

	keys := rs.KeyRules()
	if keys == nil {
		t.Error("KeyRules should not be nil")
	}

	ctx := context.Background()
	input := map[string]any{"opt": "val", "unsafeOpt": "x"}
	var out map[string]any
	errs := rs.Apply(ctx, input, &out)
	if errs != nil {
		t.Fatalf("Apply: %s", errs)
	}
}
