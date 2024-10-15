package jsonapi

import (
	"context"
	"encoding/json"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

type SingleRuleSet[T any] struct {
	datumRuleSet *DatumRuleSet[T]
	metaRuleSet  rules.RuleSet[map[string]any]
	required     bool
	rules.NoConflict[SingleDatumEnvelope[T]]
}

func NewSingleRuleSet[T any](typeName string, attributesRuleSet rules.RuleSet[T]) *SingleRuleSet[T] {
	return &SingleRuleSet[T]{
		datumRuleSet: NewDatumRuleSet(typeName, attributesRuleSet),
		metaRuleSet:  MetaRuleSet,
	}
}

func (ruleSet *SingleRuleSet[T]) clone() *SingleRuleSet[T] {
	return &SingleRuleSet[T]{
		datumRuleSet: ruleSet.datumRuleSet,
		metaRuleSet:  ruleSet.metaRuleSet,
		required:     ruleSet.required,
	}
}

func (ruleSet *SingleRuleSet[T]) WithRelationship(relName string, relRuleSet rules.RuleSet[Relationship]) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.datumRuleSet = newRuleSet.datumRuleSet.WithRelationship(relName, relRuleSet)
	return newRuleSet
}

func (ruleSet *SingleRuleSet[T]) WithUnknownRelationships() *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.datumRuleSet = newRuleSet.datumRuleSet.WithUnknownRelationships()
	return newRuleSet
}

func (ruleSet *SingleRuleSet[T]) WithRequired() *SingleRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	newRuleSet := ruleSet.clone()
	newRuleSet.required = true
	return newRuleSet
}

func (ruleSet *SingleRuleSet[T]) Required() bool {
	return ruleSet.required
}

func (ruleSet *SingleRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	// ObjectRuleSet is capable of decoding raw JSON but in this case we want to decode the JSON
	// ahead of time into a map so we can assign fields.
	// In the future if support is added upstream we can switch to using that.
	var decodedInput any
	if inputStr, ok := input.(string); ok {
		if err := json.Unmarshal([]byte(inputStr), &decodedInput); err != nil {
			return errors.Collection(errors.Errorf(errors.CodeEncoding, ctx, "Body must be Json encoded"))
		}
		input = decodedInput
	} else if inputMap, ok := input.(map[string]any); ok {
		decodedInput = inputMap
	}

	bodyValidator := rules.Struct[SingleDatumEnvelope[T]]()
	bodyValidator = bodyValidator.WithKey("data", ruleSet.datumRuleSet.Any())
	bodyValidator = bodyValidator.WithKey("meta", ruleSet.metaRuleSet.Any())

	bodyValidator = bodyValidator.WithDynamicBucket(atMembersKeyRule, "AtMembers")
	bodyValidator = bodyValidator.WithDynamicBucket(extKeyRule, "ExtensionMembers")

	err := bodyValidator.Apply(ctx, input, output)

	if err != nil {
		return err
	}

	if decodedInput != nil {
		inputMap := decodedInput.(map[string]any)
		data := inputMap["data"].(map[string]any)
		attributes, ok := data["attributes"].(map[string]any)
		if ok {
			fields := make(fieldListMap)
			for key := range attributes {
				fields[key] = true
			}

			if outputEnvelope, ok := output.(*SingleDatumEnvelope[T]); ok {
				outputEnvelope.Data.Fields = fields
			}
		}
	}

	return nil
}

func (ruleSet *SingleRuleSet[T]) Evaluate(ctx context.Context, value SingleDatumEnvelope[T]) errors.ValidationErrorCollection {
	var out SingleDatumEnvelope[T]
	return ruleSet.Apply(ctx, value, &out)
}

func (ruleSet *SingleRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[SingleDatumEnvelope[T]](ruleSet)
}

func (ruleSet *SingleRuleSet[T]) String() string {
	return "SingleRuleSet"
}
