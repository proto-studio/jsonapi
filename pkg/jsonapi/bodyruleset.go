package jsonapi

import (
	"context"
	"encoding/json"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

type SingleRuleSet[T any] struct {
	datumRuleSet *DatumRuleSet[T]
	metaRuleSet  *rules.ObjectRuleSet[map[string]any, string, any]
	required     bool
	errorConfig  *errors.ErrorConfig
	rules.NoConflict[SingleDatumEnvelope[T]]
}

// NewSingleRuleSet returns a rule set for a single primary resource document with the given type and attributes validation.
func NewSingleRuleSet[T any](typeName string, attributesRuleSet rules.RuleSet[T]) *SingleRuleSet[T] {
	metaRuleSet := rules.StringMap[any]()
	return &SingleRuleSet[T]{
		datumRuleSet: NewDatumRuleSet(typeName, attributesRuleSet).WithRequired(),
		metaRuleSet:  metaRuleSet,
	}
}

// clone returns a shallow copy of the rule set for use in builder methods.
func (ruleSet *SingleRuleSet[T]) clone() *SingleRuleSet[T] {
	return &SingleRuleSet[T]{
		datumRuleSet: ruleSet.datumRuleSet,
		metaRuleSet:  ruleSet.metaRuleSet,
		required:     ruleSet.required,
		errorConfig:  ruleSet.errorConfig,
	}
}

// WithRelationship registers a relationship name and its rule set for the primary resource.
func (ruleSet *SingleRuleSet[T]) WithRelationship(relName string, relRuleSet rules.RuleSet[Relationship]) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.datumRuleSet = newRuleSet.datumRuleSet.WithRelationship(relName, relRuleSet)
	return newRuleSet
}

// WithUnknownRelationships allows any relationship name with dynamic validation.
func (ruleSet *SingleRuleSet[T]) WithUnknownRelationships() *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.datumRuleSet = newRuleSet.datumRuleSet.WithUnknownRelationships()
	return newRuleSet
}

// WithMeta registers a resource-level meta key and its rule set.
func (ruleSet *SingleRuleSet[T]) WithMeta(key string, valueRuleSet rules.RuleSet[any]) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.datumRuleSet = newRuleSet.datumRuleSet.WithMeta(key, valueRuleSet)
	return newRuleSet
}

// WithUnknownMeta allows any resource-level meta key.
func (ruleSet *SingleRuleSet[T]) WithUnknownMeta() *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.datumRuleSet = newRuleSet.datumRuleSet.WithUnknownMeta()
	return newRuleSet
}

// WithDocumentMeta registers a top-level document meta key and its rule set.
func (ruleSet *SingleRuleSet[T]) WithDocumentMeta(key string, valueRuleSet rules.RuleSet[any]) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.metaRuleSet = newRuleSet.metaRuleSet.WithKey(key, valueRuleSet)
	return newRuleSet
}

// WithUnknownDocumentMeta allows any top-level document meta key.
func (ruleSet *SingleRuleSet[T]) WithUnknownDocumentMeta() *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	newRuleSet.metaRuleSet = newRuleSet.metaRuleSet.WithUnknown()
	return newRuleSet
}

// WithRequired marks the primary data member as required.
func (ruleSet *SingleRuleSet[T]) WithRequired() *SingleRuleSet[T] {
	if ruleSet.required {
		return ruleSet
	}

	newRuleSet := ruleSet.clone()
	newRuleSet.required = true
	return newRuleSet
}

// Required reports whether the primary data member is required.
func (ruleSet *SingleRuleSet[T]) Required() bool {
	return ruleSet.required
}

// WithNil allows nil input values to pass validation.
func (ruleSet *SingleRuleSet[T]) WithNil() *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	return newRuleSet
}

// WithErrorMessage overrides error messages for this rule set.
func (ruleSet *SingleRuleSet[T]) WithErrorMessage(short, long string) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithErrorMessage(short, long)
	return newRuleSet
}

// WithDocsURI adds a documentation link to errors.
func (ruleSet *SingleRuleSet[T]) WithDocsURI(uri string) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithDocsURI(uri)
	return newRuleSet
}

// WithTraceURI adds a trace/debug link to errors.
func (ruleSet *SingleRuleSet[T]) WithTraceURI(uri string) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithTraceURI(uri)
	return newRuleSet
}

// WithErrorCode overrides the error code.
func (ruleSet *SingleRuleSet[T]) WithErrorCode(code errors.ErrorCode) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithCode(code)
	return newRuleSet
}

// WithErrorMeta adds custom metadata.
func (ruleSet *SingleRuleSet[T]) WithErrorMeta(key string, value any) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithMeta(key, value)
	return newRuleSet
}

// WithErrorCallback applies custom error processing.
func (ruleSet *SingleRuleSet[T]) WithErrorCallback(fn errors.ErrorCallback) *SingleRuleSet[T] {
	newRuleSet := ruleSet.clone()
	if newRuleSet.errorConfig == nil {
		newRuleSet.errorConfig = &errors.ErrorConfig{}
	}
	newRuleSet.errorConfig = newRuleSet.errorConfig.WithCallback(fn)
	return newRuleSet
}

// Apply decodes and validates the input (string or map) into the output envelope.
func (ruleSet *SingleRuleSet[T]) Apply(ctx context.Context, input, output any) errors.ValidationError {
	if ruleSet.errorConfig != nil {
		ctx = errors.WithErrorConfig(ctx, ruleSet.errorConfig)
	}

	// ObjectRuleSet is capable of decoding raw JSON but in this case we want to decode the JSON
	// ahead of time into a map so we can assign fields.
	// In the future if support is added upstream we can switch to using that.
	var decodedInput any
	if inputStr, ok := input.(string); ok {
		if err := json.Unmarshal([]byte(inputStr), &decodedInput); err != nil {
			return ToJSONAPIErrors(errors.Errorf(errors.CodeEncoding, ctx, "Invalid JSON encoding", "Body must be Json encoded"), SourcePointer)
		}
		input = decodedInput
	} else if inputMap, ok := input.(map[string]any); ok {
		decodedInput = inputMap
	}

	bodyValidator := rules.Struct[SingleDatumEnvelope[T]]()
	// Allow data to be nil for meta-only documents - wrap to handle nil
	dataRuleSet := rules.Interface[Datum[T]]().WithCast(func(ctx context.Context, value any) (Datum[T], errors.ValidationError) {
		if value == nil {
			// Return zero value for nil - meta-only documents are valid
			return Datum[T]{}, nil
		}
		var out Datum[T]
		errs := ruleSet.datumRuleSet.Apply(ctx, value, &out)
		return out, errs
	})
	bodyValidator = bodyValidator.WithKey("data", dataRuleSet.Any())
	bodyValidator = bodyValidator.WithKey("meta", ruleSet.metaRuleSet.Any())
	bodyValidator = bodyValidator.WithKey("links", LinksRuleSet.Any())
	bodyValidator = bodyValidator.WithKey("included", IncludedRuleSet.Any())
	// Allow jsonapi as a top-level member (JSON:API spec allows this)
	bodyValidator = bodyValidator.WithKey("jsonapi", rules.StringMap[any]().WithUnknown().Any())

	bodyValidator = bodyValidator.WithDynamicBucket(atMembersKeyRule, "AtMembers")
	bodyValidator = bodyValidator.WithDynamicBucket(extKeyRule, "ExtensionMembers")

	err := bodyValidator.Apply(ctx, input, output)

	if err != nil {
		return ToJSONAPIErrors(err, SourcePointer)
	}

	if decodedInput != nil {
		inputMap := decodedInput.(map[string]any)
		data, ok := inputMap["data"]
		if ok && data != nil {
			dataMap := data.(map[string]any)
			attributes, ok := dataMap["attributes"].(map[string]any)
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
	}

	return nil
}

// Evaluate validates a SingleDatumEnvelope value and returns any validation errors.
func (ruleSet *SingleRuleSet[T]) Evaluate(ctx context.Context, value SingleDatumEnvelope[T]) errors.ValidationError {
	var out SingleDatumEnvelope[T]
	return ruleSet.Apply(ctx, value, &out)
}

// Any returns the rule set as rules.RuleSet[any] for use with generic validators.
func (ruleSet *SingleRuleSet[T]) Any() rules.RuleSet[any] {
	return rules.WrapAny[SingleDatumEnvelope[T]](ruleSet)
}

// String returns a stable name for the rule set for error messages and debugging.
func (ruleSet *SingleRuleSet[T]) String() string {
	return "SingleRuleSet"
}
