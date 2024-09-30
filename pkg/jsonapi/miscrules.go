package jsonapi

import (
	"context"
	"reflect"

	"proto.zip/studio/validate"
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
		errs := validate.Array[ResourceIdentifierLinkage]().WithItemRuleSet(ResourceIdentifierLinkageRuleSet).Apply(ctx, value, &out)
		return ResourceLinkageCollection(out), errs
	}

	var out ResourceIdentifierLinkage
	errs := ResourceIdentifierLinkageRuleSet.Apply(ctx, value, &out)
	return out, errs
}

var ResourceLinkageRuleSet rules.RuleSet[ResourceLinkage] = rules.Interface[ResourceLinkage]().WithCast(resourceLinkageCast)

var ResourceIdentifierLinkageRuleSet rules.RuleSet[ResourceIdentifierLinkage] = validate.Object[ResourceIdentifierLinkage]().
	WithKey("type", validate.String().Any()).
	WithKey("id", validate.String().Any()).
	WithKey("lid", validate.String().Any())

var RelationshipRuleSet rules.RuleSet[Relationship] = validate.Object[Relationship]().
	WithKey("data", ResourceLinkageRuleSet.Any()).
	WithKey("meta", validate.Map[any]().WithUnknown().Any())

var RelationshipsRuleSet *rules.ObjectRuleSet[map[string]Relationship, string, Relationship] = validate.Map[Relationship]()

var IDRuleSet rules.RuleSet[string] = validate.String().WithStrict()

var MetaRuleSet rules.RuleSet[map[string]any] = validate.MapAny()
