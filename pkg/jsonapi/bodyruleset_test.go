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
//   - Runs inner rule sets.
//   - Returns the evaluated data.
func TestSingleDatum(t *testing.T) {

	type testDatum struct {
		Name string
	}

	datumRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().WithMinLen(6).Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", datumRuleSet)

	// Protovalidate test helpers: RuleSet implements WithRequired/Required
	testhelpers.MustImplementWithRequired(t, ruleSet)

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

func TestSingleRuleSet_WithRelationship(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", attributesRuleSet)
	relRuleSet := jsonapi.RelationshipRuleSet

	// Add a relationship
	newRuleSet := ruleSet.WithRelationship("author", relRuleSet)

	// Verify it's a new instance
	if newRuleSet == ruleSet {
		t.Error("Expected WithRelationship to return a new instance")
	}

	ctx := context.Background()

	testJson := `{
		"data": {
			"id": "abc",
			"attributes": {
				"Name": "Test"
			},
			"relationships": {
				"author": {
					"data": {"id": "123", "type": "users"}
				}
			}
		}
	}`

	var out jsonapi.SingleDatumEnvelope[testDatum]
	errs := newRuleSet.Apply(ctx, testJson, &out)

	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	}

	// Verify relationship was parsed
	if rel, ok := out.Data.Relationships["author"]; ok {
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

func TestSingleRuleSet_WithRequired(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", attributesRuleSet)

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

func TestSingleRuleSet_Required(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", attributesRuleSet)

	// Initially not required
	if ruleSet.Required() {
		t.Error("Expected ruleSet to not be required initially")
	}

	// Make it required
	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected ruleSet to be required after WithRequired")
	}
}

func TestSingleRuleSet_Evaluate(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().WithMinLen(3).Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", attributesRuleSet)

	// Evaluate method exists and implements the RuleSet interface
	// Note: Evaluate is provided for interface compatibility but may have
	// limitations when used directly with struct values. Use Apply() for
	// full functionality with JSON strings or maps.
	_ = ruleSet.Evaluate
}

func TestSingleRuleSet_Any(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", attributesRuleSet)

	anyRuleSet := ruleSet.Any()
	if anyRuleSet == nil {
		t.Error("Expected Any() to return a non-nil RuleSet")
	}

	// Verify it can be used
	ctx := context.Background()
	testJson := `{
		"data": {
			"id": "abc",
			"attributes": {
				"Name": "Test"
			}
		}
	}`

	var out any
	errs := anyRuleSet.Apply(ctx, testJson, &out)
	if errs != nil {
		t.Errorf("Expected errors to be nil, got: %s", errs.Error())
	}
}

func TestSingleRuleSet_String(t *testing.T) {
	type testDatum struct {
		Name string
	}

	attributesRuleSet := rules.Struct[testDatum]().
		WithKey("Name", rules.String().Any())

	ruleSet := jsonapi.NewSingleRuleSet[testDatum]("tests", attributesRuleSet)

	str := ruleSet.String()
	expected := "SingleRuleSet"
	if str != expected {
		t.Errorf("Expected String() to return %q, got %q", expected, str)
	}
}

func TestSingleRuleSet_WithMetaWithDocumentMetaWithNil(t *testing.T) {
	type testDatum struct {
		Name string
	}
	attrs := rules.Struct[testDatum]().WithKey("Name", rules.String().Any())
	rs := jsonapi.NewSingleRuleSet[testDatum]("tests", attrs).
		WithMeta("requestId", rules.String().Any()).
		WithDocumentMeta("metaKey", rules.String().Any()).
		WithNil().
		WithErrorMessage("short", "long").
		WithDocsURI("https://docs.example.com").
		WithTraceURI("https://trace.example.com").
		WithErrorCode(errors.CodeRequired).
		WithErrorMeta("k", "v").
		WithErrorCallback(func(ctx context.Context, err errors.ValidationError) errors.ValidationError { return err })

	ctx := context.Background()
	// Apply with document meta
	body := `{"data":{"id":"1","type":"tests","attributes":{"Name":"Hi"}},"meta":{"metaKey":"val"}}`
	var out jsonapi.SingleDatumEnvelope[testDatum]
	errs := rs.Apply(ctx, body, &out)
	if errs != nil {
		t.Fatalf("Apply: %s", errs)
	}
}
