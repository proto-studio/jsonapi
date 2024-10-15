package jsonapi_test

import (
	"context"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/rules"
)

// Requirements:
//   - Runs inner rule sets.
//   - Returns the evaluated data.
func TestSingleDatum(t *testing.T) {

	type testDatum struct {
		Name string
	}

	datumRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().WithMinLen(6).Any())

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

	attributesRuleSet := rules.StringMap[any]().WithUnknown()
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

// TestSingleDatum_AttributeFields ensures that the Fields map contains only the field names
// that were present in the attributes object, excluding fields not in the JSON.
func TestSingleDatum_AttributeFields(t *testing.T) {
	type testDatum struct {
		Name  string
		Age   int
		Email string // This field won't be in the JSON
	}

	datumRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any()).
		WithKey("Age", rules.Int().Any()).
		WithKey("Email", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", datumRuleSet)

	ctx := context.Background()

	testJson := `{
		"data": {
			"id": "abc",
			"type": "tests",
			"attributes": {
				"Name": "John Doe",
				"Age": 30
			}
		}
	}`

	var testFields jsonapi.SingleDatumEnvelope[testDatum]
	errs := ruleSet.Apply(ctx, testJson, &testFields)

	if errs != nil {
		t.Fatalf("Expected errors to be nil, got: %s", errs.Error())
	}

	expectedFields := []string{"Name", "Age"}

	if testFields.Data.Fields == nil {
		t.Fatalf("Expected Fields to be initialized, but it was nil")
	}
	if len(testFields.Data.Fields.Values()) != len(expectedFields) {
		t.Errorf("Expected Fields to have %d entries, got %d", len(expectedFields), len(testFields.Data.Fields.Values()))
	}

	for _, field := range expectedFields {
		if !testFields.Data.Fields.Contains(field) {
			t.Errorf("Expected Fields to contain '%s', but it was missing", field)
		}
	}

	// Check that the Email field is not in the Fields map
	if testFields.Data.Fields.Contains("Email") {
		t.Errorf("Fields map should not contain 'Email', but it was present")
	}
}
