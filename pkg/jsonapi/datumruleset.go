package jsonapi

import (
	"context"
	"reflect"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

type DatumRuleSet[T any] struct {
	idRuleSet            rules.RuleSet[string]
	typeRuleSet          *rules.ConstantRuleSet[string]
	relationshipsRuleSet *rules.ObjectRuleSet[map[string]Relationship, string, Relationship]
	attributesRuleSet    rules.RuleSet[T]
	metaRuleSet          rules.RuleSet[map[string]any]
	required             bool
	rules.NoConflict[Datum[T]]
}

func NewDatumRuleSet[T any](typeName string, attributesRuleSet rules.RuleSet[T]) *DatumRuleSet[T] {
	return &DatumRuleSet[T]{
		idRuleSet:            IDRuleSet,
		typeRuleSet:          rules.Constant[string](typeName),
		relationshipsRuleSet: RelationshipsRuleSet,
		attributesRuleSet:    attributesRuleSet,
		metaRuleSet:          MetaRuleSet,
	}
}

func (ruleSet *DatumRuleSet[T]) clone() *DatumRuleSet[T] {
	return &DatumRuleSet[T]{
		idRuleSet:            ruleSet.idRuleSet,
		typeRuleSet:          ruleSet.typeRuleSet,
		relationshipsRuleSet: ruleSet.relationshipsRuleSet,
		attributesRuleSet:    ruleSet.attributesRuleSet,
		required:             ruleSet.required,
		metaRuleSet:          ruleSet.metaRuleSet,
	}
}

func (ruleSet *DatumRuleSet[T]) WithRelationship(relName string, relRuleSet rules.RuleSet[Relationship]) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.relationshipsRuleSet = newRuleSet.relationshipsRuleSet.WithKey(relName, relRuleSet)
	return newRuleSet
}

func (ruleSet *DatumRuleSet[T]) WithUnknownRelationships() *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.relationshipsRuleSet = newRuleSet.relationshipsRuleSet.WithDynamicKey(rules.String(), RelationshipRuleSet)
	return newRuleSet
}

func (ruleSet *DatumRuleSet[T]) WithRequired() *DatumRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	newRuleSet := ruleSet.clone()
	newRuleSet.required = true
	return newRuleSet
}

func (ruleSet *DatumRuleSet[T]) Required() bool {
	return ruleSet.required
}

func (ruleSet *DatumRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationErrorCollection {
	datumValidator := rules.Struct[Datum[T]]().WithJson()
	datumValidator = datumValidator.WithKey("id", ruleSet.idRuleSet.Any())
	datumValidator = datumValidator.WithKey("type", ruleSet.typeRuleSet.Any())
	datumValidator = datumValidator.WithKey("attributes", ruleSet.attributesRuleSet.Any())
	datumValidator = datumValidator.WithKey("relationships", ruleSet.relationshipsRuleSet.Any())
	datumValidator = datumValidator.WithKey("meta", ruleSet.metaRuleSet.Any())

	errs := datumValidator.Apply(ctx, input, output)

	if errs == nil {
		t := ruleSet.typeRuleSet.Value()

		switch o := (output).(type) {
		case *Datum[T]:
			o.Type = t
		case *any:
			// Output is a pointer to *Interface{} which points to Datum[T]
			// I need to set
			switch oo := (*o).(type) {
			case Datum[T]:
				oo.Type = t
				reflect.ValueOf(output).Elem().Set(reflect.ValueOf(oo))
			case map[string]any:
				oo["Type"] = t
				reflect.ValueOf(output).Elem().Set(reflect.ValueOf(oo))
			}
		case map[string]any:
			o["Type"] = t
		}
	}

	return errs
}

func (ruleSet *DatumRuleSet[T]) Evaluate(ctx context.Context, value Datum[T]) errors.ValidationErrorCollection {
	var out Datum[T]
	return ruleSet.Apply(ctx, value, out)
}

func (ruleSet *DatumRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[Datum[T]](ruleSet)
}

func (ruleSet *DatumRuleSet[T]) String() string {
	return "DatumRuleSet"
}
