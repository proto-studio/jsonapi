package jsonapi

import (
	"encoding/json"
	"fmt"
)

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

type NilLink struct{}

func (NilLink) Href() string {
	return ""
}

func (NilLink) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (NilLink) UnmarshalJSON(data []byte) error {
	return nil
}

type Links map[string]Link

// UnmarshalJSON custom unmarshaller for the Links map
func (links *Links) UnmarshalJSON(data []byte) error {
	// Create a temporary map to hold the raw JSON data
	tempMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	// Initialize the Links map
	*links = make(Links)

	// Iterate over the temporary map and unmarshal each link
	for key, rawValue := range tempMap {
		// First, try to unmarshal as a StringLink
		var strLink StringLink
		if err := json.Unmarshal(rawValue, &strLink); err == nil {
			(*links)[key] = strLink
			continue
		}

		// If unmarshalling as a StringLink fails, try FullLink
		var fullLink FullLink
		if err := json.Unmarshal(rawValue, &fullLink); err == nil {
			(*links)[key] = &fullLink
			continue
		}

		// If neither works, return an error
		return fmt.Errorf("failed to unmarshal link for key %s", key)
	}

	return nil
}
