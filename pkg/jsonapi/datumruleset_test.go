package jsonapi_test

import (
	"context"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/rules"
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
		if len(errs) != 1 {
			t.Errorf(`Expected 1 error, got: %d`, len(errs))
		}
		expected := "/attributes/Name"
		if p := errs.First().Path(); p != expected {
			t.Errorf(`Expected path to be "%s", got: "%s"`, expected, p)
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
		if len(errs) != 1 {
			t.Errorf(`Expected 1 error, got: %d`, len(errs))
		}
		expected := "/type"
		if p := errs.First().Path(); p != expected {
			t.Errorf(`Expected path to be "%s", got: "%s"`, expected, p)
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
