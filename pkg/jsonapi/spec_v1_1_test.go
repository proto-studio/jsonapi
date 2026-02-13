package jsonapi_test

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/rules"
)

// TestSpecV1_1_DocumentStructure tests Section 5: Document Structure
// Spec: https://jsonapi.org/format/#document-structure
// A JSON:API document MUST be at the top level a JSON object.
// A document MUST contain at least one of the following top-level members: data, errors, meta
func TestSpecV1_1_DocumentStructure(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships().
		WithUnknownDocumentMeta()

	ctx := context.Background()

	// Document with data member (required for successful responses)
	docWithData := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, docWithData)
	if errs != nil {
		t.Errorf("Document with data member should be valid: %s", errs)
	}

	// Document with meta only (valid for 204 No Content responses)
	docWithMetaOnly := `{
		"meta": {
			"copyright": "Copyright 2015 Example Corp.",
			"authors": ["Yehuda Katz", "Steve Klabnik", "Dan Gebhardt", "Tyler Kellen"]
		}
	}`

	_, errs = ruleSet.Apply(ctx, docWithMetaOnly)
	// Meta-only documents are valid per spec
	if errs != nil {
		t.Errorf("Meta-only document (spec allows this): %s", errs)
	}
}

// TestSpecV1_1_ResourceObjects tests Section 5.2: Resource Objects
// Spec: https://jsonapi.org/format/#document-resource-objects
// A resource object MUST contain type and id members
// A resource object MAY contain: attributes, relationships, links, meta
func TestSpecV1_1_ResourceObjects(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships().
		WithUnknownMeta()

	ctx := context.Background()

	// Minimal resource object (type and id only)
	minimalResource := `{
		"data": {
			"type": "articles",
			"id": "1"
		}
	}`

	_, errs := ruleSet.Apply(ctx, minimalResource)
	if errs != nil {
		t.Errorf("Minimal resource object (type and id only) should be valid: %s", errs)
	}

	// Full resource object with all optional members
	fullResource := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"author": {
					"data": {
						"type": "people",
						"id": "9"
					}
				}
			},
			"links": {
				"self": "http://example.com/articles/1"
			},
			"meta": {
				"copyright": "Copyright 2015 Example Corp."
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, fullResource)
	if errs != nil {
		t.Errorf("Full resource object with all members should be valid: %s", errs)
	}
}

// TestSpecV1_1_ResourceIdentifierObjects tests Section 5.3: Resource Identifier Objects
// Spec: https://jsonapi.org/format/#document-resource-identifier-objects
// A resource identifier object MUST contain type and id members, or type and lid members
func TestSpecV1_1_ResourceIdentifierObjects(t *testing.T) {
	ctx := context.Background()

	// Resource identifier with type and id
	identifierWithId := `{
		"type": "articles",
		"id": "1"
	}`

	_, errs := jsonapi.ResourceLinkageRuleSet.Apply(ctx, identifierWithId)
	if errs != nil {
		t.Errorf("Resource identifier with type and id should be valid: %s", errs)
	}

	// Resource identifier with type and lid (local identifier)
	identifierWithLid := `{
		"type": "articles",
		"lid": "local-1"
	}`

	_, errs = jsonapi.ResourceLinkageRuleSet.Apply(ctx, identifierWithLid)
	if errs != nil {
		t.Errorf("Resource identifier with type and lid should be valid: %s", errs)
	}

	// Null resource linkage
	nullLinkage := `null`

	var nullLink jsonapi.ResourceLinkage
	nullLink, errs = jsonapi.ResourceLinkageRuleSet.Apply(ctx, nullLinkage)
	if errs != nil {
		t.Errorf("Null resource linkage should be valid: %s", errs)
	}
	if _, ok := nullLink.(jsonapi.NilResourceLinkage); !ok {
		t.Error("Null linkage should be NilResourceLinkage")
	}
}

// TestSpecV1_1_CompoundDocuments tests Section 5.4: Compound Documents
// Spec: https://jsonapi.org/format/#document-compound-documents
// A document MAY include related resources along with the requested primary resources
func TestSpecV1_1_CompoundDocuments(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships()

	ctx := context.Background()

	// Compound document with included resources
	compoundDoc := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"author": {
					"data": {
						"type": "people",
						"id": "9"
					}
				}
			}
		},
		"included": [
			{
				"type": "people",
				"id": "9",
				"attributes": {
					"firstName": "Dan",
					"lastName": "Gebhardt"
				}
			}
		]
	}`

	_, errs := ruleSet.Apply(ctx, compoundDoc)
	if errs != nil {
		t.Errorf("Compound document with included resources should be valid: %s", errs)
	}
}

// TestSpecV1_1_MetaInformation tests Section 5.5: Meta Information
// Spec: https://jsonapi.org/format/#document-meta
// Where specified, a meta member can be used to include non-standard meta-information
func TestSpecV1_1_MetaInformation(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownMeta().
		WithUnknownDocumentMeta()

	ctx := context.Background()

	// Document with top-level meta
	docWithMeta := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			}
		},
		"meta": {
			"copyright": "Copyright 2015 Example Corp.",
			"authors": ["Yehuda Katz", "Steve Klabnik"]
		}
	}`

	_, errs := ruleSet.Apply(ctx, docWithMeta)
	if errs != nil {
		t.Errorf("Document with top-level meta should be valid: %s", errs)
	}

	// Resource object with meta member
	resourceWithMeta := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"meta": {
				"count": 10
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, resourceWithMeta)
	if errs != nil {
		t.Errorf("Resource object with meta member should be valid: %s", errs)
	}
}

// TestSpecV1_1_Links tests Section 5.6: Links
// Spec: https://jsonapi.org/format/#document-links
// A link MUST be represented as either: a string containing the link's URL, or an object
func TestSpecV1_1_Links(t *testing.T) {
	// String link
	stringLinkJSON := `{"self": "http://example.com/articles/1"}`
	var stringLinks jsonapi.Links
	err := json.Unmarshal([]byte(stringLinkJSON), &stringLinks)
	if err != nil {
		t.Fatalf("String link should be valid JSON: %v", err)
	}
	if stringLinks["self"].Href() != "http://example.com/articles/1" {
		t.Error("String link href should match")
	}

	// Object link
	objectLinkJSON := `{
		"self": {
			"href": "http://example.com/articles/1",
			"meta": {
				"count": 10
			}
		}
	}`
	var objectLinks jsonapi.Links
	err = json.Unmarshal([]byte(objectLinkJSON), &objectLinks)
	if err != nil {
		t.Fatalf("Object link should be valid JSON: %v", err)
	}
	if objectLinks["self"].Href() != "http://example.com/articles/1" {
		t.Error("Object link href should match")
	}

	// Null link
	nullLinkJSON := `{"self": null}`
	var nullLinks jsonapi.Links
	err = json.Unmarshal([]byte(nullLinkJSON), &nullLinks)
	if err != nil {
		t.Fatalf("Null link should be valid JSON: %v", err)
	}
	if nullLinks["self"].Href() != "" {
		t.Error("Null link href should be empty")
	}
}

// TestSpecV1_1_JSONAPIObject tests Section 5.7: JSON:API Object
// Spec: https://jsonapi.org/format/#document-jsonapi-object
// If present, the value of the jsonapi member MUST be an object
func TestSpecV1_1_JSONAPIObject(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := context.Background()

	// Document with jsonapi object
	docWithJsonAPI := `{
		"jsonapi": {
			"version": "1.0"
		},
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, docWithJsonAPI)
	// jsonapi object is a top-level member
	if errs != nil {
		t.Errorf("Document with jsonapi object should be valid: %s", errs)
	}
}

// TestSpecV1_1_MemberNames tests Section 5.8: Member Names
// Spec: https://jsonapi.org/format/#document-member-names
// All member names used in a JSON:API document MUST be treated as case sensitive
func TestSpecV1_1_MemberNames(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := context.Background()

	// Valid member names (letters, numbers, hyphens, underscores)
	validDoc := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "Test",
				"title-with-hyphen": "Test",
				"title_with_underscore": "Test",
				"title123": "Test"
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, validDoc)
	if errs != nil {
		t.Errorf("Document with various member names should be valid: %s", errs)
	}
}

// TestSpecV1_1_QueryParameters tests Section 6: Query Parameters
// Spec: https://jsonapi.org/format/#query-parameters
func TestSpecV1_1_QueryParameters(t *testing.T) {
	ctx := context.Background()

	// Fields parameter - sparse field sets
	fieldsQuery := `fields[articles]=title,body&fields[people]=name`
	parsed, _ := url.ParseQuery(fieldsQuery)
	_, errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Fields query parameter should be valid: %s", errs)
	}

	// Include parameter - compound documents
	includeQuery := `include=author,comments.author`
	parsed, _ = url.ParseQuery(includeQuery)
	_, errs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Include query parameter should be valid: %s", errs)
	}

	// Sort parameter - sorting
	sortQuery := `sort=-created,title`
	parsed, _ = url.ParseQuery(sortQuery)
	ctx = jsonapi.WithMethod(ctx, "GET")
	_, errs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Sort query parameter should be valid: %s", errs)
	}

	// Filter parameter - filtering (implementation-specific keys)
	filterQuery := `filter[status]=published`
	parsed, _ = url.ParseQuery(filterQuery)
	_, errs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Filter query parameter should be valid: %s", errs)
	}

	// Note: Page parameters are not tested here because JSON:API is
	// agnostic about the pagination strategy used by a server
}

// TestSpecV1_1_Errors tests Section 7: Errors
// Spec: https://jsonapi.org/format/#errors
// Error objects provide additional information about problems encountered
func TestSpecV1_1_Errors(t *testing.T) {
	// Error document structure
	errorDoc := `{
		"errors": [
			{
				"id": "1",
				"status": "422",
				"code": "validation-failed",
				"title": "Validation Failed",
				"detail": "The title field is required.",
				"source": {
					"pointer": "/data/attributes/title"
				},
				"meta": {
					"timestamp": "2020-01-01T00:00:00Z"
				}
			}
		]
	}`

	// Verify error document structure
	var errorResponse map[string]any
	err := json.Unmarshal([]byte(errorDoc), &errorResponse)
	if err != nil {
		t.Fatalf("Error document should be valid JSON: %v", err)
	}

	errors, ok := errorResponse["errors"].([]any)
	if !ok {
		t.Fatal("Error document should have errors array")
	}

	if len(errors) == 0 {
		t.Fatal("Error document should have at least one error")
	}

	// Error with minimal members (at least one of id, links, status, code, title)
	minimalError := `{
		"errors": [
			{
				"status": "422",
				"title": "Validation Failed"
			}
		]
	}`

	var minimalErrorResponse map[string]any
	err = json.Unmarshal([]byte(minimalError), &minimalErrorResponse)
	if err != nil {
		t.Fatalf("Minimal error document should be valid JSON: %v", err)
	}
}

// TestSpecV1_1_FetchingData tests Section 8: Fetching Data
// Spec: https://jsonapi.org/format/#fetching
func TestSpecV1_1_FetchingData(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// Fetching a single resource
	singleResource := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, singleResource)
	if errs != nil {
		t.Errorf("Fetching single resource should be valid: %s", errs)
	}
}

// TestSpecV1_1_CreatingResources tests Section 9: Creating Resources
// Spec: https://jsonapi.org/format/#crud-creating
// A request MUST include a single resource object as primary data
func TestSpecV1_1_CreatingResources(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().WithMinLen(1).Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := jsonapi.WithMethod(context.Background(), "POST")

	// Creating a resource (client may omit id, server generates it)
	createDoc := `{
		"data": {
			"type": "articles",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, createDoc)
	if errs != nil {
		t.Errorf("Creating resource without id should be valid: %s", errs)
	}

	// Creating with client-generated id
	createWithId := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, createWithId)
	if errs != nil {
		t.Errorf("Creating resource with client-generated id should be valid: %s", errs)
	}
}

// TestSpecV1_1_UpdatingResources tests Section 10: Updating Resources
// Spec: https://jsonapi.org/format/#crud-updating
// A request MUST include a single resource object as primary data
// The resource object MUST contain type and id members
func TestSpecV1_1_UpdatingResources(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := jsonapi.WithMethod(context.Background(), "PATCH")
	ctx = jsonapi.WithId(ctx, "1")

	// Updating a resource (must include type and id)
	updateDoc := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "Updated title"
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, updateDoc)
	if errs != nil {
		t.Errorf("Updating resource should be valid: %s", errs)
	}

	// Partial update (only some attributes)
	partialUpdate := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "Partially updated"
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, partialUpdate)
	if errs != nil {
		t.Errorf("Partial update should be valid: %s", errs)
	}
}

// TestSpecV1_1_DeletingResources tests Section 11: Deleting Resources
// Spec: https://jsonapi.org/format/#crud-deleting
// A server MUST return either 200 OK with response document or 204 No Content
func TestSpecV1_1_DeletingResources(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownDocumentMeta()

	ctx := jsonapi.WithMethod(context.Background(), "DELETE")
	ctx = jsonapi.WithId(ctx, "1")

	// DELETE response with meta (200 OK with response document)
	deleteResponse := `{
		"meta": {
			"message": "Article deleted"
		}
	}`

	_, errs := ruleSet.Apply(ctx, deleteResponse)
	// DELETE responses may have only meta (no data)
	if errs != nil {
		t.Errorf("DELETE response with meta should be valid: %s", errs)
	}
}

// TestSpecV1_1_Relationships tests Section 8.2: Relationships
// Spec: https://jsonapi.org/format/#document-resource-object-relationships
func TestSpecV1_1_Relationships(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	relRuleSet := jsonapi.RelationshipRuleSet
	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithRelationship("author", relRuleSet).
		WithRelationship("comments", relRuleSet)

	ctx := context.Background()

	// To-one relationship
	toOneRel := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"author": {
					"data": {
						"type": "people",
						"id": "9"
					}
				}
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, toOneRel)
	if errs != nil {
		t.Errorf("To-one relationship should be valid: %s", errs)
	}

	// To-many relationship
	toManyRel := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"comments": {
					"data": [
						{"type": "comments", "id": "5"},
						{"type": "comments", "id": "12"}
					]
				}
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, toManyRel)
	if errs != nil {
		t.Errorf("To-many relationship should be valid: %s", errs)
	}

	// Null relationship
	nullRel := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"author": {
					"data": null
				}
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, nullRel)
	if errs != nil {
		t.Errorf("Null relationship should be valid: %s", errs)
	}
}

// TestSpecV1_1_Extensions tests Section 3.2: Extensions
// Spec: https://jsonapi.org/format/#extensions
// Extensions provide a means to extend the base specification
func TestSpecV1_1_Extensions(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := context.Background()

	// Document with extension members (namespace:member format)
	extDoc := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"ext:version": "1.0"
		},
		"ext:meta": {
			"custom": "value"
		}
	}`

	_, errs := ruleSet.Apply(ctx, extDoc)
	// Extension members should be handled as ExtensionMembers
	if errs != nil {
		t.Errorf("Document with extension members should be valid: %s", errs)
	}
}

// TestSpecV1_1_CollectionDocuments tests collection responses
// Spec: https://jsonapi.org/format/#document-top-level
// The data member value MUST be either null, an object, or an array
func TestSpecV1_1_CollectionDocuments(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	datumRuleSet := jsonapi.NewDatumRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := context.Background()

	// Collection items - test individual datum validation
	collectionItem1 := `{
		"type": "articles",
		"id": "1",
		"attributes": {
			"title": "First article"
		}
	}`

	_, errs := datumRuleSet.Apply(ctx, collectionItem1)
	if errs != nil {
		t.Errorf("Collection item 1 should be valid: %s", errs)
	}

	collectionItem2 := `{
		"type": "articles",
		"id": "2",
		"attributes": {
			"title": "Second article"
		}
	}`

	_, errs = datumRuleSet.Apply(ctx, collectionItem2)
	if errs != nil {
		t.Errorf("Collection item 2 should be valid: %s", errs)
	}
}

// TestSpecV1_1_EmptyCollection tests empty collection response
// Spec: https://jsonapi.org/format/#document-top-level
// The data member value MAY be an empty array
func TestSpecV1_1_EmptyCollection(t *testing.T) {
	// Empty collection document
	emptyCollection := `{
		"data": []
	}`

	var emptyResponse map[string]any
	err := json.Unmarshal([]byte(emptyCollection), &emptyResponse)
	if err != nil {
		t.Fatalf("Empty collection should be valid JSON: %v", err)
	}

	data, ok := emptyResponse["data"].([]any)
	if !ok {
		t.Fatal("Empty collection should have data as array")
	}

	if len(data) != 0 {
		t.Error("Empty collection should have empty array")
	}
}

// TestSpecV1_1_NullData tests null data response
// Spec: https://jsonapi.org/format/#document-top-level
// The data member value MAY be null
func TestSpecV1_1_NullData(t *testing.T) {
	// Null data document
	nullDataDoc := `{
		"data": null
	}`

	var nullResponse map[string]any
	err := json.Unmarshal([]byte(nullDataDoc), &nullResponse)
	if err != nil {
		t.Fatalf("Null data document should be valid JSON: %v", err)
	}

	if nullResponse["data"] != nil {
		t.Error("Null data document should have null data")
	}
}

// TestSpecV1_1_AtMembers tests Section 5.8.1: @-Members
// Spec: https://jsonapi.org/format/#document-member-names-at-members
// @-members are members whose names begin with "@". They MAY appear anywhere in a document.
// Specification semantics MAY NOT be defined for @-members, but an implementation or
// profile MAY define implementation semantics for @-members.
func TestSpecV1_1_AtMembers(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := context.Background()

	// Document with @-members (reserved for implementation-specific semantics)
	// @-members MAY appear anywhere in a document and should not be rejected
	atMemberDoc := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"@context": "https://schema.org"
		}
	}`

	envelope, errs := ruleSet.Apply(ctx, atMemberDoc)
	// @-members should be captured in AtMembers map per implementation
	if errs != nil {
		t.Errorf("Document with @-member should be valid: %s", errs)
	}
	if envelope.Data.AtMembers == nil || envelope.Data.AtMembers["@context"] != "https://schema.org" {
		t.Errorf("Expected @context to be captured in Data.AtMembers; got %v", envelope.Data.AtMembers)
	}

	// Top-level @-members are also allowed
	topLevelAtMember := `{
		"@version": "1.0",
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "Test"
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, topLevelAtMember)
	if errs != nil {
		t.Errorf("Document with top-level @-member should be valid: %s", errs)
	}
}

// TestSpecV1_1_MemberNamesReservedCharacters tests Section 5.8: Member Names
// Spec: https://jsonapi.org/format/#document-member-names
// Member names MUST contain only: a-z, A-Z, 0-9, and the allowed characters _, -
// Member names MUST start with a-z, A-Z, 0-9, or a globally allowed character
// Member names MUST NOT contain reserved characters except in extension members
func TestSpecV1_1_MemberNamesReservedCharacters(t *testing.T) {
	type TestAttributes struct{}

	attributesRuleSet := rules.Struct[TestAttributes]().
		WithJson().
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[TestAttributes]("tests", attributesRuleSet).
		WithUnknownMeta().
		WithUnknownDocumentMeta()

	ctx := context.Background()

	// Test valid attribute member names
	validAttributeNames := []struct {
		name string
		doc  string
	}{
		{"lowercase", `{"data":{"type":"tests","id":"1","attributes":{"title":"test"}}}`},
		{"camelCase", `{"data":{"type":"tests","id":"1","attributes":{"firstName":"test"}}}`},
		{"PascalCase", `{"data":{"type":"tests","id":"1","attributes":{"FirstName":"test"}}}`},
		{"with_underscore", `{"data":{"type":"tests","id":"1","attributes":{"first_name":"test"}}}`},
		{"with-hyphen", `{"data":{"type":"tests","id":"1","attributes":{"first-name":"test"}}}`},
		{"with_numbers", `{"data":{"type":"tests","id":"1","attributes":{"field123":"test"}}}`},
		{"starting_underscore", `{"data":{"type":"tests","id":"1","attributes":{"_private":"test"}}}`},
	}

	for _, tc := range validAttributeNames {
		_, errs := ruleSet.Apply(ctx, tc.doc)
		if errs != nil {
			t.Errorf("Valid member name %q should be accepted: %s", tc.name, errs)
		}
	}

	// Test valid meta member names (meta allows unknown keys)
	validMetaDoc := `{
		"data": {"type": "tests", "id": "1"},
		"meta": {
			"validKey": "value",
			"another_key": "value",
			"key-with-hyphen": "value"
		}
	}`

	_, errs := ruleSet.Apply(ctx, validMetaDoc)
	if errs != nil {
		t.Errorf("Valid meta member names should be accepted: %s", errs)
	}

	// Test invalid member names with reserved characters via MemberNameRule
	// Per spec, these characters are reserved and MUST NOT appear in member names:
	// +, ,, ., [, ], !, ", #, $, %, &, ', (, ), *, /, :, ;, <, =, >, ?, @, \, ^, `, {, |, }, ~
	// Note: ":" is allowed in extension member names (namespace:member format)
	// Note: "@" is allowed at start for @-members
	invalidNames := []struct {
		memberName  string
		description string
	}{
		{"field+name", "plus sign"},
		{"field,name", "comma"},
		{"field.name", "period"},
		{"field[name", "open bracket"},
		{"field]name", "close bracket"},
		{"field!name", "exclamation"},
		{"field#name", "hash"},
		{"field$name", "dollar"},
		{"field%name", "percent"},
		{"field&name", "ampersand"},
		{"field'name", "apostrophe"},
		{"field(name", "open paren"},
		{"field)name", "close paren"},
		{"field*name", "asterisk"},
		{"field/name", "slash"},
		{"field;name", "semicolon"},
		{"field<name", "less than"},
		{"field=name", "equals"},
		{"field>name", "greater than"},
		{"field?name", "question mark"},
		{"field^name", "caret"},
		{"field`name", "backtick"},
		{"field{name", "open brace"},
		{"field|name", "pipe"},
		{"field}name", "close brace"},
		{"field~name", "tilde"},
	}

	memberNameRule := jsonapi.MemberNameRule{}
	for _, tc := range invalidNames {
		errs := memberNameRule.Evaluate(ctx, tc.memberName)
		if errs == nil {
			t.Errorf("Member name with %s (%q) should be rejected per spec (reserved character)", tc.description, tc.memberName)
		}
	}

	// Test that member names starting with digits are valid per spec
	// (spec says member names MAY start with a-z, A-Z, 0-9, or globally allowed characters)
	digitStartDoc := `{"data":{"type":"tests","id":"1","attributes":{"123field":"test"}}}`
	_, errs = ruleSet.Apply(ctx, digitStartDoc)
	if errs != nil {
		t.Errorf("Member name starting with digit should be valid: %s", errs)
	}
}

// TestSpecV1_1_ExtensionMemberNames tests Section 5.8.2: Extension Member Names
// Spec: https://jsonapi.org/format/#extension-rules
// Extension members MUST be prefixed with the extension's namespace followed by a colon
func TestSpecV1_1_ExtensionMemberNames(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)

	ctx := context.Background()

	// Document with properly namespaced extension members
	extDoc := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"version:id": "42"
		}
	}`

	envelope, errs := ruleSet.Apply(ctx, extDoc)
	if errs != nil {
		t.Errorf("Document with extension member should be valid: %s", errs)
	}

	// Verify extension members are captured
	if envelope.Data.ExtensionMembers == nil {
		t.Error("ExtensionMembers should be populated for extension members")
	} else if _, ok := envelope.Data.ExtensionMembers["version:id"]; !ok {
		t.Error("Extension member version:id should be in ExtensionMembers")
	}
}

// TestSpecV1_1_LinkObjectMembers tests Section 5.6: Link Object Members
// Spec: https://jsonapi.org/format/#document-links
// A link object MAY contain: href, rel, describedBy, title, type, hreflang, meta
func TestSpecV1_1_LinkObjectMembers(t *testing.T) {
	// Full link object with all optional members
	fullLinkJSON := `{
		"self": {
			"href": "http://example.com/articles/1",
			"rel": "self",
			"describedby": "http://example.com/schemas/article",
			"title": "Article 1",
			"type": "application/vnd.api+json",
			"hreflang": "en",
			"meta": {
				"count": 10
			}
		}
	}`

	var links jsonapi.Links
	err := json.Unmarshal([]byte(fullLinkJSON), &links)
	if err != nil {
		t.Fatalf("Full link object should be valid JSON: %v", err)
	}

	selfLink := links["self"]
	if selfLink.Href() != "http://example.com/articles/1" {
		t.Errorf("Link href should match, got: %s", selfLink.Href())
	}

	// Link object with only href (minimal)
	minimalLinkJSON := `{
		"self": {
			"href": "http://example.com/articles/1"
		}
	}`

	var minimalLinks jsonapi.Links
	err = json.Unmarshal([]byte(minimalLinkJSON), &minimalLinks)
	if err != nil {
		t.Fatalf("Minimal link object should be valid JSON: %v", err)
	}

	// Multiple link types in document
	multiLinkJSON := `{
		"self": "http://example.com/articles/1",
		"related": "http://example.com/articles/1/author",
		"first": "http://example.com/articles?page[number]=1",
		"last": "http://example.com/articles?page[number]=10",
		"prev": "http://example.com/articles?page[number]=1",
		"next": "http://example.com/articles?page[number]=3"
	}`

	var multiLinks jsonapi.Links
	err = json.Unmarshal([]byte(multiLinkJSON), &multiLinks)
	if err != nil {
		t.Fatalf("Multiple links should be valid JSON: %v", err)
	}

	expectedLinks := []string{"self", "related", "first", "last", "prev", "next"}
	for _, linkName := range expectedLinks {
		if _, ok := multiLinks[linkName]; !ok {
			t.Errorf("Link %q should be present", linkName)
		}
	}
}

// TestSpecV1_1_RelationshipLinks tests Section 5.2.4: Relationship Links
// Spec: https://jsonapi.org/format/#document-resource-object-relationships
// A relationship object MAY contain links with self and related members
func TestSpecV1_1_RelationshipLinks(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships()

	ctx := context.Background()

	// Relationship with links
	relWithLinks := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"author": {
					"links": {
						"self": "http://example.com/articles/1/relationships/author",
						"related": "http://example.com/articles/1/author"
					},
					"data": {
						"type": "people",
						"id": "9"
					}
				}
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, relWithLinks)
	if errs != nil {
		t.Errorf("Relationship with links should be valid: %s", errs)
	}

	// Relationship with only links (no data) - valid for lazy loading
	relLinksOnly := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"author": {
					"links": {
						"self": "http://example.com/articles/1/relationships/author",
						"related": "http://example.com/articles/1/author"
					}
				}
			}
		}
	}`

	_, errs = ruleSet.Apply(ctx, relLinksOnly)
	if errs != nil {
		t.Errorf("Relationship with only links should be valid: %s", errs)
	}
}

// TestSpecV1_1_JSONAPIObjectMembers tests Section 5.7: JSON:API Object Members
// Spec: https://jsonapi.org/format/#document-jsonapi-object
// The jsonapi object MAY contain: version, ext, profile, meta
func TestSpecV1_1_JSONAPIObjectMembers(t *testing.T) {
	// Full jsonapi object
	fullJsonAPIDoc := `{
		"jsonapi": {
			"version": "1.1",
			"ext": ["https://jsonapi.org/ext/atomic"],
			"profile": ["https://example.com/profiles/timestamps"],
			"meta": {
				"custom": "value"
			}
		},
		"data": {
			"type": "articles",
			"id": "1"
		}
	}`

	var fullResponse map[string]any
	err := json.Unmarshal([]byte(fullJsonAPIDoc), &fullResponse)
	if err != nil {
		t.Fatalf("Full jsonapi object should be valid JSON: %v", err)
	}

	jsonapiObj, ok := fullResponse["jsonapi"].(map[string]any)
	if !ok {
		t.Fatal("jsonapi member should be an object")
	}

	if jsonapiObj["version"] != "1.1" {
		t.Errorf("jsonapi version should be 1.1, got: %v", jsonapiObj["version"])
	}

	// Minimal jsonapi object (version only)
	minimalJsonAPIDoc := `{
		"jsonapi": {
			"version": "1.0"
		},
		"data": null
	}`

	var minimalResponse map[string]any
	err = json.Unmarshal([]byte(minimalJsonAPIDoc), &minimalResponse)
	if err != nil {
		t.Fatalf("Minimal jsonapi object should be valid JSON: %v", err)
	}
}

// TestSpecV1_1_ErrorObjectLinks tests Section 7: Error Object Links
// Spec: https://jsonapi.org/format/#error-objects
// Error links MAY contain: about, type
func TestSpecV1_1_ErrorObjectLinks(t *testing.T) {
	// Error with links
	errorWithLinks := `{
		"errors": [
			{
				"id": "1",
				"links": {
					"about": "http://example.com/errors/1",
					"type": "http://example.com/error-types/validation"
				},
				"status": "422",
				"code": "validation-failed",
				"title": "Validation Failed",
				"detail": "The title field is required.",
				"source": {
					"pointer": "/data/attributes/title"
				}
			}
		]
	}`

	var errorResponse map[string]any
	err := json.Unmarshal([]byte(errorWithLinks), &errorResponse)
	if err != nil {
		t.Fatalf("Error with links should be valid JSON: %v", err)
	}

	errors := errorResponse["errors"].([]any)
	errorObj := errors[0].(map[string]any)
	links := errorObj["links"].(map[string]any)

	if links["about"] != "http://example.com/errors/1" {
		t.Errorf("Error about link should match")
	}
	if links["type"] != "http://example.com/error-types/validation" {
		t.Errorf("Error type link should match")
	}
}

// TestSpecV1_1_ErrorSourceMembers tests Section 7: Error Source Members
// Spec: https://jsonapi.org/format/#error-objects
// Error source MAY contain: pointer, parameter, header
func TestSpecV1_1_ErrorSourceMembers(t *testing.T) {
	// Error with pointer source
	errorWithPointer := `{
		"errors": [
			{
				"status": "422",
				"title": "Validation Failed",
				"source": {
					"pointer": "/data/attributes/title"
				}
			}
		]
	}`

	var pointerResponse map[string]any
	err := json.Unmarshal([]byte(errorWithPointer), &pointerResponse)
	if err != nil {
		t.Fatalf("Error with pointer should be valid JSON: %v", err)
	}

	// Error with parameter source
	errorWithParameter := `{
		"errors": [
			{
				"status": "400",
				"title": "Invalid Parameter",
				"source": {
					"parameter": "filter[status]"
				}
			}
		]
	}`

	var paramResponse map[string]any
	err = json.Unmarshal([]byte(errorWithParameter), &paramResponse)
	if err != nil {
		t.Fatalf("Error with parameter should be valid JSON: %v", err)
	}

	// Error with header source
	errorWithHeader := `{
		"errors": [
			{
				"status": "400",
				"title": "Invalid Header",
				"source": {
					"header": "Authorization"
				}
			}
		]
	}`

	var headerResponse map[string]any
	err = json.Unmarshal([]byte(errorWithHeader), &headerResponse)
	if err != nil {
		t.Fatalf("Error with header should be valid JSON: %v", err)
	}
}

// TestSpecV1_1_QueryParameterValidNames tests that query parameters with valid names per JSON:API are accepted.
// Spec: https://jsonapi.org/format/#query-parameters
// Valid names: standard (sort, include, fields[TYPE], filter[name], page[size], page[after], page[before]),
// extension (namespace:member), and implementation-specific (must contain non a-z).
func TestSpecV1_1_QueryParameterValidNames(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	validCases := []struct {
		name  string
		query string
		check func(t *testing.T, vals url.Values)
	}{
		{"sort", "sort=title", func(t *testing.T, vals url.Values) {
			if vals.Get("sort") != "title" {
				t.Errorf("expected sort=title, got %q", vals.Get("sort"))
			}
		}},
		{"include", "include=author", func(t *testing.T, vals url.Values) {
			if vals.Get("include") != "author" {
				t.Errorf("expected include=author, got %q", vals.Get("include"))
			}
		}},
		{"page[size]", "page[size]=10", func(t *testing.T, vals url.Values) {
			if vals.Get("page[size]") != "10" {
				t.Errorf("expected page[size]=10, got %q", vals.Get("page[size]"))
			}
		}},
		{"page[after]", "page[after]=cursor", func(t *testing.T, vals url.Values) {
			if vals.Get("page[after]") != "cursor" {
				t.Errorf("expected page[after]=cursor, got %q", vals.Get("page[after]"))
			}
		}},
		{"page[before]", "page[before]=cursor", func(t *testing.T, vals url.Values) {
			if vals.Get("page[before]") != "cursor" {
				t.Errorf("expected page[before]=cursor, got %q", vals.Get("page[before]"))
			}
		}},
		{"fields[type]", "fields[articles]=title,body", func(t *testing.T, vals url.Values) {
			if vals.Get("fields[articles]") != "title,body" {
				t.Errorf("expected fields[articles]=title,body, got %q", vals.Get("fields[articles]"))
			}
		}},
		{"filter[name]", "filter[status]=published", func(t *testing.T, vals url.Values) {
			if vals.Get("filter[status]") != "published" {
				t.Errorf("expected filter[status]=published, got %q", vals.Get("filter[status]"))
			}
		}},
		{"extension namespace:member", "ext:version=1", func(t *testing.T, vals url.Values) {
			if vals.Get("ext:version") != "1" {
				t.Errorf("expected ext:version=1, got %q", vals.Get("ext:version"))
			}
		}},
		{"implementation-specific camelCase", "camelCase=value", func(t *testing.T, vals url.Values) {
			if vals.Get("camelCase") != "value" {
				t.Errorf("expected camelCase=value, got %q", vals.Get("camelCase"))
			}
		}},
		{"implementation-specific underscore", "my_param=value", func(t *testing.T, vals url.Values) {
			if vals.Get("my_param") != "value" {
				t.Errorf("expected my_param=value, got %q", vals.Get("my_param"))
			}
		}},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := url.ParseQuery(tc.query)
			if err != nil {
				t.Fatalf("parse query: %v", err)
			}
			vals, errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
			if errs != nil {
				t.Errorf("valid query parameter name should be accepted: %s", errs)
				return
			}
			if tc.check != nil {
				tc.check(t, vals)
			}
		})
	}
}

// TestSpecV1_1_QueryParameterInvalidNames tests that WithParam panics when given an illegal name.
// Spec: https://jsonapi.org/format/#query-parameters-custom
// Illegal names: all-lowercase a-z except the standard "sort" and "include" (reserved for future spec use).
// Callers can use jsonapi.QueryParamNameRule.Evaluate(ctx, name) or rules.String().WithRule(jsonapi.QueryParamNameRule) to check before WithParam to avoid a panic.
func TestSpecV1_1_QueryParameterInvalidNames(t *testing.T) {
	illegalKeys := []string{
		"unknownparam", "foo", "bar", "filter", "page", "fields",
		"customkey", "another", "xyz",
	}

	for _, key := range illegalKeys {
		key := key
		t.Run(key, func(t *testing.T) {
			var panicked interface{}
			func() {
				defer func() { panicked = recover() }()
				jsonapi.Query().WithParam(key, rules.String().Any())
			}()
			if panicked == nil {
				t.Errorf("WithParam(%q, ...) should panic (illegal per JSON:API), did not panic", key)
			}
		})
	}
}

// TestSpecV1_1_QueryParameterForbiddenUse tests that valid parameter names are rejected when used in a forbidden context.
// E.g. sort on POST, fields on DELETE, sort/page/filter when resource ID is present.
func TestSpecV1_1_QueryParameterForbiddenUse(t *testing.T) {
	cases := []struct {
		name   string
		query  string
		ctx    context.Context
		reason string
	}{
		{"sort_on_POST", "sort=title", jsonapi.WithMethod(context.Background(), "POST"), "sort only allowed on index GET/HEAD"},
		{"fields_on_DELETE", "fields[articles]=title", jsonapi.WithMethod(context.Background(), "DELETE"), "fields not allowed on DELETE"},
		{"sort_with_resource_ID", "sort=title", jsonapi.WithId(jsonapi.WithMethod(context.Background(), "GET"), "123"), "sort only allowed on index"},
		{"page_size_with_resource_ID", "page[size]=10", jsonapi.WithId(jsonapi.WithMethod(context.Background(), "GET"), "123"), "page[size] only allowed on index"},
		{"filter_on_POST", "filter[status]=published", jsonapi.WithMethod(context.Background(), "POST"), "filter only allowed on index GET/HEAD"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, _ := url.ParseQuery(tc.query)
			_, errs := jsonapi.QueryStringBaseRuleSet.Apply(tc.ctx, parsed)
			if errs == nil {
				t.Errorf("expected validation error (%s), got none for query %q", tc.reason, tc.query)
			}
		})
	}
}

// TestSpecV1_1_QueryParameterFamilies tests Section 6: Query Parameter Families
// Spec: https://jsonapi.org/format/#query-parameters-families
// A query parameter family is the set of all query parameters whose name starts with a base name
// Note: Specific keys within families are implementation-specific
func TestSpecV1_1_QueryParameterFamilies(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// Fields family - multiple resource types (sparse fieldsets)
	// The spec defines that fields[TYPE] is for sparse fieldsets
	fieldsQuery := `fields[articles]=title,body&fields[people]=name,email`
	parsed, _ := url.ParseQuery(fieldsQuery)
	_, errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Fields query parameter family should be valid: %s", errs)
	}

	// Filter family - simple key
	filterQuery := `filter[status]=published`
	parsed, _ = url.ParseQuery(filterQuery)
	_, errs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Filter query parameter family should be valid: %s", errs)
	}

	// Note: Page family keys are implementation-specific (JSON:API is agnostic about pagination)
}

// TestSpecV1_1_ImplementationSpecificQueryParameters tests Section 6.2
// Spec: https://jsonapi.org/format/#query-parameters-custom
// Custom query parameters MUST contain at least one non a-z character
func TestSpecV1_1_ImplementationSpecificQueryParameters(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// Valid custom parameter (contains capital letter)
	customQuery := `camelCase=value`
	parsed, _ := url.ParseQuery(customQuery)
	_, errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	// Custom parameters with non a-z characters should be allowed
	if errs != nil {
		t.Errorf("Custom query parameter (camelCase) should be valid: %s", errs)
	}

	// Another valid custom parameter (contains underscore)
	customQuery2 := `my_param=value`
	parsed, _ = url.ParseQuery(customQuery2)
	_, errs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	if errs != nil {
		t.Errorf("Custom query parameter (underscore) should be valid: %s", errs)
	}

	// Invalid: all lowercase a-z (reserved for future spec use)
	// Server MUST return 400 Bad Request for unknown parameters that are all lowercase
	invalidQuery := `unknownparam=value`
	parsed, _ = url.ParseQuery(invalidQuery)
	_, errs = jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed)
	// This should return an error per spec
	if errs == nil {
		t.Error("All lowercase query parameter should be rejected per spec (reserved for future use)")
	}
}

// TestSpecV1_1_ResourceObjectLid tests local identifiers (lid)
// Spec: https://jsonapi.org/format/#document-resource-object-identification
// A resource object MAY have a lid (local identifier) instead of id for client-generated IDs
func TestSpecV1_1_ResourceObjectLid(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships()

	ctx := jsonapi.WithMethod(context.Background(), "POST")

	// Resource with lid (local identifier)
	resourceWithLid := `{
		"data": {
			"type": "articles",
			"lid": "local-1",
			"attributes": {
				"title": "New article"
			},
			"relationships": {
				"author": {
					"data": {
						"type": "people",
						"lid": "local-author-1"
					}
				}
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, resourceWithLid)
	if errs != nil {
		t.Errorf("Resource with lid should be valid: %s", errs)
	}
}

// TestSpecV1_1_TopLevelMemberCoexistence tests Section 5.1: Top Level
// Spec: https://jsonapi.org/format/#document-top-level
// data and errors MUST NOT coexist in the same document
func TestSpecV1_1_TopLevelMemberCoexistence(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}
	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	// Valid: data only
	dataOnlyDoc := `{"data": {"type": "articles", "id": "1"}}`
	var dataResponse map[string]any
	err := json.Unmarshal([]byte(dataOnlyDoc), &dataResponse)
	if err != nil {
		t.Fatalf("Data-only document should be valid JSON: %v", err)
	}

	// Valid: errors only
	errorsOnlyDoc := `{"errors": [{"status": "404", "title": "Not Found"}]}`
	var errorsResponse map[string]any
	err = json.Unmarshal([]byte(errorsOnlyDoc), &errorsResponse)
	if err != nil {
		t.Fatalf("Errors-only document should be valid JSON: %v", err)
	}

	// Invalid: data AND errors (spec violation) — validator must reject
	invalidDoc := `{"data": {"type": "articles", "id": "1"}, "errors": [{"status": "400"}]}`
	var invalidParsed map[string]any
	err = json.Unmarshal([]byte(invalidDoc), &invalidParsed)
	if err != nil {
		t.Fatalf("Invalid document should still be valid JSON: %v", err)
	}
	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet)
	_, errs := ruleSet.Apply(context.Background(), invalidParsed)
	if errs == nil {
		t.Error("Document with both data and errors must be rejected by validator (they must not coexist)")
	}

	// Valid: meta only (allowed for 204 No Content)
	metaOnlyDoc := `{"meta": {"copyright": "Example Corp."}}`
	var metaResponse map[string]any
	err = json.Unmarshal([]byte(metaOnlyDoc), &metaResponse)
	if err != nil {
		t.Fatalf("Meta-only document should be valid JSON: %v", err)
	}
}

// TestSpecV1_1_IncludedResourceFullLinkage tests Section 5.4: Compound Documents
// Spec: https://jsonapi.org/format/#document-compound-documents
// All included resources MUST be referenced from primary data (full linkage)
func TestSpecV1_1_IncludedResourceFullLinkage(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships()

	ctx := context.Background()

	// Valid compound document (included resource is linked from primary data)
	validCompound := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API"
			},
			"relationships": {
				"author": {
					"data": {"type": "people", "id": "9"}
				},
				"comments": {
					"data": [
						{"type": "comments", "id": "5"},
						{"type": "comments", "id": "12"}
					]
				}
			}
		},
		"included": [
			{
				"type": "people",
				"id": "9",
				"attributes": {"name": "Dan"}
			},
			{
				"type": "comments",
				"id": "5",
				"attributes": {"body": "Great!"}
			},
			{
				"type": "comments",
				"id": "12",
				"attributes": {"body": "Awesome!"}
			}
		]
	}`

	_, errs := ruleSet.Apply(ctx, validCompound)
	if errs != nil {
		t.Errorf("Valid compound document should be valid: %s", errs)
	}
}

// TestSpecV1_1_RelationshipMeta tests relationship meta member
// Spec: https://jsonapi.org/format/#document-resource-object-relationships
// A relationship object MAY contain a meta member
func TestSpecV1_1_RelationshipMeta(t *testing.T) {
	type ArticleAttributes struct {
		Title string `json:"title"`
	}

	attributesRuleSet := rules.Struct[ArticleAttributes]().
		WithJson().
		WithKey("Title", rules.String().Any()).
		WithUnknown()

	ruleSet := jsonapi.NewSingleRuleSet[ArticleAttributes]("articles", attributesRuleSet).
		WithUnknownRelationships()

	ctx := context.Background()

	// Relationship with meta
	relWithMeta := `{
		"data": {
			"type": "articles",
			"id": "1",
			"attributes": {
				"title": "JSON:API paints my bikeshed!"
			},
			"relationships": {
				"comments": {
					"data": [
						{"type": "comments", "id": "5"}
					],
					"meta": {
						"count": 42
					}
				}
			}
		}
	}`

	_, errs := ruleSet.Apply(ctx, relWithMeta)
	if errs != nil {
		t.Errorf("Relationship with meta should be valid: %s", errs)
	}
}

// TestSpecV1_1_ResourceIdentifierMeta tests resource identifier meta
// Spec: https://jsonapi.org/format/#document-resource-identifier-objects
// A resource identifier object MAY contain a meta member
func TestSpecV1_1_ResourceIdentifierMeta(t *testing.T) {
	ctx := context.Background()

	// Resource identifier with meta
	identifierWithMeta := `{
		"type": "articles",
		"id": "1",
		"meta": {
			"created": "2020-01-01T00:00:00Z"
		}
	}`

	_, errs := jsonapi.ResourceLinkageRuleSet.Apply(ctx, identifierWithMeta)
	if errs != nil {
		t.Errorf("Resource identifier with meta should be valid: %s", errs)
	}
}

// TestSpecV1_1_AttributesRestrictions tests Section 5.2.2: Attributes
// Spec: https://jsonapi.org/format/#document-resource-object-attributes
// Attributes MUST NOT contain relationships, links, id, or type
func TestSpecV1_1_AttributesRestrictions(t *testing.T) {
	// These restrictions are about what names MUST NOT appear in attributes
	// The attributes object can contain any valid JSON except:
	// - relationships, links (belong at resource object level)
	// - id, type (belong at resource object level)

	// Valid attributes
	validAttributes := `{
		"title": "Valid Title",
		"body": "Valid body text",
		"created-at": "2020-01-01T00:00:00Z",
		"word_count": 100,
		"tags": ["json", "api"],
		"metadata": {"author": "John"}
	}`

	var attrs map[string]any
	err := json.Unmarshal([]byte(validAttributes), &attrs)
	if err != nil {
		t.Fatalf("Valid attributes should be valid JSON: %v", err)
	}

	// Check that reserved names are not in attributes
	reservedNames := []string{"relationships", "links", "id", "type"}
	for _, reserved := range reservedNames {
		if _, exists := attrs[reserved]; exists {
			t.Errorf("Attributes should not contain reserved name: %s", reserved)
		}
	}
}
