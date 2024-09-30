package jsonapi

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

func resourceLinkageCast(ctx context.Context, value any) (ResourceLinkage, errors.ValidationErrorCollection) {
	if value == nil {
		// nil and nil linkage are not the same so we need a way to differentiate
		return NilResourceLinkage{}, nil
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

var ResourceLinkageRuleSet rules.RuleSet[ResourceLinkage] = rules.Interface[ResourceLinkage]().WithCast(resourceLinkageCast)

var ResourceIdentifierLinkageRuleSet rules.RuleSet[ResourceIdentifierLinkage] = rules.Struct[ResourceIdentifierLinkage]().
	WithKey("type", rules.String().Any()).
	WithKey("id", rules.String().Any()).
	WithKey("lid", rules.String().Any())

var RelationshipRuleSet rules.RuleSet[Relationship] = rules.Struct[Relationship]().
	WithKey("data", ResourceLinkageRuleSet.Any()).
	WithKey("meta", rules.StringMap[any]().WithUnknown().Any())

var RelationshipsRuleSet *rules.ObjectRuleSet[map[string]Relationship, string, Relationship] = rules.StringMap[Relationship]()

var IDRuleSet rules.RuleSet[string] = rules.String().WithStrict()

var MetaRuleSet rules.RuleSet[map[string]any] = rules.StringMap[any]()
