package jsonapi

import (
	"proto.zip/studio/validate"
	"proto.zip/studio/validate/pkg/rules/objects"
)

func NewRequestValidator[T any](attributesValidator *objects.ObjectRuleSet[T]) *objects.ObjectRuleSet[SingleDatum[T]] {
	datumValidator := validate.Object[Datum[T]]()
	datumValidator = datumValidator.WithKey("id", validate.String().WithStrict().Any())
	datumValidator = datumValidator.WithKey("type", validate.String().WithStrict().Any())
	datumValidator = datumValidator.WithKey("attributes", attributesValidator.WithRequired().Any())

	envelopeValidator := validate.Object[SingleDatum[T]]()
	envelopeValidator = envelopeValidator.WithKey("data", datumValidator.WithRequired().Any())

	return envelopeValidator
}
