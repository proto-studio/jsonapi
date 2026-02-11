package jsonapi_test

import (
	"encoding/json"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
)

func TestFullLink_Href(t *testing.T) {
	link := &jsonapi.FullLink{
		HrefValue: "https://example.com/resource",
	}

	href := link.Href()
	expected := "https://example.com/resource"

	if href != expected {
		t.Errorf("Expected Href() to return %q, got %q", expected, href)
	}
}

func TestStringLink_Href(t *testing.T) {
	link := jsonapi.StringLink("https://example.com/resource")

	href := link.Href()
	expected := "https://example.com/resource"

	if href != expected {
		t.Errorf("Expected Href() to return %q, got %q", expected, href)
	}
}

func TestNilLink_Href(t *testing.T) {
	link := jsonapi.NilLink{}

	href := link.Href()
	expected := ""

	if href != expected {
		t.Errorf("Expected Href() to return %q, got %q", expected, href)
	}
}

func TestNilLink_MarshalJSON(t *testing.T) {
	link := jsonapi.NilLink{}

	data, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("Unexpected error marshaling NilLink: %v", err)
	}

	expected := "null"
	if string(data) != expected {
		t.Errorf("Expected MarshalJSON to return %q, got %q", expected, string(data))
	}
}

func TestNilLink_UnmarshalJSON(t *testing.T) {
	link := jsonapi.NilLink{}

	err := link.UnmarshalJSON([]byte("null"))
	if err != nil {
		t.Errorf("Unexpected error unmarshaling NilLink: %v", err)
	}

	// Should also handle empty JSON
	err = link.UnmarshalJSON([]byte(""))
	if err != nil {
		t.Errorf("Unexpected error unmarshaling empty JSON: %v", err)
	}
}

func TestLinks_UnmarshalJSON_StringLink(t *testing.T) {
	jsonData := `{"self": "https://example.com/resource"}`

	var links jsonapi.Links
	err := json.Unmarshal([]byte(jsonData), &links)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling Links: %v", err)
	}

	selfLink, ok := links["self"]
	if !ok {
		t.Fatalf("Expected 'self' link to be present")
	}

	if _, ok := selfLink.(jsonapi.StringLink); !ok {
		t.Errorf("Expected 'self' link to be StringLink")
	}

	if selfLink.Href() != "https://example.com/resource" {
		t.Errorf("Expected Href() to return 'https://example.com/resource', got %q", selfLink.Href())
	}
}

func TestLinks_UnmarshalJSON_FullLink(t *testing.T) {
	jsonData := `{
		"self": {
			"href": "https://example.com/resource",
			"meta": {"version": "1.0"}
		}
	}`

	var links jsonapi.Links
	err := json.Unmarshal([]byte(jsonData), &links)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling Links: %v", err)
	}

	selfLink, ok := links["self"]
	if !ok {
		t.Fatalf("Expected 'self' link to be present")
	}

	fullLink, ok := selfLink.(*jsonapi.FullLink)
	if !ok {
		t.Errorf("Expected 'self' link to be *FullLink")
	}

	if fullLink.Href() != "https://example.com/resource" {
		t.Errorf("Expected Href() to return 'https://example.com/resource', got %q", fullLink.Href())
	}

	if fullLink.Meta["version"] != "1.0" {
		t.Errorf("Expected Meta version to be '1.0', got %v", fullLink.Meta["version"])
	}
}

func TestLinks_UnmarshalJSON_MixedLinks(t *testing.T) {
	jsonData := `{
		"self": "https://example.com/resource",
		"related": {
			"href": "https://example.com/related",
			"meta": {"count": 10}
		}
	}`

	var links jsonapi.Links
	err := json.Unmarshal([]byte(jsonData), &links)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling Links: %v", err)
	}

	if len(links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(links))
	}

	// Check string link
	selfLink, ok := links["self"]
	if !ok {
		t.Fatalf("Expected 'self' link to be present")
	}
	if _, ok := selfLink.(jsonapi.StringLink); !ok {
		t.Errorf("Expected 'self' link to be StringLink")
	}

	// Check full link
	relatedLink, ok := links["related"]
	if !ok {
		t.Fatalf("Expected 'related' link to be present")
	}
	if _, ok := relatedLink.(*jsonapi.FullLink); !ok {
		t.Errorf("Expected 'related' link to be *FullLink")
	}
}

func TestLinks_UnmarshalJSON_InvalidLink(t *testing.T) {
	// Test with data that can't be unmarshaled as either StringLink or FullLink
	// The current implementation will try StringLink first, then FullLink
	// If both fail, it should return an error
	jsonData := `{
		"self": 12345
	}`

	var links jsonapi.Links
	err := json.Unmarshal([]byte(jsonData), &links)
	if err == nil {
		t.Errorf("Expected error when unmarshaling invalid link (number), got nil")
	}
}

func TestLinks_UnmarshalJSON_EmptyMap(t *testing.T) {
	jsonData := `{}`

	var links jsonapi.Links
	err := json.Unmarshal([]byte(jsonData), &links)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling empty Links: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("Expected empty links map, got %d links", len(links))
	}
}
