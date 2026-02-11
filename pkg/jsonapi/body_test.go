package jsonapi_test

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
)

// Requirements:
// - Marshals extensions.
// - Respects field filters.
// - Returns all fields when no filter is present.
func TestMarshalJSON(t *testing.T) {
	type ExampleAttributes struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	exampleAttr := ExampleAttributes{Name: "John Doe", Email: "john.doe@example.com", Age: 30}

	tests := []struct {
		name     string
		datum    any
		expected string
	}{
		{
			name: "Datum[ExampleAttributes] with FieldList and ExtensionMembers",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "123",
				Type:       "example",
				Attributes: exampleAttr,
				Fields:     jsonapi.NewFieldList("name", "email"),
				ExtensionMembers: map[string]any{
					"test:customField1": "customValue1",
					"test:customField2": 42,
				},
			},
			expected: `{
				"id":"123",
				"type":"example",
				"attributes":{"name":"John Doe","email":"john.doe@example.com"},
				"test:customField1":"customValue1",
				"test:customField2":42
			}`,
		},
		{
			name: "Datum[ExampleAttributes] without FieldList and with ExtensionMembers",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "123",
				Type:       "example",
				Attributes: exampleAttr,
				Fields:     nil,
				ExtensionMembers: map[string]any{
					"test:customField1": "customValue1",
					"test:customField2": 42,
				},
			},
			expected: `{
				"id":"123",
				"type":"example",
				"attributes":{"name":"John Doe","email":"john.doe@example.com","age":30},
				"test:customField1":"customValue1",
				"test:customField2":42
			}`,
		},
		{
			name: "Datum[*ExampleAttributes] with FieldList",
			datum: jsonapi.Datum[*ExampleAttributes]{
				ID:         "124",
				Type:       "example",
				Attributes: &exampleAttr,
				Fields:     jsonapi.NewFieldList("name", "age"),
			},
			expected: `{
				"id":"124",
				"type":"example",
				"attributes":{"name":"John Doe","age":30}
			}`,
		},
		{
			name: "Datum[*ExampleAttributes] without FieldList and with ExtensionMembers",
			datum: jsonapi.Datum[*ExampleAttributes]{
				ID:         "124",
				Type:       "example",
				Attributes: &exampleAttr,
				Fields:     nil,
				ExtensionMembers: map[string]any{
					"test:customField1": "customValue1",
					"test:customField2": 42,
				},
			},
			expected: `{
				"id":"124",
				"type":"example",
				"attributes":{"name":"John Doe","email":"john.doe@example.com","age":30},
				"test:customField1":"customValue1",
				"test:customField2":42
			}`,
		},
		{
			name: "Datum[map[string]any] without FieldList",
			datum: jsonapi.Datum[map[string]any]{
				ID:   "125",
				Type: "example",
				Attributes: map[string]any{
					"name":  "Bob Smith",
					"email": "bob.smith@example.com",
					"age":   35,
				},
			},
			expected: `{
				"id":"125",
				"type":"example",
				"attributes":{"name":"Bob Smith","email":"bob.smith@example.com","age":35}
			}`,
		},
		{
			name: "Datum[ExampleAttributes] with empty FieldList",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "126",
				Type:       "example",
				Attributes: exampleAttr,
				Fields:     jsonapi.NewFieldList(),
			},
			expected: `{
				"id":"126",
				"type":"example"
			}`,
		},
		{
			name: "Datum with Relationships and Fields filtering",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "127",
				Type:       "example",
				Attributes: exampleAttr,
				Relationships: map[string]jsonapi.Relationship{
					"author": {
						Data: jsonapi.ResourceIdentifierLinkage{
							ID:   "456",
							Type: "users",
						},
					},
					"comments": {
						Data: jsonapi.ResourceLinkageCollection{
							{ID: "789", Type: "comments"},
						},
					},
				},
				Fields: jsonapi.NewFieldList("name", "author"),
			},
			expected: `{
				"id":"127",
				"type":"example",
				"attributes":{"name":"John Doe"},
				"relationships":{"author":{"data":{"id":"456","type":"users"}}}
			}`,
		},
		{
			name: "Datum with Links and Meta",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "128",
				Type:       "example",
				Attributes: exampleAttr,
				Links: jsonapi.Links{
					"self": jsonapi.StringLink("http://example.com/128"),
				},
				Meta: map[string]any{
					"version": "1.0",
				},
			},
			expected: `{
				"id":"128",
				"type":"example",
				"attributes":{"name":"John Doe","email":"john.doe@example.com","age":30},
				"links":{"self":"http://example.com/128"},
				"meta":{"version":"1.0"}
			}`,
		},
		{
			name: "Datum[map[string]any] with Fields filtering",
			datum: jsonapi.Datum[map[string]any]{
				ID:   "129",
				Type: "example",
				Attributes: map[string]any{
					"name":  "Alice",
					"email": "alice@example.com",
					"age":   25,
				},
				Fields: jsonapi.NewFieldList("name", "age"),
			},
			expected: `{
				"id":"129",
				"type":"example",
				"attributes":{"name":"Alice","age":25}
			}`,
		},
		{
			name: "Datum with Fields filtering that excludes all attributes",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "130",
				Type:       "example",
				Attributes: exampleAttr,
				Fields:     jsonapi.NewFieldList("nonexistent"),
			},
			expected: `{
				"id":"130",
				"type":"example"
			}`,
		},
		{
			name: "Datum with Relationships but Fields filters them out",
			datum: jsonapi.Datum[ExampleAttributes]{
				ID:         "131",
				Type:       "example",
				Attributes: exampleAttr,
				Relationships: map[string]jsonapi.Relationship{
					"author": {
						Data: jsonapi.ResourceIdentifierLinkage{
							ID:   "456",
							Type: "users",
						},
					},
				},
				Fields: jsonapi.NewFieldList("name"), // Doesn't include "author"
			},
			expected: `{
				"id":"131",
				"type":"example",
				"attributes":{"name":"John Doe"}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := json.Marshal(tt.datum)
			if err != nil {
				t.Fatalf("Unexpected error during marshalling: %v", err)
			}

			if !jsonEqual(tt.expected, string(actual)) {
				t.Errorf("Test '%s' failed:\nExpected JSON: %s\nGot JSON: %s", tt.name, tt.expected, string(actual))
			}
		})
	}
}

// jsonEqual compares two JSON strings for equality regardless of formatting
func jsonEqual(a, b string) bool {
	var o1, o2 any

	if err := json.Unmarshal([]byte(a), &o1); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

// TestDatumUnmarshalJSON tests the UnmarshalJSON method of the Datum struct
func TestDatumUnmarshalJSON(t *testing.T) {
	type ExampleAttributes struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age,omitempty"`
	}

	// Example JSON data
	jsonData := `{
		"id": "123",
		"type": "example",
		"attributes": {"name": "John Doe", "email": "john.doe@example.com"},
		"links": {"self": "http://example.com/self"},
		"meta": {"version": "1.0"},
		"test:customField1": "customValue1",
		"test:customField2": 42
	}`

	var datum jsonapi.Datum[ExampleAttributes]
	err := json.Unmarshal([]byte(jsonData), &datum)
	if err != nil {
		t.Fatalf("Unexpected error during unmarshalling: %v", err)
	}

	// Verify ID and Type
	if datum.ID != "123" {
		t.Errorf("Expected ID to be '123', got '%s'", datum.ID)
	}
	if datum.Type != "example" {
		t.Errorf("Expected Type to be 'example', got '%s'", datum.Type)
	}

	// Verify Attributes
	expectedAttributes := ExampleAttributes{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
	if !reflect.DeepEqual(datum.Attributes, expectedAttributes) {
		t.Errorf("Expected Attributes to be %+v, got %+v", expectedAttributes, datum.Attributes)
	}

	// Verify Links
	expectedLinks := jsonapi.Links{"self": jsonapi.StringLink("http://example.com/self")}
	if !reflect.DeepEqual(datum.Links, expectedLinks) {
		t.Errorf("Expected Links to be %+v, got %+v", expectedLinks, datum.Links)
	}

	// Verify Meta
	expectedMeta := map[string]any{"version": "1.0"}
	if !reflect.DeepEqual(datum.Meta, expectedMeta) {
		t.Errorf("Expected Meta to be %+v, got %+v", expectedMeta, datum.Meta)
	}

	// Verify Fields
	expectedFields := []string{"email", "name"}

	actualFields := datum.Fields.Values()
	sort.Slice(actualFields, func(i, j int) bool {
		return actualFields[i] < actualFields[j]
	})

	if !reflect.DeepEqual(actualFields, expectedFields) {
		t.Errorf("Fields do not match. Got %v, want %v", actualFields, expectedFields)
	}

	// Verify ExtensionMembers
	expectedExtensionMembers := map[string]any{
		"test:customField1": "customValue1",
		"test:customField2": float64(42),
	}
	if !reflect.DeepEqual(datum.ExtensionMembers, expectedExtensionMembers) {
		t.Errorf("Expected ExtensionMembers to be %+v, got %+v", expectedExtensionMembers, datum.ExtensionMembers)
	}
}
