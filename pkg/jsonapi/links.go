package jsonapi

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
	Title       string         `json:"title,omitempty" validate:"title"`
	Type        string         `json:"type,omitempty" validate:"type"`
	HrefLang    string         `json:"hreflang,omitempty" validate:"hreflang"`
	Rel         string         `json:"rel,omitempty" validate:"rel"`
	DescribedBy string         `json:"describedby,omitempty" validate:"describedby"`
	Meta        map[string]any `json:"meta,omitempty" validate:"meta"`
}

type Links map[string]Link
