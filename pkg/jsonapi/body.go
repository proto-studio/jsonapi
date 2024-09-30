package jsonapi

import (
	"encoding/json"
	"reflect"
	"strings"
)

type Datum[T any] struct {
	ID               string                  `json:"id" validate:"id"`
	Type             string                  `json:"type" validate:"type"`
	Attributes       T                       `json:"attributes" validate:"attributes"`
	Links            Links                   `json:"links,omitempty"`
	Relationships    map[string]Relationship `json:"relationships,omitempty" validate:"relationships"`
	Meta             map[string]any          `json:"meta,omitempty" validate:"meta"`
	ExtensionMembers map[string]any          `json:"-"`
	Fields           FieldList               `json:"-"`
}

// MarshalJSON implements the json.Marshaler interface for Datum[T].
// Output is filtered by Fields if present and extension members are copied into the resulting Json.
func (d Datum[T]) MarshalJSON() ([]byte, error) {
	// Create a map to hold the final JSON object
	result := make(map[string]any)

	// Add non-Attributes fields
	result["id"] = d.ID
	result["type"] = d.Type
	if len(d.Links) > 0 {
		result["links"] = d.Links
	}
	if len(d.Relationships) > 0 {
		result["relationships"] = d.Relationships
	}
	if len(d.Meta) > 0 {
		result["meta"] = d.Meta
	}

	// Handle Attributes field
	if d.Fields == nil {
		// If Fields is nil, marshal Attributes as is
		result["attributes"] = d.Attributes
	} else {
		// If Fields is not nil, only serialize the fields in the FieldList
		attrMap := make(map[string]any)
		attrValue := reflect.ValueOf(d.Attributes)
		attrType := attrValue.Type()

		for i := 0; i < attrType.NumField(); i++ {
			field := attrType.Field(i)
			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				fieldName = field.Name
			}
			if d.Fields.Contains(fieldName) {
				attrMap[fieldName] = attrValue.Field(i).Interface()
			}
		}
		if len(attrMap) > 0 {
			result["attributes"] = attrMap
		}
	}

	// Copy all key-value pairs from ExtensionMembers to the parent JSON object
	for key, value := range d.ExtensionMembers {
		result[key] = value
	}

	// Marshal the result map to JSON
	return json.Marshal(result)
}

// UnmarshalJSON implements the json.Unmarshaler interface for Datum
func (d *Datum[T]) UnmarshalJSON(data []byte) error {
	// Unmarshal the data into a map of json.RawMessage
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}

	// Initialize attribute fields map
	attributeFields := make(fieldListMap, len(rawData))

	// Reflect on the Datum struct to identify fields based on JSON tags
	datumValue := reflect.ValueOf(d).Elem()
	datumType := datumValue.Type()

	// Map to hold the struct fields by their JSON tag name
	structFields := make(map[string]reflect.Value)

	for i := 0; i < datumType.NumField(); i++ {
		field := datumType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Get the first part of the JSON tag before any comma (e.g., "id,omitempty" -> "id")
		tagName := strings.Split(jsonTag, ",")[0]
		structFields[tagName] = datumValue.Field(i)
	}

	// Handle each field in rawData
	for key, value := range rawData {
		if fieldValue, exists := structFields[key]; exists {
			// Unmarshal the value into the corresponding field
			if fieldValue.CanSet() {
				if err := json.Unmarshal(value, fieldValue.Addr().Interface()); err != nil {
					return err
				}

				// If the field is "attributes", capture the fields present in the attributes JSON
				if key == "attributes" {
					var attrMap map[string]json.RawMessage
					if err := json.Unmarshal(value, &attrMap); err != nil {
						return err
					}
					for attrKey := range attrMap {
						attributeFields[attrKey] = true
					}
				}
			}
		} else {
			// Handle ExtensionMembers if the key contains a ":" and it's not at the start
			if idx := strings.Index(key, ":"); idx > 0 {
				var rawValue any
				if err := json.Unmarshal(value, &rawValue); err != nil {
					return err
				}
				if d.ExtensionMembers == nil {
					d.ExtensionMembers = make(map[string]any)
				}
				d.ExtensionMembers[key] = rawValue
			}
		}
	}

	// Set the Fields property to reflect the fields actually present in the JSON
	d.Fields = attributeFields

	return nil
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
