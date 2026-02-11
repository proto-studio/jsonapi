package jsonapi_test

import (
	"context"
	"net/url"
	"testing"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
	"proto.zip/studio/validate/pkg/errors"
)

// Cursor Pagination Profile Tests
// Profile URL: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/
//
// This profile defines three query parameters:
// - page[size]: number of results the client would like to see
// - page[before]: cursor to get results before
// - page[after]: cursor to get results after

// TestCursorPagination_PageSize tests the page[size] parameter
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#query-parameters-page-size
// "If page[size] is provided, it MUST be a positive integer"
func TestCursorPagination_PageSize(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// Valid page size
	t.Run("valid page size", func(t *testing.T) {
		query := `page[size]=10`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Valid page[size]=10 should be accepted: %s", errs)
		}
		if vals.Get("page[size]") != "10" {
			t.Errorf("Expected page[size]=10, got %q", vals.Get("page[size]"))
		}
	})

	// page[size]=1 is the minimum valid value
	t.Run("minimum page size", func(t *testing.T) {
		query := `page[size]=1`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("page[size]=1 should be valid: %s", errs)
		}
	})

	// page[size]=0 is invalid (must be positive)
	t.Run("zero page size rejected", func(t *testing.T) {
		query := `page[size]=0`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[size]=0 should be rejected (must be positive integer)")
		}
	})

	// Negative values are invalid
	t.Run("negative page size rejected", func(t *testing.T) {
		query := `page[size]=-1`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[size]=-1 should be rejected (must be positive integer)")
		}
	})
}

// TestCursorPagination_MaxPageSize tests the max page size limit
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#query-parameters-page-size
// "If page[size] exceeds the server-defined max page size, the server MUST respond
// according to the rules for the max page size exceeded error"
func TestCursorPagination_MaxPageSize(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// The library has a default max page size of 100
	t.Run("at max page size", func(t *testing.T) {
		query := `page[size]=100`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("page[size]=100 (max) should be valid: %s", errs)
		}
	})

	t.Run("exceeds max page size", func(t *testing.T) {
		query := `page[size]=101`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[size]=101 should be rejected (exceeds max page size)")
		}
	})
}

// TestCursorPagination_PageAfter tests the page[after] parameter
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#query-parameters-page-after-and-page-before
// "The page[after] parameter is typically sent by the client to get the next page"
func TestCursorPagination_PageAfter(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	t.Run("valid cursor", func(t *testing.T) {
		// Example from profile: GET /people?page[size]=100&page[after]=abcde
		query := `page[size]=100&page[after]=abcde`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("page[after]=abcde should be valid: %s", errs)
		}
		if vals.Get("page[after]") != "abcde" {
			t.Errorf("Expected page[after]=abcde, got %q", vals.Get("page[after]"))
		}
	})

	t.Run("cursor with special characters", func(t *testing.T) {
		query := `page[after]=cursor_123-abc`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Cursor with special characters should be valid: %s", errs)
		}
	})

	t.Run("opaque cursor string", func(t *testing.T) {
		// Profile: "A cursor is a string, created by the server using whatever method it likes"
		query := `page[after]=someOpaqueString`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Opaque cursor string should be valid: %s", errs)
		}
	})
}

// TestCursorPagination_PageBefore tests the page[before] parameter
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#query-parameters-page-after-and-page-before
// "page[before] is used to get the prior page"
func TestCursorPagination_PageBefore(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	t.Run("valid cursor", func(t *testing.T) {
		query := `page[before]=xyz`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("page[before]=xyz should be valid: %s", errs)
		}
		if vals.Get("page[before]") != "xyz" {
			t.Errorf("Expected page[before]=xyz, got %q", vals.Get("page[before]"))
		}
	})

	t.Run("backwards pagination", func(t *testing.T) {
		// Profile: "Replacing page[after] with page[before] would allow the client to paginate backwards"
		query := `page[size]=10&page[before]=cursor123`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Backwards pagination should be valid: %s", errs)
		}
	})
}

// TestCursorPagination_RangePagination tests using both page[after] and page[before]
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#query-parameters-page-after-and-page-before
// "to find all people between cursors abcde and fghij (exclusive), the client could request:
// GET /people?page[after]=abcde&page[before]=fghij"
func TestCursorPagination_RangePagination(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	t.Run("range with both cursors", func(t *testing.T) {
		query := `page[after]=abcde&page[before]=fghij`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Range pagination with both cursors should be valid: %s", errs)
		}
		if vals.Get("page[after]") != "abcde" {
			t.Errorf("Expected page[after]=abcde, got %q", vals.Get("page[after]"))
		}
		if vals.Get("page[before]") != "fghij" {
			t.Errorf("Expected page[before]=fghij, got %q", vals.Get("page[before]"))
		}
	})

	t.Run("range with size", func(t *testing.T) {
		query := `page[size]=50&page[after]=start&page[before]=end`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Range pagination with size should be valid: %s", errs)
		}
	})
}

// TestCursorPagination_IndexOnly tests that pagination only works on index requests
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#concepts-sorting-requirement
// "Pagination only applies to an ordered list of results"
func TestCursorPagination_IndexOnly(t *testing.T) {
	// page[size] only on GET index
	t.Run("page[size] forbidden on POST", func(t *testing.T) {
		ctx := jsonapi.WithMethod(context.Background(), "POST")
		query := `page[size]=10`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[size] should be forbidden on POST")
		}
	})

	t.Run("page[size] forbidden with ID", func(t *testing.T) {
		ctx := jsonapi.WithMethod(context.Background(), "GET")
		ctx = jsonapi.WithId(ctx, "123")
		query := `page[size]=10`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[size] should be forbidden when fetching single resource")
		}
	})

	t.Run("page[after] forbidden with ID", func(t *testing.T) {
		ctx := jsonapi.WithMethod(context.Background(), "GET")
		ctx = jsonapi.WithId(ctx, "123")
		query := `page[after]=cursor`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[after] should be forbidden when fetching single resource")
		}
	})

	t.Run("page[before] forbidden with ID", func(t *testing.T) {
		ctx := jsonapi.WithMethod(context.Background(), "GET")
		ctx = jsonapi.WithId(ctx, "123")
		query := `page[before]=cursor`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("page[before] should be forbidden when fetching single resource")
		}
	})
}

// TestCursorPagination_HEAD tests that pagination works on HEAD requests
// Profile allows HEAD requests for pagination (same as GET for index)
func TestCursorPagination_HEAD(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "HEAD")

	t.Run("page[size] allowed on HEAD", func(t *testing.T) {
		query := `page[size]=10`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("page[size] should be allowed on HEAD index: %s", errs)
		}
	})

	t.Run("page[after] allowed on HEAD", func(t *testing.T) {
		query := `page[after]=cursor`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("page[after] should be allowed on HEAD index: %s", errs)
		}
	})
}

// TestCursorPagination_WithSort tests pagination combined with sort
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#concepts-sorting-requirement
// "Pagination only applies to an ordered list of results"
func TestCursorPagination_WithSort(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	t.Run("pagination with sort", func(t *testing.T) {
		query := `sort=-created,title&page[size]=10&page[after]=cursor`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs != nil {
			t.Errorf("Pagination with sort should be valid: %s", errs)
		}
		if vals.Get("sort") != "-created,title" {
			t.Errorf("Expected sort=-created,title, got %q", vals.Get("sort"))
		}
		if vals.Get("page[size]") != "10" {
			t.Errorf("Expected page[size]=10, got %q", vals.Get("page[size]"))
		}
	})
}

// TestCursorPagination_InvalidCursorFormat tests invalid cursor error handling
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#error-cases-invalid-parameter-value-error
// "If their value is not a valid cursor, the server MUST respond according to the
// rules for the invalid query parameter error"
func TestCursorPagination_EmptyCursor(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// The profile requires cursors to be provided if the parameter is used
	// An empty cursor is implementation-specific, but the library requires min length 1
	t.Run("empty page[after] rejected", func(t *testing.T) {
		query := `page[after]=`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("Empty page[after] should be rejected")
		}
	})

	t.Run("empty page[before] rejected", func(t *testing.T) {
		query := `page[before]=`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Errorf("Empty page[before] should be rejected")
		}
	})
}

// TestCursorPagination_FullExample tests the full example from the profile
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/
// Example: GET /people?page[size]=100&page[after]=abcde
func TestCursorPagination_FullExample(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	// Example from the profile introduction
	query := `page[size]=100&page[after]=abcde`
	parsed, _ := url.ParseQuery(query)
	var vals url.Values
	errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
	if errs != nil {
		t.Errorf("Profile example should be valid: %s", errs)
	}

	if vals.Get("page[size]") != "100" {
		t.Errorf("Expected page[size]=100, got %q", vals.Get("page[size]"))
	}
	if vals.Get("page[after]") != "abcde" {
		t.Errorf("Expected page[after]=abcde, got %q", vals.Get("page[after]"))
	}
}

// TestCursorPagination_NonIntegerSize tests non-integer page size
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#query-parameters-page-size
// "the value MUST be a sequence of characters matching the regular expression ^[0-9]+$"
func TestCursorPagination_NonIntegerSize(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	testCases := []struct {
		name  string
		query string
	}{
		{"decimal", `page[size]=10.5`},
		{"string", `page[size]=ten`},
		{"empty", `page[size]=`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, _ := url.ParseQuery(tc.query)
			var vals url.Values
			errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
			if errs == nil {
				t.Errorf("Non-integer page[size] '%s' should be rejected", tc.query)
			}
		})
	}
}

// TestCursorPagination_ErrorCodes tests that appropriate error codes are returned
// Profile: https://jsonapi.org/profiles/ethanresnick/cursor-pagination/#error-cases
func TestCursorPagination_ErrorCodes(t *testing.T) {
	ctx := jsonapi.WithMethod(context.Background(), "GET")

	t.Run("invalid page size returns proper error", func(t *testing.T) {
		query := `page[size]=0`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Fatalf("Expected error for invalid page size")
		}

		// Should have a validation error
		found := false
		for _, err := range errors.Unwrap(errs) {
			if ve, ok := err.(errors.ValidationError); ok && (ve.Code() == errors.CodeMin || ve.Code() == errors.CodeType) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected CodeMin or CodeType error for invalid page size, got: %v", errs)
		}
	})

	t.Run("exceeds max returns proper error", func(t *testing.T) {
		query := `page[size]=1000`
		parsed, _ := url.ParseQuery(query)
		var vals url.Values
		errs := jsonapi.QueryStringBaseRuleSet.Apply(ctx, parsed, &vals)
		if errs == nil {
			t.Fatalf("Expected error for page size exceeding max")
		}

		// Should have a max error
		found := false
		for _, err := range errors.Unwrap(errs) {
			if ve, ok := err.(errors.ValidationError); ok && ve.Code() == errors.CodeMax {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected CodeMax error for exceeding max page size, got: %v", errs)
		}
	})
}
