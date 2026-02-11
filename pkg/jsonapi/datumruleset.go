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
	linksRuleSet         *rules.ObjectRuleSet[map[string]Link, string, Link]
	metaRuleSet          *rules.ObjectRuleSet[map[string]any, string, any]
	required             bool
	errorConfig          *errors.ErrorConfig
	rules.NoConflict[Datum[T]]
}

// NewDatumRuleSet returns a rule set for a single resource object with the given type and attributes validation.
func NewDatumRuleSet[T any](typeName string, attributesRuleSet rules.RuleSet[T]) *DatumRuleSet[T] {
	metaRuleSet := rules.StringMap[any]()
	return &DatumRuleSet[T]{
		idRuleSet:            IDRuleSet,
		typeRuleSet:          rules.Constant[string](typeName),
		relationshipsRuleSet: RelationshipsRuleSet,
		attributesRuleSet:    attributesRuleSet,
		linksRuleSet:         LinksRuleSet,
		metaRuleSet:          metaRuleSet,
	}
}

// clone returns a shallow copy of the rule set for use in builder methods.
func (ruleSet *DatumRuleSet[T]) clone() *DatumRuleSet[T] {
	return &DatumRuleSet[T]{
		idRuleSet:            ruleSet.idRuleSet,
		typeRuleSet:          ruleSet.typeRuleSet,
		relationshipsRuleSet: ruleSet.relationshipsRuleSet,
		attributesRuleSet:    ruleSet.attributesRuleSet,
		linksRuleSet:         ruleSet.linksRuleSet,
		required:             ruleSet.required,
		metaRuleSet:          ruleSet.metaRuleSet,
		errorConfig:          ruleSet.errorConfig,
	}
}

// WithRelationship registers a relationship name and its rule set.
func (ruleSet *DatumRuleSet[T]) WithRelationship(relName string, relRuleSet rules.RuleSet[Relationship]) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.relationshipsRuleSet = newRuleSet.relationshipsRuleSet.WithKey(relName, relRuleSet)
	return newRuleSet
}

// WithUnknownRelationships allows any relationship name with dynamic validation.
func (ruleSet *DatumRuleSet[T]) WithUnknownRelationships() *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.relationshipsRuleSet = newRuleSet.relationshipsRuleSet.WithDynamicKey(rules.String(), RelationshipRuleSet)
	return newRuleSet
}

// WithMeta registers a meta key and its rule set for the resource object.
func (ruleSet *DatumRuleSet[T]) WithMeta(key string, valueRuleSet rules.RuleSet[any]) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.metaRuleSet = newRuleSet.metaRuleSet.WithKey(key, valueRuleSet)
	return newRuleSet
}

// WithUnknownMeta allows any meta key on the resource object.
func (ruleSet *DatumRuleSet[T]) WithUnknownMeta() *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.metaRuleSet = newRuleSet.metaRuleSet.WithUnknown()
	return newRuleSet
}

// WithRequired marks the resource object as required when used as primary data.
func (ruleSet *DatumRuleSet[T]) WithRequired() *DatumRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	newRuleSet := ruleSet.clone()
	newRuleSet.required = true
	return newRuleSet
}

// Required reports whether the resource object is required when used as primary data.
func (ruleSet *DatumRuleSet[T]) Required() bool {
	return ruleSet.required
}

// WithNil allows nil input values to pass validation.
func (ruleSet *DatumRuleSet[T]) WithNil() *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	// Note: DatumRuleSet itself doesn't directly handle nil, but we can mark it as allowing nil
	// The actual nil handling is done by the struct validator
	return newRuleSet
}

// WithErrorMessage overrides error messages for this rule set.
func (ruleSet *DatumRuleSet[T]) WithErrorMessage(short, long string) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithErrorMessage(short, long)
	return newRuleSet
}

// WithDocsURI adds a documentation link to errors.
func (ruleSet *DatumRuleSet[T]) WithDocsURI(uri string) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithDocsURI(uri)
	return newRuleSet
}

// WithTraceURI adds a trace/debug link to errors.
func (ruleSet *DatumRuleSet[T]) WithTraceURI(uri string) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithTraceURI(uri)
	return newRuleSet
}

// WithErrorCode overrides the error code.
func (ruleSet *DatumRuleSet[T]) WithErrorCode(code errors.ErrorCode) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithCode(code)
	return newRuleSet
}

// WithErrorMeta adds custom metadata.
func (ruleSet *DatumRuleSet[T]) WithErrorMeta(key string, value any) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithMeta(key, value)
	return newRuleSet
}

// WithErrorCallback applies custom error processing.
func (ruleSet *DatumRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *DatumRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithCallback(fn)
	return newRuleSet
}

// Apply validates the input (resource object) and decodes it into output.
func (ruleSet *DatumRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationError {
	if ruleSet.errorConfig != nil {
		ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)
	}

	datumValidator := rules.Struct[Datum[T]]().WithJson()
	datumValidator = datumValidator.WithKey("id", ruleSet.idRuleSet.Any())
	datumValidator = datumValidator.WithKey("lid", rules.String().Any())
	datumValidator = datumValidator.WithKey("type", ruleSet.typeRuleSet.Any())
	datumValidator = datumValidator.WithKey("attributes", ruleSet.attributesRuleSet.Any())
	datumValidator = datumValidator.WithKey("relationships", ruleSet.relationshipsRuleSet.Any())
	datumValidator = datumValidator.WithKey("links", ruleSet.linksRuleSet.Any())
	datumValidator = datumValidator.WithKey("meta", ruleSet.metaRuleSet.Any())

	datumValidator = datumValidator.WithDynamicBucket(atMembersKeyRule, "AtMembers")
	datumValidator = datumValidator.WithDynamicBucket(extKeyRule, "ExtensionMembers")

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

// Evaluate validates a Datum value and returns any validation errors.
func (ruleSet *DatumRuleSet[T]) Evaluate(ctx context.Context, value Datum[T]) errors.ValidationError {
	var out Datum[T]
	return ruleSet.Apply(ctx, value, &out)
}

// Any returns the rule set as rules.RuleSet[any] for use with generic validators.
func (ruleSet *DatumRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[Datum[T]](ruleSet)
}

// String returns a stable name for the rule set for error messages and debugging.
func (ruleSet *DatumRuleSet[T]) String() string {
	return "DatumRuleSet"
}
