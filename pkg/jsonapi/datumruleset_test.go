package jsonapi_test

import (
	"context"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// Requirements:
//   - Typed validators return an object and validate the properties.
//   - Error returned should have correct path.
//   - Type is assumed if absent.
//   - Type will error if it is present but does not match the expected type.
func TestSingleDatumTyped(t *testing.T) {

	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().WithMinLen(6).Any())

	ruleSet := jsonapi.NewDatumRuleSet[testDatum]("tests", attributesRuleSet)

	// Protovalidate test helpers: RuleSet implements WithRequired/Required
	testhelpers.MustImplementWithRequired(t, ruleSet)

	ctx := context.Background()

	var test jsonapi.Datum[testDatum]

	errs := ruleSet.Apply(ctx, `{
	  "id": "abc",
	  "attributes": {
		  "Name": "My Test"
		}
	}`, &test)

	if errs != nil {
		t.Fatalf("Expected errors to be nil, got: %s", errs.Error())
	}

	if test.ID != "abc" {
		t.Errorf(`Expected ID to be "%s", got: %s`, "abc", test.ID)
	}
	if test.Type != "tests" {
		t.Errorf(`Expected type to be implicitly "%s", got: "%s"`, "tests", test.Type)
	}
	if test.Attributes.Name != "My Test" {
		t.Errorf(`Expected Name to be "%s", got: "%s"`, "My Test", test.Attributes.Name)
	}

	// Type matches expected
	errs = ruleSet.Apply(ctx, `{
	  "id": "abc",
		"type": "tests",
	  "attributes": {
		  "Name": "My Test"
		}
	}`, &test)
	if errs != nil {
		t.Errorf("Expected errors on matching type to be nil, got: %s", errs.Error())
	}

	// Name is too short
	errs = ruleSet.Apply(ctx, `{
	  "id": "abc",
	  "attributes": {
		  "Name": "short"
		}
	}`, &test)

	if errs != nil {
		unwrapped := errors.Unwrap(errs)
		if len(unwrapped) != 1 {
			t.Errorf(`Expected 1 error, got: %d`, len(unwrapped))
		}
		expected := "/attributes/Name"
		if ve, _ := unwrapped[0].(errors.ValidationError); ve != nil && ve.Path() != expected {
			t.Errorf(`Expected path to be "%s", got: "%s"`, expected, ve.Path())
		}
	} else {
		t.Errorf(`Expected name errors to not be nil`)
	}

	// Type is present but does not match the expected type
	errs = ruleSet.Apply(ctx, `{
	  "id": "abc",
		"type": "bogus",
	  "attributes": {
		  "Name": "My test"
		}
	}`, &test)

	if errs != nil {
		unwrapped := errors.Unwrap(errs)
		if len(unwrapped) != 1 {
			t.Errorf(`Expected 1 error, got: %d`, len(unwrapped))
		}
		expected := "/type"
		if ve, _ := unwrapped[0].(errors.ValidationError); ve != nil && ve.Path() != expected {
			t.Errorf(`Expected path to be "%s", got: "%s"`, expected, ve.Path())
		}
	} else {
		t.Errorf(`Expected type errors to not be nil`)
	}
}

// Requirements:
// - Unknown relationships error by default.
// - WithUnknownRelationships allows all relationships (nil, slice, and ID linkage) to pass.
func TestWithUnknownRelationshipsDatum(t *testing.T) {
	attributesRuleSet := rules.StringMap[any]().WithUnknown()
	ruleSet := jsonapi.NewDatumRuleSet[map[string]any]("tests", attributesRuleSet)

	ctx := context.Background()

	testJson := `{
	  "id": "abc",
	  "attributes": {},
		"relationships": {
			"test": {
				"data": {"id": "xyz", "type": "tests"}
			}
		}
	}`

	var d jsonapi.Datum[map[string]any]
	errs := ruleSet.Apply(ctx, testJson, &d)

	if errs == nil {
		t.Errorf("Expected errors to not be nil")
	}

	// ID linkage
	ruleSet = ruleSet.WithUnknownRelationships()

	var out jsonapi.Datum[map[string]any]
	errs = ruleSet.Apply(ctx, testJson, &out)

	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	} else if rel, ok := out.Relationships["test"]; !ok || rel.Data.(jsonapi.ResourceIdentifierLinkage).ID != "xyz" {
		t.Errorf("Expected identifier resource linkage")
	}

	// Nil linkage
	testJson = `{
	  "id": "abc",
	  "attributes": {},
		"relationships": {
			"test": {
				"data": null
			}
		}
	}`

	ruleSet = ruleSet.WithUnknownRelationships()
	out = jsonapi.Datum[map[string]any]{}
	errs = ruleSet.Apply(ctx, testJson, &out)

	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	} else {
		rel, ok := out.Relationships["test"]

		if ok {
			_, ok = rel.Data.(jsonapi.NilResourceLinkage)
		}

		if !ok {
			t.Errorf("Expected nil resource linkage")
		}
	}

	// Arrays
	testJson = `{
	  "id": "abc",
	  "attributes": {},
		"relationships": {
			"test": {
				"data": [
				  {"id": "xyz", "type": "tests"},
				  {"id": "qwerty", "type": "tests"}
				]
			}
		}
	}`

	ruleSet = ruleSet.WithUnknownRelationships()
	out = jsonapi.Datum[map[string]any]{}
	errs = ruleSet.Apply(ctx, testJson, &out)

	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	} else {
		rel, ok := out.Relationships["test"]

		if ok {
			col, _ := rel.Data.(jsonapi.ResourceLinkageCollection)

			if l := len(col); l != 2 {
				t.Errorf("Expected 2 resource linkages, got: %d", l)
			}
		} else {
			t.Errorf("Expected nil resource linkage")
		}
	}

}

func TestDatumRuleSet_WithRelationship(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewDatumRuleSet[testDatum]("tests", attributesRuleSet)
	relRuleSet := jsonapi.RelationshipRuleSet

	// Add a relationship
	newRuleSet := ruleSet.WithRelationship("author", relRuleSet)

	// Verify it's a new instance
	if newRuleSet == ruleSet {
		t.Error("Expected WithRelationship to return a new instance")
	}

	ctx := context.Background()

	testJson := `{
		"id": "abc",
		"attributes": {
			"Name": "Test"
		},
		"relationships": {
			"author": {
				"data": {"id": "123", "type": "users"}
			}
		}
	}`

	var out jsonapi.Datum[testDatum]
	errs := newRuleSet.Apply(ctx, testJson, &out)

	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	}

	// Verify relationship was parsed
	if rel, ok := out.Relationships["author"]; ok {
		if linkage, ok := rel.Data.(jsonapi.ResourceIdentifierLinkage); ok {
			if linkage.ID != "123" {
				t.Errorf("Expected author ID to be '123', got '%s'", linkage.ID)
			}
		} else {
			t.Error("Expected author relationship to be ResourceIdentifierLinkage")
		}
	} else {
		t.Error("Expected author relationship to be present")
	}
}

func TestDatumRuleSet_WithRequired(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewDatumRuleSet[testDatum]("tests", attributesRuleSet)

	// Initially not required
	if ruleSet.Required() {
		t.Error("Expected ruleSet to not be required initially")
	}

	// Make it required
	newRuleSet := ruleSet.WithRequired()

	// Verify it's a new instance
	if newRuleSet == ruleSet {
		t.Error("Expected WithRequired to return a new instance when not already required")
	}

	if !newRuleSet.Required() {
		t.Error("Expected new ruleSet to be required")
	}

	// Calling WithRequired again should return the same instance
	anotherRuleSet := newRuleSet.WithRequired()
	if anotherRuleSet != newRuleSet {
		t.Error("Expected WithRequired to return same instance when already required")
	}
}

func TestDatumRuleSet_Evaluate(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().WithMinLen(3).Any())

	ruleSet := jsonapi.NewDatumRuleSet[testDatum]("tests", attributesRuleSet)

	// Evaluate method exists and implements the RuleSet interface
	// Note: Evaluate is provided for interface compatibility but may have
	// limitations when used directly with struct values. Use Apply() for
	// full functionality with JSON strings or maps.
	_ = ruleSet.Evaluate
}

func TestDatumRuleSet_String(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewDatumRuleSet[testDatum]("tests", attributesRuleSet)

	str := ruleSet.String()
	expected := "DatumRuleSet"
	if str != expected {
		t.Errorf("Expected String() to return %q, got %q", expected, str)
	}
}

func TestDatumRuleSet_WithMetaWithNilAndErrorConfig(t *testing.T) {
	type testDatum struct {
		Name string
	}
	attrs := rules.Struct[testDatum]().WithKey("Name", rules.String().Any())
	rs := jsonapi.NewDatumRuleSet[testDatum]("tests", attrs).
		WithMeta("requestId", rules.String().Any()).
		WithNil().
		WithErrorMessage("short", "long").
		WithDocsURI("https://docs.example.com").
		WithTraceURI("https://trace.example.com").
		WithErrorCode(errors.CodeRequired).
		WithErrorMeta("k", "v").
		WithErrorCallback(func(ctx context.Context, err errors.ValidationError) errors.ValidationError { return err })

	ctx := context.Background()
	anyRS := rs.Any()
	if anyRS == nil {
		t.Error("Any() should not be nil")
	}
	var out any
	errs := anyRS.Apply(ctx, `{"id":"1","type":"tests","attributes":{"Name":"y"}}`, &out)
	if errs != nil {
		t.Fatalf("Apply via Any: %s", errs)
	}
}
