package jsonapi

import (
	"context"
	"encoding/json"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

// linkCast converts a raw value (string, map, or nil) into a Link for validation.
func linkCast(ctx context.Context, value any) (Link, errors.ValidationError) {
	// Link can be a string (StringLink), an object (FullLink), or null (NilLink)
	if value == nil {
		return NilLink{}, nil
	}

	// Convert to JSON bytes to use the custom UnmarshalJSON
	var jsonBytes []byte
	var err error

	if strValue, ok := value.(string); ok {
		// String link
		return StringLink(strValue), nil
	} else if mapValue, ok := value.(map[string]any); ok {
		// Full link - marshal the map to JSON so we can use FullLink unmarshaling
		jsonBytes, err = json.Marshal(mapValue)
		if err != nil {
			return nil, errors.Errorf(errors.CodeEncoding, ctx, "Link marshal failed", "Failed to marshal link: %v", err)
		}

		var fullLink FullLink
		if err := json.Unmarshal(jsonBytes, &fullLink); err != nil {
			return nil, errors.Errorf(errors.CodeEncoding, ctx, "Invalid link format", "Invalid link format: %v", err)
		}
		return &fullLink, nil
	}

	return nil, errors.Errorf(errors.CodeEncoding, ctx, "Invalid link type", "Link must be a string, object, or null")
}

var LinkRuleSet rules.RuleSet[Link] = rules.Interface[Link]().WithCast(linkCast)

var LinksRuleSet *rules.ObjectRuleSet[map[string]Link, string, Link] = rules.StringMap[Link]().WithDynamicKey(rules.String(), LinkRuleSet)
