package jsonapi

type ResourceLinkage interface {
	doNotExtend()
}

type Relationship struct {
	Links Links           `json:"links,omitempty" validate:"links"`
	Data  ResourceLinkage `json:"data,omitempty" validate:"data"`
	Meta  map[string]any  `json:"meta,omitempty" validate:"meta"`
}

type ResourceIdentifierLinkage struct {
	Type string         `json:"type" validate:"type"`
	ID   string         `json:"id,omitempty" validate:"id"`
	LID  string         `json:"lid,omitempty" validate:"lid"`
	Meta map[string]any `json:"meta,omitempty" validate:"meta"`
}

// doNotExtend prevents external types from satisfying ResourceLinkage without the intended methods.
func (ResourceIdentifierLinkage) doNotExtend() {}

type NilResourceLinkage struct{}

// MarshalJSON implements json.Marshaler for NilResourceLinkage and returns the JSON null literal.
func (n NilResourceLinkage) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler for NilResourceLinkage and accepts any input.
func (NilResourceLinkage) UnmarshalJSON(data []byte) error {
	return nil
}

// doNotExtend prevents external types from satisfying ResourceLinkage without the intended methods.
func (NilResourceLinkage) doNotExtend() {}

type ResourceLinkageCollection []ResourceIdentifierLinkage

// doNotExtend prevents external types from satisfying ResourceLinkage without the intended methods.
func (ResourceLinkageCollection) doNotExtend() {}
