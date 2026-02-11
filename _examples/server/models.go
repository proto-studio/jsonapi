package main

import (
	"context"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

const baseURL = "http://localhost:8080"

// StoreAttributes is the attributes object for a store resource.
// validate tags map validator keys (lowercase) to these fields.
type StoreAttributes struct {
	Name    string `json:"name" validate:"name"`
	Address string `json:"address" validate:"address"`
}

// PetAttributes is the attributes object for a pet resource.
// validate tags map validator keys (lowercase) to these fields.
type PetAttributes struct {
	Name    string `json:"name" validate:"name"`
	Species string `json:"species" validate:"species"`
}

// StoreAttributesRuleSet validates store attributes (name required 1-100, address optional max 200).
// Keys must match the JSON:API request body: "name", "address" (lowercase).
func StoreAttributesRuleSet() rules.RuleSet[StoreAttributes] {
	return rules.Struct[StoreAttributes]().
		WithJson().
		WithKey("name", rules.String().WithMinLen(1).WithMaxLen(100).Any()).
		WithKey("address", rules.String().WithMaxLen(200).Any())
}

// PetAttributesRuleSet validates pet attributes (name 1-80, species required enum).
// Keys must match the JSON:API request body: "name", "species" (lowercase).
func PetAttributesRuleSet() rules.RuleSet[PetAttributes] {
	return rules.Struct[PetAttributes]().
		WithJson().
		WithKey("name", rules.String().WithMinLen(1).WithMaxLen(80).Any()).
		WithKey("species", rules.String().WithMinLen(1).WithRuleFunc(func(ctx context.Context, s string) errors.ValidationError {
			switch s {
			case "dog", "cat", "bird":
				return nil
			default:
				return errors.Errorf(errors.CodePattern, ctx, "Invalid species", "species must be one of: dog, cat, bird")
			}
		}).Any())
}

// StoreRuleSet is the single-resource rule set for stores (create/update body).
func StoreRuleSet() *jsonapi.SingleRuleSet[StoreAttributes] {
	return jsonapi.NewSingleRuleSet("stores", StoreAttributesRuleSet()).
		WithRelationship("pets", jsonapi.RelationshipRuleSet)
}

// PetRuleSet is the single-resource rule set for pets (create/update body).
func PetRuleSet() *jsonapi.SingleRuleSet[PetAttributes] {
	return jsonapi.NewSingleRuleSet("pets", PetAttributesRuleSet()).
		WithRelationship("store", jsonapi.RelationshipRuleSet)
}
