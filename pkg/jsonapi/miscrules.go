package jsonapi

import (
	"context"
	"encoding/json"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// resourceLinkageCast converts a raw value (nil, map, slice, or JSON string) into a ResourceLinkage for validation.
func resourceLinkageCast(ctx context.Context, value any) (ResourceLinkage, errors.ValidationError) {
	if value == nil {
		// nil and nil linkage are not the same so we need a way to differentiate
		return NilResourceLinkage{}, nil
	}

	// Handle raw JSON string input
	if strValue, ok := value.(string); ok {
		var parsed any
		if err := json.Unmarshal([]byte(strValue), &parsed); err != nil {
			return nil, errors.Errorf(errors.CodeEncoding, ctx, "Invalid JSON", "Invalid JSON: %v", err)
		}
		// Recurse with parsed value
		return resourceLinkageCast(ctx, parsed)
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		var out []ResourceIdentifierLinkage
		errs := rules.Slice[ResourceIdentifierLinkage]().WithItemRuleSet(ResourceIdentifierLinkageRuleSet).Apply(ctx, value, &out)
		return ResourceLinkageCollection(out), errs
	}

	var out ResourceIdentifierLinkage
	errs := ResourceIdentifierLinkageRuleSet.Apply(ctx, value, &out)
	return out, errs
}

// resourceLinkageRuleSetImpl is a custom rule set that handles nil properly
type resourceLinkageRuleSetImpl struct{}

// Apply validates input into a ResourceLinkage and writes the result to output; handles nil as NilResourceLinkage.
func (r *resourceLinkageRuleSetImpl) Apply(ctx context.Context, input, output any) errors.ValidationError {
	var result ResourceLinkage

	// Handle nil specially - convert to NilResourceLinkage{}
	if input == nil {
		result = NilResourceLinkage{}
	} else {
		// For non-nil values, use the cast function
		var errs errors.ValidationError
		result, errs = resourceLinkageCast(ctx, input)
		if errs != nil {
			return errs
		}
	}

	// Set the result using reflection to handle different output types
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr {
		return errors.Error(errors.CodeType, ctx, "*ResourceLinkage", reflect.TypeOf(output).String())
	}

	elem := outputVal.Elem()
	if !elem.CanSet() {
		return errors.Error(errors.CodeType, ctx, "ResourceLinkage", reflect.TypeOf(output).String())
	}

	// Set the result
	elem.Set(reflect.ValueOf(result))
	return nil
}

// Evaluate is not used for interface types and returns nil.
func (r *resourceLinkageRuleSetImpl) Evaluate(ctx context.Context, value ResourceLinkage) errors.ValidationError {
	return nil
}

// Any returns the rule set as rules.RuleSet[any].
func (r *resourceLinkageRuleSetImpl) Any() rules.RuleSet[any] {
	return rules.WrapAny[ResourceLinkage](r)
}

// String returns a stable name for the rule set.
func (r *resourceLinkageRuleSetImpl) String() string {
	return "ResourceLinkageRuleSet"
}

// Replaces reports whether this rule set replaces another; always false.
func (r *resourceLinkageRuleSetImpl) Replaces(x rules.Rule[ResourceLinkage]) bool {
	return false
}

// Required reports whether the value is required; Relationship data may be null so returns false.
func (r *resourceLinkageRuleSetImpl) Required() bool {
	return false
}

var ResourceLinkageRuleSet rules.RuleSet[ResourceLinkage] = &resourceLinkageRuleSetImpl{}

var ResourceIdentifierLinkageRuleSet rules.RuleSet[ResourceIdentifierLinkage] = rules.Struct[ResourceIdentifierLinkage]().
	WithKey("type", rules.String().Any()).
	WithKey("id", rules.String().Any()).
	WithKey("lid", rules.String().Any()).
	WithKey("meta", rules.StringMap[any]().WithUnknown().Any())

// relationshipRuleSetImpl is a custom rule set that handles null relationship data properly.
type relationshipRuleSetImpl struct{}

// Apply validates a relationship object and handles null data by temporarily removing it for Struct validation.
func (r *relationshipRuleSetImpl) Apply(ctx context.Context, input, output any) errors.ValidationError {
	// Check if input has null data field
	var hadNullData bool
	if inputMap, ok := input.(map[string]any); ok {
		if dataVal, exists := inputMap["data"]; exists && dataVal == nil {
			hadNullData = true
			// Remove null data field temporarily to avoid Struct rule set rejection
			delete(inputMap, "data")
		}
	}
	
	// Use Struct rule set for validation (without data field if it was null)
	validator := rules.Struct[Relationship]().
		WithKey("data", relationshipDataRuleSet.Any()).
		WithKey("links", LinksRuleSet.Any()).
		WithKey("meta", rules.StringMap[any]().WithUnknown().Any())
	
	errs := validator.Apply(ctx, input, output)
	
	// After validation, if input had null data, set it to NilResourceLinkage{}
	if errs == nil && hadNullData {
		if relPtr, ok := output.(*Relationship); ok {
			relPtr.Data = NilResourceLinkage{}
		}
		// Restore null in input map for potential future use
		if inputMap, ok := input.(map[string]any); ok {
			inputMap["data"] = nil
		}
	}
	
	return errs
}

// Evaluate validates a Relationship value and returns any validation errors.
func (r *relationshipRuleSetImpl) Evaluate(ctx context.Context, value Relationship) errors.ValidationError {
	var out Relationship
	return r.Apply(ctx, value, &out)
}

// Any returns the rule set as rules.RuleSet[any].
func (r *relationshipRuleSetImpl) Any() rules.RuleSet[any] {
	return rules.WrapAny[Relationship](r)
}

// String returns a stable name for the rule set.
func (r *relationshipRuleSetImpl) String() string {
	return "RelationshipRuleSet"
}

// Replaces reports whether this rule set replaces another; always false.
func (r *relationshipRuleSetImpl) Replaces(x rules.Rule[Relationship]) bool {
	return false
}

// Required reports whether the relationship is required; returns false.
func (r *relationshipRuleSetImpl) Required() bool {
	return false
}

// Create a rule set for Relationship data that handles nil properly
var relationshipDataRuleSet = rules.Interface[ResourceLinkage]().WithCast(func(ctx context.Context, value any) (ResourceLinkage, errors.ValidationError) {
	if value == nil {
		return NilResourceLinkage{}, nil
	}
	var out ResourceLinkage
	errs := ResourceLinkageRuleSet.Apply(ctx, value, &out)
	return out, errs
})

var RelationshipRuleSet rules.RuleSet[Relationship] = &relationshipRuleSetImpl{}

var RelationshipsRuleSet *rules.ObjectRuleSet[map[string]Relationship, string, Relationship] = rules.StringMap[Relationship]()

var IDRuleSet rules.RuleSet[string] = rules.String().WithStrict()

var MetaRuleSet rules.RuleSet[map[string]any] = rules.StringMap[any]()

// IncludedResourceRuleSet validates a single included resource object
// Included resources can have any type of attributes, so we validate the basic structure
var IncludedResourceRuleSet rules.RuleSet[map[string]any] = rules.StringMap[any]().
	WithKey("type", rules.String().Any()).
	WithKey("id", rules.String().Any()).
	WithUnknown()

// IncludedRuleSet validates the included array in a compound document
var IncludedRuleSet rules.RuleSet[[]any] = rules.Slice[any]().WithItemRuleSet(IncludedResourceRuleSet.Any())
