package jsonapi

type Datum[T any] struct {
	ID               string                  `json:"id" validate:"id"`
	Type             string                  `json:"type" validate:"type"`
	Attributes       T                       `json:"attributes" validate:"attributes"`
	Links            Links                   `json:"links,omitempty"`
	Relationships    map[string]Relationship `json:"relationships,omitempty" validate:"relationships"`
	Meta             map[string]any          `json:"meta,omitempty" validate:"meta"`
	ExtensionMembers map[string]any          `json:"-"`
}

type SingleDatumEnvelope[T any] struct {
	Data             Datum[T]       `json:"data" validate:"data"`
	Links            Links          `json:"links,omitempty"`
	Meta             map[string]any `json:"meta,omitempty" validate:"meta"`
	ExtensionMembers map[string]any `json:"-"`
}

type DatumCollectionEnvelope[T any] struct {
	Data             []Datum[T]     `json:"data" validate:"data"`
	Links            Links          `json:"links,omitempty"`
	Meta             map[string]any `json:"meta,omitempty" validate:"meta"`
	ExtensionMembers map[string]any `json:"-"`
}
