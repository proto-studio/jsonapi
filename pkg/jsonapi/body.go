package jsonapi

type Datum[T any] struct {
	ID         string `json:"id" validate:"id"`
	Type       string `json:"type" validate:"type"`
	Attributes T      `json:"attributes" validate:"attributes"`
	Links      Links  `json:"links,omitempty"`
}

type Link interface {
	Href() string
}

type FullLink struct {
	HrefValue   string         `json:"href" validate:"href"`
	Meta        map[string]any `json:"meta,omitempty" validate:"meta"`
	Rel         string         `json:"rel,omitempty" validate:"rel"`
	DescribedBy string         `json:"describedby,omitempty" validate:"describedby"`
	Title       string         `json:"title,omitempty" validate:"title"`
	Type        string         `json:"type,omitempty" validate:"type"`
	HrefLang    string         `json:"hreflang,omitempty" validate:"hreflang"` // TODO: this can also be an array of strings
}

func (link *FullLink) Href() string {
	return link.HrefValue
}

type StringLink string

func (str StringLink) Href() string {
	return string(str)
}

type StructuredLink struct {
	Href        string         `json:"href" validate:"href"`
	Title       string         `json:"title" validate:"title"`
	Type        string         `json:"type" validate:"type"`
	HrefLang    string         `json:"hreflang" validate:"hreflang"`
	Rel         string         `json:"rel" validate:"rel"`
	DescribedBy string         `json:"describedby" validate:"describedby"`
	Meta        map[string]any `json:"meta" validate:"meta"`
}

type Links map[string]Link

type SingleDatum[T any] struct {
	Data  Datum[T] `json:"data" validate:"data"`
	Links Links    `json:"links"`
}

type DatumCollection[T any] struct {
	Data  []Datum[T] `json:"data" validate:"data"`
	Links Links      `json:"links"`
}
