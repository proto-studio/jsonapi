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
	Type string `json:"type" validate:"type"`
	ID   string `json:"id,omitempty" validate:"id"`
	LID  string `json:"lid,omitempty" validate:"lid"`
}

func (ResourceIdentifierLinkage) doNotExtend() {}

type NilResourceLinkage struct{}

func (n NilResourceLinkage) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (NilResourceLinkage) doNotExtend() {}

type ResourceLinkageCollection []ResourceIdentifierLinkage

func (ResourceLinkageCollection) doNotExtend() {}
