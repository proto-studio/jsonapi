package jsonapi_test

import (
	"context"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate"
)

// Requirements:
//   - Runs inner rule sets.
//   - Returns the evaluated data.
func TestSingleDatum(t *testing.T) {

	type testDatum struct {
		Name string
	}

	datumRuleSet := validate.Object[testDatum]().
		WithKey("Name", validate.String().WithMinLen(6).Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", datumRuleSet)

	ctx := context.Background()

	var test jsonapi.SingleDatumEnvelope[testDatum]
	errs := ruleSet.Apply(ctx, `{
	  "data": {
			"id": "abc",
			"attributes": {
				"Name": "My Test"
			}
		}
	}`, &test)

	if errs != nil {
		t.Fatalf("Expected errors to be nil, got: %s", errs.Error())
	}

	if test.Data.Type != "tests" {
		t.Errorf(`Expected ID to be "%s", got: %s`, "tests", test.Data.Type)
	}
	if test.Data.ID != "abc" {
		t.Errorf(`Expected ID to be "%s", got: %s`, "abc", test.Data.ID)
	}
	if test.Data.Attributes.Name != "My Test" {
		t.Errorf(`Expected Name to be "%s", got: "%s"`, "My Test", test.Data.Attributes.Name)
	}

	// Type matches expected
	errs = ruleSet.Apply(ctx, `{
	  "data": {
			"id": "abc",
			"type": "tests",
			"attributes": {
				"Name": "My Test"
			}
		}
	}`, &test)

	if errs != nil {
		t.Errorf("Expected errors on matching type to be nil, got: %s", errs.Error())
	}
}

// Requirements:
// - Unknown relationships error by default.
// - WithUnknownRelationships allows all relationships (nil, slice, and ID linkage) to pass.
func TestWithUnknownRelationshipsBody(t *testing.T) {

	attributesRuleSet := validate.MapAny().WithUnknown()
	ruleSet := jsonapi.NewSingleRuleSet[map[string]any]("tests", attributesRuleSet)

	ctx := context.Background()

	testJson := `{
	  "data": {
			"id": "abc",
			"attributes": {},
			"relationships": {
				"test": {
					"data": {"id": "xyz", "type": "tests"}
				}
			}
    }
	}`

	var out jsonapi.SingleDatumEnvelope[map[string]any]
	errs := ruleSet.Apply(ctx, testJson, &out)

	if errs == nil {
		t.Errorf("Expected errors to not be nil")
	}

	// ID linkage
	ruleSet = ruleSet.WithUnknownRelationships()
	out = jsonapi.SingleDatumEnvelope[map[string]any]{}
	errs = ruleSet.Apply(ctx, testJson, &out)

	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	} else if rel, ok := out.Data.Relationships["test"]; !ok || rel.Data.(jsonapi.ResourceIdentifierLinkage).ID != "xyz" {
		t.Errorf("Expected identifier resource linkage")
	}
}
