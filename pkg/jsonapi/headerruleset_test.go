package jsonapi

import (
	"context"
	"net/http"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules"
)

func TestHeaderRuleSet_ContentTypeRequired(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	empty := http.Header{}
	err := rs.Apply(ctx, empty, nil)
	if err == nil {
		t.Fatal("expected error when Content-Type missing")
	}
	list := ErrorsFromValidationError(err, SourceHeader)
	if len(list) == 0 {
		t.Fatal("expected at least one error")
	}
	if list[0].Source == nil || list[0].Source.Header == "" {
		t.Errorf("expected source.header set (e.g. Content-Type), got %v", list[0].Source)
	}
}

func TestHeaderRuleSet_ContentTypeWrongMediaType(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	err := rs.Apply(ctx, h, nil)
	if err == nil {
		t.Fatal("expected error for wrong media type")
	}
	list := ErrorsFromValidationError(err, SourceHeader)
	if len(list) == 0 {
		t.Fatal("expected at least one error")
	}
}

func TestHeaderRuleSet_ContentTypeValid(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	h := http.Header{}
	h.Set("Content-Type", MediaTypeJSONAPI)
	err := rs.Apply(ctx, h, nil)
	if err != nil {
		t.Fatalf("expected no error for valid Content-Type: %v", err)
	}
}

func TestHeaderRuleSet_ContentTypeWithExtAndProfile(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	h := http.Header{}
	h.Set("Content-Type", `application/vnd.api+json; ext="https://jsonapi.org/ext/version"; profile="https://example.com/profile"`)
	err := rs.Apply(ctx, h, nil)
	if err != nil {
		t.Fatalf("expected no error for valid Content-Type with ext and profile: %v", err)
	}
}

func TestHeaderRuleSet_ContentTypeDisallowedParam(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	h := http.Header{}
	h.Set("Content-Type", "application/vnd.api+json; charset=utf-8")
	err := rs.Apply(ctx, h, nil)
	if err == nil {
		t.Fatal("expected error for disallowed charset parameter")
	}
	list := ErrorsFromValidationError(err, SourceHeader)
	if len(list) == 0 {
		t.Fatal("expected at least one error")
	}
}

func TestHeaderRuleSet_WithExt(t *testing.T) {
	// WithExt validates the ext parameter value with a rule set
	allowedExt := "https://jsonapi.org/ext/version"
	rs := Headers().WithExt(rules.String().WithRuleFunc(func(ctx context.Context, s string) errors.ValidationError {
		if s != allowedExt {
			return errors.Errorf(errors.CodeNotAllowed, ctx, "ext not allowed", "ext must be %q", allowedExt)
		}
		return nil
	}).Any())
	ctx := context.Background()

	h := http.Header{}
	h.Set("Content-Type", `application/vnd.api+json; ext="https://jsonapi.org/ext/version"`)
	err := rs.Apply(ctx, h, nil)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	h.Set("Content-Type", `application/vnd.api+json; ext="https://other.org/ext"`)
	err = rs.Apply(ctx, h, nil)
	if err == nil {
		t.Fatal("expected error for disallowed ext")
	}
}

func TestHeaderRuleSet_WithProfile(t *testing.T) {
	allowedProfile := "https://example.com/profile"
	rs := Headers().WithProfile(rules.String().WithRuleFunc(func(ctx context.Context, s string) errors.ValidationError {
		if s != allowedProfile {
			return errors.Errorf(errors.CodeNotAllowed, ctx, "profile not allowed", "profile must be %q", allowedProfile)
		}
		return nil
	}).Any())
	ctx := context.Background()

	h := http.Header{}
	h.Set("Content-Type", `application/vnd.api+json; profile="https://example.com/profile"`)
	err := rs.Apply(ctx, h, nil)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	h.Set("Content-Type", `application/vnd.api+json; profile="https://other.com/profile"`)
	err = rs.Apply(ctx, h, nil)
	if err == nil {
		t.Fatal("expected error for disallowed profile")
	}
}

func TestHeaderRuleSet_WithHeader(t *testing.T) {
	rs := Headers().WithHeader("X-Request-Id", rules.String().WithMinLen(1).Any())
	ctx := context.Background()

	h := http.Header{}
	h.Set("Content-Type", MediaTypeJSONAPI)
	h.Set("X-Request-Id", "abc-123")
	err := rs.Apply(ctx, h, nil)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	h.Set("X-Request-Id", "")
	err = rs.Apply(ctx, h, nil)
	if err == nil {
		t.Fatal("expected error for empty X-Request-Id")
	}
	list := ErrorsFromValidationError(err, SourceHeader)
	if len(list) == 0 {
		t.Fatal("expected at least one error")
	}
	if list[0].Source == nil || list[0].Source.Header != "X-Request-Id" {
		t.Errorf("expected source.header = X-Request-Id, got %v", list[0].Source)
	}
}

func TestHeaderRuleSet_ErrorsUseSourceHeader(t *testing.T) {
	rs := Headers().WithHeader("Accept", rules.String().WithMinLen(10).Any())
	ctx := context.Background()
	h := http.Header{}
	h.Set("Content-Type", MediaTypeJSONAPI)
	h.Set("Accept", "short")
	err := rs.Apply(ctx, h, nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	list := ErrorsFromValidationError(err, SourceHeader)
	if len(list) == 0 {
		t.Fatal("expected errors")
	}
	for _, e := range list {
		if e.Source == nil {
			t.Error("expected Source set on error")
			continue
		}
		if e.Source.Header == "" && e.Source.Pointer != "" && e.Source.Parameter != "" {
			t.Errorf("expected source.header set for header error, got %+v", e.Source)
		}
	}
}

func TestHeaderRuleSet_Apply_InputJsonAPIHeader(t *testing.T) {
	rs := Headers()
	ctx := context.Background()

	// *Header input: valid (no ext/profile)
	in := &Header{Version: Version_1_1}
	err := rs.Apply(ctx, in, nil)
	if err != nil {
		t.Fatalf("expected no error for *Header input: %v", err)
	}

	// Header value input: with ext and profile
	inVal := Header{
		Version: Version_1_1,
		Ext:     []Extension{{URI: "https://jsonapi.org/ext/version"}},
		Profile: []Profile{{URI: "https://example.com/profile"}},
	}
	err = rs.Apply(ctx, inVal, nil)
	if err != nil {
		t.Fatalf("expected no error for Header input with ext/profile: %v", err)
	}

	// *Header input that would fail Content-Type check (we only set Content-Type from ext/profile; no way to pass wrong media type from Header)
	// So only test valid cases for Header input.
}

func TestHeaderRuleSet_Apply_OutputJsonAPIHeader(t *testing.T) {
	rs := Headers()
	ctx := context.Background()

	h := http.Header{}
	h.Set("Content-Type", `application/vnd.api+json; ext="https://jsonapi.org/ext/version"; profile="https://example.com/profile"`)
	var out Header
	err := rs.Apply(ctx, h, &out)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if out.Version != Version_1_1 {
		t.Errorf("expected Version %q, got %q", Version_1_1, out.Version)
	}
	if len(out.Ext) != 1 || out.Ext[0].URI != "https://jsonapi.org/ext/version" {
		t.Errorf("expected Ext with one URI, got %+v", out.Ext)
	}
	if len(out.Profile) != 1 || out.Profile[0].URI != "https://example.com/profile" {
		t.Errorf("expected Profile with one URI, got %+v", out.Profile)
	}
}

func TestHeaderRuleSet_Apply_HeaderRoundtrip(t *testing.T) {
	rs := Headers()
	ctx := context.Background()

	in := &Header{
		Version: Version_1_1,
		Ext:     []Extension{{URI: "https://jsonapi.org/ext/version"}},
		Profile: []Profile{{URI: "https://example.com/profile"}},
	}
	var out Header
	err := rs.Apply(ctx, in, &out)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if out.Version != Version_1_1 {
		t.Errorf("expected Version %q, got %q", Version_1_1, out.Version)
	}
	if len(out.Ext) != 1 || out.Ext[0].URI != in.Ext[0].URI {
		t.Errorf("expected Ext %+v, got %+v", in.Ext, out.Ext)
	}
	if len(out.Profile) != 1 || out.Profile[0].URI != in.Profile[0].URI {
		t.Errorf("expected Profile %+v, got %+v", in.Profile, out.Profile)
	}
}

func TestHeaderRuleSet_WithContentRequired(t *testing.T) {
	rs := Headers().WithContentRequired(false)
	ctx := context.Background()
	empty := http.Header{}
	err := rs.Apply(ctx, empty, nil)
	if err != nil {
		t.Fatalf("WithContentRequired(false): expected no error when Content-Type missing, got %v", err)
	}
	if rs.Required() {
		t.Error("expected Required false after WithContentRequired(false)")
	}
}

func TestHeaderRuleSet_Apply_MapStringSlice(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	h := map[string][]string{"Content-Type": {MediaTypeJSONAPI}}
	err := rs.Apply(ctx, h, nil)
	if err != nil {
		t.Fatalf("Apply with map[string][]string: %v", err)
	}
}

func TestHeaderRuleSet_Apply_InvalidType(t *testing.T) {
	rs := Headers()
	ctx := context.Background()
	err := rs.Apply(ctx, 12345, nil)
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestHeaderRuleSet_RequiredStringReplacesAny(t *testing.T) {
	rs := Headers()
	if !rs.Required() {
		t.Error("expected Required true by default")
	}
	if s := rs.String(); s != "HeaderRuleSet" {
		t.Errorf("String(): got %q", s)
	}
	_ = rs.Replaces(nil)
	anyRS := rs.Any()
	if anyRS == nil {
		t.Error("Any() should not be nil")
	}
}
