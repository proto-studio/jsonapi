package jsonapi_test

import (
	"context"
	"encoding/json"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
)

// Requirements:
// - NilResourceLinkage should serialize to "null"
func TestNilJson(t *testing.T) {
	n := jsonapi.NilResourceLinkage{}

	// This should JSON serialize to the word "null"
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Failed to serialize NilResourceLinkage: %v", err)
	}

	expected := "null"
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}
func TestIdentifierLinkageCast(t *testing.T) {
	linkage := map[string]string{
		"type": "tests",
		"id":   "123",
	}

	val, errs := jsonapi.ResourceLinkageRuleSet.Run(context.Background(), linkage)

	if errs != nil {
		t.Errorf("Unexpected error running rule set: %s", errs.Error())
	} else if val == nil {
		t.Errorf("Expected value to not be nil")
	} else if l, ok := val.(jsonapi.ResourceIdentifierLinkage); ok {
		if l.ID != "123" {
			t.Errorf(`Expected ID to be "%s", got: "%s"`, "123", l.ID)
		}
	} else {
		t.Errorf("Expected value to be ResourceIdentifierLinkage")
	}
}

// Requirements:
// - Linkage validator should return NilResourceLinkage not nil
func TestNilLinkage(t *testing.T) {
	val, errs := jsonapi.ResourceLinkageRuleSet.Run(context.Background(), nil)

	if errs != nil {
		t.Errorf("Unexpected error running rule set: %s", errs.Error())
	} else if val == nil {
		t.Errorf("Expected value to not be nil")
	} else if _, ok := val.(jsonapi.NilResourceLinkage); !ok {
		t.Errorf("Expected value to be NilResourceLinkage")
	}
}

func TestLinkageCollection(t *testing.T) {
	linkage := map[string]string{
		"type": "tests",
		"id":   "123",
	}

	linkages := []map[string]string{
		linkage,
		linkage,
	}

	val, errs := jsonapi.ResourceLinkageRuleSet.Run(context.Background(), linkages)

	if errs != nil {
		t.Errorf("Unexpected error running rule set: %s", errs.Error())
	} else if val == nil {
		t.Errorf("Expected value to not be nil")
	} else if c, ok := val.(jsonapi.ResourceLinkageCollection); ok {

		if len(c) != 2 {
			t.Errorf("Expected %d linkages, got %d", 2, len(c))
		}

	} else {
		t.Errorf("Expected value to be ResourceLinkageCollection")
	}
}
